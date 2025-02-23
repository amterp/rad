package core

import ts "github.com/tree-sitter/go-tree-sitter"

func ErrIndexOutOfBounds(i *Interpreter, idxNode *ts.Node, idx int64, length int64) {
	i.errorf(idxNode, "Index out of bounds: %d (length %d)", idx, length)
}

func (i *Interpreter) CheckForErrors() {
	invalidNodes := i.sd.Tree.FindInvalidNodes()
	if len(invalidNodes) > 0 {
		for _, node := range invalidNodes {
			// TODO print all errors up front instead of exiting here
			i.errorf(node, "Invalid syntax")
		}
	}
}
