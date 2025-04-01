/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package decode

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
)

func Test_commonStruct(t *testing.T) {
	tests := []struct {
		d          *Deployable
		want       *DecodedDeployable
		wantErrors bool
	}{
		{
			d: &Deployable{
				Name:    "test",
				Type:    addrs.RType,
				ForEach: nil,
				Count:   nil,
				Config: &hclsyntax.Body{Attributes: hclsyntax.Attributes{
					"key": &hclsyntax.Attribute{
						Name: "key",
						Expr: &hclsyntax.LiteralValueExpr{
							Val: cty.StringVal("testing"),
						},
					},
				},
				},
			},
			want: &DecodedDeployable{
				Name: "test",
				Type: addrs.RType,
				Config: map[string]cty.Value{
					"resource.test": cty.ObjectVal(map[string]cty.Value{
						"key": cty.StringVal("testing"),
					},
					),
				},
			},
			wantErrors: false,
		},

		{
			d: &Deployable{
				Name: "test",
				Type: addrs.RType,
				Config: &hclsyntax.Body{Attributes: hclsyntax.Attributes{
					"key": &hclsyntax.Attribute{
						Name: "key",
						Expr: &hclsyntax.LiteralValueExpr{
							Val: cty.StringVal("testing"),
						},
					},
				},
					Blocks: hclsyntax.Blocks{
						&hclsyntax.Block{
							Type: "red",
							Body: &hclsyntax.Body{Attributes: hclsyntax.Attributes{
								"key": &hclsyntax.Attribute{
									Name: "key",
									Expr: &hclsyntax.LiteralValueExpr{
										Val: cty.StringVal("testing"),
									},
								},
							},
							},
						},
					},
				},
			},
			want: &DecodedDeployable{
				Name: "test",
				Type: addrs.RType,
				Config: map[string]cty.Value{
					"resource.test": cty.ObjectVal(map[string]cty.Value{
						"key": cty.StringVal("testing"),
						"red": cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("testing")}),
					},
					),
				},
			},

			wantErrors: false,
		},

		{
			d: &Deployable{
				Name: "test",
				Type: addrs.MType,
				Config: &hclsyntax.Body{Attributes: hclsyntax.Attributes{
					"key": &hclsyntax.Attribute{
						Name: "key",
						Expr: &hclsyntax.LiteralValueExpr{
							Val: cty.StringVal("testing"),
						},
					},
				},
					Blocks: hclsyntax.Blocks{
						&hclsyntax.Block{
							Type: "red",
							Body: &hclsyntax.Body{Attributes: hclsyntax.Attributes{
								"key": &hclsyntax.Attribute{
									Name: "key",
									Expr: &hclsyntax.LiteralValueExpr{
										Val: cty.StringVal("testing"),
									},
								},
							},
							},
						},
					},
				},
			},
			want: &DecodedDeployable{
				Name: "test",
				Type: addrs.MType,
				Config: map[string]cty.Value{
					"module.test": cty.ObjectVal(map[string]cty.Value{
						"key": cty.StringVal("testing"),
						"red": cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("testing")}),
					},
					),
				},
			},

			wantErrors: false,
		},
	}

	for _, test := range tests {
		d, diags := test.d.Decode(&hcl.EvalContext{})
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Error received and was not expected: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Error not received but was expected")
		} else if !reflect.DeepEqual(d, test.want) && !test.wantErrors {
			t.Errorf("Wanted \n%s got  \n%s", test.want, d)
		}
	}
}
