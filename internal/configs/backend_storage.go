/*
This file was inspired from https://github.com/opentofu/opentofu
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: MPL-2.0
Licesne: https://www.mozilla.org/en-US/MPL/2.0/
*/
package configs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	// "kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
)

// var variables VariableList

const (
	secretKind = "kube_secret"
	stateless  = "stateless"
)

type BackendStorage struct {
	Kind      string // `json:"Name"`
	Used      bool
	DeclRange hcl.Range // `json:"DeclRange"`
}

// type VariableMap map[string]*Variable

// func (v *Variable) addr() addrs.Variable {
// 	return addrs.Variable{
// 		Name: v.Name,
// 	}
// }

var storageCounter = 0

var inputStorageBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "kube_secret",
		},
		{
			Type: "stateless",
		},
	},
}

// Decode variable and verify the type of the variable matches the default value defined
// If no type is defined each value will be accepted
// Decode the value into a golang cty.value
func (v *BackendStorage) decode(ctx *hcl.EvalContext) (*decode.DecodedBackendStorage, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	dS := &decode.DecodedBackendStorage{
		Kind:      v.Kind,
		DeclRange: v.DeclRange,
	}

	return dS, diags
}

func isValidStorageOption(block *hcl.Block) bool {
	return block.Type == stateless || block.Type == secretKind
}

// Each variable block can contain type,description and default value
// Checks for those
// None of those are required
func decodeStorageBlock(block *hcl.Block) (*BackendStorage, hcl.Diagnostics) {
	var storage *BackendStorage = &BackendStorage{
		Kind: secretKind,
	}

	if block == nil {
		return storage, nil
	}

	content, diags := block.Body.Content(inputStorageBlockSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	if len(content.Blocks) < 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Storage block must have at least one block within it",
			Detail:   fmt.Sprintf("Block %s has no definition within it, valid options are [\"stateless\", \"kube_secret\"]", block.Type),
			Subject:  &block.DefRange,
		})
		return nil, diags
	}

	if len(content.Blocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Storage block must have at only one block within it",
			Detail:   fmt.Sprintf("Block %s has 2 or more blocks within it.", block.Type),
			Subject:  &block.DefRange,
		})
		return nil, diags
	}

	if isValidStorageOption(content.Blocks[0]) {
		storage.Kind = content.Blocks[0].Type
		storage.Used = true
		storage.DeclRange = content.Blocks[0].DefRange
	}

	return storage, diags
}

func DecodeBackendStorageBlocks(blocks hcl.Blocks) (*BackendStorage, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(blocks) == 0 {
		return decodeStorageBlock(nil)
	}

	storageCounter++
	if len(blocks) >= 1 && storageCounter > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one storage block is allowed",
			Detail:   "One storage block is allowed, please remove unnecessary or duplicated storage blocks",
			Subject:  &blocks[0].DefRange,
		})
		return nil, diags
	}

	return decodeStorageBlock(blocks[0])
}
