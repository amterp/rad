package core

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"
)

type RadError struct {
	Span *rl.Span // source location where the error originated (nil if unknown)
	msg  RadString
	Code rl.Error
}

func NewError(msg RadString) *RadError {
	return &RadError{
		msg: msg,
	}
}

func NewErrorStrf(msg string, args ...interface{}) *RadError { // todo make a constructor forcing a Rad error code
	return &RadError{
		msg: NewRadString(fmt.Sprintf(msg, args...)),
	}
}

func (e *RadError) SetCode(code rl.Error) *RadError {
	e.Code = code
	return e
}

func (e *RadError) SetSpan(span *rl.Span) *RadError {
	e.Span = span
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
