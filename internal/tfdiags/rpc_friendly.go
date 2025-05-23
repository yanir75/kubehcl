/*
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfdiags

import (
	"encoding/gob"
)

type rpcFriendlyDiag struct {
	Severity_ Severity
	Summary_  string
	Detail_   string
	Subject_  *SourceRange
	Context_  *SourceRange
}

// rpcFriendlyDiag transforms a given diagnostic so that is more friendly to
// RPC.
//
// In particular, it currently returns an object that can be serialized and
// later re-inflated using gob. This definition may grow to include other
// serializations later.
func makeRPCFriendlyDiag(diag Diagnostic) Diagnostic {
	desc := diag.Description()
	source := diag.Source()
	return &rpcFriendlyDiag{
		Severity_: diag.Severity(),
		Summary_:  desc.Summary,
		Detail_:   desc.Detail,
		Subject_:  source.Subject,
		Context_:  source.Context,
	}
}

func (d *rpcFriendlyDiag) Severity() Severity {
	return d.Severity_
}

func (d *rpcFriendlyDiag) Description() Description {
	return Description{
		Summary: d.Summary_,
		Detail:  d.Detail_,
	}
}

func (d *rpcFriendlyDiag) Source() Source {
	return Source{
		Subject: d.Subject_,
		Context: d.Context_,
	}
}

func (d rpcFriendlyDiag) FromExpr() *FromExpr {
	// RPC-friendly diagnostics cannot preserve expression information because
	// expressions themselves are not RPC-friendly.
	return nil
}

func (d rpcFriendlyDiag) ExtraInfo() interface{} {
	// RPC-friendly diagnostics always discard any "extra information".
	return nil
}

func init() {
	gob.Register((*rpcFriendlyDiag)(nil))
}
