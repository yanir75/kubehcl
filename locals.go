package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
)

type Local struct {
	Name      string
	Value     hcl.Expression
	DeclRange hcl.Range
}

type Locals []*Local


func (l *Local) addr() addrs.Local{
	return addrs.Local{
		Name: l.Name,
	}
}
// var inputLocalsBlockSchema = &hcl.BodySchema{

// }

func (locals Locals) getMapValues(ctx *hcl.EvalContext) map[string]cty.Value {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)
	var diags hcl.Diagnostics

	for _, local := range locals {
		value, valDiag := local.Value.Value(ctx)
		diags = append(diags, valDiag...)
		vals[local.Name] = value
	}
	vars["local"] = cty.ObjectVal(vals)
	return vars
}

func decodeLocalsBlock(block *hcl.Block) (Locals,hcl.Diagnostics) {
	var locals Locals
	attrs, diags := block.Body.JustAttributes()
	for _, attr := range attrs {
		local := &Local{
			Name:      attr.Name,
			Value:     attr.Expr,
			DeclRange: attr.NameRange,
		}
		locals = append(locals, local)
	}

	return locals,diags
}

func decodeLocalsBlocks(ctx *hcl.EvalContext,blocks hcl.Blocks) (Locals,hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var locals Locals

	for _, block := range blocks {
		localsD,localDiags := decodeLocalsBlock(block)
		diags = append(diags, localDiags...)
		locals = append(locals, localsD...)
	}
	for _, local := range locals {
		if addrMap.add(local.addr().String(),local) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Locals must have different names",
				Detail:   fmt.Sprintf("Two locals have the same name: %s", local.Name),
				// Context: names[variable.Name],
			})
		}
	}
	return locals,diags
}
