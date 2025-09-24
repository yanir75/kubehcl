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
)

func Test_Storage(t *testing.T) {
	tests := []struct {
		d          *hcl.Block
		want       *BackendStorage
		wantErrors bool
	}{
		{
			d: &hcl.Block{
				Type:   "backend_storage",
				Labels: []string{},
				Body: &hclsyntax.Body{
					Blocks: hclsyntax.Blocks{
						&hclsyntax.Block{Type: "stateless"},
					},
				},
			},

			want: &BackendStorage{
				Kind: "stateless",
				Used: true,
			},
			wantErrors: false,
		},
		{
			d: &hcl.Block{
				Type:   "backend_storage",
				Labels: []string{},
				Body: &hclsyntax.Body{
					Blocks: hclsyntax.Blocks{
						&hclsyntax.Block{Type: "kube_secret"},
					},
				},
			},

			want: &BackendStorage{
				Kind: "kube_secret",
				Used: true,
			},
			wantErrors: false,
		},

		{
			d: nil,

			want: &BackendStorage{
				Kind: "kube_secret",
				Used: false,
			},
			wantErrors: false,
		},
	}

	for _, test := range tests {
		want, diags := decodeStorageBlock(test.d)
		if diags.HasErrors() && !test.wantErrors {
			t.Errorf("Don't want errors but received: %s", diags.Errs())
		} else if !diags.HasErrors() && test.wantErrors {
			t.Errorf("Want errors but did not receive any")
		} else {
			if !reflect.DeepEqual(want, test.want) {
				t.Errorf("Values are not equal")
			}
		}
	}
}
