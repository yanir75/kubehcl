/* 
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/lang/marks"
)

func TestRedactIfSensitive(t *testing.T) {
	testCases := map[string]struct {
		value interface{}
		marks []cty.ValueMarks
		want  string
	}{
		"sensitive string": {
			value: "foo",
			marks: []cty.ValueMarks{cty.NewValueMarks(marks.Sensitive)},
			want:  "(sensitive value)",
		},
		"marked non-sensitive string": {
			value: "foo",
			marks: []cty.ValueMarks{cty.NewValueMarks("boop")},
			want:  `"foo"`,
		},
		"sensitive string with other marks": {
			value: "foo",
			marks: []cty.ValueMarks{cty.NewValueMarks("boop"), cty.NewValueMarks(marks.Sensitive)},
			want:  "(sensitive value)",
		},
		"sensitive number": {
			value: 12345,
			marks: []cty.ValueMarks{cty.NewValueMarks(marks.Sensitive)},
			want:  "(sensitive value)",
		},
		"non-sensitive number": {
			value: 12345,
			marks: []cty.ValueMarks{},
			want:  "12345",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := redactIfSensitive(tc.value, tc.marks...)
			if got != tc.want {
				t.Errorf("wrong result, got %v, want %v", got, tc.want)
			}
		})
	}
}

