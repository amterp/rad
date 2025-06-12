package core

import "fmt"

type ExpectedOutput int

const (
	Zero ExpectedOutput = iota
	One
	NoConstraint
)

func (e ExpectedOutput) String() string {
	switch e {
	case Zero:
		return "no output"
	case One:
		return "1 output"
	case NoConstraint:
		return "output or no output"
	default:
		panic(fmt.Sprintf("Bug! Unhandled value: %d", e))
	}
}

func (e ExpectedOutput) Acceptable(actual int) bool {
	switch e {
	case Zero:
		return actual == 0
	case One:
		return actual == 1
	case NoConstraint:
		return true
	default:
		panic(fmt.Sprintf("Bug! Unhandled value: %d", e))
	}
}
