package rts

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

type Position struct {
	Row int
	Col int
}

func NewPosition(p ts.Point) Position {
	return Position{Row: int(p.Row), Col: int(p.Column)}
}

type Node interface {
	Src() string
	// Indexes in the original source code.
	StartByte() int
	EndByte() int // exclusive
	StartPos() Position
	EndPos() Position
}

func NodeName[T Node]() interface{} {
	var zero T
	switch any(zero).(type) {
	case *Shebang:
		return "shebang"
	case *FileHeader:
		return "file_header"
	case *StringNode:
		return "string"
	default:
		return ""
	}
}

type BaseNode struct {
	src       string
	startByte int
	endByte   int
	startPos  Position
	endPos    Position
}

func newBaseNode(src string, node *ts.Node) BaseNode {
	return BaseNode{
		src:       src[node.StartByte():node.EndByte()],
		startByte: int(node.StartByte()),
		endByte:   int(node.EndByte()),
		startPos:  NewPosition(node.StartPosition()),
		endPos:    NewPosition(node.EndPosition()),
	}
}

func (n *BaseNode) Src() string {
	return n.src
}

func (n *BaseNode) StartByte() int {
	return n.startByte
}

func (n *BaseNode) EndByte() int {
	return n.endByte
}

func (n *BaseNode) StartPos() Position {
	return n.startPos
}

func (n *BaseNode) EndPos() Position {
	return n.endPos
}

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
