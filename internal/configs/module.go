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
	"strings"


	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/logging"
)

//TODO: remove decode folder from each module and add module caching no reason to decode a module 10 times if it exists in a folder
//TODO: Seperate the decoding file and folder from module part
// var maxGoRountines = 10
var RepoConfigFile *string
var parser = hclparse.NewParser()

const (
	INDEXVARSFILE = "index.hclvars"
	DOWNLOADINDEXFILE = "index.yaml"
)

func Parser() *hclparse.Parser {
	return parser
}

var inputConfig = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "kube_resource",
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
		{
			Type: "backend_storage",
			// LabelNames: []string{"Kind","Name"},
		},
	},
}

var ext string = ".hcl"

var varsFile string = "kubehcl.tfvars"

// Merge multiple modules into one
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
	if m.BackendStorage != nil {
		if !m.BackendStorage.Used {
			m.BackendStorage = o.BackendStorage
		}
	} else {
		m.BackendStorage = o.BackendStorage
	}
}

// Verify that each input has value
func (m *Module) verify() hcl.Diagnostics {
	name := m.Name
	if m.Name == "" {
		name = "root"
	}
	logging.KubeLogger.Info(fmt.Sprintf("Verifying inputs for module: %s", name))

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

func decodeVars(vals []string) (VariableMap, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var variables VariableMap = make(map[string]*Variable)

	for _, val := range vals {
		srcHCL, diagsParse := parser.ParseHCL([]byte(val), "commandline arguments")
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
	}

	return variables, diags
}

// Decode tfvars file into variables
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

func (call *ModuleCall) decodeCallWithFolder(source,folderName string,appFs afero.Fs)(*Module,hcl.Diagnostics){
		var diags hcl.Diagnostics
		
		if string(folderName[len(folderName)-1]) != "/" {
			folderName = folderName + "/"
		}

		if string(source[:2]) == "./" {
			source = source[2:]
		}

		source = folderName + source
		module, modDiags := decodeFolder(source,appFs)

		module.Source = source
		if folderName == source || folderName == source+"./" || folderName+"./" == source {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Circle folder",
				Detail:   fmt.Sprintf("Folder can't be used as a module causes a loop: %s", source),
			})
			return &Module{}, diags
		}

		diags = append(diags, modDiags...)

		// check if variable is declared correctly
		// _,diag := module.Inputs.Decode(nil)
		// diags = append(diags, diag...)
		for _, input := range module.Inputs {
			if input.HasDefault {
				_, diag := input.Default.Value(nil)
				diags = append(diags, diag...)
			}
		}

		attrs, attrDiags := call.Config.JustAttributes()
		diags = append(diags, attrDiags...)

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
		return module,diags
}

func parseSource(source string,r *hcl.Range)(string,string,hcl.Diagnostics){
	strs := strings.Split(source,"/")
	if len(strs) != 4 {
		return "","",hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary: "Source is invalid",
				Detail: fmt.Sprintf("Format of source is repo://<repoName>/<tag or name> got %s",source),
				Subject: r,
			},
		}
	}
	return strs[2],strs[3],hcl.Diagnostics{}
}

func (call *ModuleCall) decodeCallWithRepo(source,prev string)(*Module,hcl.Diagnostics){
		version,_ := call.DecodeVersion(&hcl.EvalContext{})
		repoName,tag,diags := parseSource(source,call.Source.Range().Ptr())
		if diags.HasErrors(){
			return &Module{},diags
		}

		appFs,diags :=	Pull(version,*RepoConfigFile,repoName,tag,false)
		if diags.HasErrors() {
			return &Module{},diags
		}
		module, modDiags := decodeFolder(tag,appFs)

		module.Source = tag
		if prev == tag  {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Circle repo",
				Detail:   fmt.Sprintf("Repo can't be used as a module causes a loop: %s", source),
			})
			return &Module{}, diags
		}

		diags = append(diags, modDiags...)

		// check if variable is declared correctly
		// _,diag := module.Inputs.Decode(nil)
		// diags = append(diags, diag...)
		for _, input := range module.Inputs {
			if input.HasDefault {
				_, diag := input.Default.Value(nil)
				diags = append(diags, diag...)
			}
		}

		attrs, attrDiags := call.Config.JustAttributes()
		diags = append(diags, attrDiags...)

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
		module.Scope = appFs
		return module,diags
}

