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
)

func Test_Variable(t *testing.T) {
	tests := []struct {
		d          []*hcl.Block
		want       VariableMap
		wantErrors bool
	}{
		{
			d: []*hcl.Block{
				&hcl.Block{
					Type:   "variable",
					Labels: []string{"foo"},
					Body: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"default": &hclsyntax.Attribute{
								Name: "default",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("bar")},
							},
							"type": &hclsyntax.Attribute{
								Name: "type",
								Expr: &hclsyntax.ScopeTraversalExpr{Traversal: hcl.Traversal{hcl.TraverseRoot{Name: "string"}}},
							},
						},
					},
				},
				&hcl.Block{
					Type:   "variable",
					Labels: []string{"bar"},

					Body: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"default": &hclsyntax.Attribute{
								Name: "default",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")})},
							},
						},
					},
				},
			},
			want: VariableMap{"foo": &Variable{
				Name:       "foo",
				Default:    &hclsyntax.LiteralValueExpr{Val: cty.StringVal("bar")},
				Type:       cty.String,
				HasDefault: true,
			},
				"bar": &Variable{
					Name:       "bar",
					Default:    &hclsyntax.LiteralValueExpr{Val: cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")})},
					HasDefault: true,
				},
			},
			wantErrors: false,
		},
	}

	for _, test := range tests {
		want, diags := DecodeVariableBlocks(test.d)
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Don't want errors but received: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Want errors but did not receive any")
		} else {
			for _, variable := range test.want {
				varWant := want[variable.Name]
				if !reflect.DeepEqual(variable, varWant) {
					t.Errorf("Values are not equal")
				}
			}
		}
	}
}
