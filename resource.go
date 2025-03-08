package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	// "kubehcl.sh/kubehcl/internal/dag"
)

func decodeDependsOn(attr *hcl.Attribute) ([]hcl.Traversal, hcl.Diagnostics) {
	var ret []hcl.Traversal
	exprs, diags := hcl.ExprList(attr.Expr)

	for _, expr := range exprs {
		traversal, travDiags := hcl.AbsTraversalForExpr(expr)
		diags = append(diags, travDiags...)
		if len(traversal) != 0 {
			ret = append(ret, traversal)
		}
	}

	return ret, diags
}

type Resource struct {
	Name      string
	ForEach   cty.Value
	Count     cty.Value
	DependsOn []hcl.Traversal
	DeclRange hcl.Range
	Body      hcl.Body
}

type ResourceList []*Resource

var inputResourceBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "for_each", Required: false},
		{Name: "count", Required: false},
		{Name: "depends_on", Required: false},
	},
}

var resourceList ResourceList

func decodeResourceBlock(block *hcl.Block) (*Resource, hcl.Diagnostics) {
	var resource *Resource = &Resource{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, body, diags := block.Body.PartialContent(inputResourceBlockSchema)
	resource.Body = body
	if attr, exists := content.Attributes["count"]; exists {
		val, varDiags := attr.Expr.Value(createContext())
		diags = append(diags, varDiags...)
		if count, err := convert.Convert(val, cty.Number); err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Cannot convert value to int`,
				Detail:   fmt.Sprintf("Cannot convert this value to int : %s", attr.Expr),
				Subject:  &attr.NameRange,
			})
		} else {
			resource.Count = count
		}
	}
	if attr, exists := content.Attributes["for_each"]; exists {
		if resource.Count != cty.NilVal {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
				Context:  &content.Attributes["count"].NameRange,
			})
		}
		val, varDiags := attr.Expr.Value(createContext())
		diags = append(diags, varDiags...)
		var for_each cty.Value = cty.NilVal
		var err error
		if for_each, err = convert.Convert(val, cty.Map(cty.DynamicPseudoType)); err != nil {
			if for_each, err = convert.Convert(val, cty.Set(cty.DynamicPseudoType)); err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  `Cannot convert value to map or set of string`,
					Subject:  &attr.NameRange,
				})
			}
		}
		resource.ForEach = for_each
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		traversal, dependsDiags := decodeDependsOn(attr)
		diags = append(diags, dependsDiags...)
		resource.DependsOn = append(resource.DependsOn, traversal...)
	}

	return resource, diags

}

func decodeResourceBlocks(blocks hcl.Blocks) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, block := range blocks {
		resource, rDiags := decodeResourceBlock(block)
		diags = append(diags, rDiags...)
		resourceList = append(resourceList, resource)

	}

	return diags
}