// Decode module into decoded module
// This decodes module and also modules inside that module
// There are few parameters
// Depth of the module
// Folder to decode
// Namespace to add to each resource if not exists
// previous module context all vars and locals to apply to the variables of the new module
func (m *Module) decode(releaseName string, depth int, folderName string, varsF string, vals []string, prevCtx *hcl.EvalContext,appFs afero.Fs) (*decode.DecodedModule, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if m.Scope != nil {
		appFs = m.Scope
	} else {
		m.Scope = appFs
	}
	decodedModule := &decode.DecodedModule{
		Depth:     depth,
		Name:      m.Name,
		DependsOn: m.DependsOn,
	}

	if depth != 0 {
		if m.BackendStorage.Used {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "backend_storage block is not allowed a module",
				Subject:  &m.BackendStorage.DeclRange,
			})
		}
	}

	if diags.HasErrors() {
		return &decode.DecodedModule{}, diags
	}

	if depth == 0 {
		varsFromFile, varFileDiags := decodeVarsFile(folderName, varsF)
		diags = append(diags, varFileDiags...)
		vars, varsDiags := decodeVars(vals)
		diags = append(diags, varsDiags...)
		for key, variable := range vars {
			if varFromFile, exists := varsFromFile[key]; exists {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variable declared in vars file and commandline",
					Detail:   fmt.Sprintf("Declare the variable in the file or at commandline not both: %s", variable.Name),
					Subject:  &varFromFile.DeclRange,
				})
			}
			varsFromFile[key] = variable
		}

		for key, variable := range varsFromFile {
			if input, exists := m.Inputs[key]; exists {
				input.Default = variable.Default
				input.HasDefault = true
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variable declared in vars but not in file",
					Detail:   fmt.Sprintf("Declare the variable in the file or remove it from vars file variable: %s", variable.Name),
					Subject:  &variable.DeclRange,
				})
			}
		}
		diags = append(diags, m.verify()...)
	}

	var modules ModuleList
	for _, call := range m.ModuleCalls {
		source, sourceDiags := call.DecodeSource(&hcl.EvalContext{})
		
		diags = append(diags, sourceDiags...)
		if strings.HasPrefix(source,"repo://"){
			module,callDiags :=call.decodeCallWithRepo(source,folderName)
			modules = append(modules, module)
			diags = append(diags, callDiags...)
			if module.Name == "" {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary: "This was pulled through a repo",
					Detail: fmt.Sprintf("This module was pull from repo, please validate the source %s",source),
				})
				return &decode.DecodedModule{},diags
			}
			
		} else {
			module,callDiags :=call.decodeCallWithFolder(source,folderName,appFs)
			modules = append(modules, module)
			diags = append(diags, callDiags...)
			if module.Name == "" {
				return &decode.DecodedModule{},diags
			}
		}
	}

	if diags.HasErrors() {
		return &decode.DecodedModule{}, diags
	}

	decodedVariables, decodeVarDiags := m.Inputs.Decode(prevCtx)
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
	if releaseName != "" {
		DecodedAnnotations = append(DecodedAnnotations, &decode.DecodedAnnotation{
			Name:  "kubehcl.sh/managed",
			Value: cty.StringVal("This resource is managed by kubehcl"),
		})
		DecodedAnnotations = append(DecodedAnnotations, &decode.DecodedAnnotation{
			Name:  "kubehcl.sh/release",
			Value: cty.StringVal(releaseName),
		})
	}
	decodedModule.Annotations = DecodedAnnotations

	DecodedResources, decodeResourcesDiags := m.Resources.Decode(ctx)
	diags = append(diags, decodeResourcesDiags...)
	decodedModule.Resources = DecodedResources

	DecodedModuleCalls, decodeModuleCallDiags := m.ModuleCalls.Decode(ctx)
	diags = append(diags, decodeModuleCallDiags...)
	decodedModule.ModuleCalls = DecodedModuleCalls

	for _, module := range modules {
		dm, dmDiags := module.decode(releaseName, depth+1, module.Source, "", make([]string, 0), ctx,appFs)
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

	if depth == 0 {
		storage, storageDiags := m.BackendStorage.decode(ctx)
		diags = append(diags, storageDiags...)
		decodedModule.BackendStorage = storage
	}

	return decodedModule, diags
}

func decodeHclBytes(src []byte, fileName string, addrMap addrs.AddressMap) (Module, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	srcHCL, diagsParse := parser.ParseHCL(src, fileName)
	diags = append(diags, diagsParse...)

	b, blockDiags := srcHCL.Body.Content(inputConfig)
	diags = append(diags, blockDiags...)

	var vars VariableMap
	variables, varDiags := DecodeVariableBlocks(b.Blocks.OfType("variable"))
	vars = variables

	diags = append(diags, varDiags...)

	var locals Locals
	localList, localDiags := DecodeLocalsBlocks(b.Blocks.OfType("locals"), addrMap)
	locals = localList

	diags = append(diags, localDiags...)

	var defaultAnnotaions Annotations

	annotations, annotationsDiags := DecodeAnnotationsBlocks(b.Blocks.OfType("default_annotations"), addrMap)
	defaultAnnotaions = annotations

	diags = append(diags, annotationsDiags...)

	var resources ResourceList

	resourcesList, resourceDiags := DecodeResourceBlocks(b.Blocks.OfType("kube_resource"), addrMap)
	resources = resourcesList

	diags = append(diags, resourceDiags...)

	storageBlock, storageDiags := DecodeBackendStorageBlocks(b.Blocks.OfType("backend_storage"))
	diags = append(diags, storageDiags...)

	var modules ModuleCallList

	moduleList, moduleDiags := DecodeModuleBlocks(b.Blocks.OfType("module"), addrMap)
	modules = moduleList

	diags = append(diags, moduleDiags...)

	return Module{
		BackendStorage: storageBlock,
		Inputs:         vars,
		Locals:         locals,
		Annotations:    defaultAnnotaions,
		Resources:      resources,
		ModuleCalls:    modules,
	}, diags
}

