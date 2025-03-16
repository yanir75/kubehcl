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
