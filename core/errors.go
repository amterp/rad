package core

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

func ErrIndexOutOfBounds(i *Interpreter, node *ts.Node, idx int64, length int64) {
	i.errorf(node, "Index out of bounds: %d (length %d)", idx, length)
}

type RadPanic struct {
	ErrV        RadValue
	ShellResult *shellResult // For shell command errors, contains exit code/stdout/stderr
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

func (p *RadPanic) Err() *RadError {
	err, _ := p.ErrV.Val.(*RadError)
	return err
}

func (p *RadPanic) Panic() {
	panic(p)
}
