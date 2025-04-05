package core

import ts "github.com/tree-sitter/go-tree-sitter"

type RslTypeVisitor struct {
	i    *Interpreter
	node *ts.Node

	VisitBool   func(RslValue, bool)
	VisitInt    func(RslValue, int64)
	VisitFloat  func(RslValue, float64)
	VisitString func(RslValue, RslString)
	VisitList   func(RslValue, *RslList)
	VisitMap    func(RslValue, *RslMap)
	Default     func(RslValue)
}

func NewTypeVisitor(i *Interpreter, node *ts.Node) *RslTypeVisitor {
	return &RslTypeVisitor{
		i:    i,
		node: node,
	}
}

func (v *RslTypeVisitor) UnhandledTypeError(val RslValue) {
	v.i.errorf(v.node, "Unsupported type: %s", TypeAsString(val))
}
