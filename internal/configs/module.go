/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package configs

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"

	// "sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

//TODO remove decode folder from each module and add module caching no reason to decode a module 10 times if it exists in a folder
// var maxGoRountines = 10

var parser = hclparse.NewParser()

//	type Task struct {
//		function func()
//	}
func Parser() *hclparse.Parser {
	return parser
}

var inputConfig = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "resource",
			LabelNames: []string{"Name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"Name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"Name"},
		},
		{
			Type: "locals",
			// LabelNames: []string{"Kind","Name"},
		},
		{
			Type: "default_annotations",
			// LabelNames: []string{"Kind","Name"},
		},
	},
}

var ext string = ".hcl"

var varsFile string = "kubehcl.tfvars"

const (
	rootName = "root"
)

type Module struct {
	Name        string
	Inputs      VariableMap
	Locals      Locals
	Annotations Annotations
	Resources   ResourceList
	ModuleCalls ModuleCallList
	DependsOn   []hcl.Traversal
	Source      string
}

type ModuleList []*Module

func (m *Module) merge(o *Module) {
	if m.Inputs == nil {
		m.Inputs = o.Inputs
	} else {
		maps.Copy(m.Inputs, o.Inputs)
	}
	m.Locals = append(m.Locals, o.Locals...)
	m.Annotations = append(m.Annotations, o.Annotations...)
	m.Resources = append(m.Resources, o.Resources...)
	m.ModuleCalls = append(m.ModuleCalls, o.ModuleCalls...)
}

func (m *Module) verify() hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, input := range m.Inputs {
		if !input.HasDefault {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Variable has no value",
				Detail:   fmt.Sprintf("Variable %s has no value", input.Name),
				Subject:  &input.DeclRange,
			})
		}
	}
	return diags
}

