package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Local struct {
	Name      string
	Value     cty.Value
	DeclRange hcl.Range
}

type Locals []*Local

var locals Locals

// var inputLocalsBlockSchema = &hcl.BodySchema{

// }

func (locals Locals) getMapValues() map[string]cty.Value {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)

	for _, local := range locals {
		vals[local.Name] = local.Value
	}
	vars["local"] = cty.ObjectVal(vals)
	return vars
}

func decodeLocalsBlock(block *hcl.Block) hcl.Diagnostics {

	attrs, diags := block.Body.JustAttributes()
	names := make(map[string]bool)
	for _, attr := range attrs {
		value, valDiag := attr.Expr.Value(createContext())
		diags = append(diags, valDiag...)
		local := &Local{
			Name:      attr.Name,
			Value:     value,
			DeclRange: attr.NameRange,
		}
		if exists := names[local.Name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Variables must have different names",
				Detail:   fmt.Sprintf("Two variables have the same name: %s", local.Name),
				// Context: names[variable.Name],
			})
		}
		locals = append(locals, local)
		names[local.Name] = true
	}

	return diags
}

func decodeLocalsBlocks(blocks hcl.Blocks) hcl.Diagnostics {
	var diags hcl.Diagnostics
	names := make(map[string]bool)
	for _, block := range blocks {
		varDiags := decodeLocalsBlock(block)
		diags = append(diags, varDiags...)
	}
	for _, local := range locals {
		if exists := names[local.Name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Variables must have different names",
				Detail:   fmt.Sprintf("Two variables have the same name: %s", local.Name),
				// Context: names[variable.Name],
			})
		}
		names[local.Name] = true
	}
	return diags
}
