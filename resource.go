package main

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

var resources ResourceList

type Resource struct {
	Name string
	Description string
	Default cty.Value
	Type cty.Type
	DeclRange hcl.Range
	HasValue bool // for checking if needed request from the user
}

type ResourceList []*Resource

var inputResourceBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "for_each", Required: false},
		{Name: "count", Required: false},
		{Name: "depends_on", Required: false},
	},
}
