package configs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

// var variables VariableList

type Variable struct {
	Name        string
	Description string
	Default     hcl.Expression
	Type        cty.Type
	DeclRange   hcl.Range
	HasDefault  bool // for checking if needed request from the user
}

type VariableList []*Variable

func (v *Variable) addr() addrs.Variable {
	return addrs.Variable{
		Name: v.Name,
	}
}

var inputVariableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "type", Required: true},
		{Name: "default", Required: false},
		{Name: "description", Required: false},
	},
}

func (v *Variable) decode() (*decode.DecodedVariable, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	dV := &decode.DecodedVariable{
		Name:        v.Name,
		Description: v.Description,
		Type:        v.Type,
		DeclRange:   v.DeclRange,
	}
	val, valDiags := v.Default.Value(nil)
	diags = append(diags, valDiags...)
	
	if v.Type != cty.NilType {
		var err error
		val, err = convert.Convert(val, v.Type)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid default value for variable",
				Detail:   fmt.Sprintf("This default value is not compatible with the variable's type constraint: %s.", err),
				Subject:  v.Default.Range().Ptr(),
			})
			val = cty.DynamicVal
		}
	}
	dV.Default = val

	return dV, diags
}

func (v VariableList) Decode() (decode.DecodedVariableList, hcl.Diagnostics) {
	var dVars decode.DecodedVariableList
	var diags hcl.Diagnostics
	for _, variable := range v {
		dV, varDiags := variable.decode()
		diags = append(diags, varDiags...)
		dVars = append(dVars, dV)
	}

	return dVars, diags
}

func DecodeVariableBlocks(blocks hcl.Blocks, addrMap addrs.AddressMap) (VariableList, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	var variables VariableList
	for _, block := range blocks {
		variable, varDiags := decodeVariableBlock(block)
		diags = append(diags, varDiags...)
		if variable != nil {
			variables = append(variables, variable)
			if addrMap.Add(variable.addr().String(), variable) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variables must have different names",
					Detail:   fmt.Sprintf("Two variables have the same name: %s", variable.Name),
					Subject:  &block.DefRange,
					// Context: names[variable.Name],
				})

			}
		}
	}
	return variables, diags
}

func decodeVariableBlock(block *hcl.Block) (*Variable, hcl.Diagnostics) {
	var variable *Variable = &Variable{
		Name:       block.Labels[0],
		DeclRange:  block.DefRange,
		HasDefault: false,
	}

	content, diags := block.Body.Content(inputVariableBlockSchema)

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
		variable.Default = attr.Expr
		variable.HasDefault = true
	}

	return variable, diags
}
