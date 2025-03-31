/* 
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfdiags

// Sourceless creates and returns a diagnostic with no source location
// information. This is generally used for operational-type errors that are
// caused by or relate to the environment where OpenTofu is running rather
// than to the provided configuration.
func Sourceless(severity Severity, summary, detail string) Diagnostic {
	return diagnosticBase{
		severity: severity,
		summary:  summary,
		detail:   detail,
	}
}

