package core

import ts "github.com/tree-sitter/go-tree-sitter"

func ErrIndexOutOfBounds(i *Interpreter, node *ts.Node, idx int64, length int64) {
	i.errorf(node, "Index out of bounds: %d (length %d)", idx, length)
}
