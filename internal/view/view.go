/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package view

// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/colorstring"
	"github.com/zclconf/go-cty/cty"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/format"
	"kubehcl.sh/kubehcl/internal/terminal"
	"kubehcl.sh/kubehcl/internal/tfdiags"
)

type Status int

const (
	ADDED Status = iota
	REMOVED 
	MODIFIED 
)

type Line struct {
	Name string
	ToValue any
	FromValue any
	Status Status
}
type ResourceChanges struct {
	Name string
	Lines []*Line
}

type CmpAttr struct {
	Name string
	FromAttr any 
	ToAttr any
	Status Status
}
type CompareResources struct{
	Current runtime.Object
	Wanted runtime.Object
}

type ViewArgs struct {
	// NoColor is used to disable the use of terminal color codes in all
	// output.
	NoColor bool

	// CompactWarnings is used to coalesce duplicate warnings, to reduce the
	// level of noise when multiple instances of the same warning are raised
	// for a configuration.
	CompactWarnings     bool
	ConsolidateWarnings bool
	ConsolidateErrors   bool

	// Concise is used to reduce the level of noise in the output and display
	// only the important details.
	Concise bool

	// ShowSensitive is used to display the value of variables marked as sensitive.
}

// View is the base layer for command views, encapsulating a set of I/O
// streams, a colorize implementation, and implementing a human friendly view
// for diagnostics.
type View struct {
	streams  *terminal.Streams
	colorize *colorstring.Colorize

	compactWarnings     bool
	consolidateWarnings bool
	consolidateErrors   bool

	// When this is true it's a hint that OpenTofu is being run indirectly
	// via a wrapper script or other automation and so we may wish to replace
	// direct examples of commands to run with more conceptual directions.
	// However, we only do this on a best-effort basis, typically prioritizing
	// the messages that users are most likely to see.
	runningInAutomation bool

	// Concise is used to reduce the level of noise in the output and display
	// only the important details.
	concise bool

	// showSensitive is used to display the value of variables marked as sensitive.
	showSensitive bool

	// This unfortunate wart is required to enable rendering of diagnostics which
	// have associated source code in the configuration. This function pointer
	// will be dereferenced as late as possible when rendering diagnostics in
	// order to access the config loader cache.
	configSources func() map[string]*hcl.File
}

// Initialize a View with the given streams, a disabled colorize object, and a
// no-op configSources callback.
func NewView(streams *terminal.Streams) *View {
	return &View{
		streams: streams,
		colorize: &colorstring.Colorize{
			Colors:  colorstring.DefaultColors,
			Disable: true,
			Reset:   true,
		},
		configSources: func() map[string]*hcl.File { return nil },
	}
}

// SetRunningInAutomation modifies the view's "running in automation" flag,
// which causes some slight adjustments to certain messages that would normally
// suggest specific OpenTofu commands to run, to make more conceptual gestures
// instead for situations where the user isn't running OpenTofu directly.
//
// For convenient use during initialization (in conjunction with NewView),
// SetRunningInAutomation returns the receiver after modifying it.
func (v *View) SetRunningInAutomation(new bool) *View {
	v.runningInAutomation = new
	return v
}

func (v *View) RunningInAutomation() bool {
	return v.runningInAutomation
}

// Configure applies the global view configuration flags.
func (v *View) Configure(view *ViewArgs) {
	v.colorize.Disable = view.NoColor
	v.compactWarnings = view.CompactWarnings
	v.consolidateWarnings = view.ConsolidateWarnings
	v.consolidateErrors = view.ConsolidateErrors
	v.concise = view.Concise
}

// SetConfigSources overrides the default no-op callback with a new function
// pointer, and should be called when the config loader is initialized.
func (v *View) SetConfigSources(cb func() map[string]*hcl.File) {
	v.configSources = cb
}

