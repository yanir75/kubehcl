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

	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl/v2"
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
			Inputs: decode.DecodedVariableMap{
				"foo": &decode.DecodedVariable{
				Name: "foo",
				Description: "Ports of the container",
				Default: cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})}),
				Type: cty.List(cty.Map(cty.String)),
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
			Locals: decode.DecodedLocalsMap{},
			Annotations: decode.DecodedAnnotationsMap{
				"bar": &decode.DecodedAnnotation{
				Name: "bar",
				Value: cty.StringVal("foo"),
				DeclRange: hcl.Range{
					Filename: "testcases/example/main.hcl",
					Start: hcl.Pos{
					Line: 63,
					Column: 3,
					Byte: 1011,
					},
					End: hcl.Pos{
					Line: 63,
					Column: 6,
					Byte: 1014,
					},
				},
				},
				"foo": &decode.DecodedAnnotation{
				Name: "foo",
				Value: cty.StringVal("bar"),
				DeclRange: hcl.Range{
					Filename: "testcases/example/main.hcl",
					Start: hcl.Pos{
					Line: 59,
					Column: 3,
					Byte: 972,
					},
					End: hcl.Pos{
					Line: 59,
					Column: 6,
					Byte: 975,
					},
				},
				},
				"kubehcl.sh/managed": &decode.DecodedAnnotation{
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
				"kubehcl.sh/release": &decode.DecodedAnnotation{
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
			Resources: decode.DecodedResourceMap{
				"foo": &decode.DecodedResource{
				DecodedDeployable: decode.DecodedDeployable{
					Name: "foo",
					Config: map[string]cty.Value{
					"kube_resource.foo[bar]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"bar":cty.StringVal("foo"), "foo":cty.StringVal("bar"), "kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("foo")}), "name":cty.StringVal("bar")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.StringVal("2"), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("foo")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("foo")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("foo"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})})})})})})})}),
					"kube_resource.foo[foo]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"bar":cty.StringVal("foo"), "foo":cty.StringVal("bar"), "kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("bar")}), "name":cty.StringVal("foo")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.StringVal("2"), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("bar")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app":cty.StringVal("bar")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("bar"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})})})})})})})}),
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
							Byte: 749,
							},
							End: hcl.Pos{
							Line: 48,
							Column: 23,
							Byte: 755,
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
							Byte: 755,
							},
							End: hcl.Pos{
							Line: 48,
							Column: 28,
							Byte: 760,
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
							Byte: 762,
							},
							End: hcl.Pos{
							Line: 48,
							Column: 43,
							Byte: 775,
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
							Byte: 775,
							},
							End: hcl.Pos{
							Line: 48,
							Column: 53,
							Byte: 785,
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
				"namespace": &decode.DecodedResource{
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
			},
			ModuleCalls: decode.DecodedModuleCallMap{
				"test": &decode.DecodedModuleCall{
				DecodedDeployable: decode.DecodedDeployable{
					Name: "test",
					Config: map[string]cty.Value{
					"module.test": cty.ObjectVal(map[string]cty.Value{"foo":cty.TupleVal([]cty.Value{cty.StringVal("service1"), cty.StringVal("service2")}), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})}), "source":cty.StringVal("./modules/starter")}),
					},
					Type: "m",
					DependsOn: []hcl.Traversal{ // p0
					hcl.Traversal{
						hcl.TraverseRoot{
						Name: "kube_resource",
						SrcRange: hcl.Range{
							Filename: "testcases/example/main.hcl",
							Start: hcl.Pos{
							Line: 55,
							Column: 17,
							Byte: 920,
							},
							End: hcl.Pos{
							Line: 55,
							Column: 30,
							Byte: 933,
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
							Byte: 933,
							},
							End: hcl.Pos{
							Line: 55,
							Column: 40,
							Byte: 943,
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
						Byte: 790,
					},
					End: hcl.Pos{
						Line: 51,
						Column: 14,
						Byte: 803,
					},
					},
				},
				Source: "./modules/starter",
				},
			},
			Modules: decode.DecodedModuleMap{
				"test": &decode.DecodedModule{
				Name: "test",
				Inputs: decode.DecodedVariableMap{
					"foo": &decode.DecodedVariable{
					Name: "foo",
					Description: "",
					Default: cty.ListVal([]cty.Value{cty.StringVal("service1"), cty.StringVal("service2")}),
					Type: cty.List(cty.String),
					DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
						Line: 53,
						Column: 16,
						Byte: 856,
						},
						End: hcl.Pos{
						Line: 53,
						Column: 40,
						Byte: 880,
						},
					},
					},
					"ports": &decode.DecodedVariable{
					Name: "ports",
					Description: "",
					Default: cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})}),
					Type: cty.NilType,
					DeclRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
						Line: 54,
						Column: 16,
						Byte: 896,
						},
						End: hcl.Pos{
						Line: 54,
						Column: 23,
						Byte: 903,
						},
					},
					},
				},
				Locals: decode.DecodedLocalsMap{
					"other_option": &decode.DecodedLocal{
					Name: "other_option",
					Value: cty.ObjectVal(map[string]cty.Value{"service1":cty.ObjectVal(map[string]cty.Value{"targetPort":cty.StringVal("80")}), "service2":cty.ObjectVal(map[string]cty.Value{"targetPort":cty.StringVal("80")})}),
					DeclRange: hcl.Range{
						Filename: "testcases/example/modules/starter/main.hcl",
						Start: hcl.Pos{
						Line: 9,
						Column: 3,
						Byte: 146,
						},
						End: hcl.Pos{
						Line: 9,
						Column: 15,
						Byte: 158,
						},
					},
					},
					"service_ports": &decode.DecodedLocal{
					Name: "service_ports",
					Value: cty.ObjectVal(map[string]cty.Value{"0":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service1"), "targetPort":cty.StringVal("80")}), "1":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service2"), "targetPort":cty.StringVal("80")})}),
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
				},
				Annotations: decode.DecodedAnnotationsMap{
					"kubehcl.sh/managed": &decode.DecodedAnnotation{
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
					"kubehcl.sh/release": &decode.DecodedAnnotation{
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
				Resources: decode.DecodedResourceMap{
					"bar": &decode.DecodedResource{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "bar",
						Config: map[string]cty.Value{
						"kube_resource.bar": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("foobar")}), "name":cty.StringVal("foobar")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(3), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("foobar")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("foobar")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("foobar"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})})})})})})})}),
						},
						Type: "r",
						DependsOn: nil,
						DeclRange: hcl.Range{
						Filename: "testcases/example/modules/starter/main.hcl",
						Start: hcl.Pos{
							Line: 71,
							Column: 1,
							Byte: 1321,
						},
						End: hcl.Pos{
							Line: 71,
							Column: 20,
							Byte: 1340,
						},
						},
					},
					Depth: 0,
					Dependencies: nil,
					DependenciesAppended: nil,
					},
					"foo": &decode.DecodedResource{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "foo",
						Config: map[string]cty.Value{
						"kube_resource.foo[0]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")}), "name":cty.StringVal("service1")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(3), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("service1"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})})})})})})})}),
						"kube_resource.foo[1]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("apps/v1"), "kind":cty.StringVal("Deployment"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")}), "name":cty.StringVal("service2")}), "spec":cty.ObjectVal(map[string]cty.Value{"replicas":cty.NumberIntVal(3), "selector":cty.ObjectVal(map[string]cty.Value{"matchLabels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")})}), "template":cty.ObjectVal(map[string]cty.Value{"metadata":cty.ObjectVal(map[string]cty.Value{"labels":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")})}), "spec":cty.ObjectVal(map[string]cty.Value{"containers":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"image":cty.StringVal("nginx:1.14.2"), "name":cty.StringVal("service2"), "ports":cty.ListVal([]cty.Value{cty.MapVal(map[string]cty.Value{"containerPort":cty.StringVal("80")})})})})})})})}),
						},
						Type: "r",
						DependsOn: nil,
						DeclRange: hcl.Range{
						Filename: "testcases/example/modules/starter/main.hcl",
						Start: hcl.Pos{
							Line: 32,
							Column: 1,
							Byte: 570,
						},
						End: hcl.Pos{
							Line: 32,
							Column: 20,
							Byte: 589,
						},
						},
					},
					Depth: 0,
					Dependencies: nil,
					DependenciesAppended: nil,
					},
					"service": &decode.DecodedResource{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "service",
						Config: map[string]cty.Value{
						"kube_resource.service[0]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("v1"), "kind":cty.StringVal("Service"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "name":cty.StringVal("service1")}), "spec":cty.ObjectVal(map[string]cty.Value{"ports":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service1"), "port":cty.StringVal("9367"), "targetPort":cty.StringVal("80")})}), "selector":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service1")})})}),
						"kube_resource.service[1]": cty.ObjectVal(map[string]cty.Value{"apiVersion":cty.StringVal("v1"), "kind":cty.StringVal("Service"), "metadata":cty.ObjectVal(map[string]cty.Value{"annotations":cty.ObjectVal(map[string]cty.Value{"kubehcl.sh/managed":cty.StringVal("This resource is managed by kubehcl"), "kubehcl.sh/release":cty.StringVal("test")}), "name":cty.StringVal("service2")}), "spec":cty.ObjectVal(map[string]cty.Value{"ports":cty.TupleVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("service2"), "port":cty.StringVal("9367"), "targetPort":cty.StringVal("80")})}), "selector":cty.ObjectVal(map[string]cty.Value{"app.kubernetes.io/name":cty.StringVal("service2")})})}),
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
								Byte: 548,
								},
								End: hcl.Pos{
								Line: 29,
								Column: 30,
								Byte: 561,
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
								Byte: 561,
								},
								End: hcl.Pos{
								Line: 29,
								Column: 34,
								Byte: 565,
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
							Byte: 236,
						},
						End: hcl.Pos{
							Line: 16,
							Column: 24,
							Byte: 259,
						},
						},
					},
					Depth: 0,
					Dependencies: nil,
					DependenciesAppended: nil,
					},
				},
				ModuleCalls: decode.DecodedModuleCallMap{
					"secret": &decode.DecodedModuleCall{
					DecodedDeployable: decode.DecodedDeployable{
						Name: "secret",
						Config: map[string]cty.Value{
						"module.secret": cty.ObjectVal(map[string]cty.Value{"source":cty.StringVal("./modules/secret")}),
						},
						Type: "m",
						DependsOn: []hcl.Traversal{ // p1
						hcl.Traversal{
							hcl.TraverseRoot{
							Name: "kube_resource",
							SrcRange: hcl.Range{
								Filename: "testcases/example/modules/starter/main.hcl",
								Start: hcl.Pos{
								Line: 68,
								Column: 17,
								Byte: 1299,
								},
								End: hcl.Pos{
								Line: 68,
								Column: 30,
								Byte: 1312,
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
								Byte: 1312,
								},
								End: hcl.Pos{
								Line: 68,
								Column: 34,
								Byte: 1316,
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
							Byte: 1235,
						},
						End: hcl.Pos{
							Line: 66,
							Column: 16,
							Byte: 1250,
						},
						},
					},
					Source: "./modules/secret",
					},
				},
				Modules: decode.DecodedModuleMap{
					"secret": &decode.DecodedModule{
					Name: "secret",
					Inputs: decode.DecodedVariableMap{},
					Locals: decode.DecodedLocalsMap{},
					Annotations: decode.DecodedAnnotationsMap{
						"kubehcl.sh/managed": &decode.DecodedAnnotation{
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
						"kubehcl.sh/release": &decode.DecodedAnnotation{
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
					Resources: decode.DecodedResourceMap{
						"secret": &decode.DecodedResource{
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
					ModuleCalls: decode.DecodedModuleCallMap{},
					Modules: decode.DecodedModuleMap{},
					BackendStorage: nil,
					Depth: 2,
					DependsOn: []hcl.Traversal{ // p1
						hcl.Traversal{
						hcl.TraverseRoot{
							Name: "kube_resource",
							SrcRange: hcl.Range{
							Filename: "testcases/example/modules/starter/main.hcl",
							Start: hcl.Pos{
								Line: 68,
								Column: 17,
								Byte: 1299,
							},
							End: hcl.Pos{
								Line: 68,
								Column: 30,
								Byte: 1312,
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
								Byte: 1312,
							},
							End: hcl.Pos{
								Line: 68,
								Column: 34,
								Byte: 1316,
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
				DependsOn: []hcl.Traversal{ // p0
					hcl.Traversal{
					hcl.TraverseRoot{
						Name: "kube_resource",
						SrcRange: hcl.Range{
						Filename: "testcases/example/main.hcl",
						Start: hcl.Pos{
							Line: 55,
							Column: 17,
							Byte: 920,
						},
						End: hcl.Pos{
							Line: 55,
							Column: 30,
							Byte: 933,
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
							Byte: 933,
						},
						End: hcl.Pos{
							Line: 55,
							Column: 40,
							Byte: 943,
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
		{
			d: "testcases/example1",
			want: &decode.DecodedModule{},
			wantErrors: true,
		},
		{
			d: "testcases/example2",
			want: &decode.DecodedModule{},
			wantErrors: true,
		},
		{
			d: "testcases/example3",
			want: &decode.DecodedModule{},
			wantErrors: true,
		},
		{
			d: "testcases/example4",
			want: &decode.DecodedModule{},
			wantErrors: true,
		},
		{
			d: "testcases/example5",
			want: &decode.DecodedModule{},
			wantErrors: true,
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
			strs := deep.Equal(want,test.want)	
			if len(strs) > 0{
				t.Errorf("Modules are not equal %s",strs)
			}
	}
}
}
