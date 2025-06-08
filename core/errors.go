package core

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

func ErrIndexOutOfBounds(i *Interpreter, node *ts.Node, idx int64, length int64) {
	i.errorf(node, "Index out of bounds: %d (length %d)", idx, length)
}

func (i *Interpreter) MaybePropagateError(node *ts.Node, err RadValue) {
	if err.IsErrorToPropagate() {
		panic(i.NewRadPanic(node, err))
	}
}

type RadPanic struct {
	ErrV RadValue
}

func (i *Interpreter) NewRadPanic(node *ts.Node, err RadValue) *RadPanic {
	unwrapped := err.RequireError(i, node)
	if unwrapped.Node == nil {
		unwrapped.Node = node
	}
	return &RadPanic{
		ErrV: err,
	}
}

func (p RadPanic) Err() *RadError {
	err, _ := p.ErrV.Val.(*RadError)
	return err
}