// Diagnostics renders a set of warnings and errors in human-readable form.
// Warnings are printed to stdout, and errors to stderr.
func (v *View) Diagnostics(diags tfdiags.Diagnostics) {
	diags.Sort()
	
	if len(diags) == 0 {
		return
	}

	if v.consolidateWarnings {
		diags = diags.Consolidate(1, tfdiags.Warning)
	}
	if v.consolidateErrors {
		diags = diags.Consolidate(1, tfdiags.Error)
	}

	// Since warning messages are generally competing
	if v.compactWarnings {
		// If the user selected compact warnings and all of the diagnostics are
		// warnings then we'll use a more compact representation of the warnings
		// that only includes their summaries.
		// We show full warnings if there are also errors, because a warning
		// can sometimes serve as good context for a subsequent error.
		useCompact := true
		for _, diag := range diags {
			if diag.Severity() != tfdiags.Warning {
				useCompact = false
				break
			}
		}
		if useCompact {
			msg := format.DiagnosticWarningsCompact(diags, v.colorize)
			msg = "\n" + msg + "\nTo see the full warning notes, run kubehcl without -compact-warnings.\n"
			_, _ = v.streams.Print(msg)
			return
		}
	}

	for _, diag := range diags {
		var msg string
		if v.colorize.Disable {
			msg = format.DiagnosticPlain(diag, v.configSources(), v.streams.Stderr.Columns())
		} else {
			msg = format.Diagnostic(diag, v.configSources(), v.colorize, v.streams.Stderr.Columns())
		}

		if diag.Severity() == tfdiags.Error {
			_, _ = v.streams.Eprint(msg)
		} else {
			_, _ = v.streams.Print(msg)
		}
	}
}

// HelpPrompt is intended to be called from commands which fail to parse all
// of their CLI arguments successfully. It refers users to the full help output
// rather than rendering it directly, which can be overwhelming and confusing.
func (v *View) HelpPrompt(command string) {
	_, _ = v.streams.Eprintf(helpPrompt, command)
}

const helpPrompt = `
For more help on using this command, run:
  kubehcl %s -help
`

// outputColumns returns the number of text character cells any non-error
// output should be wrapped to.
//
// This is the number of columns to use if you are calling v.streams.Print or
// related functions.
// func (v *View) outputColumns() int {
// 	return v.streams.Stdout.Columns()
// }

// errorColumns returns the number of text character cells any error
// output should be wrapped to.
//
// This is the number of columns to use if you are calling v.streams.Eprint
// or related functions.
// func (v *View) errorColumns() int {
// 	return v.streams.Stderr.Columns()
// }

// outputHorizRule will call v.streams.Println with enough horizontal line
// characters to fill an entire row of output.
//
// If UI color is enabled, the rule will get a dark grey coloring to try to
// visually de-emphasize it.
// func (v *View) outputHorizRule() {
// 	v.streams.Println(format.HorizontalRule(v.colorize, v.outputColumns()))
// }

func (v *View) SetShowSensitive(showSensitive bool) {
	v.showSensitive = showSensitive
}

// Prints the diagnostic received with the arguments.
// This will print them in the same format as opentofu format
func (v *View) DiagPrinter (diags hcl.Diagnostics, viewDef *ViewArgs) {
	v.SetConfigSources(configs.Parser().Files)
	var d tfdiags.Diagnostics
	d = d.Append(diags)
	v.Configure(viewDef)
	v.Diagnostics(d)
}

func inferType(value interface{})cty.Value {

	switch tt:=value.(type){
	case string:
		return cty.StringVal(tt)
	case int64:
		return cty.NumberIntVal(tt)
	case float64:
		return cty.NumberFloatVal(tt)
	case bool:
		return cty.BoolVal(tt)
	case []any:
		var vals []cty.Value
		for _,val := range tt {
			vals = append(vals, inferType(val))
		}
		return cty.ListVal(vals)

	case map[string]any:
		valMap := make(map[string]cty.Value)
		for key,val := range tt {
			valMap[key] = inferType(val)
		}
		return cty.ObjectVal(valMap)
	default:
		panic("Unknown type")
	}
}

// func determineAttr(attrName string ,current *resource.Info,wanted *resource.Info) string{
// 	if current == nil {
// 		return colorstring.Color(fmt.Sprintf("[bold][green]+[reset] %s",attrName))
// 	}

// 	if wanted == nil {
// 		return colorstring.Color(fmt.Sprintf("[bold][red]-[reset] %s",attrName))
// 	}

