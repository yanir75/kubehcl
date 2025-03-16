package configs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
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
	decode.Deployable
}

type ResourceList []*Resource

func (r ResourceList) Decode(ctx *hcl.EvalContext) (decode.DecodedResourceList, hcl.Diagnostics) {
	var dR decode.DecodedResourceList
	var diags hcl.Diagnostics
	for _, variable := range r {
		dV, varDiags := variable.decode(ctx)
		diags = append(diags, varDiags...)
		dR = append(dR, dV)
	}

	return dR, diags
}

func (r *Resource) decode(ctx *hcl.EvalContext) (*decode.DecodedResource, hcl.Diagnostics) {
	deployable, diags := r.Deployable.Decode(ctx)
	res := &decode.DecodedResource{DecodedDeployable: *deployable}

	return res, diags
}

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
		// Name:      block.Labels[0],
		// DeclRange: block.DefRange,
	
	}

	resource.Name = block.Labels[0]
	resource.DeclRange = block.DefRange
	resource.Type = addrs.RType
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
		if _, countExists := content.Attributes["count"]; countExists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
				Context:  &content.Attributes["count"].NameRange,
			})
		}
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

	if attr, exists := content.Attributes["depends_on"]; exists {
		traversal, dependsDiags := decodeDependsOn(attr)
		diags = append(diags, dependsDiags...)
		resource.DependsOn = append(resource.DependsOn, traversal...)
	}

	return resource, diags

}

func DecodeResourceBlocks(blocks hcl.Blocks, addrMap addrs.AddressMap) (ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var resourceList ResourceList
	for _, block := range blocks {
		resource, rDiags := decodeResourceBlock(block)
		diags = append(diags, rDiags...)
		resourceList = append(resourceList, resource)
		if addrMap.Add(resource.addr().String(), resource) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Resources must have different names",
				Detail:   fmt.Sprintf("Two Resources have the same name: %s", resource.Name),
				Subject:  &block.DefRange,
				// Context: names[variable.Name],
			})
		}
	}

	return resourceList, diags
}

func (r Resource) addr() addrs.Resource {
	return addrs.Resource{
		Name:         r.Name,
		ResourceMode: addrs.RMode,
	}
}

// // for _,block := range body.Blocks {
// // 	block.
// // }
