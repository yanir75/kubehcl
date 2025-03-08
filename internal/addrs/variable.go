// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

type Variable struct {
	referenceable
	Name string
}

// uniqueKeySigil implements UniqueKey.
func (v Variable) uniqueKeySigil() {

}

func (v Variable) String() string {
	return "var." + v.Name
}

func (v Variable) UniqueKey() UniqueKey {
	return v
}