// 	switch w:=wanted.Object.(type) {
// 	case *unstructured.Unstructured:
// 		switch c:=current.Object.(type){
// 		case *unstructured.Unstructured:
// 			if _,exists := c.Object[attrName]; !exists {
// 				return colorstring.Color(fmt.Sprintf("[bold][green]+[reset] %s",attrName))
// 			}

// 			if _,exists := w.Object[attrName]; !exists {
// 				return colorstring.Color(fmt.Sprintf("[bold][red]-[reset] %s",attrName))
// 			}

// 			if reflect.DeepEqual(c.Object[attrName],w.Object[attrName]) {
// 				return ""
// 			} else {
// 				return colorstring.Color(fmt.Sprintf("[bold][yellow]~[reset] %s",attrName))
// 			}

// 		}
// 	}
// 	panic("Didn't return any value")
// }




// func printResourceDiff(name string,current *resource.Info,wanted *resource.Info) {
//     f := hclwrite.NewEmptyFile()
// 	if wanted != nil  && current != nil {
// 		cur := current.Object.(*unstructured.Unstructured)
// 		// wan := wanted.Object.(*unstructured.Unstructured)
// 		fmt.Printf("%s",cur.man)
// 	}
	
// 	if current != nil {
// 		switch tt:= current.Object.(type){
// 		case *unstructured.Unstructured:
// 			removeUnnecessaryFields(tt.Object)
// 		}
// 	}

// 	if wanted != nil {
// 		switch tt:= wanted.Object.(type){
// 		case *unstructured.Unstructured:
// 			removeUnnecessaryFields(tt.Object)
// 		}
// 	}
// 	if wanted != nil {
// 		switch tt:=wanted.Object.(type){
// 		case *unstructured.Unstructured:
// 				block := f.Body().AppendNewBlock(name,[]string{})
// 				body := block.Body()
// 				if attr:=determineAttr("kind",current,wanted); attr!= "" {
// 					body.SetAttributeValue(attr,inferType(tt.Object["kind"]))
// 				}

// 				if attr:=determineAttr("apiVersion",current,wanted); attr!= "" {
// 					body.SetAttributeValue(attr,inferType(tt.Object["apiVersion"]))
// 				}

// 				for key,value := range tt.Object{
// 					if attr:=determineAttr(key,current,wanted); attr!= "" {
// 						body.SetAttributeValue(attr,inferType(value))
// 					}
// 				}				// fmt.Printf("key: %s value: %s\n",key,value)

// 		}
// 	}

// 	if wanted == nil {
// 		switch tt:=current.Object.(type){
// 		case *unstructured.Unstructured:
// 				block := f.Body().AppendNewBlock(name,[]string{})
// 				body := block.Body()
// 				if attr:=determineAttr("kind",current,wanted); attr!= "" {
// 					body.SetAttributeValue(attr,inferType(tt.Object["kind"]))
// 				}

// 				if attr:=determineAttr("apiVersion",current,wanted); attr!= "" {
// 					body.SetAttributeValue(attr,inferType(tt.Object["apiVersion"]))
// 				}
// 				for key,value := range tt.Object{
// 					if attr:=determineAttr(key,current,wanted); attr!= "" {
// 						body.SetAttributeValue(attr,inferType(value))
// 					}
// 				}				// fmt.Printf("key: %s value: %s\n",key,value)

// 		}
// 	}

// 	hclStr := string(f.Bytes())
// 	splitStr := strings.Split(hclStr, "\n")
// 	splitStr[0] = colorstring.Color(fmt.Sprintf("[bold][green]+[reset] %s",splitStr[0]))
// 	fmt.Printf("%s",strings.Join(splitStr,"\n"))
// 	// f.WriteTo(os.Stdout)

// }

// func PlanPrinter(wanted,current map[string]kube.ResourceList){
// 	// var buf bytes.Buffer
// 	fmt.Println("Kubehcl will use the following symbols for each action and attribute")
// 	fmt.Println()
// 	fmt.Println(colorstring.Color("[bold][green]+[reset] create"))
// 	fmt.Println(colorstring.Color("[bold][yellow]~[reset] replace"))
// 	fmt.Println(colorstring.Color("[bold][red]-[reset] destroy"))
// 	fmt.Println()
// 	fmt.Println("Kubehcl will perform the following actions:")
// 	fmt.Println()

