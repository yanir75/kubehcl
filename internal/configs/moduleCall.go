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
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
	// "kubehcl.sh/kubehcl/internal/dag"
)

type ModuleCall struct {
	decode.Deployable
	Source hcl.Expression
}

type ModuleCallList []*ModuleCall

func (m *ModuleCall) DecodeSource(ctx *hcl.EvalContext) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	val, valDdiags := m.Source.Value(ctx)
	diags = append(diags, valDdiags...)
	if val.Type() != cty.String {
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Source must be string",
			Detail:      fmt.Sprintf("Required string and you entered type %s", val.Type().FriendlyName()),
			Subject:     m.Source.Range().Ptr(),
			Expression:  m.Source,
			EvalContext: ctx,
		})
	}
	return val.AsString(), diags
}

func (r ModuleCallList) Decode(ctx *hcl.EvalContext) (decode.DecodedModuleCallList, hcl.Diagnostics) {
	var dR decode.DecodedModuleCallList
	var diags hcl.Diagnostics
	for _, variable := range r {
		dV, varDiags := variable.decode(ctx)
		diags = append(diags, varDiags...)
		dR = append(dR, dV)
	}

	return dR, diags
}

func (r *ModuleCall) decode(ctx *hcl.EvalContext) (*decode.DecodedModuleCall, hcl.Diagnostics) {
	deployable, diags := r.Deployable.Decode(ctx)
	source, sourceDiags := r.DecodeSource(ctx)
	diags = append(diags, sourceDiags...)
	dM := &decode.DecodedModuleCall{
		Source:            source,
		DecodedDeployable: *deployable,
	}

	return dM, diags
}

var inputModuleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "for_each",
		},
		{
			Name: "count",
		},
		{
			Name: "depends_on",
		},
		{
			Name:     "source",
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{},
}

func decodeModuleBlock(block *hcl.Block) (*ModuleCall, hcl.Diagnostics) {
	var Module *ModuleCall = &ModuleCall{
		// Name:      block.Labels[0],
		// DeclRange: block.DefRange,
	}
	Module.Name = block.Labels[0]
	Module.DeclRange = block.DefRange
	Module.Type = addrs.MType

	content, remain, diags := block.Body.PartialContent(inputModuleBlockSchema)
	Module.Config = remain
	if attr, exists := content.Attributes["count"]; exists {
		// val, varDiags := attr.Expr.Value(ctx)
		// diags = append(diags, varDiags...)
		// if count, err := convert.Convert(val, cty.Number); err != nil {
		// diags = append(diags, &hcl.Diagnostic{
		// Severity: hcl.DiagError,
		// Summary:  `Cannot convert value to int`,
		// Detail:   fmt.Sprintf("Cannot convert this value to int : %s", attr.Expr),
		// Subject:  &attr.NameRange,
		// })
		// } else {
		// resource.Count = count
		// }
		Module.Count = attr.Expr
	}
	if attr, exists := content.Attributes["for_each"]; exists {
		if _, countExists := content.Attributes["count"]; countExists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
				Context:  &content.Attributes["count"].NameRange,
			})
			Module.ForEach = attr.Expr
			// val, varDiags := attr.Expr.Value(ctx)
			// diags = append(diags, varDiags...)
			// ty :=val.Type()
			// var isAllowedType bool
			// allowedTypesMessage := "map, or set of strings"

			// isAllowedType = ty.IsMapType() || ty.IsSetType() || ty.IsObjectType()
			// if val.IsKnown() && !isAllowedType {
			// 	diags = diags.Append(&hcl.Diagnostic{
			// 		Severity:    hcl.DiagError,
			// 		Summary:     "Invalid for_each argument",
			// 		Detail:      fmt.Sprintf(`The given "for_each" argument value is unsuitable: the "for_each" argument must be a %s, and you have provided a value of type %s.`, allowedTypesMessage, ty.FriendlyName()),
			// 		Subject:     attr.Expr.Range().Ptr(),
			// 		Expression:  attr.Expr,
			// 		EvalContext: ctx,
			// 	})
			// }
			// resource.ForEach = val
		}
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		traversal, dependsDiags := decodeDependsOn(attr)
		diags = append(diags, dependsDiags...)
		Module.DependsOn = append(Module.DependsOn, traversal...)
	}

	if attr, exists := content.Attributes["source"]; exists {
		Module.Source = attr.Expr
	}

	return Module, diags

}

