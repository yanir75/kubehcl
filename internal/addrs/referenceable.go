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

// Referenceable is an interface implemented by all address types that can
// appear as references in configuration language expressions.
type Referenceable interface {
	// All implementations of this interface must be covered by the type switch
	// in lang.Scope.buildEvalContext.
	referenceableSigil()

	// All Referenceable address types must have unique keys.
	UniqueKeyer

	// String produces a string representation of the address that could be
	// parsed as a HCL traversal and passed to ParseRef to produce an identical
	// result.
	String() string
}

type referenceable struct {
}

func (r referenceable) referenceableSigil() {
}