// 	for key,value := range wanted {
// 		currentList := current[key]
// 		for i,val := range value {
// 			if i< len(currentList){
// 				printResourceDiff(key,currentList[i],val)
// 			} else {
// 				printResourceDiff(key,nil,val)
// 			}
// 		}
// 	}
// }


func (v *View) PlanPrinter(m map[string]*CompareResources,viewDef *ViewArgs) {
	v.SetConfigSources(configs.Parser().Files)
	v.Configure(viewDef)
	if v.colorize.Disable {
		v.plainPlanPrinter(m)
	} else {
		v.planColoredPrinter(m)
	}
}

func determineAttrDiff(fromVal,toVal any,key string)[]*CmpAttr{
	cmpAttrList := []*CmpAttr{}

	if fromVal == nil && toVal != nil {
		cmpAttr := &CmpAttr{
				Name: key,
				FromAttr: fromVal,
				ToAttr: toVal,
				Status: ADDED,
			}
		cmpAttrList = append(cmpAttrList, cmpAttr)
		return cmpAttrList
	}

	if fromVal != nil && toVal == nil {
		cmpAttr := &CmpAttr{
				Name: key,
				FromAttr: fromVal,
				ToAttr: toVal,
				Status: REMOVED,
			}
		cmpAttrList = append(cmpAttrList, cmpAttr)
		return cmpAttrList
	}

	switch fromValTT:=fromVal.(type) {
		case map[string]any:
			toValTT := toVal.(map[string]any)
			for item,fromValMap := range fromValTT {
				if toValMap,ok := toValTT[item]; ok {
					if !reflect.DeepEqual(toValMap,fromValMap){
						cmpAttrList = append(cmpAttrList, determineAttrDiff(fromValMap,toValMap,fmt.Sprintf("%s.%s",key,item))...)
					}
				} else {
						cmpAttrList = append(cmpAttrList, determineAttrDiff(fromValMap,nil,fmt.Sprintf("%s.%s",key,item))...)	
				}
			}

			for item,toValMap := range toValTT {
				if _,ok := fromValTT[item]; !ok {
					cmpAttrList = append(cmpAttrList, determineAttrDiff(nil,toValMap,fmt.Sprintf("%s.%s",key,item))...)	
				}
			}
			return cmpAttrList

		case []any:
			toValTT := toVal.([]any)
			if len(toValTT) < len(fromValTT) {
				index := 0
				for _,item := range toValTT {
					if !reflect.DeepEqual(fromValTT[index],item){
						cmpAttrList = append(cmpAttrList, determineAttrDiff(fromValTT[index],item,fmt.Sprintf("%s[%d]",key,index))...)
					}
					index+=1
				}
				for ; index<len(fromValTT); index++ {
					cmpAttrList = append(cmpAttrList, determineAttrDiff(fromValTT[index],nil,fmt.Sprintf("%s[%d]",key,index))...)
				}
				return cmpAttrList
			} else {
				index := 0
				for _,item := range fromValTT {
					if !reflect.DeepEqual(toValTT[index],item){
						cmpAttrList = append(cmpAttrList, determineAttrDiff(item,toValTT[index],fmt.Sprintf("%s[%d]",key,index))...)
					}
					index+=1
				}
				for ; index<len(toValTT); index++ {
					cmpAttrList = append(cmpAttrList, determineAttrDiff(nil,toValTT[index],fmt.Sprintf("%s[%d]",key,index))...)
				}
				return cmpAttrList
			}

		default:
			if !reflect.DeepEqual(fromVal,toVal){
				cmpAttrList = append(cmpAttrList, &CmpAttr{
					Name: key,
					FromAttr: fromVal,
					ToAttr: toVal,
					Status: MODIFIED,
				})
				return cmpAttrList
			}
	}
	return cmpAttrList
}

