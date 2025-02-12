package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)


type Local struct {
	Name string
	Value cty.Value
	DeclRange hcl.Range
}

// var inputLocalsBlockSchema = &hcl.BodySchema{
	
// }
func decodeLocalsBlock(block *hcl.Block) (*Local,hcl.Diagnostics){
	var local *Local = &Local{
		DeclRange : block.DefRange,
	}

	attrs,diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		fmt.Printf("%s","errors")
	}
	for _,attr := range attrs {
		test,_ := attr.Expr.Value(createContext(variables))
		local.Name = attr.Name
		local.Value = test
	}
	


	return local,diags
}
