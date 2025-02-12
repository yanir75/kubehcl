package main

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
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
		annotaions = append(annotaions, &Annotation{
			Name:attr.Name,
			Value: value,
			DeclRange: attr.NameRange,
		})
	}
	


	return annotaions,diags
}