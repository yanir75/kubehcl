package main

import (
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
	"kubehcl.sh/kubehcl/internal/configs"
	"kubehcl.sh/kubehcl/internal/decode"
)

var maxGoRountines = 10

var parser = hclparse.NewParser()

// type Task struct {
// 	function func()
// }


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

var varsFile string = "kubehcl.vars"

const (
	rootName = "root"
)

type Module struct {
	Name        string
	Inputs      configs.VariableMap
	Locals      configs.Locals
	Annotations configs.Annotations
	Resources   configs.ResourceList
	ModuleCalls configs.ModuleCallList
}

type ModuleList []*Module

func (m *Module) merge(o *Module) {
	if m.Inputs == nil {
		m.Inputs = o.Inputs
	}else {
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
				Summary:  "Variable has no input",
				Detail:   fmt.Sprintf("Variable %s has no value", input.Name),
			})
		}
	}
	return diags
}

func (m *Module) decode(depth int) (*decode.DecodedModule, hcl.Diagnostics) {
	decodedModule := &decode.DecodedModule{
		Depth: depth,
		Name:  m.Name,
	}
	var diags hcl.Diagnostics
	var modules []*Module
	for _, call := range m.ModuleCalls {
		source,sourceDiags := call.DecodeSource(&hcl.EvalContext{})
		attrs,attrDiags :=call.Config.JustAttributes()
		diags = append(diags, attrDiags...)
		diags = append(diags, sourceDiags...)
		module,modDiags :=decodeFolder(source)
		diags = append(diags, modDiags...)
		for _,attr := range attrs {
			variable  := &configs.Variable{
				Name: attr.Name,
				Default: attr.Expr,
				HasDefault: true,
				DeclRange: attr.Expr.Range(),
			}
			if existingVar,exists := module.Inputs[variable.Name];!exists{
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary: "Variable not declared in module",
					Detail: fmt.Sprintf("Assigned a value to variable which was not declared in the module: %s",variable.Name),
					Subject: &variable.DeclRange,
				})
			} else {
				if existingVar.Type != cty.NilType {
					variable.Type = existingVar.Type
				}
				module.Inputs[variable.Name] = variable
			}
		}
		modules = append(modules, module)
		
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

	for _,module:= range modules {
		dm,dmDiags :=module.decode(depth+1)
		diags = append(diags, dmDiags...)
		decodedModule.Modules = append(decodedModule.Modules, dm)
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
	var vars configs.VariableMap
	// tasks <- Task{func() {
		variables, varDiags := configs.DecodeVariableBlocks(b.Blocks.OfType("variable"))
		vars = variables
		// l.Lock()
		// defer l.Unlock()
		diags = append(diags, varDiags...)
		// wg.Done()

	// }}

	// get ctx

	// decode locals
	var locals configs.Locals
	// tasks <- Task{func() {
		localList, localDiags := configs.DecodeLocalsBlocks(b.Blocks.OfType("locals"), addrMap)
		locals = localList
		// l.Lock()
		// defer l.Unlock()
		diags = append(diags, localDiags...)
		// wg.Done()

	// }}
	
	var defaultAnnotaions configs.Annotations

	// tasks <- Task{func() {
		annotations, annotationsDiags := configs.DecodeAnnotationsBlocks(b.Blocks.OfType("default_annotations"), addrMap)
		defaultAnnotaions = annotations
		// l.Lock()
		// defer l.Unlock()
		diags = append(diags, annotationsDiags...)
		// wg.Done()

	// }}

	var resources configs.ResourceList

	// tasks <- Task{func() {
		resourcesList, resourceDiags := configs.DecodeResourceBlocks(b.Blocks.OfType("resource"), addrMap)
		resources = resourcesList
		// l.Lock()
		// defer l.Unlock()
		diags = append(diags, resourceDiags...)
		// wg.Done()

	// }}

	var modules configs.ModuleCallList

	// tasks <- Task{func() {
		moduleList, moduleDiags := configs.DecodeModuleBlocks(b.Blocks.OfType("module"), addrMap)
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
	var addrMap addrs.AddressMap = addrs.AddressMap{}
	files, err := os.ReadDir(folderName)
	var diags hcl.Diagnostics
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
	for _,module := range moduleList {
		deployable.merge(module)
	}

	verDiags := deployable.verify()
	diags = append(diags, verDiags...)

	// dM, decodeModuleDiags := deployable.decode(depth)
	// diags = append(diags, decodeModuleDiags...)
	// dM.Name = name
	return deployable, diags
}

func decodeFolderAndModules(folderName string, name string, depth int)(*decode.DecodedModule, hcl.Diagnostics){
	mod,diags :=decodeFolder(folderName)
	dm,decodeDiags := mod.decode(0)
	diags = append(diags, decodeDiags...)
	return dm,diags
}