// Decode a single file into a module format
func decodeFile(fileName string, addrMap addrs.AddressMap,appFs afero.Fs) (Module, hcl.Diagnostics) {
	logging.KubeLogger.Info(fmt.Sprintf("Decoding file %s", fileName))

	input, err := appFs.Open(fileName)
	if err != nil {
		fmt.Printf("%s", err)
	}

	defer func() {
		err = input.Close()
		if err != nil {
			panic("Couldn't close the file")
		}
	}()

	src, err := io.ReadAll(input)
	if err != nil {
		fmt.Printf("%s", err)
	}

	return decodeHclBytes(src, fileName, addrMap)
}

// Decode a folder into a module format, this goes over each file in the folder and decodes the files, afterwards it merges the modules.
func decodeFolder(folderName string,appFs afero.Fs) (*Module, hcl.Diagnostics) {
	logging.KubeLogger.Info(fmt.Sprintf("Decoding folder %s", folderName))
	var diags hcl.Diagnostics

	var addrMap = addrs.AddressMap{}
	files, err := afero.ReadDir(appFs,folderName)
	var deployable = &Module{}
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid Source",
			Detail:   fmt.Sprintf("Can not open the given folder %s", folderName),
		})
	}
	var moduleList []*Module
	for _, f := range files {
		if filepath.Ext(f.Name()) == ext {

			dep, decodeDiags := decodeFile(folderName+"/"+f.Name(), addrMap,appFs)

			moduleList = append(moduleList, &dep)
			diags = append(diags, decodeDiags...)

		}
	}

	for _, module := range moduleList {
		deployable.merge(module)
	}

	return deployable, diags
}

func decodeIndexFile(fileName string) (map[string]string, hcl.Diagnostics) {
	logging.KubeLogger.Info(fmt.Sprintf("Decoding file %s", fileName))
	input, err := os.Open(fileName)
	if err != nil {
		return make(map[string]string), hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("File %s is missing", INDEXVARSFILE),
				Detail:   fmt.Sprintf("%s file must be created and populated with the relevant info", INDEXVARSFILE),
			},
		}
	}

	defer func() {
		err = input.Close()
		if err != nil {
			panic("Couldn't close the file")
		}
	}()
	var diags hcl.Diagnostics

	src, err := io.ReadAll(input)
	if err != nil {
		fmt.Printf("%s", err)
	}

	srcHCL, diagsParse := parser.ParseHCL(src, fileName)
	diags = append(diags, diagsParse...)
	attrs, attrDiags := srcHCL.Body.JustAttributes()
	diags = append(diags, attrDiags...)
	var variables VariableMap = make(map[string]*Variable)
	requiredAttrs := []string{"name", "version"}
	for _, item := range requiredAttrs {
		if _, ok := attrs[item]; !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("%s must contain name", INDEXVARSFILE),
				Detail:   fmt.Sprintf("File %s contains descriptive keys for the module, %s is a required key", INDEXVARSFILE, item),
				Subject:  srcHCL.Body.MissingItemRange().Ptr(),
			})
		}
	}

	for _, attr := range attrs {
		variables[attr.Name] = &Variable{
			Name:       attr.Name,
			Default:    attr.Expr,
			HasDefault: true,
			DeclRange:  *attr.Expr.Range().Ptr(),
		}
	}
	decodedVariables, decodeDiags := variables.Decode(&hcl.EvalContext{})
	diags = append(diags, decodeDiags...)
	annotations := make(map[string]string)
	for _, variable := range decodedVariables {
		var str string
		err := gocty.FromCtyValue(variable.Default, &str)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Variables in %s must be strings", INDEXVARSFILE),
				Detail:   fmt.Sprintf("Variable %s is not string error: %s", variable.Name, err.Error()),
				Subject:  &variable.DeclRange,
			})
		} else {
			annotations[variable.Name] = str
		}
	}

	return annotations, diags
}

// Decode both folder and module into a decoded module
func DecodeFolderAndModules(releaseName string, folderName string, name string, varF string, vals []string, depth int) (*decode.DecodedModule, hcl.Diagnostics) {
	if depth == 0 {
		_, diags := decodeIndexFile(folderName + "/" + INDEXVARSFILE)
		if diags.HasErrors() {
			return &decode.DecodedModule{}, diags
		}
	}
	appFs := afero.NewOsFs()
	mod, diags := decodeFolder(folderName,appFs)
	dm, decodeDiags := mod.decode(releaseName, 0, folderName, varF, vals, &hcl.EvalContext{},appFs)
	diags = append(diags, decodeDiags...)
	return dm, diags
}
