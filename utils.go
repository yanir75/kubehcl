package main

import (
	"maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/function"
)
const (
	ERROR = hcl.DiagError
	WARNING = hcl.DiagWarning
	INVALID = hcl.DiagInvalid
)



func createContext() *hcl.EvalContext{
	vals := variables.getMapValues()
	functions := getFunctions()
	locals := locals.getMapValues()
	maps.Copy(locals,vals)
	// fmt.Printf("%s",vals["var"].AsValueMap())
	return &hcl.EvalContext{
		Variables: vals,
		Functions: functions,
	}
}

func getFunctions() map[string]function.Function {
	return make(map[string]function.Function)
}