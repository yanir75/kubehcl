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

func decodeUnknownBody(ctx *hcl.EvalContext, body *hclsyntax.Body) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	attrMap := make(map[string]cty.Value)
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {
			if len(block.Labels) > 0 {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Subject: &block.TypeRange,
					Summary: "Block shouldn't have labels",
					Detail: fmt.Sprintf("Block has labels: %s and type: \"%s\"",strings.Join(block.Labels,", "),block.Type),
					Context: &block.LabelRanges[0],
				})
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
		} else {
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
