package rl

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

func GetChildren(node *ts.Node, fieldName string) []ts.Node {
	return node.ChildrenByFieldName(fieldName, node.Walk())
}

func GetChild(node *ts.Node, fieldName string) *ts.Node {
	return node.ChildByFieldName(fieldName)
}

func GetSrc(node *ts.Node, src string) string {
	return src[node.StartByte():node.EndByte()]
}

// spanFromNode builds a Span from a tree-sitter node within the rl
// package, where no file path is available - the typing resolver runs
// before/independently of the converter and doesn't have file context.
// The file field is left blank; consumers that need it (currently none
// on the read path) can fill it in from the surrounding source unit.
func spanFromNode(node *ts.Node) Span {
	if node == nil {
		return Span{}
	}
	start := node.StartPosition()
	end := node.EndPosition()
	return Span{
		StartByte: int(node.StartByte()),
		EndByte:   int(node.EndByte()),
		StartRow:  int(start.Row),
		StartCol:  int(start.Column),
		EndRow:    int(end.Row),
		EndCol:    int(end.Column),
	}
}
