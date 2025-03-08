package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/dag"
)

type Graph struct {
	dag.AcyclicGraph
	Resources ResourceList
}

func (g Graph) init() hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, r := range resourceList {
		resource := addrs.Resource{
			Name:         r.Name,
			ResourceMode: addrs.RMode,
		}
		if g.HasVertex(resource) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Resources must have different names",
				Detail:   fmt.Sprintf("Two resources have the same name: %s", r.Name),
				Subject:  &r.DeclRange,
				// Context: names[variable.Name],
			})
		}

		g.Add(
			resource,
		)

	}

	for _, r := range resourceList {
		if r.DependsOn != nil {
			edges, dependsOnDiags := checkCircularDependencies(r)
			diags = append(diags, dependsOnDiags...)
			for _, edge := range edges {
				g.Connect(dag.BasicEdge(addrs.Resource{
					Name:         r.Name,
					ResourceMode: addrs.RMode,
				},
					addrs.Resource{
						Name:         edge,
						ResourceMode: addrs.RMode,
					},
				))
			}
		}
	}

	for _, cycle := range g.Cycles() {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Circular dependencies",
			Detail:   fmt.Sprintf("Resources %s %s have circular dependencies", cycle[0], cycle[1]),
			// Context: names[variable.Name],
		})
	}
	return diags
}

func checkCircularDependencies(resource *Resource) ([]string, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	// var dag dag.AcyclicGraph
	var rNames = []string{}
	for _, traversal := range resource.DependsOn {
		if len(traversal) != 2 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid reference",
				Detail:   `Depends on must have resource/module and the name of the module`,
				Subject:  hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()).Ptr(),
			})
			return rNames, diags
		}

		var resourceType, resourceName string
		switch tt := traversal[0].(type) {
		case hcl.TraverseRoot:
			resourceType = tt.Name
		case hcl.TraverseAttr:
			resourceType = tt.Name
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unknown error",
				Detail:   `Unknown error`,
				Subject:  hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()).Ptr(),
			})
		}
		switch tt := traversal[1].(type) {
		case hcl.TraverseAttr:
			resourceName = tt.Name
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid address",
				Detail:   "A resource name is require.",
				Subject:  traversal[1].SourceRange().Ptr(),
			})
		}
		if resourceType != "resource" {
			panic("not implemented!")
		}
		rNames = append(rNames, resourceName)
	}

	return rNames, diags
}
