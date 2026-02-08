package core

import (
	"github.com/amterp/rad/rts/rl"
)

func ErrIndexOutOfBounds(i *Interpreter, node rl.Node, idx int64, length int64) {
	// Use panic so fallback operator (??) can catch this error
	span := nodeSpanPtr(node)
	errVal := newRadValue(i, node, NewErrorStrf("Index out of bounds: %d (length %d)", idx, length).SetCode(rl.ErrIndexOutOfBounds).SetSpan(span))
	i.NewRadPanic(node, errVal).Panic()
}

type RadPanic struct {
	ErrV        RadValue
	ShellResult *shellResult // For shell command errors, contains exit code/stdout/stderr
}

func (i *Interpreter) NewRadPanic(node rl.Node, err RadValue) *RadPanic {
	unwrapped := err.RequireError(i, node)
	if unwrapped.Span == nil {
		unwrapped.Span = nodeSpanPtr(node)
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

// nodeSpanPtr returns a pointer to the node's span, or nil if node is nil.
func nodeSpanPtr(node rl.Node) *rl.Span {
	if node == nil {
		return nil
	}
	s := node.Span()
	return &s
}
