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
		return "an output"
	case NoConstraint:
		return "output or no output"
	default:
		return fmt.Sprintf("Bug! Unhandled value: %d", e)
	}
}

var EXPECT_ONE_OUTPUT = NewEvalCtx(One)

type EvalCtx struct {
	ExpectedOutput ExpectedOutput
}

func NewEvalCtx(ExpectValue ExpectedOutput) EvalCtx {
	return EvalCtx{
		ExpectedOutput: ExpectValue,
	}
}
