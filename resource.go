package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"kubehcl.sh/kubehcl/internal/addrs"

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
	ForEach   hcl.Expression
	Count     hcl.Expression
	DependsOn []hcl.Traversal
	DeclRange hcl.Range
	Config      hcl.Body
}

type ResourceList []*Resource



var inputResourceBlockSchema = &hcl.BodySchema{
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

func decodeResourceBlock(block *hcl.Block) (*Resource, hcl.Diagnostics) {
	var resource *Resource = &Resource{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, remain, diags := block.Body.PartialContent(inputResourceBlockSchema)
	resource.Config = remain
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
		resource.Count = attr.Expr
	}
	if attr, exists := content.Attributes["for_each"]; exists {
		if _, countExists :=  content.Attributes["count"]; countExists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
				Context:  &content.Attributes["count"].NameRange,
			})
		resource.ForEach = attr.Expr
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
		resource.DependsOn = append(resource.DependsOn, traversal...)
	}

	return resource, diags

}

func decodeResourceBlocks(blocks hcl.Blocks) (ResourceList,hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var resourceList ResourceList
	for _, block := range blocks {
		resource, rDiags := decodeResourceBlock(block)
		diags = append(diags, rDiags...)
		resourceList = append(resourceList, resource)
		if addrMap.add(resource.addr().String(),resource) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Resources must have different names",
				Detail:   fmt.Sprintf("Two Resources have the same name: %s", resource.Name),
				Subject:  &block.DefRange,
				// Context: names[variable.Name],
			})
		}
	}

	return resourceList,diags
}


func (r Resource) addr() addrs.Resource{
	return addrs.Resource{
		Name: r.Name,
		ResourceMode: addrs.RMode,
	}
}

func decodeUnknownBody(ctx *hcl.EvalContext,body *hclsyntax.Body) (cty.Value,hcl.Diagnostics){
	var diags hcl.Diagnostics
	attrMap := make(map[string]cty.Value)
	if len(body.Blocks) > 0{
		for _,block := range body.Blocks {
			m,blockDiags :=decodeUnknownBody(ctx,block.Body)
			diags = append(diags, blockDiags...)
			attrMap[block.Type] = m
		}
	}
	for _,attr := range body.Attributes {
		val,attrDiags :=attr.Expr.Value(ctx)
		diags = append(diags, attrDiags...)

		attrMap[attr.Name] = val
	}
	return cty.ObjectVal(attrMap),diags
}



func (r *Resource) decodeResource(ctx *hcl.EvalContext) hcl.Diagnostics{
	var diags hcl.Diagnostics
	body, ok := r.Config.(*hclsyntax.Body)

	if !ok {
		panic("should always be ok")
	}
	for _,attrS:= range inputResourceBlockSchema.Attributes{
			delete(body.Attributes,attrS.Name)
	}
	if r.Count != nil{
		count,countDiags := decodeCountExpr(ctx,r.Count)
		diags = append(diags, countDiags...)
		for i :=cty.NumberIntVal(1); i.LessThanOrEqualTo(count) == cty.True; i=i.Add(cty.NumberIntVal(1)) {
			ctx.Variables["count"] = cty.ObjectVal(map[string]cty.Value{ "index" : i})
			Attributes,countDiags := decodeUnknownBody(ctx,body)
			diags = append(diags, countDiags...)
			val, err := convert.Convert(i,cty.String)
			if err != nil {
				panic("Always can convert int")
			}
			deployMap[fmt.Sprintf("%s[%s]",r.addr().String(),val.AsString())] = Attributes

		}
	} else if r.ForEach != nil{
		forEach,forEachDiags := decodeCountExpr(ctx,r.Count)
		diags = append(diags, forEachDiags...)
		ty:= forEach.Type()
		if ty.IsMapType() || ty.IsObjectType(){
			for key,val := range forEach.AsValueMap(){
				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{ "key" : cty.StringVal(key),"value": val})
				Attributes,forEachDiags := decodeUnknownBody(ctx,body)
				diags = append(diags, forEachDiags...)
				deployMap[fmt.Sprintf("%s[%s]",r.addr().String(),key)] = Attributes
			}
		} else{
			for _,val := range forEach.AsValueSet().Values(){
				ctx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{ "key" : val,"value": val})
				Attributes,forEachDiags := decodeUnknownBody(ctx,body)
				diags = append(diags, forEachDiags...)
				deployMap[fmt.Sprintf("%s[%s]",r.addr().String(),val.AsString())] = Attributes
			}
		}
	} else {
		Attributes,regDiags := decodeUnknownBody(ctx,body)
		diags = append(diags, regDiags...)
		deployMap[r.addr().String()] = Attributes
	}
	return diags
}
	// // for _,block := range body.Blocks {
	// // 	block.
	// // }