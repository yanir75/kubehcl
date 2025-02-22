package main

import (
	"maps"

	"github.com/hashicorp/hcl/v2"
)
const (
	ERROR = hcl.DiagError
	WARNING = hcl.DiagWarning
	INVALID = hcl.DiagInvalid
)



func createContext() *hcl.EvalContext{
	vals := variables.getMapValues()
	locals := locals.getMapValues()
	maps.Copy(vals,locals)
	// fmt.Printf("%s",vals["var"].AsValueMap())
	return &hcl.EvalContext{
		Variables: vals,
		// Functions: makeBaseFunctionTable("./"),
	}
}



// func checkForBlocks (block *hcl.Block) *hcl.Diagnostic {
// 	content,_,_ :=block.Body.PartialContent(&hcl.BodySchema{})
// 	if len(content.Blocks) > 0 {
// 		return &hcl.Diagnostic{
// 			Severity: hcl.DiagError,
// 			Summary:  "No blocks allowed in default_annotation block",
// 			Detail:   fmt.Sprintf("Couldn't convert value to string, this is value of type: %s",content.Blocks[0].Type),
// 			Subject:  &content.Blocks[0].DefRange,
// 		}
// 	}

// 	return &hcl.Diagnostic{
// 		Summary:  "No blocks are found",
// 		Detail:   fmt.Sprintf("There are 0 blocks in  %s",block.Type),
// 		Subject:  &block.DefRange,
// 	}
// }