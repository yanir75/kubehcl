/* 
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"reflect"

	"github.com/zclconf/go-cty/cty"
)

// TypeType is a capsule type used to represent a cty.Type as a cty.Value. This
// is used by the `type()` console function to smuggle cty.Type values to the
// REPL session, where it can be displayed to the user directly.
var TypeType = cty.Capsule("type", reflect.TypeOf(cty.Type{}))

