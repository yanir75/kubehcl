package main

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)


type Local struct {
	Name string
	Value cty.Value
	DeclRange hcl.Range
}

type Locals []*Local

var locals Locals
// var inputLocalsBlockSchema = &hcl.BodySchema{
	
// }

func (locals Locals) getMapValues() map[string]cty.Value{
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)

	for _,local := range locals {
		vals[local.Name] = local.Value
	}
	vars["local"] = cty.ObjectVal(vals)
	return vars
}

func decodeLocalsBlock(block *hcl.Block) (hcl.Diagnostics){

	attrs,diags := block.Body.JustAttributes()
	for _,attr := range attrs {
		value,valDiag :=  attr.Expr.Value(createContext())
		diags = append(diags,valDiag...)
		locals = append(locals, &Local{
			Name:attr.Name,
			Value: value,
			DeclRange: attr.NameRange,
		})
	}
	


	return diags
}
