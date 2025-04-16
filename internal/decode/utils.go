/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package decode

import (
	"fmt"
	"maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

const (
	ERROR   = hcl.DiagError
	WARNING = hcl.DiagWarning
	INVALID = hcl.DiagInvalid
)

// Creates hcl context with the variables and locals to decode all values within the blocks
func CreateContext(variables DecodedVariableList, locals DecodedLocals) (*hcl.EvalContext, hcl.Diagnostics) {
	variableMap, diags := variables.getMapValues()
	localMap := locals.getMapValues()
	maps.Copy(variableMap, localMap)
	// fmt.Printf("%s",vals["var"].AsValueMap())
	return &hcl.EvalContext{
		Variables: variableMap,
		Functions: makeBaseFunctionTable("./"),
	}, diags
}

// Decode count expression into a cty.value with type number
func decodeCountExpr(ctx *hcl.EvalContext, expr hcl.Expression) (cty.Value, hcl.Diagnostics) {
	val, diags := expr.Value(ctx)
	if countVal, err := convert.Convert(val, cty.Number); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Cannot convert value to int`,
			Detail:   fmt.Sprintf("Cannot convert this value to int : %s", expr),
			Subject:  expr.Range().Ptr(),
		})
	} else if countVal.LessThan(cty.NumberIntVal(0)) == cty.True {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Count is lower than 0`,
			Detail:   fmt.Sprintf("Expression results in a result lower than 0 : %s", expr),
			Subject:  expr.Range().Ptr(),
		})
	}
	return val, diags
}

// Decode for each expression into a set,object or map type
func decodeForExpr(ctx *hcl.EvalContext, expr hcl.Expression) (cty.Value, hcl.Diagnostics) {
	val, diags := expr.Value(ctx)
	ty := val.Type()
	var isAllowedType bool
	allowedTypesMessage := "map, or set of strings"

	isAllowedType = ty.IsMapType() || ty.IsSetType() || ty.IsObjectType()
	if val.IsKnown() && !isAllowedType {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Invalid for_each argument",
			Detail:      fmt.Sprintf(`The given "for_each" argument value is unsuitable: the "for_each" argument must be a %s, and you have provided a value of type %s.`, allowedTypesMessage, ty.FriendlyName()),
			Subject:     expr.Range().Ptr(),
			Expression:  expr,
			EvalContext: ctx,
		})
	}
	return val, diags
}

