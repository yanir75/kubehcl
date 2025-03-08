package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

var variables VariableList

type Variable struct {
	Name        string
	Description string
	Default     cty.Value
	Type        cty.Type
	DeclRange   hcl.Range
	HasValue    bool // for checking if needed request from the user
}

type VariableList []*Variable

var inputVariableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "type", Required: true},
		{Name: "default", Required: false},
		{Name: "description", Required: false},
	},
}

func (varList VariableList) getMapValues() map[string]cty.Value {
	vals := make(map[string]cty.Value)
	vars := make(map[string]cty.Value)

	for _, variable := range varList {
		vals[variable.Name] = variable.Default
	}
	vars["var"] = cty.ObjectVal(vals)
	return vars
}

func decodeVariableBlocks(blocks hcl.Blocks) hcl.Diagnostics {

	var diags hcl.Diagnostics
	names := make(map[string]bool)
	for _, block := range blocks {
		variable, varDiags := decodeVariableBlock(block)
		diags = append(diags, varDiags...)
		if variable != nil {
			variables = append(variables, variable)
			if _, exists := names[variable.Name]; exists {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variables must have different names",
					Detail:   fmt.Sprintf("Two variables have the same name: %s", variable.Name),
					Subject:  &block.DefRange,
					// Context: names[variable.Name],
				})

			}
			names[variable.Name] = true
		}
	}
	return diags
}

func decodeVariableBlock(block *hcl.Block) (*Variable, hcl.Diagnostics) {
	var variable *Variable = &Variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
		HasValue:  false,
	}

	content, diags := block.Body.Content(inputVariableBlockSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	// if len(content.Blocks) > 0 {
	// 	diags = append(diags,&hcl.Diagnostic{
	// 		Severity: hcl.DiagError,
	// 		Summary: "Too many blocks in variable it doesn't suppose to have blocks",
	// 		Detail: fmt.Sprintf("Too many blocks in %s",block.Labels),
	// 		Subject: &block.DefRange,
	// 	})
	// 	return nil, diags
	// }

	if attr, exists := content.Attributes["type"]; exists {
		t, _, valDiags := typeexpr.TypeConstraintWithDefaults(attr.Expr)
		diags = append(diags, valDiags...)
		variable.Type = t
	}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &variable.Description)
		diags = append(diags, valDiags...)
	}

	if attr, exists := content.Attributes["default"]; exists {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)

		if variable.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, variable.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid default value for variable",
					Detail:   fmt.Sprintf("This default value is not compatible with the variable's type constraint: %s.", err),
					Subject:  attr.Expr.Range().Ptr(),
				})
				val = cty.DynamicVal
			}
		}

		variable.Default = val
		variable.HasValue = true
	}

	return variable, diags
}
