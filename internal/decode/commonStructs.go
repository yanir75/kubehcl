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
	Name      string
	ForEach   hcl.Expression
	Count     hcl.Expression
	Config    hcl.Body
	Type      string
	DependsOn []hcl.Traversal
	DeclRange hcl.Range
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

func expandDynamicBlock(block *hclsyntax.Block,ctx *hcl.EvalContext) (cty.Value,hcl.Diagnostics){
	var diags hcl.Diagnostics
	var blocks hclsyntax.Blocks
	var exBlock expandable
	var contentBlock *hclsyntax.Block
	body := block.Body
	if len(block.Labels) > 1 {
		diags = append(diags,&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Dynamic block must have exactly one label`,
			Detail:   fmt.Sprintf(`Your dynamic block has more than 1 label %s`,strings.Join(block.Labels, ",")),
			Subject:  &block.TypeRange,
			Context: &block.LabelRanges[0],
		})
		return cty.NilVal,diags
	}

	for _,b := range body.Blocks{
		if b.Type != "content" {
			diags = append(diags,&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Dynamic block must have 1 block exactly named content`,
				Detail:   fmt.Sprintf(`Your dynamic block has more than 1 block %s`,b.Type),
				Subject:  &b.TypeRange,
			})
			return cty.NilVal,diags
		} else {
			contentBlock = b
		}
	}
	
	if len(body.Blocks) != 1 {
		diags = append(diags,&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Dynamic block must have 1 block exactly named content`,
			Subject:  &block.TypeRange,
		})
		return cty.NilVal,diags
	}

	if attr, exists :=body.Attributes["for_each"]; !exists {
		diags = append(diags,&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Dynamic block must have for_each`,
			Subject:  &block.TypeRange,
		})
		return cty.NilVal,diags
	} else {
		forEach,valDiags := decodeForExpr(ctx,attr.Expr)
		diags = append(diags, valDiags...)
		exBlock.ForEach = forEach
	}
	ty := exBlock.ForEach.Type()
	if ty.IsMapType() || ty.IsObjectType(){
		for key,val := range exBlock.ForEach.AsValueMap(){
			ctx.Variables[block.Type] = cty.ObjectVal(map[string]cty.Value{ "key" : cty.StringVal(key),"value": val})
			var b hclsyntax.Block
			b.OpenBraceRange = block.OpenBraceRange
			b.CloseBraceRange = block.CloseBraceRange
			b.Type = block.Labels[0]
			b.TypeRange = block.TypeRange
			b.Body = &hclsyntax.Body{}
			b.Body.SrcRange = contentBlock.Body.SrcRange
			b.Body.EndRange = contentBlock.Body.EndRange
			for _,attr := range contentBlock.Body.Attributes {
				b.Body.Attributes[attr.Name] = attr
			}
			b.Body.Blocks = contentBlock.Body.Blocks
		}
	} else if ty.IsSetType() {
		for _,val := range exBlock.ForEach.AsValueSet().Values(){
			ctx.Variables[block.Type] = cty.ObjectVal(map[string]cty.Value{ "key" : val,"value": val})
			var b hclsyntax.Block
			b.OpenBraceRange = block.OpenBraceRange
			b.CloseBraceRange = block.CloseBraceRange
			b.Type = block.Labels[0]
			b.TypeRange = block.TypeRange
			b.Body = &hclsyntax.Body{Attributes: hclsyntax.Attributes{}}
			b.Body.SrcRange = contentBlock.Body.SrcRange
			b.Body.EndRange = contentBlock.Body.EndRange
			for _,attr := range contentBlock.Body.Attributes {
				b.Body.Attributes[attr.Name] = attr
			}
			b.Body.Blocks = contentBlock.Body.Blocks
			blocks = append(blocks, &b)
		}
	}
	valList := []cty.Value{}
	for _,b := range blocks {
		val,decodeDiags := decodeUnknownBody(ctx,b.Body)	
		diags = append(diags, decodeDiags...)
		valList = append(valList, val)
	}
	if len(valList) == 0 {
		return cty.NilVal,diags
	} 
	return cty.ListVal(valList),diags
}

func decodeUnknownBody(ctx *hcl.EvalContext, body *hclsyntax.Body) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	attrMap := make(map[string]cty.Value)
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {
			// if len(block.Labels) > 0 {
			// 	diags = append(diags, &hcl.Diagnostic{
			// 		Severity: hcl.DiagError,
			// 		Subject:  &block.TypeRange,
			// 		Summary:  "Block shouldn't have labels",
			// 		Detail:   fmt.Sprintf("Block has labels: %s and type: \"%s\"", strings.Join(block.Labels, ", "), block.Type),
			// 		Context:  &block.LabelRanges[0],
			// 	})
			// }
			if block.Type == "dynamic" {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Subject:  &block.TypeRange,
					Summary:  "Dynamic blocks will be converted to lists instead of maps",
					// Detail:   fmt.Sprintf("Block has labels: %s and type: \"%s\"", strings.Join(block.Labels, ", "), block.Type),
					Context:  &block.LabelRanges[0],
				})

				val,expandDiags :=expandDynamicBlock(block,ctx)
				diags = append(diags, expandDiags...)
				if len(block.Labels) > 0 && val!= cty.NilVal{
					attrMap[block.Labels[0]]= val
				}
			}
			m, blockDiags := decodeUnknownBody(ctx, block.Body)
			diags = append(diags, blockDiags...)
			attrMap[block.Type] = m
		}

	}
	for _, attr := range body.Attributes {
		val, attrDiags := attr.Expr.Value(ctx)
		diags = append(diags, attrDiags...)
		attrMap[attr.Name] = val
	}
	return cty.ObjectVal(attrMap), diags
}

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
		for i := cty.NumberIntVal(1); i.LessThanOrEqualTo(count) == cty.True; i = i.Add(cty.NumberIntVal(1)) {
			ctx.Variables["count"] = cty.ObjectVal(map[string]cty.Value{"index": i})
			Attributes, countDiags := decodeUnknownBody(ctx, body)
			diags = append(diags, countDiags...)
			val, err := convert.Convert(i, cty.String)
			if err != nil {
				panic("Always can convert int")
			}
			deployMap[fmt.Sprintf("%s[%s]", r.addr().String(), val.AsString())] = Attributes
			delete(ctx.Variables, "count")
		}
	} else if r.ForEach != nil {
		forEach, forEachDiags := decodeForExpr(ctx, r.ForEach)
		diags = append(diags, forEachDiags...)
		ty := forEach.Type()
		if ty.IsMapType() || ty.IsObjectType() {
			for key, val := range forEach.AsValueMap() {
				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal(key), "value": val})
				Attributes, forEachDiags := decodeUnknownBody(ctx, body)
				diags = append(diags, forEachDiags...)
				deployMap[fmt.Sprintf("%s[%s]", r.addr().String(), key)] = Attributes
				delete(ctx.Variables, "each")
			}
		} else if ty.IsSetType(){
			for _, val := range forEach.AsValueSet().Values() {
				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"key": val, "value": val})
				Attributes, forEachDiags := decodeUnknownBody(ctx, body)
				diags = append(diags, forEachDiags...)
				deployMap[fmt.Sprintf("%s[%s]", r.addr().String(), val.AsString())] = Attributes
				delete(ctx.Variables, "each")

			}
		}
	} else {
		Attributes, regDiags := decodeUnknownBody(ctx, body)
		diags = append(diags, regDiags...)
		deployMap[r.addr().String()] = Attributes
	}
	dR.Config = deployMap
	return dR, diags
}
