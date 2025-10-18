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

func Test_Deployable(t *testing.T) {
	Test := []Deployable{
		{
			Name: "kubehcl",
			Type: MType,
		},
		{
			Name: "version",
			Type: RType,
		},
		{
			Name: "application",
			Type: RType,
		},
	}

	for i := 0; i < len(Test)-1; i++ {
		if Test[i].UniqueKey() == Test[i+1] {
			t.Errorf("2 different default keys are equal: %s, %s", Test[i].String(), Test[i+1].String())
		}

	}

	for i := 1; i < len(Test); i++ {

		if !strings.HasPrefix(Test[i].String(), "module.") && !strings.HasPrefix(Test[i].String(), "kube_resource.") {
			t.Errorf("Deployable addr must start with resource/module this starts with: %s", Test[i].String())
		}
	}
}
