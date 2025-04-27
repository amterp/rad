package core

import (
	"fmt"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslTypeVisitorAcceptor interface {
	Accept(visitor *RslTypeVisitor)
}

// RslTypeVisitor is my best attempt at trying to work around the fact that Go
// doesn't have an exhaustive switch on sealed types/enums. This visitor is
// verbose, but should make it easy to update existing code when adding
// a new type to RSL.
type RslTypeVisitor struct {
	i    *Interpreter
	node *ts.Node

	visitBool    func(RslValue, bool)
	visitInt     func(RslValue, int64)
	visitFloat   func(RslValue, float64)
	visitString  func(RslValue, RslString)
	visitList    func(RslValue, *RslList)
	visitMap     func(RslValue, *RslMap)
	visitFn      func(RslValue, RslFn)
	defaultVisit func(RslValue)
}

func NewTypeVisitor(i *Interpreter, node *ts.Node) *RslTypeVisitor {
	return &RslTypeVisitor{
		i:    i,
		node: node,
	}
}

func NewTypeVisitorUnsafe() *RslTypeVisitor {
	return &RslTypeVisitor{}
}

func (v *RslTypeVisitor) UnhandledTypeError(val RslValue) {
	if v.i == nil {
		panic(fmt.Sprintf("Bug! Unhandled type: %T", val.Type()))
	}
	v.i.errorf(v.node, "Unsupported type: %s", TypeAsString(val))
}

func (v *RslTypeVisitor) ForBool(handler func(RslValue, bool)) *RslTypeVisitor {
	v.visitBool = handler
	return v
}

func (v *RslTypeVisitor) ForInt(handler func(RslValue, int64)) *RslTypeVisitor {
	v.visitInt = handler
	return v
}

func (v *RslTypeVisitor) ForFloat(handler func(RslValue, float64)) *RslTypeVisitor {
	v.visitFloat = handler
	return v
}

func (v *RslTypeVisitor) ForString(handler func(RslValue, RslString)) *RslTypeVisitor {
	v.visitString = handler
	return v
}

func (v *RslTypeVisitor) ForList(handler func(RslValue, *RslList)) *RslTypeVisitor {
	v.visitList = handler
	return v
}

func (v *RslTypeVisitor) ForMap(handler func(RslValue, *RslMap)) *RslTypeVisitor {
	v.visitMap = handler
	return v
}

func (v *RslTypeVisitor) ForFn(handler func(RslValue, RslFn)) *RslTypeVisitor {
	v.visitFn = handler
	return v
}

func (v *RslTypeVisitor) ForDefault(handler func(RslValue)) *RslTypeVisitor {
	v.defaultVisit = handler
	return v
}

func (v *RslTypeVisitor) Visit(acceptor RslTypeVisitorAcceptor) {
	acceptor.Accept(v)
}
