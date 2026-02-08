package core

import (
	"github.com/amterp/rad/rts/rl"
)

type JsonFieldVar struct {
	Name string
	Path JsonPath
	Span *rl.Span // source location for error reporting
}

type JsonPath struct {
	Segments []JsonPathSegment
}

type JsonPathSegment struct {
	Identifier  string
	SegmentSpan rl.Span
	IdxSegments []JsonPathSegmentIdx
}

type JsonPathSegmentIdx struct {
	Span rl.Span
	Idx  *RadValue // e.g. json.names[0]; nil = wildcard []
}

// NewJsonFieldVarSimple creates a JsonFieldVar with a simple single-segment path.
// Used by the AST-based rad block interpreter for plain field names.
func NewJsonFieldVarSimple(name string, span rl.Span) *JsonFieldVar {
	return &JsonFieldVar{
		Name: name,
		Path: JsonPath{
			Segments: []JsonPathSegment{
				{Identifier: name, SegmentSpan: span},
			},
		},
		Span: &span,
	}
}

// NewJsonFieldVar creates a JsonFieldVar from explicit path segments.
func NewJsonFieldVar(i *Interpreter, name string, span rl.Span, segments []JsonPathSegment) *JsonFieldVar {
	_ = i // reserved for future validation
	return &JsonFieldVar{
		Name: name,
		Path: JsonPath{Segments: segments},
		Span: &span,
	}
}
