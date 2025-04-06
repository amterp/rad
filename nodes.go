package rts

import (
	"github.com/amterp/rts/rsl"
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
	start := node.ChildByFieldName(rsl.F_START)
	end := node.ChildByFieldName(rsl.F_END)
	contentStart := start.EndByte()
	contentEnd := end.StartByte()
	return &StringNode{
		BaseNode:  newBaseNode(src, node),
		RawLexeme: src[contentStart:contentEnd],
	}, true
}

type CallNode struct {
	BaseNode
	Name     string
	NameNode *ts.Node
}

func newCallNode(node *ts.Node, completeSrc string) (*CallNode, bool) {
	nameNode := node.ChildByFieldName(rsl.F_FUNC)
	if nameNode == nil {
		return nil, false
	}

	name := completeSrc[nameNode.StartByte():nameNode.EndByte()]
	return &CallNode{
		BaseNode: newBaseNode(completeSrc, node),
		Name:     name,
		NameNode: nameNode,
	}, true
}
