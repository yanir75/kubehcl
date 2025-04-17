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
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

func Test_ModuleCall(t *testing.T) {
	tests := []struct {
		d          hcl.Blocks
		want       ModuleCallList
		wantErrors bool
	}{
		{
			d: hcl.Blocks{
				&hcl.Block{
					Type:   "module",
					Labels: []string{"foo"},
					Body: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"kind": &hclsyntax.Attribute{
								Name: "kind",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("bar")},
							},
							"version": &hclsyntax.Attribute{
								Name: "version",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.NumberIntVal(5)},
							},
							"source": &hclsyntax.Attribute{
								Name: "source",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("./test")},
							},
						},
						Blocks: hclsyntax.Blocks{
							&hclsyntax.Block{
								Type: "test",
								Body: &hclsyntax.Body{
									Attributes: hclsyntax.Attributes{
										"foo": &hclsyntax.Attribute{Name: "foo",Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("bar")}},
									},
								},
							},
						},
					},
				},
				&hcl.Block{
					Type:   "module",
					Labels: []string{"bar"},

					Body: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"default": &hclsyntax.Attribute{
								Name: "default",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")})},
							},
							"source": &hclsyntax.Attribute{
								Name: "source",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("./test")},
							},
						},
					},
				},
			},
			want: ModuleCallList{&ModuleCall{ 
				Deployable: decode.Deployable{
					Name: "foo",
					Config: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"kind": &hclsyntax.Attribute{Name: "kind",Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("bar")}},
							"version": &hclsyntax.Attribute{Name: "version",Expr: &hclsyntax.LiteralValueExpr{Val: cty.NumberIntVal(5)}},
							"source": &hclsyntax.Attribute{
								Name: "source",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("./test")},
							},
						},
						Blocks: hclsyntax.Blocks{
							&hclsyntax.Block{
								Type: "test",
								Body: &hclsyntax.Body{
									Attributes: hclsyntax.Attributes{
										"foo": &hclsyntax.Attribute{Name: "foo",Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("bar")}},
									},
								},
							},
						},
					},
				},
			},
			&ModuleCall{ 
				Deployable: decode.Deployable{
					Name: "bar",
					Config: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"default": &hclsyntax.Attribute{Name: "kind",Expr: &hclsyntax.LiteralValueExpr{Val: cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")})}},
							"source": &hclsyntax.Attribute{
								Name: "source",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("./test")},
							},
						},
						
					},
				},
			},
		
		},
		wantErrors: false,
		},
			
	}

	for _, test := range tests {
		want, diags := DecodeModuleBlocks(test.d,addrs.AddressMap{})
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Don't want errors but received: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Want errors but did not receive any")
		} else {
			if len(want) != len(test.want){
				t.Errorf("Length of the results is not equal")
				for i,r := range want {
					if !reflect.DeepEqual(test.want[i],r) {
						t.Errorf("Module Calls are not equal %s , %s",want[i].Name,r.Name)
					} 
				}
			}
		}
	}
}
