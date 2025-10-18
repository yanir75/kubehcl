/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package addrs

import (
	"strings"
	"testing"
)

func Test_DefaultAnnotation(t *testing.T) {
	Test := []DefaultAnnotation{
		{
			Name: "kubehcl",
		},
		{
			Name: "version",
		},
		{
			Name: "application",
		},
	}

	for i := 0; i < len(Test)-1; i++ {
		if Test[i].UniqueKey() == Test[i+1] {
			t.Errorf("2 different default keys are equal: %s, %s", Test[i].String(), Test[i+1].String())
		}

	}

	for i := 0; i < len(Test); i++ {

		if !strings.HasPrefix(Test[i].String(), "tag.") {
			t.Errorf("Annotation addr must start with tag this starts with: %s", Test[i].String())
		}
	}
}
