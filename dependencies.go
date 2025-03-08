package main

// import (
// 	"fmt"

// 	"github.com/hashicorp/hcl/v2"
// 	"kubehcl.sh/kubehcl/internal/addrs"
// 	"kubehcl.sh/kubehcl/internal/dag"
// )

// type Graph struct {
// 	dag.AcyclicGraph
// 	Resources ResourceList
// }

// type rangeName struct {
// 	Name string
// 	Range hcl.Range
// }
// const rootNodeName = "root"
// var rootNode graphNodeRoot
// type graphNodeRoot struct{}

// func (n graphNodeRoot) Name() string {
// 	return rootNodeName
// }
// func addRootNodeToGraph(g *Graph) {
// 	// We always add the root node. This is a singleton so if it's already
// 	// in the graph this will do nothing and just retain the existing root node.
// 	//
// 	// Note that rootNode is intentionally added by value and not by pointer
// 	// so that all root nodes will be equal to one another and therefore
// 	// coalesce when two valid graphs get merged together into a single graph.
// 	g.Add(rootNode)

// 	// Everything that doesn't already depend on at least one other node will
// 	// depend on the root node, except the root node itself.
// 	for _, v := range g.Vertices() {
// 		if v == dag.Vertex(rootNode) {
// 			continue
// 		}

// 		if g.UpEdges(v).Len() == 0 {
// 			g.Connect(dag.BasicEdge(rootNode, v))
// 		}
// 	}
// }

// func (g *Graph) Init() hcl.Diagnostics {
// 	var diags hcl.Diagnostics
// 	for _, r := range resourceList {
// 		// resource := addrs.Resource{
// 		// 	Name:         r.Name,
// 		// 	ResourceMode: addrs.RMode,
// 		// }
// 		if g.HasVertex(r) {
// 			diags = append(diags, &hcl.Diagnostic{
// 				Severity: hcl.DiagError,
// 				Summary:  "Resources must have different names",
// 				Detail:   fmt.Sprintf("Two resources have the same name: %s", r.Name),
// 				Subject:  &r.DeclRange,
// 				// Context: names[variable.Name],
// 			})
// 		}

// 		g.Add(
// 			r,
// 		)

// 	}

// 	for _, r := range resourceList {
// 		if r.DependsOn != nil {
// 			edges, dependsOnDiags := getName(r)
// 			diags = append(diags, dependsOnDiags...)
// 			if diags.HasErrors(){
// 				return diags
// 			}
			
// 			for _, edge := range edges {
// 				resource := addrs.Resource{
// 					Name:         edge.Name,
// 					ResourceMode: addrs.RMode,
// 				}
				
// 				if val,exists :=addrMap[resource.String()]; !exists {
// 					diags = append(diags, &hcl.Diagnostic{
// 						Severity: hcl.DiagError,
// 						Summary:  "Resource does not exist",
// 						Detail:   fmt.Sprintf("There isn't a resource with this name %s", r.Name),
// 						Subject: &edge.Range,
// 						// Context: names[variable.Name],
// 					})
// 				} else {
// 					g.Connect(dag.BasicEdge(r,
// 						val,
// 					))
// 				}
// 			}
// 		}
// 	}
// 	for _, cycle := range g.Cycles() {
// 		diags = append(diags, &hcl.Diagnostic{
// 			Severity: hcl.DiagError,
// 			Summary:  "Circular dependencies",
// 			Detail:   fmt.Sprintf("Resources %s %s have circular dependencies", cycle[0], cycle[1]),
// 			// Context: names[variable.Name],
// 		})
// 	}

// 	addRootNodeToGraph(g)

// 	if err:=g.Validate(); err != nil {
// 		fmt.Printf("%s",err)
// 		panic("Should not be here")
// 	}

// 	return diags
// }

// func getName(resource *Resource) ([]rangeName, hcl.Diagnostics) {

// 	var diags hcl.Diagnostics
// 	// var dag dag.AcyclicGraph
// 	var rNames = []rangeName{}
// 	for _, traversal := range resource.DependsOn {
// 		if len(traversal) != 2 {
// 			diags = append(diags, &hcl.Diagnostic{
// 				Severity: hcl.DiagError,
// 				Summary:  "Invalid reference",
// 				Detail:   `Depends on must have resource/module and the name of the module`,
// 				Subject:  hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()).Ptr(),
// 			})
// 			return rNames, diags
// 		}

// 		var resourceType, resourceName string
// 		switch tt := traversal[0].(type) {
// 		case hcl.TraverseRoot:
// 			resourceType = tt.Name
// 		case hcl.TraverseAttr:
// 			resourceType = tt.Name
// 		default:
// 			diags = append(diags, &hcl.Diagnostic{
// 				Severity: hcl.DiagError,
// 				Summary:  "Unknown error",
// 				Detail:   `Unknown error`,
// 				Subject:  hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()).Ptr(),
// 			})
// 		}
// 		switch tt := traversal[1].(type) {
// 		case hcl.TraverseAttr:
// 			resourceName = tt.Name
// 		default:
// 			diags = append(diags, &hcl.Diagnostic{
// 				Severity: hcl.DiagError,
// 				Summary:  "Invalid address",
// 				Detail:   "A resource name is require.",
// 				Subject:  traversal[1].SourceRange().Ptr(),
// 			})
// 		}
// 		if resourceType != "resource" {
// 			diags = append(diags, &hcl.Diagnostic{
// 				Severity: hcl.DiagError,
// 				Summary:  "This type is not supported",
// 				Detail:   fmt.Sprintf("Allowed types are [resource] got: %s",resourceType),
// 				Subject:  traversal[0].SourceRange().Ptr(),
// 			})
// 		}
// 		rNames = append(rNames, rangeName{
// 			Name: resourceName,
// 			Range: hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()),
// 		})
// 	}

// 	return rNames, diags
// }
