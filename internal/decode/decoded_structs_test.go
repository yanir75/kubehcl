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
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func Test_getMapValuesVariable(t *testing.T) {
	tests := []struct {
		d    DecodedVariableMap
		want map[string]cty.Value
	}{
		{
			d: DecodedVariableMap{
				"test": &DecodedVariable{
					Name:    "test",
					Default: cty.StringVal("test"),
					Type:    cty.String,
				},
				"bla": &DecodedVariable{
					Name:    "foo",
					Default: cty.StringVal("bar"),
					Type:    cty.String,
				},
				"ra": &DecodedVariable{
					Name:    "bar",
					Default: cty.MustParseNumberVal("5"),
					Type:    cty.Number,
				},
			},
			want: map[string]cty.Value{
				"var": cty.ObjectVal(map[string]cty.Value{
					"test": cty.StringVal("test"),
					"foo":  cty.StringVal("bar"),
					"bar":  cty.MustParseNumberVal("5"),
				}),
			},
		},
	}

	for _, test := range tests {
		v, _ := test.d.getMapValues()
		if !reflect.DeepEqual(v, test.want) {
			t.Errorf("Maps are not equal")
		}
	}
}

func Test_getMapValuesLocal(t *testing.T) {
	tests := []struct {
		d    DecodedLocalsMap
		want map[string]cty.Value
	}{
		{
			d: DecodedLocalsMap{
				"test": &DecodedLocal{
					Name:  "test",
					Value: cty.StringVal("test"),
				},
				"foo": &DecodedLocal{
					Name:  "foo",
					Value: cty.StringVal("bar"),
				},
				"bar": &DecodedLocal{
					Name:  "bar",
					Value: cty.MustParseNumberVal("5"),
				},
			},
			want: map[string]cty.Value{
				"local": cty.ObjectVal(map[string]cty.Value{
					"test": cty.StringVal("test"),
					"foo":  cty.StringVal("bar"),
					"bar":  cty.MustParseNumberVal("5"),
				}),
			},
		},
	}

	for _, test := range tests {
		v := test.d.getMapValues()
		if !reflect.DeepEqual(v, test.want) {
			t.Errorf("Maps are not equal")
		}
	}
}
