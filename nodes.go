package rts

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

type Shebang struct {
	BaseNode
}

func newShebang(src string, node *ts.Node) (*Shebang, bool) {
	return &Shebang{
		BaseNode: newBaseNode(src, node),
	}, true
}

type FileHeader struct {
	BaseNode
	Contents string
}

func newFileHeader(src string, node *ts.Node) (*FileHeader, bool) {
	contentsNode := node.ChildByFieldName("contents")
	if contentsNode == nil {
		// would be strange`
		return nil, false
	}

	return &FileHeader{
		BaseNode: newBaseNode(src, node),
		Contents: src[contentsNode.StartByte():contentsNode.EndByte()],
	}, true
}

type StringNode struct {
	BaseNode
	RawLexeme string // Literal src, excluding delimiters, ws, comments, etc
}

func newStringNode(src string, node *ts.Node) (*StringNode, bool) {
	start := node.ChildByFieldName("start")
	end := node.ChildByFieldName("end")
	contentStart := start.EndByte()
	contentEnd := end.StartByte()
	return &StringNode{
		BaseNode:  newBaseNode(src, node),
		RawLexeme: src[contentStart:contentEnd],
	}, true
}
