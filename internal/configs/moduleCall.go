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
	"github.com/zclconf/go-cty/cty"
	"kubehcl.sh/kubehcl/internal/addrs"
	"kubehcl.sh/kubehcl/internal/decode"
	// "kubehcl.sh/kubehcl/internal/dag"
)

type ModuleCallList []*ModuleCall

// Decode the source of a module, source means the folder which contains the module
func (m *ModuleCall) DecodeSource(ctx *hcl.EvalContext) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	val, valDdiags := m.Source.Value(ctx)
	diags = append(diags, valDdiags...)
	if val.Type() != cty.String {
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Source must be string",
			Detail:      fmt.Sprintf("Required string and you entered type %s", val.Type().FriendlyName()),
			Subject:     m.Source.Range().Ptr(),
			Expression:  m.Source,
			EvalContext: ctx,
		})
	}
	return val.AsString(), diags
}

// Decode the multiple module calls to the module
func (r ModuleCallList) Decode(ctx *hcl.EvalContext) (decode.DecodedModuleCallList, hcl.Diagnostics) {
	var dR decode.DecodedModuleCallList
	var diags hcl.Diagnostics
	for _, call := range r {
		dV, varDiags := call.decode(ctx)
		diags = append(diags, varDiags...)
		dR = append(dR, dV)
	}

	return dR, diags
}

// Decode the module call which returns the source of the module and decoded deployable
func (r *ModuleCall) decode(ctx *hcl.EvalContext) (*decode.DecodedModuleCall, hcl.Diagnostics) {
	deployable, diags := r.Decode(ctx)
	source, sourceDiags := r.DecodeSource(ctx)
	diags = append(diags, sourceDiags...)
	dM := &decode.DecodedModuleCall{
		Source:            source,
		DecodedDeployable: *deployable,
	}

	return dM, diags
}

var inputModuleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "for_each",
		},
		{
			Name: "count",
		},
		{
			Name: "depends_on",
		},
		{
			Name:     "source",
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{},
}

// Decode module block
// Each block can contain for_each, count depends on and must contain source
// Other is the remaining body of the module
func decodeModuleBlock(block *hcl.Block) (*ModuleCall, hcl.Diagnostics) {
	var Module = &ModuleCall{
		// Name:      block.Labels[0],
		// DeclRange: block.DefRange,
	}
	Module.Name = block.Labels[0]
	Module.DeclRange = block.DefRange
	Module.Type = addrs.MType

	content, remain, diags := block.Body.PartialContent(inputModuleBlockSchema)
	Module.Config = remain
	if attr, exists := content.Attributes["count"]; exists {

		Module.Count = attr.Expr
	}
	if attr, exists := content.Attributes["for_each"]; exists {
		if _, countExists := content.Attributes["count"]; countExists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
				Context:  &content.Attributes["count"].NameRange,
			})
			Module.ForEach = attr.Expr

		}
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		traversal, dependsDiags := decodeDependsOn(attr)
		diags = append(diags, dependsDiags...)
		Module.DependsOn = append(Module.DependsOn, traversal...)
	}

	if attr, exists := content.Attributes["source"]; exists {
		Module.Source = attr.Expr
	}

	return Module, diags

}

//	func getFatherModule(call *ModuleCall,addrMap addrs.AddressMap) *ModuleCall{
//		if len(strings.Split(call.addr().String(),".")) < 3 {
//			return nil
//		}
//		return addrMap[strings.Join(strings.Split(call.addr().String(), ".")[:2],".")].(*ModuleCall)
//	}
//
// Decode multiple blocks of a module call
func DecodeModuleBlocks(blocks hcl.Blocks, addrMap addrs.AddressMap) (ModuleCallList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var callList ModuleCallList
	for _, block := range blocks {
		call, rDiags := decodeModuleBlock(block)
		diags = append(diags, rDiags...)
		callList = append(callList, call)
		// fatherCall := getFatherModule(call,addrMap)
		// if fatherCall != nil {
		// 	call.DependsOn = append(call.DependsOn, fatherCall.DependsOn...)
		// }
		if addrMap.Add(call.addr().String(), call) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Modules must have different names",
				Detail:   fmt.Sprintf("Two Modules have the same name: %s", call.Name),
				Subject:  &block.DefRange,
				// Context: names[variable.Name],
			})
		}
	}

	return callList, diags
}

func (r ModuleCall) addr() addrs.ModuleCall {
	return addrs.ModuleCall{
		Name: r.Name,
	}
}
