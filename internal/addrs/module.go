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

type ModuleCall struct {
	referenceable
	Name string
	Mode string
}

// uniqueKeySigil implements UniqueKey.
func (m ModuleCall) uniqueKeySigil() {

}

func (m ModuleCall) String() string {
	return "module." + m.Name
}

func (m ModuleCall) UniqueKey() UniqueKey {
	return m
}

func (m ModuleCall) Equals(o ModuleCall) bool {
	return m.String() == o.String()
}

const (
	mLocal  = "l"
	mRemote = "r"
)

