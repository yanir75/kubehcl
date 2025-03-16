// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

type Deployable struct {
	referenceable
	Name string
	Type string
}

// uniqueKeySigil implements UniqueKey.
func (d Deployable) uniqueKeySigil() {

}

func (d Deployable) String() string {
	switch d.Type {
	case RType:
		return "resource." + d.Name
	case MType:
		return "module." + d.Name
	default:
		panic("shouldn't get here")
	}
}

func (d Deployable) UniqueKey() UniqueKey {
	return d
}

const (
	RType = "r"
	MType = "m"
)
