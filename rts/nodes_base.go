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
	Node() *ts.Node
	CompleteSrc() string
	Src() string
	// Indexes in the original source code.
	// Zero indexed, so add +1 to get human readable values.
	// todo wrap in own Range object instead?
	StartByte() int
	EndByte() int // inclusive
	StartPos() Position
	EndPos() Position // inclusive
}

func NodeName[T Node]() string {
	var zero T
	switch any(zero).(type) {
	case *Shebang:
		return "shebang"
	case *FileHeader:
		return "file_header"
	case *ArgBlock:
		return "arg_block"
	case *CmdBlock:
		return "cmd_block"
	case *StringNode:
		return "string"
	default:
		return ""
	}
}

type BaseNode struct {
	node      *ts.Node
	src       string // complete src
	startByte int
	endByte   int
	startPos  Position
	endPos    Position
}

func newBaseNode(src string, node *ts.Node) BaseNode {
	return BaseNode{
		node:      node,
		src:       src,
		startByte: int(node.StartByte()),
		endByte:   int(node.EndByte()),
		startPos:  NewPosition(node.StartPosition()),
		endPos:    NewPosition(node.EndPosition()),
	}
}

func (n *BaseNode) Node() *ts.Node {
	return n.node
}

func (n *BaseNode) CompleteSrc() string {
	return n.src
}

func (n *BaseNode) Src() string {
	return n.src[n.startByte:n.endByte]
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
