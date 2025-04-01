/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

import (
	"fmt"
	"testing"
)

// TestUniqueKeyer aims to ensure that all of the types that have unique keys
// will continue to meet the UniqueKeyer contract under future changes.
//
// If you add a new implementation of UniqueKey, consider adding a test case
// for it here.
func TestUniqueKeyer(t *testing.T) {
	tests := []UniqueKeyer{
		Local{Name: "index"},
		DefaultAnnotation{Name: "key"},
		ModuleCall{Name: "workspace"},
		Deployable{Name: "module"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s", test), func(t *testing.T) {
			a := test.UniqueKey()
			b := test.UniqueKey()

			// The following comparison will panic if the unique key is not
			// of a comparable type.
			if a != b {
				t.Fatalf("the two unique keys are not equal\na: %#v\b: %#v", a, b)
			}
		})
	}
}
