package core

import ts "github.com/tree-sitter/go-tree-sitter"

func GetSrc(src string, node *ts.Node) string {
	return src[node.StartByte():node.EndByte()]
}
