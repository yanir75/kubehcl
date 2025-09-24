/*
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terminal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStreamsFmtHelpers(t *testing.T) {
	streams, close := StreamsForTesting(t)

	_,_ = streams.Print("stdout print ", 5, "\n")
	_,_ = streams.Eprint("stderr print ", 6, "\n")
	_,_ = streams.Println("stdout println", 7)
	_,_ = streams.Eprintln("stderr println", 8)
	_,_ = streams.Printf("stdout printf %d\n", 9)
	_,_ = streams.Eprintf("stderr printf %d\n", 10)

	outp := close(t)

	gotOut := outp.Stdout()
	wantOut := `stdout print 5
stdout println 7
stdout printf 9
`
	if diff := cmp.Diff(wantOut, gotOut); diff != "" {
		t.Errorf("wrong stdout\n%s", diff)
	}

	gotErr := outp.Stderr()
	wantErr := `stderr print 6
stderr println 8
stderr printf 10
`
	if diff := cmp.Diff(wantErr, gotErr); diff != "" {
		t.Errorf("wrong stderr\n%s", diff)
	}
}
