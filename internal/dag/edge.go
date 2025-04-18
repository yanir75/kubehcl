/*
// SPDX-License-Identifier: MPL-2.0
This file was copied from https://github.com/opentofu/opentofu and retains its' original license: https://www.mozilla.org/en-US/MPL/2.0/
*/
// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dag

// Edge represents an edge in the graph, with a source and target vertex.
type Edge interface {
	Source() Vertex
	Target() Vertex

	Hashable
}

// BasicEdge returns an Edge implementation that simply tracks the source
// and target given as-is.
func BasicEdge(source, target Vertex) Edge {
	return &basicEdge{S: source, T: target}
}

// basicEdge is a basic implementation of Edge that has the source and
// target vertex.
type basicEdge struct {
	S, T Vertex
}

func (e *basicEdge) Hashcode() interface{} {
	return [...]interface{}{e.S, e.T}
}

func (e *basicEdge) Source() Vertex {
	return e.S
}

func (e *basicEdge) Target() Vertex {
	return e.T
}
