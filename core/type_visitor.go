package core

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"
)

type RadTypeVisitorAcceptor interface {
	Accept(visitor *RadTypeVisitor)
}

// RadTypeVisitor is my best attempt at trying to work around the fact that Go
// doesn't have an exhaustive switch on sealed types/enums. This visitor is
// verbose, but should make it easy to update existing code when adding
// a new type to Rad.
type RadTypeVisitor struct {
	i    *Interpreter
	node rl.Node

	visitBool    func(RadValue, bool)
	visitInt     func(RadValue, int64)
	visitFloat   func(RadValue, float64)
	visitString  func(RadValue, RadString)
	visitList    func(RadValue, *RadList)
	visitMap     func(RadValue, *RadMap)
	visitFn      func(RadValue, RadFn)
	visitNull    func(RadValue, RadNull)
	visitError   func(RadValue, *RadError)
	defaultVisit func(RadValue)
}

func NewTypeVisitor(i *Interpreter, node rl.Node) *RadTypeVisitor {
	return &RadTypeVisitor{
		i:    i,
		node: node,
	}
}

func NewTypeVisitorUnsafe() *RadTypeVisitor {
	return &RadTypeVisitor{}
}

func (v *RadTypeVisitor) UnhandledTypeError(val RadValue) {
	if v.i == nil {
		panic(fmt.Sprintf("Bug! Unhandled type: %T", val.Type()))
	}
	v.i.emitErrorf(rl.ErrTypeMismatch, v.node, "Unsupported type: %s", TypeAsString(val))
}

func (v *RadTypeVisitor) ForBool(handler func(RadValue, bool)) *RadTypeVisitor {
	v.visitBool = handler
	return v
}

func (v *RadTypeVisitor) ForInt(handler func(RadValue, int64)) *RadTypeVisitor {
	v.visitInt = handler
	return v
}

func (v *RadTypeVisitor) ForFloat(handler func(RadValue, float64)) *RadTypeVisitor {
	v.visitFloat = handler
	return v
}

func (v *RadTypeVisitor) ForString(handler func(RadValue, RadString)) *RadTypeVisitor {
	v.visitString = handler
	return v
}

func (v *RadTypeVisitor) ForList(handler func(RadValue, *RadList)) *RadTypeVisitor {
	v.visitList = handler
	return v
}

func (v *RadTypeVisitor) ForMap(handler func(RadValue, *RadMap)) *RadTypeVisitor {
	v.visitMap = handler
	return v
}

func (v *RadTypeVisitor) ForNull(handler func(RadValue, RadNull)) *RadTypeVisitor {
	v.visitNull = handler
	return v
}

func (v *RadTypeVisitor) ForFn(handler func(RadValue, RadFn)) *RadTypeVisitor {
	v.visitFn = handler
	return v
}

func (v *RadTypeVisitor) ForError(handler func(RadValue, *RadError)) *RadTypeVisitor {
	v.visitError = handler
	return v
}

func (v *RadTypeVisitor) ForDefault(handler func(RadValue)) *RadTypeVisitor {
	v.defaultVisit = handler
	return v
}

func (v *RadTypeVisitor) Visit(acceptor RadTypeVisitorAcceptor) {
	acceptor.Accept(v)
}
