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
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func Test_utils(t *testing.T) {
	tests := []struct {
		variables          DecodedVariableMap
		locals             DecodedLocalsMap
		CountExpr          string
		ForExpr            string
		wantFor, wantCount cty.Value
		wantErrors         bool
	}{
		{
			variables: DecodedVariableMap{
				"test": &DecodedVariable{
					Name:    "test",
					Default: cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("value")}),
				},
			},
			locals: DecodedLocalsMap{
				"test":&DecodedLocal{
					Name:  "test",
					Value: cty.NumberIntVal(1),
				},
			},
			CountExpr:  "local.test",
			ForExpr:    "var.test",
			wantFor:    cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("value")}),
			wantCount:  cty.NumberIntVal(1),
			wantErrors: false,
		},
		{
			variables: DecodedVariableMap{
				"foo":&DecodedVariable{
					Name:    "foo",
					Default: cty.SetVal([]cty.Value{cty.StringVal("bar")}),
				},
				"bar":&DecodedVariable{
					Name:    "bar",
					Default: cty.ObjectVal(map[string]cty.Value{"key": cty.NumberIntVal(5)}),
				},
			},
			locals:     DecodedLocalsMap{},
			CountExpr:  "var.bar[\"key\"]",
			ForExpr:    "var.foo",
			wantFor:    cty.SetVal([]cty.Value{cty.StringVal("bar")}),
			wantCount:  cty.NumberIntVal(5),
			wantErrors: false,
		},
		{
			variables: DecodedVariableMap{
				"test":&DecodedVariable{
					Name: "test",
					Default: cty.ObjectVal(map[string]cty.Value{
						"key": cty.StringVal("value"),
						"object": cty.ObjectVal(map[string]cty.Value{
							"inKey": cty.StringVal("test"),
						},
						),
					},
					),
				},
			},
			locals: DecodedLocalsMap{
				"test":&DecodedLocal{
					Name:  "test",
					Value: cty.NumberIntVal(1),
				},
				"bla":&DecodedLocal{
					Name:  "tester",
					Value: cty.NumberIntVal(1),
				},
			},
			CountExpr: "local.test - local.tester",
			ForExpr:   "var.test[\"object\"]",
			wantFor: cty.ObjectVal(map[string]cty.Value{
				"inKey": cty.StringVal("test"),
			},
			),
			wantCount:  cty.NumberIntVal(0),
			wantErrors: false,
		},

		{
			variables: DecodedVariableMap{
				"bla":&DecodedVariable{
					Name: "test",
					Default: cty.ObjectVal(map[string]cty.Value{
						"key": cty.StringVal("value"),
						"object": cty.ObjectVal(map[string]cty.Value{
							"inKey": cty.StringVal("test"),
						},
						),
					},
					),
				},
			},
			locals: DecodedLocalsMap{
				"test":&DecodedLocal{
					Name:  "test",
					Value: cty.NumberIntVal(1),
				},
				"foo":&DecodedLocal{
					Name:  "tester",
					Value: cty.NumberIntVal(1),
				},
			},
			CountExpr: "local.test - local.tester - 1",
			ForExpr:   "var.test[\"object\"][\"inKey\"]",
			wantFor: cty.ObjectVal(map[string]cty.Value{
				"inKey": cty.StringVal("test"),
			},
			),
			wantCount:  cty.NumberIntVal(0),
			wantErrors: true,
		},
	}

	for _, test := range tests {

		ctx, diags := CreateContext(test.variables, test.locals)
		if diags.HasErrors() {
			t.Errorf("Error received and was not expected: %s", diags.Errs())
		} else {
			countExpr, diags := hclsyntax.ParseExpression([]byte(test.CountExpr), "", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Errorf("Test is not written correctly %s", test.CountExpr)
			}
			countExprHCL, ok := countExpr.(hcl.Expression)
			if !ok {
				t.Errorf("Cannot convert expression to hcl expression %s", test.CountExpr)
			}
			count, countDiags := decodeCountExpr(ctx, countExprHCL)
			if countDiags.HasErrors() && !test.wantErrors {
				t.Errorf("Error received and was not expected: %s", countDiags.Errs())
			} else if !countDiags.HasErrors() && test.wantErrors {
				t.Errorf("Error not received but was expected")
			} else if count.Equals(test.wantCount) != cty.True && !test.wantErrors {
				t.Errorf("Wanted %s got %s", test.wantCount, count)
			}

			forExpr, diags := hclsyntax.ParseExpression([]byte(test.ForExpr), "", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Errorf("Test is not written correctly %s", test.ForExpr)
			}
			forExprHCL, ok := forExpr.(hcl.Expression)
			if !ok {
				t.Errorf("Cannot convert expression to hcl expression %s", test.ForExpr)
			}
			forVal, forDiags := decodeForExpr(ctx, forExprHCL)
			if forDiags.HasErrors() && !test.wantErrors {
				t.Errorf("Error received and was not expected: %s", forDiags.Errs())
			} else if !forDiags.HasErrors() && test.wantErrors {
				t.Errorf("Error not received but was expected")
			} else if forVal.Equals(test.wantFor) != cty.True && !test.wantErrors {
				t.Errorf("Wanted %s got %s", test.wantFor, forVal)
			}
		}
		// count = decodeCountExpr(hclsyntax.ParseExpression([]byte(test.Expr), "", hcl.Pos{Line: 1, Column: 1}))
	}
}