func decodeVarsFile(folderName, fileName string) (VariableMap, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var variables VariableMap = make(map[string]*Variable)
	if fileName == "" {
		fileName = varsFile
	}
	if string(folderName[len(folderName)-1]) != "/" {
		folderName = folderName + "/"
	}
	fullName := folderName + fileName
	if _, err := os.Stat(fullName); errors.Is(err, os.ErrNotExist) {
		return VariableMap{}, diags
	}

	f, err := os.Open(fullName)
	if err != nil {
		fmt.Printf("%s", err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("%s", err)
	}

	srcHCL, diagsParse := parser.ParseHCL(data, fileName)
	diags = append(diags, diagsParse...)
	attrs, attrDiags := srcHCL.Body.JustAttributes()
	diags = append(diags, attrDiags...)
	for _, attr := range attrs {
		variables[attr.Name] = &Variable{
			Name:       attr.Name,
			Default:    attr.Expr,
			HasDefault: true,
			DeclRange:  attr.NameRange,
		}
	}

	return variables, diags

}

func (m *Module) decode(depth int, folderName string) (*decode.DecodedModule, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	decodedModule := &decode.DecodedModule{
		Depth:     depth,
		Name:      m.Name,
		DependsOn: m.DependsOn,
	}

	if depth == 0 {
		vars, varFileDiags := decodeVarsFile(folderName, "")
		diags = append(diags, varFileDiags...)
		for key, variable := range vars {
			if input, exists := m.Inputs[key]; exists {
				input.Default = variable.Default
				input.HasDefault = true
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variable declared in vars but not in file",
					Detail:   fmt.Sprintf("Declare the variable in the file or remove it from vars file variabe: %s", variable.Name),
					Subject:  &variable.DeclRange,
				})
			}
		}
		diags = append(diags, m.verify()...)
	}

	if diags.HasErrors() {
		return &decode.DecodedModule{}, diags
	}

	var modules ModuleList
	for _, call := range m.ModuleCalls {
		source, sourceDiags := call.DecodeSource(&hcl.EvalContext{})
		if string(folderName[len(folderName)-1]) != "/" {
			folderName = folderName + "/"
		}

		if string(source[:2]) == "./" {
			source = source[2:]
		}
		source = folderName + source
		attrs, attrDiags := call.Config.JustAttributes()
		diags = append(diags, attrDiags...)
		diags = append(diags, sourceDiags...)
		module, modDiags := decodeFolder(source)
		module.Source = source
		if folderName == source || folderName == source+"./" || folderName+"./" == source {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Circle folder",
				Detail:   fmt.Sprintf("Folder can't be used as a module causes a loop: %s", source),
			})
			return &decode.DecodedModule{}, diags
		}

		diags = append(diags, modDiags...)
		for _, attr := range attrs {
			if attr.Name == "depends_on" {
				continue
			}
			variable := &Variable{
				Name:       attr.Name,
				Default:    attr.Expr,
				HasDefault: true,
				DeclRange:  attr.Expr.Range(),
			}
			if existingVar, exists := module.Inputs[variable.Name]; !exists {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variable not declared in module",
					Detail:   fmt.Sprintf("Assigned a value to variable which was not declared in the module: %s", variable.Name),
					Subject:  &variable.DeclRange,
				})
			} else {
				if existingVar.Type != cty.NilType {
					variable.Type = existingVar.Type
				}
				module.Inputs[variable.Name] = variable
			}
		}
		for _, input := range module.Inputs {
			if !input.HasDefault {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variable requires value",
					Detail:   fmt.Sprintf("Need to assign a value to variable which was declared in the module: %s", input.Name),
					Subject:  &input.DeclRange,
				})
			}
		}
		module.Name = call.Name
		module.DependsOn = call.DependsOn
		modules = append(modules, module)

	}

	if diags.HasErrors() {
		return &decode.DecodedModule{}, diags
	}

	decodedVariables, decodeVarDiags := m.Inputs.Decode()
	diags = append(diags, decodeVarDiags...)
	decodedModule.Inputs = decodedVariables

	ctx, ctxDiags := decode.CreateContext(decodedVariables, decode.DecodedLocals{})
	diags = append(diags, ctxDiags...)

	DecodedLocals, decodeLocalsDiags := m.Locals.Decode(ctx)
	diags = append(diags, decodeLocalsDiags...)
	decodedModule.Locals = DecodedLocals

	ctx, ctxDiags = decode.CreateContext(decodedVariables, DecodedLocals)
	diags = append(diags, ctxDiags...)

	DecodedAnnotations, decodeAnnotationsDiags := m.Annotations.Decode(ctx)
	diags = append(diags, decodeAnnotationsDiags...)
	decodedModule.Annotations = DecodedAnnotations

	DecodedResources, decodeResourcesDiags := m.Resources.Decode(ctx)
	diags = append(diags, decodeResourcesDiags...)
	decodedModule.Resources = DecodedResources

	DecodedModuleCalls, decodeModuleCallDiags := m.ModuleCalls.Decode(ctx)
	diags = append(diags, decodeModuleCallDiags...)
	decodedModule.ModuleCalls = DecodedModuleCalls

	for _, module := range modules {
		dm, dmDiags := module.decode(depth+1, module.Source)
		diags = append(diags, dmDiags...)
		decodedModule.Modules = append(decodedModule.Modules, dm)
	}
	/*
		Adding annotations to each decoded resource only if it has metadata defined beforehand
	*/
	for _, resource := range DecodedResources {
		for res, resInfo := range resource.Config {
			if resInfo.Type().IsObjectType() || resInfo.Type().IsMapType() {
				resInfoMap := resInfo.AsValueMap()
				if val, exists := resInfoMap["metadata"]; exists {
					if val.Type().IsObjectType() || val.Type().IsMapType() {
						metadata := val.AsValueMap()
						if annotations, exists := metadata["annotations"]; exists {
							if annotations.Type().IsObjectType() || annotations.Type().IsMapType() {
								annotationsMap := val.AsValueMap()
								for _, v := range DecodedAnnotations {
									if _, exists := annotationsMap[v.Name]; !exists {
										annotationsMap[v.Name] = v.Value
									}
									metadata["annotations"] = cty.ObjectVal(annotationsMap)
								}
							}
						} else {
							annotationsMap := make(map[string]cty.Value)
							for _, v := range DecodedAnnotations {
								if _, exists := annotationsMap[v.Name]; !exists {
									annotationsMap[v.Name] = v.Value
								}

							}
							metadata["annotations"] = cty.ObjectVal(annotationsMap)
						}
						resInfoMap["metadata"] = cty.ObjectVal(metadata)
					}
				}
				resource.Config[res] = cty.ObjectVal(resInfoMap)
			}
		}
	}

	return decodedModule, diags
}

