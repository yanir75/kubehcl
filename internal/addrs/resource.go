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
		return "module." + r.Name
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
	InModule = "module"
	RMode    = "resource"
)
