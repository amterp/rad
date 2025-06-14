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
	indexingNodes := i.getChildren(leftNode, rl.F_INDEXING)
	if len(indexingNodes) != 0 {
		i.errorf(leftNode, "Json paths must be defined to plain identifiers")
	}
	leftIdentifierNode := i.getChild(leftNode, rl.F_ROOT)

	var segments []JsonPathSegment

	segmentNodes := i.getChildren(jsonPathNode, rl.F_SEGMENT)
	for _, segmentNode := range segmentNodes {
		identifierNode := i.getChild(&segmentNode, rl.F_KEY)
		identifierStr := i.sd.Src[identifierNode.StartByte():identifierNode.EndByte()]
		indexNodes := i.getChildren(&segmentNode, rl.F_INDEX)

		var idxSegments []JsonPathSegmentIdx
		for _, indexNode := range indexNodes {
			idxExprNode := i.getChild(&indexNode, rl.F_EXPR)
			if idxExprNode == nil {
				idxSegments = append(idxSegments, JsonPathSegmentIdx{IdxNode: &indexNode})
			} else {
				idx := i.eval(idxExprNode).Val
				idx.RequireType(i, idxExprNode, fmt.Sprintf("Json path indexes must be ints, was %s", TypeAsString(idx)), RadIntT)
				idxSegments = append(idxSegments, JsonPathSegmentIdx{IdxNode: &indexNode, Idx: &idx})
			}
		}

		segments = append(
			segments,
			JsonPathSegment{Identifier: identifierStr, SegmentNode: &segmentNode, IdxSegments: idxSegments},
		)
	}

	identifierStr := i.sd.Src[leftIdentifierNode.StartByte():leftIdentifierNode.EndByte()]
	return &JsonFieldVar{
		Name: identifierStr,
		Path: JsonPath{Segments: segments},
	}
}