func determineMapDiff(current,wanted map[string]any)map[string]*CmpAttr{
	diffMap := make(map[string]*CmpAttr)
	for key,curVal := range current {
		if wanVal,ok := wanted[key]; ok {
			attrsCmp := determineAttrDiff(curVal,wanVal,key)
			for _,item := range attrsCmp {
				if _,inMap := diffMap[item.Name];inMap {
					panic("Item should not appear twice"+item.Name)
				}
				diffMap[item.Name] = item
			}
		} else {
			diffMap[key] = &CmpAttr{
				Name: key,
				FromAttr: curVal,
				ToAttr: nil,
				Status: REMOVED}
		}
	}

	for key,wanVal := range wanted {
		if _,ok := current[key]; !ok {
			diffMap[key] = &CmpAttr{
				Name: key,
				FromAttr: nil,
				ToAttr: wanVal,
				Status: ADDED}
		} 
	}
	return diffMap

}



func generateResourceChanges(m map[string]*CompareResources) []*ResourceChanges{

	resourceChanges := []*ResourceChanges{}
	for name,cmp := range m {
		var mFrom map[string]any = make(map[string]any)
		var mTo map[string]any = make(map[string]any)
		if cmp.Current != nil {
			current := cmp.Current.(*unstructured.Unstructured)
			mFrom = current.Object
		}
		if cmp.Wanted != nil {
			wanted := cmp.Wanted.(*unstructured.Unstructured)
			mTo = wanted.Object
		}
		diffMap := determineMapDiff(mFrom,mTo)
		if len(diffMap) > 0 {
			resourceChange := &ResourceChanges{
								Name: name,
								Lines: []*Line{},
							}
			for key,value := range diffMap {
				if value.FromAttr == nil {
					resourceChange.Lines = append(resourceChange.Lines, 
						&Line{
							Name: key,
							ToValue: value.ToAttr,
							FromValue: nil,
							Status: value.Status,
						},
					)
				} else if value.ToAttr == nil {
					resourceChange.Lines = append(resourceChange.Lines, 
						&Line{
							Name: key,
							ToValue: nil,
							FromValue: value.FromAttr,
							Status: value.Status,
						},
					)
					} else {
						resourceChange.Lines = append(resourceChange.Lines, 
							&Line{
								Name: key,
								ToValue: value.ToAttr,
								FromValue: value.FromAttr,
								Status: value.Status,
							},
						)
					}
			}
			resourceChanges = append(resourceChanges, resourceChange)
		}
	}

	return resourceChanges
}

func (v *View) plainPlanPrinter(m map[string]*CompareResources){

}

func (v *View) planColoredPrinter(m map[string]*CompareResources){
	changes := generateResourceChanges(m)
	_,_ = v.streams.Println("Kubehcl will use the following symbols for each action and attribute")
	_,_ = v.streams.Println()
	_,_ = v.streams.Println(colorstring.Color("[bold][green]+[reset] create"))
	_,_ = v.streams.Println(colorstring.Color("[bold][yellow]~[reset] modify"))
	_,_ = v.streams.Println(colorstring.Color("[bold][red]-[reset] remove"))
	_,_ = v.streams.Println()
	_,_ = v.streams.Println("Kubehcl will perform the following actions:")
	_,_ = v.streams.Println()
	for _,change := range changes {
		_,_ = v.streams.Printf("%s {",change.Name)
		_,_ = v.streams.Println()
		_,_ = v.streams.Println()
	// 	for _,line := range change.Lines {
	// 		switch line.Status {
	// 			case ADDED:
	// 				_,_ = v.streams.Print(colorstring.Color("[bold][green]  +[reset]"))
	// 			case MODIFIED:
	// 				_,_ = v.streams.Print(colorstring.Color("[bold][yellow]  ~[reset]"))
	// 			default:
	// 				_,_ = v.streams.Print(colorstring.Color("[bold][red]  -[reset]"))
	// 		}
	// 		_,_ = v.streams.Printf("%s",line.Value)
	// 		_,_ = v.streams.Println()
	// 	}
		_,_ = v.streams.Println()
		_,_ = v.streams.Println("}")
		_,_ = v.streams.Println()
	// 	_,_ = v.streams.Println()

	} 
}
