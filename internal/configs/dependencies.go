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
	"strings"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
)

type Graph struct {
	dag.AcyclicGraph
	DecodedModule *decode.DecodedModule
}

const rootNodeName = "root"

var rootNode graphNodeRoot

type graphNodeRoot struct{}

func (n graphNodeRoot) Name() string {
	return rootNodeName
}

func addRootNodeToGraph(g *Graph) {
	// We always add the root node. This is a singleton so if it's already
	// in the graph this will do nothing and just retain the existing root node.
	//
	// Note that rootNode is intentionally added by value and not by pointer
	// so that all root nodes will be equal to one another and therefore
	// coalesce when two valid graphs get merged together into a single graph.
	g.Add(rootNode)

	// Everything that doesn't already depend on at least one other node will
	// depend on the root node, except the root node itself.
	for _, v := range g.Vertices() {
		if v == dag.Vertex(rootNode) {
			continue
		}

		if g.UpEdges(v).Len() == 0 {
			g.Connect(dag.BasicEdge(rootNode, v))
		}
	}
}

func getResourceNames(m *decode.DecodedModule) decode.DecodedResourceList {
	return getResourceName(m, decode.DecodedResourceList{}, "")
}

// Get the resources out of all the modules reassign their name accordingly within the module.
// For example a resource named foo inside module called bar will be called module.bar.resource.foo
func getResourceName(m *decode.DecodedModule, rList decode.DecodedResourceList, currentName string) decode.DecodedResourceList {
	for _, r := range m.Resources {
		r.Name = currentName + r.Addr().String()
		r.Depth = m.Depth
		rList = append(rList, r)
		var dependencies []decode.DependsOn
		var moduleDependencies []decode.DependsOn
		if m.DependsOn != nil {
			moduleDependencies = append(moduleDependencies, decode.DependsOn{
				Depth: m.Depth - 1,
				Trav:  m.DependsOn,
			})

			for _, module := range m.Modules {
				module.Dependencies = append(module.Dependencies, moduleDependencies...)
			}
		}
		if r.DependsOn != nil {
			dependencies = append(dependencies, decode.DependsOn{
				Depth: r.Depth,
				Trav:  r.DependsOn,
			})
		}
		dependencies = append(dependencies, m.Dependencies...)
		dependencies = append(dependencies, moduleDependencies...)
		r.Dependencies = dependencies
	}

	// for _, module := range m.Modules {
	// 	module.DependsOn = append(module.DependsOn, m.DependsOn...)
	// }

	for _, mod := range m.Modules {
		rList = append(rList, getResourceName(mod, decode.DecodedResourceList{}, currentName+ModuleType+"."+mod.Name+".")...)
	}
	return rList
}

func addEdges(g *Graph, r *decode.DecodedResource, resourceMap map[string]*decode.DecodedResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	var edges []rangeName
	edges, diags = getName(r)
	if diags.HasErrors() {
		return diags
	}
	for _, edge := range edges {
		added := false

		startName := strings.Split(r.Name, ".")
		fullName := strings.Join(startName[:edge.Depth*2], ".")
		if fullName != "" {
			fullName = fullName + "." + edge.Type + "." + edge.Name
		} else {
			fullName = edge.Type + "." + edge.Name
		}

		if edge.Type == ResourceType {
			if res, exists := resourceMap[fullName]; exists && res.Depth == edge.Depth {
				g.Connect(dag.BasicEdge(r, res))
				added = true
			}
		}

		if edge.Type == ModuleType {
			for _, vertex := range g.Vertices() {

				res := vertex.(*decode.DecodedResource)
				sName := strings.Split(res.Name, ".")
				if edge.Depth < res.Depth {
					fName := strings.Join(sName[:edge.Depth*2+2], ".")
					if fullName == fName {
						g.Connect(dag.BasicEdge(r, res))
						added = true
					}
				}
			}
		}

		if !added {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Dependency does not exist",
				Detail:   fmt.Sprintf("Entered %s which does not exist, please make sure resources in depends_on list exist", edge.Name),
				Subject:  &edge.Range,
			})
		}
	}

	return diags
}

// Creates a DAG based on dependencies for later purposes such as walking over the graph and activating a function on each resource.
func (g *Graph) Init() hcl.Diagnostics {
	var diags hcl.Diagnostics
	rList := getResourceNames(g.DecodedModule)
	// for _,r := range rList {
	// 	// fmt.Printf("%s\n",r.Name)
	// }
	resourceMap := make(map[string]*decode.DecodedResource)
	for _, r := range rList {
		if g.HasVertex(r) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Resources must have different names",
				Detail:   fmt.Sprintf("Two resources have the same name: %s", r.Name),
				Subject:  &r.DeclRange,
				// Context: names[variable.Name],
			})
		}
		resourceMap[r.Name] = r
		g.Add(
			r,
		)
	}

	// TODO: fix this diags to a much better suited solution
	for _, r := range rList {
		if len(r.Dependencies) > 0 {
			rDiags := addEdges(g, r, resourceMap)
			added := true
			for _, diag := range diags {
				if rDiags[0].Subject.Start == diag.Subject.Start && rDiags[0].Subject.End == diag.Subject.End {
					added = false
					break
				}
			}
			if added {
				diags = append(diags, rDiags...)
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

	addRootNodeToGraph(g)
	if diags.HasErrors(){
		return diags
	}

	if err := g.Validate(); err != nil {
		fmt.Printf("%s", err)
		panic("Should not be here")
	}

	return diags
}

type rangeName struct {
	Name  string
	Type  string
	Depth int
	Range hcl.Range
}

// Get the name of the resource based on the depends on attribute in resource and module blocks
// Decode it into strings and put it in the graph accordingly
func getName(resource *decode.DecodedResource) ([]rangeName, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	// var dag dag.AcyclicGraph
	var rNames = []rangeName{}
	for _, dependsOn := range resource.Dependencies {
		for _, traversal := range dependsOn.Trav {
			if len(traversal) < 2 {
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
			if resourceType != ResourceType && resourceType != ModuleType {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "This type is not supported",
					Detail:   fmt.Sprintf("Allowed types are [%s,%s] got: %s", ResourceType, ModuleType, resourceType),
					Subject:  traversal[0].SourceRange().Ptr(),
				})
			}

			rNames = append(rNames, rangeName{
				Name:  resourceName,
				Type:  resourceType,
				Range: hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()),
				Depth: dependsOn.Depth,
			})
		}
	}

	return rNames, diags
}

const (
	ResourceType = "kube_resource"
	ModuleType   = "module"
)
