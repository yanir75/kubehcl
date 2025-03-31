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

package addrs

type Local struct {
	referenceable
	Name string
}

// uniqueKeySigil implements UniqueKey.
func (l Local) uniqueKeySigil() {

}

func (l Local) String() string {
	return "local." + l.Name
}

func (l Local) UniqueKey() UniqueKey {
	return l
}

