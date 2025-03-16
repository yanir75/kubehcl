package configs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

type Local struct {
	Name      string
	Value     hcl.Expression
	DeclRange hcl.Range
}

type Locals []*Local

func (l *Local) addr() addrs.Local {
	return addrs.Local{
		Name: l.Name,
	}
}

// var inputLocalsBlockSchema = &hcl.BodySchema{

// }

func (l *Local) decode(ctx *hcl.EvalContext) (*decode.DecodedLocal, hcl.Diagnostics) {
	dL := &decode.DecodedLocal{
		Name:      l.Name,
		DeclRange: l.DeclRange,
	}
	value, diags := l.Value.Value(ctx)

	dL.Value = value
	return dL, diags
}

func (v Locals) Decode(ctx *hcl.EvalContext) (decode.DecodedLocals, hcl.Diagnostics) {
	var dVars decode.DecodedLocals
	var diags hcl.Diagnostics
	for _, variable := range v {
		dV, varDiags := variable.decode(ctx)
		diags = append(diags, varDiags...)
		dVars = append(dVars, dV)
	}

	return dVars, diags
}

// func (locals Locals) getMapValues(ctx *hcl.EvalContext) map[string]cty.Value {
// 	vals := make(map[string]cty.Value)
// 	vars := make(map[string]cty.Value)
// 	var diags hcl.Diagnostics

// 	for _, local := range locals {
// 		value, valDiag := local.Value.Value(ctx)
// 		diags = append(diags, valDiag...)
// 		vals[local.Name] = value
// 	}
// 	vars["local"] = cty.ObjectVal(vals)
// 	return vars
// }

func decodeLocalsBlock(block *hcl.Block) (Locals, hcl.Diagnostics) {
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

	return locals, diags
}

func DecodeLocalsBlocks(blocks hcl.Blocks, addrMap addrs.AddressMap) (Locals, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var locals Locals

	for _, block := range blocks {
		localsD, localDiags := decodeLocalsBlock(block)
		diags = append(diags, localDiags...)
		locals = append(locals, localsD...)
	}
	for _, local := range locals {
		if addrMap.Add(local.addr().String(), local) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Locals must have different names",
				Detail:   fmt.Sprintf("Two locals have the same name: %s", local.Name),
				// Context: names[variable.Name],
			})
		}
	}
	return locals, diags
}
