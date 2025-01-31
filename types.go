package rts

import ts "github.com/tree-sitter/go-tree-sitter"

type Position struct {
	Row int
	Col int
}

func NewPosition(p ts.Point) Position {
	return Position{Row: int(p.Row), Col: int(p.Column)}
}

type BaseNode struct {
	Src string
	// Indexes in the original source code.
	StartByte int
	EndByte   int // exclusive
	StartPos  Position
	EndPos    Position
}

type Shebang struct {
	BaseNode
}

type FileHeader struct {
	BaseNode
	Contents string
}
