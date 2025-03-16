// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

type Resource struct {
	referenceable
	Name         string
	ResourceMode string
}

// uniqueKeySigil implements UniqueKey.
func (r Resource) uniqueKeySigil() {

}

func (r Resource) String() string {
	switch r.ResourceMode {
	case InModule:
		return "module.resource." + r.Name
	case RMode:
		return "resource." + r.Name

	default:
		panic("Can't reach here")
	}
}

func (r Resource) UniqueKey() UniqueKey {
	return r
}

func (r Resource) Equals(o Resource) bool {
	return r.String() == o.String()
}

const (
	InModule = "I"
	RMode    = "R"
)
