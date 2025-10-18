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
	"testing"
)

func Test_AddressMap(t *testing.T) {
	Test := AddressMap{}
	b := Test.Add("item", "test")
	if b {
		t.Errorf("Address map says item exists even though it does not.")
	}
	b = Test.Add("item", "test")
	if !b {
		t.Errorf("Address map says item does not exist even though it does.")
	}
}
