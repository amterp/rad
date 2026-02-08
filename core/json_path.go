package core

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type JsonFieldVar struct {
	Name string
	Path JsonPath
	Node *ts.Node
}

type JsonPath struct {
	Segments []JsonPathSegment
}

type JsonPathSegment struct {
	Identifier  string
	SegmentNode *ts.Node
	IdxSegments []JsonPathSegmentIdx
}

type JsonPathSegmentIdx struct {
	IdxNode *ts.Node  // e.g. json.names[]
	Idx     *RadValue // e.g. json.names[0]
}

func NewJsonFieldVar(i *Interpreter, leftNode, jsonPathNode *ts.Node) *JsonFieldVar {
	indexingNodes := rl.GetChildren(leftNode, rl.F_INDEXING, i.cursor)
	if len(indexingNodes) != 0 {
		i.emitError(rl.ErrInvalidSyntax, leftNode, "Json paths must be defined to plain identifiers")
	}
	leftIdentifierNode := rl.GetChild(leftNode, rl.F_ROOT)

	var segments []JsonPathSegment

	segmentNodes := rl.GetChildren(jsonPathNode, rl.F_SEGMENT, i.cursor)
	for _, segmentNode := range segmentNodes {
		identifierNode := rl.GetChild(&segmentNode, rl.F_KEY)
		identifierStr := i.GetSrcForNode(identifierNode)
		indexNodes := rl.GetChildren(&segmentNode, rl.F_INDEX, i.cursor)

		var idxSegments []JsonPathSegmentIdx
		for _, indexNode := range indexNodes {
			idxExprNode := rl.GetChild(&indexNode, rl.F_EXPR)
			if idxExprNode == nil {
				idxSegments = append(idxSegments, JsonPathSegmentIdx{IdxNode: &indexNode})
			} else {
				idx := i.eval(idxExprNode).Val
				idx.RequireType(i, idxExprNode, fmt.Sprintf("Json path indexes must be ints, was %s", TypeAsString(idx)), rl.RadIntT)
				idxSegments = append(idxSegments, JsonPathSegmentIdx{IdxNode: &indexNode, Idx: &idx})
			}
		}

		segments = append(
			segments,
			JsonPathSegment{Identifier: identifierStr, SegmentNode: &segmentNode, IdxSegments: idxSegments},
		)
	}

	identifierStr := i.GetSrcForNode(leftIdentifierNode)
	return &JsonFieldVar{
		Name: identifierStr,
		Path: JsonPath{Segments: segments},
	}
}
