package core

import (
	"github.com/amterp/rad/rts/raderr"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadError struct {
	Node            *ts.Node
	msg             RadString
	Code            raderr.Error
	ShouldPropagate bool
}

func NewError(msg RadString) *RadError {
	return &RadError{
		msg:             msg,
		ShouldPropagate: true,
	}
}

func NewErrorStr(msg string) *RadError { // todo make a constructor forcing a Rad error code
	return &RadError{
		msg:             NewRadString(msg),
		ShouldPropagate: true,
	}
}

func (e *RadError) SetCode(code raderr.Error) *RadError {
	e.Code = code
	return e
}

func (e *RadError) SetShouldPropagate(shouldPropagate bool) *RadError {
	e.ShouldPropagate = shouldPropagate
	return e
}

func (e *RadError) Msg() RadString {
	return e.msg
}

func (e *RadError) Equals(other *RadError) bool {
	return e.Msg().Equals(other.Msg())
}

func (e *RadError) Hash() string {
	return e.Msg().Plain()
}
