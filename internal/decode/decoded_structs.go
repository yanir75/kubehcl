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

	"fmt"

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

type DecodedResourceMap map[string]*DecodedResource

func (rMap DecodedResourceMap) Add(r *DecodedResource) hcl.Diagnostics {
	_, ok := rMap[r.Name]
	if ok {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "resource already exists",
				Detail:   fmt.Sprintf("Resource is declared twice %s", r.Name),
				Subject:  &r.DeclRange,
			},
		}
	}
	rMap[r.Name] = r
	return hcl.Diagnostics{}
}

type DecodedBackendStorage struct {
	Kind      string
	DeclRange hcl.Range
}

type DecodedLocal struct {
	Name      string
	Value     cty.Value
	DeclRange hcl.Range
}

type DecodedLocalsMap map[string]*DecodedLocal

type DecodedAnnotation struct {
	Name      string
	Value     cty.Value
	DeclRange hcl.Range
}

type DecodedAnnotationsMap map[string]*DecodedAnnotation

type DecodedVariable struct {
	Name        string
	Description string
	Default     cty.Value
	Type        cty.Type
	DeclRange   hcl.Range
}

type DecodedVariableMap map[string]*DecodedVariable

type DecodedModuleCall struct {
	DecodedDeployable
	Source string
}

type DecodedRepo struct {
	Name                  string
	DeclRange             hcl.Range
	Url                   string
	Protocol              string
	Username              string
	Password              string
	Timeout               int64
	CertFile              string
	KeyFile               string
	CaFile                string
	InsecureSkipTLSverify bool
	PlainHttp             bool
	RepoFile              string
	RepoCache             string
}

type DecodedRepoMap map[string]*DecodedRepo

type DecodedModule struct {
	Name           string
	Inputs         DecodedVariableMap
	Locals         DecodedLocalsMap
	Annotations    DecodedAnnotationsMap
	Resources      DecodedResourceMap
	ModuleCalls    DecodedModuleCallMap
	Modules        DecodedModuleMap
	BackendStorage *DecodedBackendStorage
	Depth          int
	DependsOn      []hcl.Traversal
	Dependencies   []DependsOn
}

type DecodedModuleMap map[string]*DecodedModule

type DecodedModuleCallMap map[string]*DecodedModuleCall

// Get variables as map[string]cty.value
func (varList DecodedVariableMap) getMapValues() (map[string]cty.Value, hcl.Diagnostics) {
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

func (locals DecodedLocalsMap) getMapValues() map[string]cty.Value {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)
	for _, local := range locals {
		vals[local.Name] = local.Value
	}
	vars["local"] = cty.ObjectVal(vals)
	return vars
}
