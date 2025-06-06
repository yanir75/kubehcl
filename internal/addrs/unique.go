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

// UniqueKey is an interface implemented by values that serve as unique map
// keys for particular addresses.
//
// All implementations of UniqueKey are comparable and can thus be used as
// map keys. Unique keys generated from different address types are always
// distinct. All functionally-equivalent keys for the same address type
// always compare equal, and likewise functionally-different values do not.
type UniqueKey interface {
	uniqueKeySigil()
}

// UniqueKeyer is an interface implemented by types that can be represented
// by a unique key.
//
// Some address types naturally comply with the expectations of a UniqueKey
// and may thus be their own unique key type. However, address types that
// are not naturally comparable can implement this interface by returning
// proxy values.
type UniqueKeyer interface {
	UniqueKey() UniqueKey
}

// func Equivalent[T UniqueKeyer](a, b T) bool {
// 	return a.UniqueKey() == b.UniqueKey()
// }
