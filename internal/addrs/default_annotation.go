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

