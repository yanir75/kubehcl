package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type Annotation struct {
	Name      string
	Value     cty.Value
	DeclRange hcl.Range
}

type Annotations []*Annotation

var annotaions Annotations

func decodeAnnotationsBlock(block *hcl.Block) hcl.Diagnostics {
	attrs, diags := block.Body.JustAttributes()
	names := make(map[string]bool)

	for _, attr := range attrs {
		value, valDiag := attr.Expr.Value(createContext())
		diags = append(diags, valDiag...)
		if val, err := convert.Convert(value, cty.String); err == nil {
			annotaions = append(annotaions, &Annotation{
				Name:      attr.Name,
				Value:     val,
				DeclRange: attr.NameRange,
			})
		} else {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Annotations must be string",
				Detail:   fmt.Sprintf("Couldn't convert value to string, this is value of type: %s", value.Type().FriendlyName()),
				Subject:  attr.Expr.Range().Ptr(),
			})
		}
		if exists := names[attr.Name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Annotations must have different names",
				Detail:   fmt.Sprintf("Two Annotations have the same name: %s", attr.Name),
				// Context: names[variable.Name],
			})
		}
		names[attr.Name] = true

	}

	return diags
}

func decodeAnnotationsBlocks(blocks hcl.Blocks) hcl.Diagnostics {
	var diags hcl.Diagnostics
	names := make(map[string]bool)
	for _, block := range blocks {
		varDiags := decodeAnnotationsBlock(block)
		diags = append(diags, varDiags...)
	}
	for _, annotation := range annotaions {
		if exists := names[annotation.Name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Annotations must have different names",
				Detail:   fmt.Sprintf("Two Annotations have the same name: %s", annotation.Name),
				// Context: names[variable.Name],
			})
		}
		names[annotation.Name] = true
	}
	return diags
}
