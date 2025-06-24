package rl

import ts "github.com/tree-sitter/go-tree-sitter"

type RadNode struct {
	Node *ts.Node
	Src  string
}

func NewRadNode(node *ts.Node, src string) *RadNode {
	return &RadNode{
		Node: node,
		Src:  src,
	}
}
