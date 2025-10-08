/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package decode

import (
	// "maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
)

type DecodedDeployable struct {
	Name      string
	Config    map[string]cty.Value
	Type      string
	DependsOn []hcl.Traversal
	DeclRange hcl.Range
}

func (d *DecodedDeployable) Addr() addrs.Deployable {
	return addrs.Deployable{
		Type: d.Type,
		Name: d.Name,
	}
}

type DependsOn struct {
	Trav  []hcl.Traversal
	Depth int
}
type DecodedResource struct {
	DecodedDeployable
	Depth                int
	Dependencies         []DependsOn
	DependenciesAppended []DependsOn
}

type DecodedResourceList []*DecodedResource

type DecodedBackendStorage struct {
	Kind      string
	DeclRange hcl.Range
}

type DecodedLocal struct {
	Name      string
	Value     cty.Value
	DeclRange hcl.Range
}

type DecodedLocals []*DecodedLocal

type DecodedAnnotation struct {
	Name      string
	Value     cty.Value
	DeclRange hcl.Range
}

type DecodedAnnotations []*DecodedAnnotation

type DecodedVariable struct {
	Name        string
	Description string
	Default     cty.Value
	Type        cty.Type
	DeclRange   hcl.Range
}

type DecodedVariableList []*DecodedVariable

type DecodedModuleCall struct {
	DecodedDeployable
	Source string
}

type DecodedRepo struct {
	Name        string         
	DeclRange   hcl.Range      
	Url string
	Username string
	Password string
	Timeout int64
	CertFile string
	KeyFile string
	CaFile string
	InsecureSkipTLSverify bool
	PlainHttp bool
	RepoFile string
	RepoCache string
}

type DecodedRepoMap map[string]*DecodedRepo

type DecodedModule struct {
	Name           string
	Inputs         DecodedVariableList
	Locals         DecodedLocals
	Annotations    DecodedAnnotations
	Resources      DecodedResourceList
	ModuleCalls    DecodedModuleCallList
	Modules        DecodedModuleList
	BackendStorage *DecodedBackendStorage
	Depth          int
	DependsOn      []hcl.Traversal
	Dependencies   []DependsOn
}

type DecodedModuleList []*DecodedModule

type DecodedModuleCallList []*DecodedModuleCall

// Get variables as map[string]cty.value
func (varList DecodedVariableList) getMapValues() (map[string]cty.Value, hcl.Diagnostics) {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)
	var diags hcl.Diagnostics
	for _, variable := range varList {
		vals[variable.Name] = variable.Default
	}
	vars["var"] = cty.ObjectVal(vals)
	return vars, diags
}

// Get locals as map[string]cty.value

func (locals DecodedLocals) getMapValues() map[string]cty.Value {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)
	for _, local := range locals {
		vals[local.Name] = local.Value
	}
	vars["local"] = cty.ObjectVal(vals)
	return vars
}
