package core

import (
	"fmt"

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
	Idx     *RslValue // e.g. json.names[0]
}

func NewJsonFieldVar(i *Interpreter, assignNode, jsonPathNode *ts.Node) *JsonFieldVar {
	leftVarPathNodes := i.getChildren(assignNode, F_LEFT)

	if len(leftVarPathNodes) != 1 {
		i.errorf(assignNode, "Expected exactly 1 assignment with JSON path")
	}
	leftNode := leftVarPathNodes[0]
	if leftNode.ChildCount() != 1 {
		i.errorf(&leftNode, "Left side of a json path assignment must be only an identifier.")
	}

	leftIdentifierNode := i.getOnlyChild(&leftNode)

	var segments []JsonPathSegment

	segmentNodes := i.getChildren(jsonPathNode, F_SEGMENT)
	for _, segmentNode := range segmentNodes {
		identifierNode := i.getChild(&segmentNode, F_KEY)
		identifierStr := i.sd.Src[identifierNode.StartByte():identifierNode.EndByte()]
		indexNodes := i.getChildren(&segmentNode, F_INDEX)

		var idxSegments []JsonPathSegmentIdx
		for _, indexNode := range indexNodes {
			idxExprNode := i.getChild(&indexNode, F_EXPR)
			if idxExprNode == nil {
				idxSegments = append(idxSegments, JsonPathSegmentIdx{IdxNode: &indexNode})
			} else {
				idx := i.evaluate(idxExprNode, 1)[0]
				idx.RequireType(i, idxExprNode, fmt.Sprintf("Json path indexes must be ints, was %s", TypeAsString(idx)), RslIntT)
				idxSegments = append(idxSegments, JsonPathSegmentIdx{IdxNode: &indexNode, Idx: &idx})
			}
		}

		segments = append(segments, JsonPathSegment{Identifier: identifierStr, SegmentNode: &segmentNode, IdxSegments: idxSegments})
	}

	identifierStr := i.sd.Src[leftIdentifierNode.StartByte():leftIdentifierNode.EndByte()]
	return &JsonFieldVar{
		Name: identifierStr,
		Path: JsonPath{Segments: segments},
	}
}
