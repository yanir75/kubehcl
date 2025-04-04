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
	"testing"

	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/lang/marks"
)

func TestTo(t *testing.T) {
	tests := []struct {
		Value    cty.Value
		TargetTy cty.Type
		Want     cty.Value
		Err      string
	}{
		{
			cty.StringVal("a"),
			cty.String,
			cty.StringVal("a"),
			``,
		},
		{
			cty.UnknownVal(cty.String),
			cty.String,
			cty.UnknownVal(cty.String),
			``,
		},
		{
			cty.NullVal(cty.String),
			cty.String,
			cty.NullVal(cty.String),
			``,
		},
		{
			// This test case represents evaluating the expression tostring(null)
			// from HCL, since null in HCL is cty.NullVal(cty.DynamicPseudoType).
			// The result in that case should still be null, but a null specifically
			// of type string.
			cty.NullVal(cty.DynamicPseudoType),
			cty.String,
			cty.NullVal(cty.String),
			``,
		},
		{
			cty.StringVal("a").Mark("boop"),
			cty.String,
			cty.StringVal("a").Mark("boop"),
			``,
		},
		{
			cty.NullVal(cty.String).Mark("boop"),
			cty.String,
			cty.NullVal(cty.String).Mark("boop"),
			``,
		},
		{
			cty.True,
			cty.String,
			cty.StringVal("true"),
			``,
		},
		{
			cty.StringVal("a"),
			cty.Bool,
			cty.DynamicVal,
			`cannot convert "a" to bool; only the strings "true" or "false" are allowed`,
		},
		{
			cty.StringVal("a").Mark("boop"),
			cty.Bool,
			cty.DynamicVal,
			`cannot convert "a" to bool; only the strings "true" or "false" are allowed`,
		},
		{
			cty.StringVal("a").Mark(marks.Sensitive),
			cty.Bool,
			cty.DynamicVal,
			`cannot convert this sensitive string to bool`,
		},
		{
			cty.StringVal("a"),
			cty.Number,
			cty.DynamicVal,
			`cannot convert "a" to number; given string must be a decimal representation of a number`,
		},
		{
			cty.StringVal("a").Mark("boop"),
			cty.Number,
			cty.DynamicVal,
			`cannot convert "a" to number; given string must be a decimal representation of a number`,
		},
		{
			cty.StringVal("a").Mark(marks.Sensitive),
			cty.Number,
			cty.DynamicVal,
			`cannot convert this sensitive string to number`,
		},
		{
			cty.NullVal(cty.String),
			cty.Number,
			cty.NullVal(cty.Number),
			``,
		},
		{
			cty.UnknownVal(cty.Bool),
			cty.String,
			cty.UnknownVal(cty.String),
			``,
		},
		{
			cty.UnknownVal(cty.String),
			cty.Bool,
			cty.UnknownVal(cty.Bool), // conversion is optimistic
			``,
		},
		{
			cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.True}),
			cty.List(cty.String),
			cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("true")}),
			``,
		},
		{
			cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.True}),
			cty.Set(cty.String),
			cty.SetVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("true")}),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("hello"), "bar": cty.True}),
			cty.Map(cty.String),
			cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("hello"), "bar": cty.StringVal("true")}),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("hello"), "bar": cty.StringVal("world").Mark("boop")}),
			cty.Map(cty.String),
			cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("hello"), "bar": cty.StringVal("world").Mark("boop")}),
			``,
		},
		{
			cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("hello"), "bar": cty.StringVal("world")}).Mark("boop"),
			cty.Map(cty.String),
			cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("hello"), "bar": cty.StringVal("world")}).Mark("boop"),
			``,
		},
		{
			cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world").Mark("boop")}),
			cty.List(cty.String),
			cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world").Mark("boop")}),
			``,
		},
		{
			cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}).Mark("boop"),
			cty.List(cty.String),
			cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")}).Mark("boop"),
			``,
		},
		{
			cty.EmptyTupleVal,
			cty.String,
			cty.DynamicVal,
			`cannot convert tuple to string`,
		},
		{
			cty.UnknownVal(cty.EmptyTuple),
			cty.String,
			cty.DynamicVal,
			`cannot convert tuple to string`,
		},
		{
			cty.EmptyObjectVal,
			cty.Object(map[string]cty.Type{"foo": cty.String}),
			cty.DynamicVal,
			`incompatible object type for conversion: attribute "foo" is required`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("to %s(%#v)", test.TargetTy.FriendlyNameForConstraint(), test.Value), func(t *testing.T) {
			f := MakeToFunc(test.TargetTy)
			got, err := f.Call([]cty.Value{test.Value})

			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
