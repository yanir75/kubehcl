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
	"strings"

	"github.com/hashicorp/hcl/v2"
	// "github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"kubehcl.sh/kubehcl/internal/addrs"
	// "kubehcl.sh/kubehcl/internal/configschema"
)

type Deployable struct {
	Name      string          `json:"Name"`
	ForEach   hcl.Expression  `json:"ForEach"`
	Count     hcl.Expression  `json:"Count"`
	Config    hcl.Body        `json:"Config"`
	Type      string          `json:"Type"`
	DependsOn []hcl.Traversal `json:"DependsOn"`
	DeclRange hcl.Range       `json:"DeclRange"`
}

var commonAttributes = &hcl.BodySchema{
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
	},
	Blocks: []hcl.BlockHeaderSchema{},
}

func (d *Deployable) addr() addrs.Deployable {
	return addrs.Deployable{
		Type: d.Type,
		Name: d.Name,
	}
}

type expandable struct {
	ForEach cty.Value
}

// Expand dynamic blocks this is experimental and and would be defined as a list of maps not multiple maps
func expandDynamicBlock(block *hclsyntax.Block, ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var blocks hclsyntax.Blocks
	var exBlock expandable
	var contentBlock *hclsyntax.Block
	body := block.Body
	if len(block.Labels) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Dynamic block must have exactly one label`,
			Detail:   fmt.Sprintf(`Your dynamic block has more than 1 label %s`, strings.Join(block.Labels, ",")),
			Subject:  &block.TypeRange,
			Context:  &block.LabelRanges[0],
		})
		return cty.NilVal, diags
	}

	for _, b := range body.Blocks {
		if b.Type != "content" {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Dynamic block must have 1 block exactly named content`,
				Detail:   fmt.Sprintf(`Your dynamic block has more than 1 block %s`, b.Type),
				Subject:  &b.TypeRange,
			})
			return cty.NilVal, diags
		} else {
			contentBlock = b
		}
	}

	if len(body.Blocks) != 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Dynamic block must have 1 block exactly named content`,
			Subject:  &block.TypeRange,
		})
		return cty.NilVal, diags
	}

	if attr, exists := body.Attributes["for_each"]; !exists {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Dynamic block must have for_each`,
			Subject:  &block.TypeRange,
		})
		return cty.NilVal, diags
	} else {
		forEach, valDiags := decodeForExpr(ctx, attr.Expr)
		diags = append(diags, valDiags...)
		exBlock.ForEach = forEach
	}
	ty := exBlock.ForEach.Type()
	if ty.IsMapType() || ty.IsObjectType() {
		for key, val := range exBlock.ForEach.AsValueMap() {
			ctx.Variables[block.Type] = cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal(key), "value": val})
			var b hclsyntax.Block
			b.OpenBraceRange = block.OpenBraceRange
			b.CloseBraceRange = block.CloseBraceRange
			b.Type = block.Labels[0]
			b.TypeRange = block.TypeRange
			b.Body = &hclsyntax.Body{}
			b.Body.SrcRange = contentBlock.Body.SrcRange
			b.Body.EndRange = contentBlock.Body.EndRange
			for _, attr := range contentBlock.Body.Attributes {
				b.Body.Attributes[attr.Name] = attr
			}
			b.Body.Blocks = contentBlock.Body.Blocks
		}
	} else if ty.IsSetType() {
		for _, val := range exBlock.ForEach.AsValueSet().Values() {
			ctx.Variables[block.Type] = cty.ObjectVal(map[string]cty.Value{"key": val, "value": val})
			var b hclsyntax.Block
			b.OpenBraceRange = block.OpenBraceRange
			b.CloseBraceRange = block.CloseBraceRange
			b.Type = block.Labels[0]
			b.TypeRange = block.TypeRange
			b.Body = &hclsyntax.Body{Attributes: hclsyntax.Attributes{}}
			b.Body.SrcRange = contentBlock.Body.SrcRange
			b.Body.EndRange = contentBlock.Body.EndRange
			for _, attr := range contentBlock.Body.Attributes {
				b.Body.Attributes[attr.Name] = attr
			}
			b.Body.Blocks = contentBlock.Body.Blocks
			blocks = append(blocks, &b)
		}
	}
	valList := []cty.Value{}
	for _, b := range blocks {
		val, decodeDiags := decodeUnknownBody(ctx, b.Body, false)
		diags = append(diags, decodeDiags...)
		valList = append(valList, val)
	}
	if len(valList) == 0 {
		return cty.NilVal, diags
	}
	return cty.ListVal(valList), diags
}

// Decode body with unknown attributes
// This is a map/body inside a resource or module blocks
// Return the attributes as map[string]cty.value
// String represents the name of the attribute and cty.value the value of the attributes
func checkIfExists(m map[string]cty.Value, key string, ran *hcl.Range) hcl.Diagnostics {
	if _, exists := m[key]; exists {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate values are not allowed",
				Detail:   fmt.Sprintf("Attribute: \"%s\" is mentioned more than once", key),
				Subject:  ran,
			},
		}
	}
	return hcl.Diagnostics{}
}

