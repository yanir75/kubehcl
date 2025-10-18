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
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	// "kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

// var variables VariableList
type Variable struct {
	Name        string         // `json:"Name"`
	Description string         // `json:"Description"`
	Default     hcl.Expression // `json:"Default"`
	Type        cty.Type       // `json:"Type"`
	DeclRange   hcl.Range      // `json:"DeclRange"`
	HasDefault  bool           // `json:"HasDefault"` // for checking if needed request from the user
}

type VariableMap map[string]*Variable

// func (v *Variable) addr() addrs.Variable {
// 	return addrs.Variable{
// 		Name: v.Name,
// 	}
// }

var inputVariableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "type", Required: false},
		{Name: "default", Required: false},
		{Name: "description", Required: false},
	},
}

func (v *Variable) checkVariableName() hcl.Diagnostics {
	invalidNames := []string{"version", "source"}
	if slices.Contains(invalidNames, v.Name) {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid variable name",
				Detail:   fmt.Sprintf("Variable name %s is reserved for internal use, please use other name", v.Name),
				Subject:  &v.DeclRange,
			},
		}
	}
	return hcl.Diagnostics{}
}

// Decode variable and verify the type of the variable matches the default value defined
// If no type is defined each value will be accepted
// Decode the value into a golang cty.value
func (v *Variable) decode(ctx *hcl.EvalContext) (*decode.DecodedVariable, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	dV := &decode.DecodedVariable{
		Name:        v.Name,
		Description: v.Description,
		Type:        v.Type,
		DeclRange:   v.DeclRange,
	}
	val, valDiags := v.Default.Value(ctx)
	diags = append(diags, valDiags...)

	if v.Type != cty.NilType {
		var err error
		val, err = convert.Convert(val, v.Type)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid value value for variable",
				Detail:   fmt.Sprintf("This value is not compatible with the variable's type constraint: %s.", err),
				Subject:  v.Default.Range().Ptr(),
			})
			val = cty.DynamicVal
		}
	}
	dV.Default = val

	return dV, diags
}

// Decode variable map
func (v VariableMap) Decode(ctx *hcl.EvalContext) (decode.DecodedVariableMap, hcl.Diagnostics) {
	dVars := make(decode.DecodedVariableMap)
	var diags hcl.Diagnostics
	for key, variable := range v {
		dV, varDiags := variable.decode(ctx)
		diags = append(diags, varDiags...)
		if _, ok := dVars[key]; ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Resource exists more than once",
				Detail:   fmt.Sprintf("Resource was already declared %s", key),
				Subject:  &dV.DeclRange,
			})
		}
		dVars[key] = dV
	}

	return dVars, diags
}

// Decode multiple variable blocks
func DecodeVariableBlocks(blocks hcl.Blocks) (VariableMap, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var variables VariableMap = make(map[string]*Variable)
	for _, block := range blocks {
		variable, varDiags := decodeVariableBlock(block)
		diags = append(diags, varDiags...)
		if variable != nil {
			if _, exists := variables[variable.Name]; exists {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variables must have different names",
					Detail:   fmt.Sprintf("Two variables have the same name: %s", variable.Name),
					Subject:  &block.DefRange,
					// Context: names[variable.Name],
				})

			} else {
				variables[variable.Name] = variable
			}
		}
	}
	return variables, diags
}

// Each variable block can contain type,description and default value
// Checks for those
// None of those are required
func decodeVariableBlock(block *hcl.Block) (*Variable, hcl.Diagnostics) {
	var variable = &Variable{
		Name:       block.Labels[0],
		DeclRange:  block.DefRange,
		HasDefault: false,
	}

	content, diags := block.Body.Content(inputVariableBlockSchema)

	if attr, exists := content.Attributes["type"]; exists {
		t, _, valDiags := typeexpr.TypeConstraintWithDefaults(attr.Expr)
		diags = append(diags, valDiags...)
		variable.Type = t
	} // else {
	// 	variable.Type = cty.DynamicPseudoType
	// }

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &variable.Description)
		diags = append(diags, valDiags...)
	}

	if attr, exists := content.Attributes["default"]; exists {
		variable.Default = attr.Expr
		variable.HasDefault = true
	}
	diags = append(diags, variable.checkVariableName()...)

	return variable, diags
}
