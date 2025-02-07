package core

import ts "github.com/tree-sitter/go-tree-sitter"

func ErrIndexOutOfBounds(i *Interpreter, idxNode *ts.Node, idx int64, length int64) {
	i.errorf(idxNode, "Index out of bounds: %d (length %d)", idx, length)
}
