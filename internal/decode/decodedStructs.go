package decode

import (
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

type DecodedResource struct {
	DecodedDeployable
}

type DecodedResourceList []*DecodedResource

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

type DecodedModule struct {
	Name        string
	Inputs      DecodedVariableList
	Locals      DecodedLocals
	Annotations DecodedAnnotations
	Resources   DecodedResourceList
	ModuleCalls DecodedModuleCallList
	Modules     DecodedModuleList
	Depth       int
}

type DecodedModuleList []*DecodedModule

type DecodedModuleCallList []*DecodedModuleCall

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

func (locals DecodedLocals) getMapValues() map[string]cty.Value {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)
	for _, local := range locals {
		vals[local.Name] = local.Value
	}
	vars["local"] = cty.ObjectVal(vals)
	return vars
}
