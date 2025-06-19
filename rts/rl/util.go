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