func DecodeModuleBlocks(blocks hcl.Blocks, addrMap addrs.AddressMap) (ModuleCallList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var moduleList ModuleCallList
	for _, block := range blocks {
		Module, rDiags := decodeModuleBlock(block)
		diags = append(diags, rDiags...)
		moduleList = append(moduleList, Module)
		if addrMap.Add(Module.addr().String(), Module) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Modules must have different names",
				Detail:   fmt.Sprintf("Two Modules have the same name: %s", Module.Name),
				Subject:  &block.DefRange,
				// Context: names[variable.Name],
			})
		}
	}

	return moduleList, diags
}

func (r ModuleCall) addr() addrs.ModuleCall {
	return addrs.ModuleCall{
		Name: r.Name,
	}
}

// func decodeUnknownBody(ctx *hcl.EvalContext,body *hclsyntax.Body) (cty.Value,hcl.Diagnostics){
// 	var diags hcl.Diagnostics
// 	attrMap := make(map[string]cty.Value)
// 	if len(body.Blocks) > 0{
// 		for _,block := range body.Blocks {
// 			m,blockDiags :=decodeUnknownBody(ctx,block.Body)
// 			diags = append(diags, blockDiags...)
// 			attrMap[block.Type] = m
// 		}
// 	}
// 	for _,attr := range body.Attributes {
// 		val,attrDiags :=attr.Expr.Value(ctx)
// 		diags = append(diags, attrDiags...)

// 		attrMap[attr.Name] = val
// 	}
// 	return cty.ObjectVal(attrMap),diags
// }

// func (r *Module) decodeModule() hcl.Diagnostics{
// 	var diags hcl.Diagnostics
// 	body, ok := r.Config.(*hclsyntax.Body)

// 	if !ok {
// 		panic("should always be ok")
// 	}
// 	for _,attrS:= range inputModuleBlockSchema.Attributes{
// 			delete(body.Attributes,attrS.Name)
// 	}
// 	if r.Count != cty.NilVal{
// 		for i :=cty.NumberIntVal(1); i.LessThanOrEqualTo(r.Count) == cty.True; i=i.Add(cty.NumberIntVal(1)) {

// 			ctx := createContext()
// 			ctx.Variables["count"] = cty.ObjectVal(map[string]cty.Value{ "index" : i})
// 			Attributes,countDiags := decodeUnknownBody(ctx,body)
// 			diags = append(diags, countDiags...)
// 			val, err := convert.Convert(i,cty.String)
// 			if err != nil {
// 				panic("Always can convert int")
// 			}
// 			deployMap[fmt.Sprintf("%s[%s]",r.addr().String(),val.AsString())] = Attributes

// 		}
// 	} else if r.ForEach != cty.NilVal{
// 		ty:= r.ForEach.Type()
// 		if ty.IsMapType() || ty.IsObjectType(){
// 			for key,val := range r.ForEach.AsValueMap(){

// 				ctx := createContext()
// 				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{ "key" : cty.StringVal(key),"value": val})
// 				Attributes,forEachDiags := decodeUnknownBody(ctx,body)
// 				diags = append(diags, forEachDiags...)
// 				deployMap[fmt.Sprintf("%s[%s]",r.addr().String(),key)] = Attributes
// 			}
// 		} else{
// 			for _,val := range r.ForEach.AsValueSet().Values(){
// 				ctx := createContext()
// 				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{ "key" : val,"value": val})
// 				Attributes,forEachDiags := decodeUnknownBody(ctx,body)
// 				diags = append(diags, forEachDiags...)
// 				deployMap[fmt.Sprintf("%s[%s]",r.addr().String(),val.AsString())] = Attributes
// 			}
// 		}
// 	} else {
// 		ctx := createContext()
// 		Attributes,regDiags := decodeUnknownBody(ctx,body)
// 		diags = append(diags, regDiags...)
// 		deployMap[r.addr().String()] = Attributes
// 	}
// 	return diags
// }
// // for _,block := range body.Blocks {
// // 	block.
// // }

