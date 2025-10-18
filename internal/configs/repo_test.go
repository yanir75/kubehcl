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

func Test_Repo(t *testing.T) {
	tests := []struct {
		d          hcl.Blocks
		want       RepoMap
		wantErrors bool
	}{
		{
			d: hcl.Blocks{
				&hcl.Block{
					Type:     "repo",
					Labels:   []string{"foo"},
					DefRange: hcl.EmptyBody().MissingItemRange(),
					Body: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"Name": &hclsyntax.Attribute{
								Name: "Name",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("foo")},
							},
							"Url": &hclsyntax.Attribute{
								Name: "Url",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("foo")},
							},
							"Protocol": &hclsyntax.Attribute{
								Name: "Protocol",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("https")},
							},
						},
					},
				},
				&hcl.Block{
					Type:     "repo",
					Labels:   []string{"bar"},
					DefRange: hcl.EmptyBody().MissingItemRange(),
					Body: &hclsyntax.Body{
						Attributes: hclsyntax.Attributes{
							"Name": &hclsyntax.Attribute{
								Name: "Name",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("foo")},
							},
							"Url": &hclsyntax.Attribute{
								Name: "Url",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("foo")},
							},
							"Protocol": &hclsyntax.Attribute{
								Name: "Protocol",
								Expr: &hclsyntax.LiteralValueExpr{Val: cty.StringVal("oci")},
							},
						},
					},
				},
			},
			want: RepoMap{
				"foo": &Repo{
					Name:                  "foo",
					DeclRange:             hcl.EmptyBody().MissingItemRange(),
					Username:              "",
					Url:                   "foo",
					Password:              "",
					Timeout:               120,
					CaFile:                "",
					CertFile:              "",
					KeyFile:               "",
					InsecureSkipTLSverify: false,
					PlainHttp:             false,
					RepoFile:              "",
					RepoCache:             "",
					Protocol:              "https",
				},
				"bar": &Repo{
					Name:                  "foo",
					DeclRange:             hcl.EmptyBody().MissingItemRange(),
					Username:              "",
					Url:                   "foo",
					Password:              "",
					Timeout:               120,
					CaFile:                "",
					CertFile:              "",
					KeyFile:               "",
					InsecureSkipTLSverify: false,
					PlainHttp:             false,
					RepoFile:              "",
					RepoCache:             "",
					Protocol:              "oci",
				},
			},
			wantErrors: false,
		},
	}

	for _, test := range tests {
		want, diags := DecodeRepoBlocks(test.d)
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Don't want errors but received: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Want errors but did not receive any")
		} else {
			if len(want) != len(test.want) {
				t.Errorf("Length of the results is not equal")
				for i, r := range want {
					if !reflect.DeepEqual(test.want[i], r) {
						t.Errorf("Resources are not equal %s , %s", want[i].Name, r.Name)
					}
				}
			}
		}
	}
}
