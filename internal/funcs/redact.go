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
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/lang/marks"
)

func redactIfSensitive(value interface{}, valueMarks ...cty.ValueMarks) string {
	if marks.Has(cty.DynamicVal.WithMarks(valueMarks...), marks.Sensitive) {
		return "(sensitive value)"
	}
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
