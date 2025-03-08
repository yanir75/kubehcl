package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/addrs"
)

type Annotation struct {
	Name      string
	Value     hcl.Expression
	DeclRange hcl.Range
}

type Annotations []*Annotation

// var annotaions Annotations

func (d *Annotation) addr() addrs.DefaultAnnotation{
	return addrs.DefaultAnnotation{
		Name: d.Name,
	}
}

func decodeAnnotationsBlock(block *hcl.Block) (Annotations,hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	// names := make(map[string]bool)
	var annotaions Annotations

	for _, attr := range attrs {
		
		// value, valDiag := attr.Expr.Value(ctx)
		// diags = append(diags, valDiag...)
		// if val, err := convert.Convert(value, cty.String); err == nil {
			annotaions = append(annotaions, &Annotation{
				Name:      attr.Name,
				Value:     attr.Expr,
				DeclRange: attr.NameRange,
			})
		// } else {
		// 	diags = append(diags, &hcl.Diagnostic{
		// 		Severity: hcl.DiagError,
		// 		Summary:  "Annotations must be string",
		// 		Detail:   fmt.Sprintf("Couldn't convert value to string, this is value of type: %s", value.Type().FriendlyName()),
		// 		Subject:  attr.Expr.Range().Ptr(),
		// 	})
		// }
		// if exists := names[attr.Name]; exists {
		// 	diags = append(diags, &hcl.Diagnostic{
		// 		Severity: hcl.DiagError,
		// 		Summary:  "Annotations must have different names",
		// 		Detail:   fmt.Sprintf("Two Annotations have the same name: %s", attr.Name),
		// 		// Context: names[variable.Name],
		// 	})
		// }
		// names[attr.Name] = true

	}

	return annotaions,diags
}

func decodeAnnotationsBlocks(blocks hcl.Blocks) (Annotations,hcl.Diagnostics) {
	var annotations Annotations
	var diags hcl.Diagnostics
	for _, block := range blocks {
		annotaionsF,annotationsDiags := decodeAnnotationsBlock(block)
		diags = append(diags, annotationsDiags...)
		annotations = append(annotations, annotaionsF...)
	}
	for _, annotation := range annotations {
		if addrMap.add(annotation.addr().String(),annotation) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Annotations must have different names",
				Detail:   fmt.Sprintf("Two Annotations have the same name: %s", annotation.Name),
				// Context: names[variable.Name],
			})
		}
	}
	return annotations,diags
}
