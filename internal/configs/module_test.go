// /*
// This file was inspired from https://github.com/opentofu/opentofu
// This file has been modified from the original version
// Changes made to fit kubehcl purposes
// This file retains its' original license
// // SPDX-License-Identifier: MPL-2.0
// Licesne: https://www.mozilla.org/en-US/MPL/2.0/
// */
package configs

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/decode"
	"kubehcl.sh/kubehcl/internal/logging"
)


func Test_Module(t *testing.T) {
	tests := []struct {
		d          string
		want       *decode.DecodedModule
		wantErrors bool
	}{
		{
			d: "testcases/example",
			want: &decode.DecodedModule{
				Name: "",
				Inputs: decode.DecodedVariableList{
					&decode.DecodedVariable{
					Name: "foo",
					Description: "Ports of the container",
					Default: cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})}),
					Type: cty.List(cty.Map(cty.Number)),
					DeclRange: hcl.Range{
						Filename: "testcases/example/variable.hcl",
						Start: hcl.Pos{
						Line: 1,
						Column: 1,
						Byte: 0,
						},
						End: hcl.Pos{
						Line: 1,
						Column: 15,
						Byte: 14,
						},
					},
					},
				},
				Locals: nil,
				Annotations: decode.DecodedAnnotations{
					&decode.DecodedAnnotation{
					Name: "foo",
					Value: cty.StringVal("bar"),
					DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
						Line: 59,
						Column: 3,
						Byte: 970,
						},
						End: hcl.Pos{
						Line: 59,
						Column: 6,
						Byte: 973,
						},
					},
					},
					&decode.DecodedAnnotation{
					Name: "bar",
					Value: cty.StringVal("foo"),
					DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
						Line: 63,
						Column: 3,
						Byte: 1009,
						},
						End: hcl.Pos{
						Line: 63,
						Column: 6,
						Byte: 1012,
						},
					},
					},
					&decode.DecodedAnnotation{
					Name: "kubehcl.sh/managed",
					Value: cty.StringVal("This resource is managed by kubehcl"),
					DeclRange: hcl.Range{
						Filename: "",
						Start: hcl.Pos{
						Line: 0,
						Column: 0,
						Byte: 0,
						},
						End: hcl.Pos{
						Line: 0,
						Column: 0,
						Byte: 0,
						},
					},
					},
					&decode.DecodedAnnotation{
					Name: "kubehcl.sh/release",
					Value: cty.StringVal("test"),
					DeclRange: hcl.Range{
						Filename: "",
						Start: hcl.Pos{
						Line: 0,
						Column: 0,
						Byte: 0,
						},
						End: hcl.Pos{
						Line: 0,
						Column: 0,
						Byte: 0,
						},
					},
					},
				},
				Resources: decode.DecodedResourceList{
					&decode.DecodedResource{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "namespace",
						Config: map[string]cty.Value{
						"kube_resource.namespace": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("v1"), "kind":cty.StringVal("Namespace"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"bar":cty.StringVal("foo"), "foo":cty.StringVal("bar"), "kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("bar")}), "name":cty.StringVal("foo")})}),
						},
						Type: "r",
						DependsOn: nil,
						DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
							Line: 1,
							Column: 1,
							Byte: 0,
						},
						End: hcl.Pos{
							Line: 1,
							Column: 26,
							Byte: 25,
						},
						},
					},
					Depth: 0,
					Dependencies: nil,
					DependenciesAppended: nil,
					},
					&decode.DecodedResource{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "foo",
						Config: map[string]cty.Value{
						"kube_resource.foo[bar]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"bar":cty.StringVal("foo"), "foo":cty.StringVal("bar"), "kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("foo")}), "name":cty.StringVal("bar")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(2), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("foo")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("foo")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("foo"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})})})})})})})}),
						"kube_resource.foo[foo]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"bar":cty.StringVal("foo"), "foo":cty.StringVal("bar"), "kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("bar")}), "name":cty.StringVal("foo")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(2), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("bar")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("bar")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("bar"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})})})})})})})}),
						},
						Type: "r",
						DependsOn: []hcl.Traversal{
						hcl.Traversal{
							hcl.TraverseRoot{
							Name: "module",
							SrcRange: hcl.Range{
								Filename: "testcases/example/main.hcl",
								Start: hcl.Pos{
								Line: 48,
								Column: 17,
								Byte: 747,
								},
								End: hcl.Pos{
								Line: 48,
								Column: 23,
								Byte: 753,
								},
							},
							},
							hcl.TraverseAttr{
							Name: "test",
							SrcRange: hcl.Range{
								Filename: "testcases/example/main.hcl",
								Start: hcl.Pos{
								Line: 48,
								Column: 23,
								Byte: 753,
								},
								End: hcl.Pos{
								Line: 48,
								Column: 28,
								Byte: 758,
								},
							},
							},
						},
						hcl.Traversal{
							hcl.TraverseRoot{
							Name: "kube_resource",
							SrcRange: hcl.Range{
								Filename: "testcases/example/main.hcl",
								Start: hcl.Pos{
								Line: 48,
								Column: 30,
								Byte: 760,
								},
								End: hcl.Pos{
								Line: 48,
								Column: 43,
								Byte: 773,
								},
							},
							},
							hcl.TraverseAttr{
							Name: "namespace",
							SrcRange: hcl.Range{
								Filename: "testcases/example/main.hcl",
								Start: hcl.Pos{
								Line: 48,
								Column: 43,
								Byte: 773,
								},
								End: hcl.Pos{
								Line: 48,
								Column: 53,
								Byte: 783,
								},
							},
							},
						},
						},
						DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
							Line: 13,
							Column: 1,
							Byte: 155,
						},
						End: hcl.Pos{
							Line: 13,
							Column: 20,
							Byte: 174,
						},
						},
					},
					Depth: 0,
					Dependencies: nil,
					DependenciesAppended: nil,
					},
				},
				ModuleCalls: decode.DecodedModuleCallList{
					&decode.DecodedModuleCall{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "test",
						Config: map[string]cty.Value{
						"module.test": cty.ObjectVal(map[string]cty.Value{"foo":cty.TupleVal([]cty.Value{cty.StringVal("service1"), cty.StringVal("service2")}), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})}), "source":cty.StringVal("./modules/starter")}),
						},
						Type: "m",
						DependsOn: []hcl.Traversal{ 
						hcl.Traversal{
							hcl.TraverseRoot{
							Name: "kube_resource",
							SrcRange: hcl.Range{
								Filename: "testcases/example/main.hcl",
								Start: hcl.Pos{
								Line: 55,
								Column: 17,
								Byte: 918,
								},
								End: hcl.Pos{
								Line: 55,
								Column: 30,
								Byte: 931,
								},
							},
							},
							hcl.TraverseAttr{
							Name: "namespace",
							SrcRange: hcl.Range{
								Filename: "testcases/example/main.hcl",
								Start: hcl.Pos{
								Line: 55,
								Column: 30,
								Byte: 931,
								},
								End: hcl.Pos{
								Line: 55,
								Column: 40,
								Byte: 941,
								},
							},
							},
						},
						},
						DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
							Line: 51,
							Column: 1,
							Byte: 788,
						},
						End: hcl.Pos{
							Line: 51,
							Column: 14,
							Byte: 801,
						},
						},
					},
					Source: "./modules/starter",
					},
				},
				Modules: decode.DecodedModuleList{
					&decode.DecodedModule{
					Name: "test",
					Inputs: decode.DecodedVariableList{
						&decode.DecodedVariable{
						Name: "ports",
						Description: "",
						Default: cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})}),
						Type: cty.NilType,
						DeclRange: hcl.Range{
							Filename: "testcases/example/main.hcl",
							Start: hcl.Pos{
							Line: 54,
							Column: 16,
							Byte: 894,
							},
							End: hcl.Pos{
							Line: 54,
							Column: 23,
							Byte: 901,
							},
						},
						},
						&decode.DecodedVariable{
						Name: "foo",
						Description: "",
						Default: cty.ListVal([]cty.Value{cty.StringVal("service1"), cty.StringVal("service2")}),
						Type: cty.List(cty.String),
						DeclRange: hcl.Range{
							Filename: "testcases/example/main.hcl",
							Start: hcl.Pos{
							Line: 53,
							Column: 16,
							Byte: 854,
							},
							End: hcl.Pos{
							Line: 53,
							Column: 40,
							Byte: 878,
							},
						},
						},
					},
					Locals: decode.DecodedLocals{
						&decode.DecodedLocal{
						Name: "service_ports",
						Value: cty.ObjectVal(map[string]cty.Value{"0":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service1"), "targetPort":cty.NumberIntVal(80)}), "1":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service2"), "targetPort":cty.NumberIntVal(80)})}),
						DeclRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
							Line: 2,
							Column: 3,
							Byte: 11,
							},
							End: hcl.Pos{
							Line: 2,
							Column: 16,
							Byte: 24,
							},
						},
						},
						&decode.DecodedLocal{
						Name: "other_option",
						Value: cty.ObjectVal(map[string]cty.Value{"service1":cty.ObjectVal(map[string]cty.Value{"targetPort":cty.NumberIntVal(80)}), "service2":cty.ObjectVal(map[string]cty.Value{"targetPort":cty.NumberIntVal(80)})}),
						DeclRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
							Line: 9,
							Column: 3,
							Byte: 144,
							},
							End: hcl.Pos{
							Line: 9,
							Column: 15,
							Byte: 156,
							},
						},
						},
					},
					Annotations: decode.DecodedAnnotations{
						&decode.DecodedAnnotation{
						Name: "kubehcl.sh/managed",
						Value: cty.StringVal("This resource is managed by kubehcl"),
						DeclRange: hcl.Range{
							Filename: "",
							Start: hcl.Pos{
							Line: 0,
							Column: 0,
							Byte: 0,
							},
							End: hcl.Pos{
							Line: 0,
							Column: 0,
							Byte: 0,
							},
						},
						},
						&decode.DecodedAnnotation{
						Name: "kubehcl.sh/release",
						Value: cty.StringVal("test"),
						DeclRange: hcl.Range{
							Filename: "",
							Start: hcl.Pos{
							Line: 0,
							Column: 0,
							Byte: 0,
							},
							End: hcl.Pos{
							Line: 0,
							Column: 0,
							Byte: 0,
							},
						},
						},
					},
					Resources: decode.DecodedResourceList{
						&decode.DecodedResource{
						DecodedDeployable: decode.DecodedDeployable{
							Name: "service",
							Config: map[string]cty.Value{
							"kube_resource.service[0]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("v1"), "kind":cty.StringVal("Service"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "name":cty.StringVal("service1")}), "spec":cty.ObjectVal(map[string]cty.Value{"ports":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service1"), "port":cty.NumberIntVal(9367), "targetPort":cty.NumberIntVal(80)})}), "selector":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")})})}),
							"kube_resource.service[1]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("v1"), "kind":cty.StringVal("Service"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "name":cty.StringVal("service2")}), "spec":cty.ObjectVal(map[string]cty.Value{"ports":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service2"), "port":cty.NumberIntVal(9367), "targetPort":cty.NumberIntVal(80)})}), "selector":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")})})}),
							},
							Type: "r",
							DependsOn: []hcl.Traversal{
							hcl.Traversal{
								hcl.TraverseRoot{
								Name: "kube_resource",
								SrcRange: hcl.Range{
									Filename: "testcases/example/modules/starter/main.hcl",
									Start: hcl.Pos{
									Line: 29,
									Column: 17,
									Byte: 542,
									},
									End: hcl.Pos{
									Line: 29,
									Column: 30,
									Byte: 555,
									},
								},
								},
								hcl.TraverseAttr{
								Name: "foo",
								SrcRange: hcl.Range{
									Filename: "testcases/example/modules/starter/main.hcl",
									Start: hcl.Pos{
									Line: 29,
									Column: 30,
									Byte: 555,
									},
									End: hcl.Pos{
									Line: 29,
									Column: 34,
									Byte: 559,
									},
								},
								},
							},
							},
							DeclRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
								Line: 16,
								Column: 1,
								Byte: 232,
							},
							End: hcl.Pos{
								Line: 16,
								Column: 24,
								Byte: 255,
							},
							},
						},
						Depth: 0,
						Dependencies: nil,
						DependenciesAppended: nil,
						},
						&decode.DecodedResource{
						DecodedDeployable: decode.DecodedDeployable{
							Name: "foo",
							Config: map[string]cty.Value{
							"kube_resource.foo[0]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")}), "name":cty.StringVal("service1")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(3), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("service1"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})})})})})})})}),
							"kube_resource.foo[1]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")}), "name":cty.StringVal("service2")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(3), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("service2"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})})})})})})})}),
							},
							Type: "r",
							DependsOn: nil,
							DeclRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
								Line: 32,
								Column: 1,
								Byte: 564,
							},
							End: hcl.Pos{
								Line: 32,
								Column: 20,
								Byte: 583,
							},
							},
						},
						Depth: 0,
						Dependencies: nil,
						DependenciesAppended: nil,
						},
						&decode.DecodedResource{
						DecodedDeployable: decode.DecodedDeployable{
							Name: "bar",
							Config: map[string]cty.Value{
							"kube_resource.bar": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("foobar")}), "name":cty.StringVal("foobar")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(3), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("foobar")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("foobar")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("foobar"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.NumberIntVal(80)})})})})})})})}),
							},
							Type: "r",
							DependsOn: nil,
							DeclRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
								Line: 71,
								Column: 1,
								Byte: 1315,
							},
							End: hcl.Pos{
								Line: 71,
								Column: 20,
								Byte: 1334,
							},
							},
						},
						Depth: 0,
						Dependencies: nil,
						DependenciesAppended: nil,
						},
					},
					ModuleCalls: decode.DecodedModuleCallList{
						&decode.DecodedModuleCall{
						DecodedDeployable: decode.DecodedDeployable{
							Name: "secret",
							Config: map[string]cty.Value{
							"module.secret": cty.ObjectVal(map[string]cty.Value{"source":cty.StringVal("./modules/secret")}),
							},
							Type: "m",
							DependsOn: []hcl.Traversal{ 
							hcl.Traversal{
								hcl.TraverseRoot{
								Name: "kube_resource",
								SrcRange: hcl.Range{
									Filename: "testcases/example/modules/starter/main.hcl",
									Start: hcl.Pos{
									Line: 68,
									Column: 17,
									Byte: 1293,
									},
									End: hcl.Pos{
									Line: 68,
									Column: 30,
									Byte: 1306,
									},
								},
								},
								hcl.TraverseAttr{
								Name: "bar",
								SrcRange: hcl.Range{
									Filename: "testcases/example/modules/starter/main.hcl",
									Start: hcl.Pos{
									Line: 68,
									Column: 30,
									Byte: 1306,
									},
									End: hcl.Pos{
									Line: 68,
									Column: 34,
									Byte: 1310,
									},
								},
								},
							},
							},
							DeclRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
								Line: 66,
								Column: 1,
								Byte: 1229,
							},
							End: hcl.Pos{
								Line: 66,
								Column: 16,
								Byte: 1244,
							},
							},
						},
						Source: "./modules/secret",
						},
					},
					Modules: decode.DecodedModuleList{
						&decode.DecodedModule{
						Name: "secret",
						Inputs: nil,
						Locals: nil,
						Annotations: decode.DecodedAnnotations{
							&decode.DecodedAnnotation{
							Name: "kubehcl.sh/managed",
							Value: cty.StringVal("This resource is managed by kubehcl"),
							DeclRange: hcl.Range{
								Filename: "",
								Start: hcl.Pos{
								Line: 0,
								Column: 0,
								Byte: 0,
								},
								End: hcl.Pos{
								Line: 0,
								Column: 0,
								Byte: 0,
								},
							},
							},
							&decode.DecodedAnnotation{
							Name: "kubehcl.sh/release",
							Value: cty.StringVal("test"),
							DeclRange: hcl.Range{
								Filename: "",
								Start: hcl.Pos{
								Line: 0,
								Column: 0,
								Byte: 0,
								},
								End: hcl.Pos{
								Line: 0,
								Column: 0,
								Byte: 0,
								},
							},
							},
						},
						Resources: decode.DecodedResourceList{
							&decode.DecodedResource{
							DecodedDeployable: decode.DecodedDeployable{
								Name: "secret",
								Config: map[string]cty.Value{
								"kube_resource.secret": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("v1"), "data":cty.ObjectVal(map[string]cty.Value{".secret-file":cty.StringVal("SGVsbG8gV29ybGQ=")}), "kind":cty.StringVal("Secret"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "name":cty.StringVal("dotfile-secret")})}),
								},
								Type: "r",
								DependsOn: nil,
								DeclRange: hcl.Range{
								Filename: "testcases/example/modules/starter/modules/secret/main.hcl",
								Start: hcl.Pos{
									Line: 1,
									Column: 1,
									Byte: 0,
								},
								End: hcl.Pos{
									Line: 1,
									Column: 23,
									Byte: 22,
								},
								},
							},
							Depth: 0,
							Dependencies: nil,
							DependenciesAppended: nil,
							},
						},
						ModuleCalls: nil,
						Modules: nil,
						BackendStorage: nil,
						Depth: 2,
						DependsOn: []hcl.Traversal{ 
							hcl.Traversal{
							hcl.TraverseRoot{
								Name: "kube_resource",
								SrcRange: hcl.Range{
								Filename: "testcases/example/modules/starter/main.hcl",
								Start: hcl.Pos{
									Line: 68,
									Column: 17,
									Byte: 1293,
								},
								End: hcl.Pos{
									Line: 68,
									Column: 30,
									Byte: 1306,
								},
								},
							},
							hcl.TraverseAttr{
								Name: "bar",
								SrcRange: hcl.Range{
								Filename: "testcases/example/modules/starter/main.hcl",
								Start: hcl.Pos{
									Line: 68,
									Column: 30,
									Byte: 1306,
								},
								End: hcl.Pos{
									Line: 68,
									Column: 34,
									Byte: 1310,
								},
								},
							},
							},
						},
						Dependencies: nil,
						},
					},
					BackendStorage: nil,
					Depth: 1,
					DependsOn: []hcl.Traversal{ 
						hcl.Traversal{
						hcl.TraverseRoot{
							Name: "kube_resource",
							SrcRange: hcl.Range{
							Filename: "testcases/example/main.hcl",
							Start: hcl.Pos{
								Line: 55,
								Column: 17,
								Byte: 918,
							},
							End: hcl.Pos{
								Line: 55,
								Column: 30,
								Byte: 931,
							},
							},
						},
						hcl.TraverseAttr{
							Name: "namespace",
							SrcRange: hcl.Range{
							Filename: "testcases/example/main.hcl",
							Start: hcl.Pos{
								Line: 55,
								Column: 30,
								Byte: 931,
							},
							End: hcl.Pos{
								Line: 55,
								Column: 40,
								Byte: 941,
							},
							},
						},
						},
					},
					Dependencies: nil,
					},
				},
				BackendStorage: &decode.DecodedBackendStorage{
					Kind: "kube_secret",
					DeclRange: hcl.Range{
					Filename: "",
					Start: hcl.Pos{
						Line: 0,
						Column: 0,
						Byte: 0,
					},
					End: hcl.Pos{
						Line: 0,
						Column: 0,
						Byte: 0,
					},
					},
				},
				Depth: 0,
				DependsOn: nil,
				Dependencies: nil,
			},
			wantErrors: false,
		},
	}

	for _, test := range tests {
		logging.SetLogger(false)
		want, diags := DecodeFolderAndModules("test",test.d,"","",[]string{},0)
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Don't want errors but received: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Want errors but did not receive any")
		} else {	
				v1,err := json.Marshal(test.want)
				if err != nil {
					t.Errorf("Couldn't marshal object")
				}
				v2,err := json.Marshal(want)
				if err != nil {
					t.Errorf("Couldn't marshal object")
				}
				require.JSONEqf(t,string(v1),string(v2),"Jsons are not equal")
	}
}
}
