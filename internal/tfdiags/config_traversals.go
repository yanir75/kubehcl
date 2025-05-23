/*
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfdiags

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/zclconf/go-cty/cty"
)

// FormatCtyPath is a helper function to produce a user-friendly string
// representation of a cty.Path. The result uses a syntax similar to the
// HCL expression language in the hope of it being familiar to users.
func FormatCtyPath(path cty.Path) string {
	var buf bytes.Buffer
	for _, step := range path {
		switch ts := step.(type) {
		case cty.GetAttrStep:
			fmt.Fprintf(&buf, ".%s", ts.Name)
		case cty.IndexStep:
			buf.WriteByte('[')
			key := ts.Key
			keyTy := key.Type()
			switch {
			case key.IsNull():
				buf.WriteString("null")
			case !key.IsKnown():
				buf.WriteString("(not yet known)")
			case keyTy == cty.Number:
				bf := key.AsBigFloat()
				buf.WriteString(bf.Text('g', -1))
			case keyTy == cty.String:
				buf.WriteString(strconv.Quote(key.AsString()))
			default:
				buf.WriteString("...")
			}
			buf.WriteByte(']')
		}
	}
	return buf.String()
}

// FormatError is a helper function to produce a user-friendly string
// representation of certain special error types that we might want to
// include in diagnostic messages.
//
// This currently has special behavior only for cty.PathError, where a
// non-empty path is rendered in a HCL-like syntax as context.
func FormatError(err error) string {
	perr, ok := err.(cty.PathError)
	if !ok || len(perr.Path) == 0 {
		return err.Error()
	}

	return fmt.Sprintf("%s: %s", FormatCtyPath(perr.Path), perr.Error())
}

// FormatErrorPrefixed is like FormatError except that it presents any path
// information after the given prefix string, which is assumed to contain
// an HCL syntax representation of the value that errors are relative to.
func FormatErrorPrefixed(err error, prefix string) string {
	perr, ok := err.(cty.PathError)
	if !ok || len(perr.Path) == 0 {
		return fmt.Sprintf("%s: %s", prefix, err.Error())
	}

	return fmt.Sprintf("%s%s: %s", prefix, FormatCtyPath(perr.Path), perr.Error())
}