func decodeFile(fileName string, addrMap addrs.AddressMap) (Module, hcl.Diagnostics) {
	// wg := sync.WaitGroup{}
	// wg.Add(5)
	input, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("%s", err)
	}
	defer input.Close()
	var diags hcl.Diagnostics

	src, err := io.ReadAll(input)
	if err != nil {
		fmt.Printf("%s", err)
	}

	srcHCL, diagsParse := parser.ParseHCL(src, fileName)
	diags = append(diags, diagsParse...)

	b, blockDiags := srcHCL.Body.Content(inputConfig)
	diags = append(diags, blockDiags...)

	// Decode variables
	var vars VariableMap
	// tasks <- Task{func() {
	variables, varDiags := DecodeVariableBlocks(b.Blocks.OfType("variable"))
	vars = variables
	// l.Lock()
	// defer l.Unlock()
	diags = append(diags, varDiags...)
	// wg.Done()

	// }}

	// get ctx

	// decode locals
	var locals Locals
	// tasks <- Task{func() {
	localList, localDiags := DecodeLocalsBlocks(b.Blocks.OfType("locals"), addrMap)
	locals = localList
	// l.Lock()
	// defer l.Unlock()
	diags = append(diags, localDiags...)
	// wg.Done()

	// }}

	var defaultAnnotaions Annotations

	// tasks <- Task{func() {
	annotations, annotationsDiags := DecodeAnnotationsBlocks(b.Blocks.OfType("default_annotations"), addrMap)
	defaultAnnotaions = annotations
	// l.Lock()
	// defer l.Unlock()
	diags = append(diags, annotationsDiags...)
	// wg.Done()

	// }}

	var resources ResourceList

	// tasks <- Task{func() {
	resourcesList, resourceDiags := DecodeResourceBlocks(b.Blocks.OfType("resource"), addrMap)
	resources = resourcesList
	// l.Lock()
	// defer l.Unlock()
	diags = append(diags, resourceDiags...)
	// wg.Done()

	// }}

	var modules ModuleCallList

	// tasks <- Task{func() {
	moduleList, moduleDiags := DecodeModuleBlocks(b.Blocks.OfType("module"), addrMap)
	modules = moduleList
	// l.Lock()
	// defer l.Unlock()
	diags = append(diags, moduleDiags...)
	// wg.Done()
	// }}

	// wg.Wait()

	return Module{
		Inputs:      vars,
		Locals:      locals,
		Annotations: defaultAnnotaions,
		Resources:   resources,
		ModuleCalls: modules,
	}, diags
}

func decodeFolder(folderName string) (*Module, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	var addrMap addrs.AddressMap = addrs.AddressMap{}
	files, err := os.ReadDir(folderName)
	var deployable *Module = &Module{}
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid Source",
			Detail:   fmt.Sprintf("Can not open the given folder %s", folderName),
		})
	}
	var moduleList []*Module
	// wg := sync.WaitGroup{}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ext {
			// wg.Add(1)
			// tasks <- Task{func(){
			dep, decodeDiags := decodeFile(folderName+"/"+f.Name(), addrMap)
			// l.Lock()
			// defer l.Unlock()
			moduleList = append(moduleList, &dep)
			diags = append(diags, decodeDiags...)
			// wg.Done()
			// }}
		}
	}
	// wg.Wait()
	for _, module := range moduleList {
		deployable.merge(module)
	}

	// dM, decodeModuleDiags := deployable.decode(depth)
	// diags = append(diags, decodeModuleDiags...)
	// dM.Name = name
	return deployable, diags
}

func DecodeFolderAndModules(folderName string, name string, depth int) (*decode.DecodedModule, hcl.Diagnostics) {
	mod, diags := decodeFolder(folderName)
	dm, decodeDiags := mod.decode(0, folderName)
	diags = append(diags, decodeDiags...)
	return dm, diags
}