func decodeUnknownBody(ctx *hcl.EvalContext, body *hclsyntax.Body, check bool) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	attrMap := make(map[string]cty.Value)
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {

			if block.Type == "dynamic" {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Subject:  &block.TypeRange,
					Summary:  "Dynamic blocks will be converted to lists instead of maps",
					// Detail:   fmt.Sprintf("Block has labels: %s and type: \"%s\"", strings.Join(block.Labels, ", "), block.Type),
					Context: &block.LabelRanges[0],
				})

				val, expandDiags := expandDynamicBlock(block, ctx)
				diags = append(diags, expandDiags...)
				if len(block.Labels) > 0 && val != cty.NilVal {
					diags = append(diags, checkIfExists(attrMap, block.Labels[0], &block.LabelRanges[0])...)
					attrMap[block.Labels[0]] = val
				}
			} else {
				m, blockDiags := decodeUnknownBody(ctx, block.Body, false)
				diags = append(diags, blockDiags...)
				diags = append(diags, checkIfExists(attrMap, block.Type, &block.TypeRange)...)
				attrMap[block.Type] = m
			}
		}

	}
	for _, attr := range body.Attributes {
		if check {
			travs := attr.Expr.Variables()
			for _, trav := range travs {
				if len(trav) > 1 {
					root := trav[0].(hcl.TraverseRoot)
					travAttr := trav[1].(hcl.TraverseAttr)
					if root.Name != "each" || (travAttr.Name != "value" && travAttr.Name != "key") {
						val, attrDiags := attr.Expr.Value(ctx)
						diags = append(diags, checkIfExists(attrMap, attr.Name, &attr.NameRange)...)
						diags = append(diags, attrDiags...)
						attrMap[attr.Name] = val
					}
				}
			}
		} else {
			val, attrDiags := attr.Expr.Value(ctx)
			diags = append(diags, checkIfExists(attrMap, attr.Name, &attr.NameRange)...)
			diags = append(diags, attrDiags...)
			attrMap[attr.Name] = val
		}
	}
	return cty.ObjectVal(attrMap), diags
}

// Decode the deployable
// Deployable can contain for each or count which will be decoded into multiple different resources
// It will return a deployable with configmap which contains all the resources available to be deployed
func (r Deployable) Decode(ctx *hcl.EvalContext) (*DecodedDeployable, hcl.Diagnostics) {
	dR := &DecodedDeployable{
		Name:      r.Name,
		Type:      r.Type,
		DependsOn: r.DependsOn,
		DeclRange: r.DeclRange,
	}
	deployMap := make(map[string]cty.Value)
	var diags hcl.Diagnostics
	body, ok := r.Config.(*hclsyntax.Body)

	if !ok {
		panic("should always be ok")
	}
	for _, attrS := range commonAttributes.Attributes {
		delete(body.Attributes, attrS.Name)
	}
	if r.Count != nil {
		count, countDiags := decodeCountExpr(ctx, r.Count)
		diags = append(diags, countDiags...)
		didOperate := false
		for i := cty.NumberIntVal(1); i.LessThanOrEqualTo(count) == cty.True; i = i.Add(cty.NumberIntVal(1)) {
			didOperate = true
			ctx.Variables["count"] = cty.ObjectVal(map[string]cty.Value{"index": i})
			Attributes, countDiags := decodeUnknownBody(ctx, body, false)
			diags = append(diags, countDiags...)
			val, err := convert.Convert(i, cty.String)
			if err != nil {
				panic("Always can convert int")
			}
			deployMap[fmt.Sprintf("%s[%s]", r.addr().String(), val.AsString())] = Attributes
			delete(ctx.Variables, "count")
		}
		// check configuration of the resource
		if !didOperate {
			ctx.Variables["count"] = cty.ObjectVal(map[string]cty.Value{"index": cty.StringVal("1")})
			_, countDiags := decodeUnknownBody(ctx, body, false)
			diags = append(diags, countDiags...)
			delete(ctx.Variables, "count")
		}
	} else if r.ForEach != nil {
		forEach, forEachDiags := decodeForExpr(ctx, r.ForEach)
		diags = append(diags, forEachDiags...)
		ty := forEach.Type()
		didOperate := false
		if ty.IsMapType() || ty.IsObjectType() {
			for key, val := range forEach.AsValueMap() {
				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal(key), "value": val})
				Attributes, forEachDiags := decodeUnknownBody(ctx, body, false)
				diags = append(diags, forEachDiags...)
				deployMap[fmt.Sprintf("%s[%s]", r.addr().String(), key)] = Attributes
				delete(ctx.Variables, "each")
				didOperate = true
			}
		} else if ty.IsSetType() {
			for _, val := range forEach.AsValueSet().Values() {
				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"key": val, "value": val})
				Attributes, forEachDiags := decodeUnknownBody(ctx, body, false)
				diags = append(diags, forEachDiags...)
				deployMap[fmt.Sprintf("%s[%s]", r.addr().String(), val.AsString())] = Attributes
				delete(ctx.Variables, "each")
				didOperate = true
			}
		}
		if !didOperate {

			ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("foo"), "value": cty.StringVal("test")})
			_, forEachDiags := decodeUnknownBody(ctx, body, true)
			diags = append(diags, forEachDiags...)
			delete(ctx.Variables, "each")

		}
	} else {
		Attributes, regDiags := decodeUnknownBody(ctx, body, false)
		diags = append(diags, regDiags...)
		deployMap[r.addr().String()] = Attributes
	}
	dR.Config = deployMap
	return dR, diags
}
