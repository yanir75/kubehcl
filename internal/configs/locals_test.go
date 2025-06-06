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
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func Test_Locals(t *testing.T) {
	tests := []struct {
		d          *hcl.Block
		want       Locals
		wantErrors bool
	}{
		{
			d: &hcl.Block{
				Type:   "locals",
				Labels: []string{},
				Body: &hclsyntax.Body{
					Attributes: hclsyntax.Attributes{
						"test": &hclsyntax.Attribute{
							Name: "testing",
							Expr: &hclsyntax.LiteralValueExpr{
								Val: cty.StringVal("test"),
							},
						},
						"test1": &hclsyntax.Attribute{
							Name: "testingadf",
							Expr: &hclsyntax.LiteralValueExpr{
								Val: cty.StringVal("testasdf"),
							},
						},
					},
				},
			},

			want: Locals{
				&Local{
					Name: "testing",
					Value: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("test"),
					},
				},
				&Local{
					Name: "testingadf",
					Value: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("testasdf"),
					},
				},
			},
			wantErrors: false,
		},

		{
			d: &hcl.Block{
				Type:   "locals",
				Labels: []string{},
				Body: &hclsyntax.Body{
					Attributes: hclsyntax.Attributes{
						"test": &hclsyntax.Attribute{
							Name: "test1",
							Expr: &hclsyntax.LiteralValueExpr{
								Val: cty.StringVal("asdf"),
							},
						},
						"test1": &hclsyntax.Attribute{
							Name: "test2",
							Expr: &hclsyntax.ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseAttr{
										Name: "var",
									},
									hcl.TraverseAttr{
										Name: "bla",
									},
								},
							},
						},
					},
				},
			},

			want: Locals{
				&Local{
					Name: "test1",
					Value: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("asdf"),
					},
				},
				&Local{
					Name: "test2",
					Value: &hclsyntax.ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseAttr{
								Name: "var",
							},
							hcl.TraverseAttr{
								Name: "bla",
							},
						},
					},
				},
			},
			wantErrors: false,
		},
	}

	for _, test := range tests {
		want, diags := decodeLocalsBlock(test.d)
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Don't want errors but received: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Want errors but did not receive any")
		} else {
			localMap := make(map[string]hcl.Expression)
			for _, local := range want {
				localMap[local.Name] = local.Value
			}
			for _, local := range test.want {
				if !reflect.DeepEqual(local.Value, localMap[local.Name]) {
					t.Errorf("Values are not equal")
				}
			}
		}
	}
}
