package configs

import (
	// "fmt"

	// "github.com/hashicorp/hcl/v2"
	// "kubehcl.sh/kubehcl/internal/addrs"
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

func getResourceName(m *decode.DecodedModule, rList decode.DecodedResourceList, currentName string) decode.DecodedResourceList {
	for _, r := range m.Resources {
		r.Name = currentName + r.Addr().String()
		r.Depth = m.Depth
		rList = append(rList, r)
		var dependencies []decode.DependsOn
		if m.DependsOn != nil {
			dependencies =append(dependencies, decode.DependsOn{
				Depth:m.Depth-1,
				Trav: m.DependsOn,
			},)
		}
		if r.DependsOn != nil {
			dependencies = append(dependencies,decode.DependsOn{
				Depth: r.Depth,
				Trav: r.DependsOn,
			},)
		}
		r.Dependencies = dependencies
	}

	for _, mod := range m.Modules {
		rList = append(rList, getResourceName(mod, decode.DecodedResourceList{}, currentName+"module."+mod.Name+".")...)
	}
	return rList
}


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

	for _,r := range rList {
		if len(r.Dependencies)>0 {
			edges,edgeDiags :=getName(r)
			diags = append(diags, edgeDiags...)
			// if diags.HasErrors(){
			// 	return diags
			// }
			for _, edge := range edges {
				added := false
				// for _, vertex := range g.Vertices(){
				// 	res := vertex.(*decode.DecodedResource)
				// 	if res.Depth != r.Depth
				// }
				// added := false

				// resource := addrs.Resource{
				// 	Name:         edge.Name,
				// 	ResourceMode: edge.Type,
				// }
				startName := strings.Split(r.Name,".")
				fullName := strings.Join(startName[:edge.Depth*2],".")
				if fullName != ""{
					fullName = fullName+"."+edge.Type+"."+edge.Name
				} else {
					fullName = edge.Type+"."+edge.Name
				}

				if edge.Type == ResourceType {
					if res,exists :=resourceMap[fullName];exists && res.Depth == edge.Depth{
						g.Connect(dag.BasicEdge(r,res))
						added = true
					}
				}
				if edge.Type == ModuleType {
					for _,vertex := range g.Vertices(){
						
						res := vertex.(*decode.DecodedResource)
						sName := strings.Split(res.Name,".")
						if edge.Depth < res.Depth  {
								fName := strings.Join(sName[:edge.Depth*2+2],".")
								if fullName == fName {
									g.Connect(dag.BasicEdge(r,res))
									added = true
								}
						}
					}
				}
				if !added{
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary: "Couldn't find the depends on resource",
						Detail: fmt.Sprintf("Entered %s which does not exist",edge.Name),
						Subject: &edge.Range,
					})	
				}
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

	if err:=g.Validate(); err != nil {
		fmt.Printf("%s",err)
		panic("Should not be here")
	}

	return diags
}

// func (g *Graph) Init() hcl.Diagnostics {
// 	var diags hcl.Diagnostics
// 	for _, r := range resourceList {
// 		// resource := addrs.Resource{
// 		// 	Name:         r.Name,
// 		// 	ResourceMode: addrs.RMode,
// 		// }
// if g.HasVertex(r) {
// 	diags = append(diags, &hcl.Diagnostic{
// 		Severity: hcl.DiagError,
// 		Summary:  "Resources must have different names",
// 		Detail:   fmt.Sprintf("Two resources have the same name: %s", r.Name),
// 		Subject:  &r.DeclRange,
// 		// Context: names[variable.Name],
// 	})
// }

// 		g.Add(
// 			r,
// 		)

// 	}

// 	for _, r := range resourceList {
// 		if r.DependsOn != nil {
			// edges, dependsOnDiags := getName(r)
// 			diags = append(diags, dependsOnDiags...)
// 			if diags.HasErrors(){
// 				return diags
// 			}

			// for _, edge := range edges {
				// resource := addrs.Resource{
				// 	Name:         edge.Name,
				// 	ResourceMode: addrs.RMode,
				// }

			// 	if val,exists :=addrMap[resource.String()]; !exists {
			// 		diags = append(diags, &hcl.Diagnostic{
			// 			Severity: hcl.DiagError,
			// 			Summary:  "Resource does not exist",
			// 			Detail:   fmt.Sprintf("There isn't a resource with this name %s", r.Name),
			// 			Subject: &edge.Range,
			// 			// Context: names[variable.Name],
			// 		})
			// 	} else {
			// 		g.Connect(dag.BasicEdge(r,
			// 			val,
			// 		))
			// 	}
			// }
		// }
	// }
	// for _, cycle := range g.Cycles() {
	// 	diags = append(diags, &hcl.Diagnostic{
	// 		Severity: hcl.DiagError,
	// 		Summary:  "Circular dependencies",
	// 		Detail:   fmt.Sprintf("Resources %s %s have circular dependencies", cycle[0], cycle[1]),
	// 		// Context: names[variable.Name],
	// 	})
	// }

	// addRootNodeToGraph(g)

	// if err:=g.Validate(); err != nil {
	// 	fmt.Printf("%s",err)
	// 	panic("Should not be here")
	// }

	// return diags
// }
type rangeName  struct{
	Name string
	Type string
	Depth int
	Range hcl.Range
}

func getName(resource *decode.DecodedResource) ([]rangeName, hcl.Diagnostics) {

	var diags hcl.Diagnostics
	// var dag dag.AcyclicGraph
	var rNames = []rangeName{}
	for _,dependsOn := range resource.Dependencies{
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
					Detail:   fmt.Sprintf("Allowed types are [resource,module] got: %s",resourceType),
					Subject:  traversal[0].SourceRange().Ptr(),
				})
			}

			rNames = append(rNames, rangeName{
				Name: resourceName,
				Type: resourceType,
				Range: hcl.RangeBetween(traversal[0].SourceRange(), traversal[len(traversal)-1].SourceRange()),
				Depth: dependsOn.Depth,
			})
		}
	}

	return rNames, diags
}

const (
	ResourceType = "resource"
	ModuleType = "module"
)
