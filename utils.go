package main

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/function"
)
const (
	ERROR = hcl.DiagError
	WARNING = hcl.DiagWarning
	INVALID = hcl.DiagInvalid
)



func createContext(vars VariableList) *hcl.EvalContext{
	vals := vars.getMapValues()
	functions := getFunctions()
	// fmt.Printf("%s",vals["var"].AsValueMap())
	return &hcl.EvalContext{
		Variables: vals,
		Functions: functions,
	}
}

func getFunctions() map[string]function.Function {
	return make(map[string]function.Function)
}