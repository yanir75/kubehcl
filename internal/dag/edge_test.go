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

package dag

import (
	"testing"
)

func TestBasicEdgeHashcode(t *testing.T) {
	e1 := BasicEdge(1, 2)
	e2 := BasicEdge(1, 2)
	if e1.Hashcode() != e2.Hashcode() {
		t.Fatalf("bad")
	}
}

func TestBasicEdgeHashcode_pointer(t *testing.T) {
	type test struct {
		Value string
	}

	v1, v2 := &test{"foo"}, &test{"bar"}
	e1 := BasicEdge(v1, v2)
	e2 := BasicEdge(v1, v2)
	if e1.Hashcode() != e2.Hashcode() {
		t.Fatalf("bad")
	}
}

