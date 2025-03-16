// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addrs

type DefaultAnnotation struct {
	referenceable
	Name string
}

// uniqueKeySigil implements UniqueKey.
func (d DefaultAnnotation) uniqueKeySigil() {

}

func (d DefaultAnnotation) String() string {
	return "tag." + d.Name
}

func (d DefaultAnnotation) UniqueKey() UniqueKey {
	return d
}
