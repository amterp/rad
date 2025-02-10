package core

import ts "github.com/tree-sitter/go-tree-sitter"

func ErrIndexOutOfBounds(i *Interpreter, idxNode *ts.Node, idx int64, length int64) {
	i.errorf(idxNode, "Index out of bounds: %d (length %d)", idx, length)
}

func (i *Interpreter) CheckForErrors() {
	var errorNodes []*ts.Node
	i.errorCheck(i.sd.Tree.Root(), &errorNodes)
	if len(errorNodes) > 0 {
		for _, node := range errorNodes {
			// TODO print all errors up front instead of exiting here
			i.errorf(node, "Invalid syntax")
		}
	}
}

func (i *Interpreter) errorCheck(node *ts.Node, errorNodes *[]*ts.Node) {
	if node.IsError() || node.IsMissing() {
		*errorNodes = append(*errorNodes, node)
	}
	childrenNodes := node.Children(node.Walk())
	for _, child := range childrenNodes {
		i.errorCheck(&child, errorNodes)
	}
}
