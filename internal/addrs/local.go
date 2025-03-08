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
