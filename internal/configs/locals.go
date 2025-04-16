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

// returns unique address of local
func (l *Local) addr() addrs.Local {
	return addrs.Local{
		Name: l.Name,
	}
}

// Decode local into a decodedlocal structure each local has name, value and range
// Name local name in string format
// Value local cty.value after being decoded into go value
// Range the defenition location in the file
func (l *Local) decode(ctx *hcl.EvalContext) (*decode.DecodedLocal, hcl.Diagnostics) {
	dL := &decode.DecodedLocal{
		Name:      l.Name,
		DeclRange: l.DeclRange,
	}
	value, diags := l.Value.Value(ctx)

	dL.Value = value
	return dL, diags
}

// Decode multiple locals
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

// Decode local block each local block can contain multiple local attributes
// Each attribute must have name and value
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

// Decode multiple local blocks
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
				Subject: &local.DeclRange,
				// Context: names[variable.Name],
			})
		}
	}
	return locals, diags
}
