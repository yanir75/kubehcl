package main

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type Annotation struct {
	Name string
	Value cty.Value
	DeclRange hcl.Range
}

type Annotations []*Annotation

func decodeAnnotationsBlock(block *hcl.Block) (Annotations,hcl.Diagnostics){
	var annotaions Annotations
	attrs,diags := block.Body.JustAttributes()

	for _,attr := range attrs {
		value,valDiag :=  attr.Expr.Value(createContext())
		diags = append(diags,valDiag...)
		if val,err :=convert.Convert(value,cty.String);err==nil {
			annotaions = append(annotaions, &Annotation{
				Name:attr.Name,
				Value: val,
				DeclRange: attr.NameRange,
			})
		} else {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Annotations must be string",
				Detail:   fmt.Sprintf("Couldn't convert value to string, this is value of type: %s",value.Type().FriendlyName()),
				Subject:  attr.Expr.Range().Ptr(),
			})
		}
	}
	


	return annotaions,diags
}