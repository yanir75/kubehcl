package configs

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"kubehcl.sh/kubehcl/internal/dag"
	"kubehcl.sh/kubehcl/internal/decode"
)

func Test_Dependencies(t *testing.T) {
	mod := &decode.DecodedModule{
		Name: rootNodeName,
		Modules: []*decode.DecodedModule{
			{
				Modules: []*decode.DecodedModule{
					{
						Name:  "foo",
						Depth: 2,
						DependsOn: []hcl.Traversal{
							{
								hcl.TraverseRoot{
									Name: "kube_resource",
								},
								hcl.TraverseAttr{
									Name: "number_2",
								},
							},
						},
						Resources: decode.DecodedResourceList{
							&decode.DecodedResource{
								Depth: 2,
								DecodedDeployable: decode.DecodedDeployable{
									Type: "r",
									Name: "number_3",
								},
							},
						},
					},
				},
				Name:  "bar",
				Depth: 1,
				DependsOn: []hcl.Traversal{
					{
						hcl.TraverseRoot{
							Name: "kube_resource",
						},
						hcl.TraverseAttr{
							Name: "number_1",
						},
					},
				},
				Resources: decode.DecodedResourceList{
					&decode.DecodedResource{
						Depth: 1,
						DecodedDeployable: decode.DecodedDeployable{
							Type: "r",
							Name: "number_2",
						},
					},
				},
			},
		},
		Resources: decode.DecodedResourceList{
			&decode.DecodedResource{
				DecodedDeployable: decode.DecodedDeployable{
					Name: "number_4",
					Type: "r",

					DependsOn: []hcl.Traversal{
						{
							hcl.TraverseRoot{
								Name: "module",
							},
							hcl.TraverseAttr{
								Name: "bar",
							},
						},
						{
							hcl.TraverseRoot{
								Name: "kube_resource",
							},
							hcl.TraverseAttr{
								Name: "number_1",
							},
						},
					},
				},
				Depth: 0,
			},
			&decode.DecodedResource{
				DecodedDeployable: decode.DecodedDeployable{
					Type: "r",

					Name: "number_1",
				},
				Depth: 0,
			},
		},
		Depth: 0,
	}

	g := &Graph{DecodedModule: mod}
	i := 0
	results := []string{"kube_resource.number_1", "module.bar.kube_resource.number_2", "module.bar.module.foo.kube_resource.number_3", "kube_resource.number_4"}
	diags := g.Init()
	if diags.HasErrors() {
		t.Errorf("Failed becuase %s", diags.Error())
	}

	diags = g.Walk(func(v dag.Vertex) hcl.Diagnostics {
		switch tt := v.(type) {
		case *decode.DecodedResource:
			if tt.Name != results[i] {
				t.Errorf("Values are not equal %s, %s", tt.Name, results[i])
			}
			i++
		}
		return nil
	})

	if diags.HasErrors() {
		t.Errorf("Failed becuase %s", diags.Error())
	}

	cirMod := &decode.DecodedModule{
		Name: rootNodeName,
		Modules: []*decode.DecodedModule{
			{
				Modules: []*decode.DecodedModule{
					{
						Name:  "foo",
						Depth: 2,
						DependsOn: []hcl.Traversal{
							{
								hcl.TraverseRoot{
									Name: "kube_resource",
								},
								hcl.TraverseAttr{
									Name: "number_2",
								},
							},
						},
						Resources: decode.DecodedResourceList{
							&decode.DecodedResource{
								Depth: 2,
								DecodedDeployable: decode.DecodedDeployable{
									Type: "r",
									Name: "number_3",
								},
							},
						},
					},
				},
				Name:  "bar",
				Depth: 1,
				DependsOn: []hcl.Traversal{
					{
						hcl.TraverseRoot{
							Name: "kube_resource",
						},
						hcl.TraverseAttr{
							Name: "number_1",
						},
					},
				},
				Resources: decode.DecodedResourceList{
					&decode.DecodedResource{
						Depth: 1,
						DecodedDeployable: decode.DecodedDeployable{
							Type: "r",
							Name: "number_2",
							DependsOn: []hcl.Traversal{
								{
									hcl.TraverseRoot{
										Name: "module",
									},
									hcl.TraverseAttr{
										Name: "foo",
									},
								},
							},
						},
					},
				},
			},
		},
		Resources: decode.DecodedResourceList{
			&decode.DecodedResource{
				DecodedDeployable: decode.DecodedDeployable{
					Name: "number_4",
					Type: "r",

					DependsOn: []hcl.Traversal{
						{
							hcl.TraverseRoot{
								Name: "module",
							},
							hcl.TraverseAttr{
								Name: "bar",
							},
						},
						{
							hcl.TraverseRoot{
								Name: "kube_resource",
							},
							hcl.TraverseAttr{
								Name: "number_1",
							},
						},
					},
				},
				Depth: 0,
			},
			&decode.DecodedResource{
				DecodedDeployable: decode.DecodedDeployable{
					Type: "r",

					Name: "number_1",
				},
				Depth: 0,
			},
		},
		Depth: 0,
	}

	g2 := &Graph{DecodedModule: cirMod}
	diags = g2.Init()
	if !diags.HasErrors() {
		t.Errorf("Should have error")
	} else {
		if !strings.Contains(diags.Error(), "circular") {
			t.Errorf("Error should contain circular dependency %s", diags.Error())
		}
	}
}
