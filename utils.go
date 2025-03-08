package main

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

type AddressMap map[string]interface{}

var addrMap AddressMap = AddressMap{}

var deployMap map[string]cty.Value = make(map[string]cty.Value)


func (m AddressMap) add(key string,value interface{}) bool{
	if _,exists :=m[key]; exists {
		return true
	}
	m[key] = value
	return false
}


func createContext(variables VariableList,locals Locals) (*hcl.EvalContext,hcl.Diagnostics) {
	variableMap,diags := variables.getMapValues()
	localMap := locals.getMapValues(&hcl.EvalContext{
		Variables: variableMap,
		Functions: makeBaseFunctionTable("./"),
	})
	maps.Copy(variableMap, localMap)
	// fmt.Printf("%s",vals["var"].AsValueMap())
	return &hcl.EvalContext{
		Variables: variableMap,
		Functions: makeBaseFunctionTable("./"),
	},diags
}

func decodeCountExpr(ctx *hcl.EvalContext,expr hcl.Expression) (cty.Value,hcl.Diagnostics){
	val, diags := expr.Value(ctx)
	if _, err := convert.Convert(val, cty.Number); err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Cannot convert value to int`,
			Detail:   fmt.Sprintf("Cannot convert this value to int : %s", expr),
			Subject:  expr.Range().Ptr(),
		})
	}
	return val,diags
}

func decodeForExpr(ctx *hcl.EvalContext,expr hcl.Expression) (cty.Value,hcl.Diagnostics){
	val, diags := expr.Value(ctx)
	ty :=val.Type()
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
return val,diags
}
// func checkForBlocks (block *hcl.Block) *hcl.Diagnostic {
// 	content,_,_ :=block.Body.PartialContent(&hcl.BodySchema{})
// 	if len(content.Blocks) > 0 {
// 		return &hcl.Diagnostic{
// 			Severity: hcl.DiagError,
// 			Summary:  "No blocks allowed in default_annotation block",
// 			Detail:   fmt.Sprintf("Couldn't convert value to string, this is value of type: %s",content.Blocks[0].Type),
// 			Subject:  &content.Blocks[0].DefRange,
// 		}
// 	}

// 	return &hcl.Diagnostic{
// 		Summary:  "No blocks are found",
// 		Detail:   fmt.Sprintf("There are 0 blocks in  %s",block.Type),
// 		Subject:  &block.DefRange,
// 	}
// }
