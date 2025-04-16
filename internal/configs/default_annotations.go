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
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"

	// "github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

type Annotation struct {
	Name      string
	Value     hcl.Expression
	DeclRange hcl.Range
}

type Annotations []*Annotation

// Returns a unique address for the default annotation
func (d *Annotation) addr() addrs.DefaultAnnotation {
	return addrs.DefaultAnnotation{
		Name: d.Name,
	}
}

// Decode returns the annotation after it has been assigned go values as a decoded annotation
// It returns the decoded value as a cty.value
// Annotation values must be strings
func (l *Annotation) decode(ctx *hcl.EvalContext) (*decode.DecodedAnnotation, hcl.Diagnostics) {
	dA := &decode.DecodedAnnotation{
		Name:      l.Name,
		DeclRange: l.DeclRange,
	}
	value, diags := l.Value.Value(ctx)
	if !value.Type().Equals(cty.String) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Annotation must be string",
			Detail:   fmt.Sprintf("Annotation has to be string but recevied: %s", typeexpr.TypeString(value.Type())),
			Subject:  l.Value.Range().Ptr(),
		},
		)
	}
	dA.Value = value
	return dA, diags
}

// Decode multiple annotations
func (v Annotations) Decode(ctx *hcl.EvalContext) (decode.DecodedAnnotations, hcl.Diagnostics) {
	var dVars decode.DecodedAnnotations
	var diags hcl.Diagnostics
	for _, variable := range v {
		dV, varDiags := variable.decode(ctx)
		diags = append(diags, varDiags...)
		dVars = append(dVars, dV)
	}

	return dVars, diags
}

// default annotation block turns into multiple annotations of name of annotation and the value of the annotation
func decodeAnnotationsBlock(block *hcl.Block) (Annotations, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()

	var annotaions Annotations

	for _, attr := range attrs {

		annotaions = append(annotaions, &Annotation{
			Name:      attr.Name,
			Value:     attr.Expr,
			DeclRange: attr.NameRange,
		})

	}

	return annotaions, diags
}

// Decode multiple default_annotation blocks and verifies that there aren't any duplicates of the key and values in the attributes
func DecodeAnnotationsBlocks(blocks hcl.Blocks, addrMap addrs.AddressMap) (Annotations, hcl.Diagnostics) {
	var annotations Annotations
	var diags hcl.Diagnostics
	for _, block := range blocks {
		annotaionsF, annotationsDiags := decodeAnnotationsBlock(block)
		diags = append(diags, annotationsDiags...)
		annotations = append(annotations, annotaionsF...)
	}
	for _, annotation := range annotations {
		if addrMap.Add(annotation.addr().String(), annotation) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Annotations must have different names",
				Detail:   fmt.Sprintf("Two Annotations have the same name: %s", annotation.Name),
				// Context: names[variable.Name],
			})
		}
	}
	return annotations, diags
}
