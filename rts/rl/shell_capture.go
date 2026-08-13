package rl

// Shell capture rules - which stream each assignment target receives, and
// which streams Rad therefore has to capture rather than pass through to the
// terminal.
//
// This lives in rl, next to the Shell node, because two callers need to agree
// exactly: the interpreter (core) binds the values, and the checker
// (rts/check) types the targets and warns about misleading ones. A checker
// that disagreed with the interpreter would report confident, wrong types -
// worse than not typing shell targets at all. core imports rl, and rts must
// not import core, so rl is the only place both can reach.

// Shell capture target names that carry meaning. When *every* target uses one
// of these, assignment is by name and order stops mattering.
const (
	ShellCaptureCode   = "code"
	ShellCaptureStdout = "stdout"
	ShellCaptureStderr = "stderr"
)

// ShellStream identifies which of a command's three results a target receives.
type ShellStream int

const (
	ShellStdout ShellStream = iota
	ShellStderr
	ShellCode
	// ShellStreamNone is for a named-mode target Rad doesn't recognize. It
	// can't happen for a well-formed named assignment (the mode requires all
	// names be recognized) and exists so callers get a total function.
	ShellStreamNone
)

// shellPositionalOrder is the order positional targets are filled:
// (stdout, stderr, code). Most-wanted first. The exit code comes last because
// Rad already promotes a non-zero exit to an error, so `catch:` covers what
// the code is usually read for - and because it needs no capture, so it costs
// nothing to leave until the end.
var shellPositionalOrder = [3]ShellStream{ShellStdout, ShellStderr, ShellCode}

// IsNamedShellAssignment reports whether ALL targets are named exactly "code",
// "stdout", or "stderr". Any other name (including "_") drops the whole
// statement to positional assignment.
func IsNamedShellAssignment(targets []Node) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		switch ShellTargetRootName(target) {
		case ShellCaptureCode, ShellCaptureStdout, ShellCaptureStderr:
		default:
			return false
		}
	}
	return true
}

// ShellStreamForTarget returns the stream the target at index i receives, given
// the statement's total target count and whether it assigns by name.
func ShellStreamForTarget(targets []Node, i int, isNamed bool) ShellStream {
	if i < 0 || i >= len(targets) {
		return ShellStreamNone
	}
	if !isNamed {
		if i >= len(shellPositionalOrder) {
			return ShellStreamNone
		}
		return shellPositionalOrder[i]
	}
	switch ShellTargetRootName(targets[i]) {
	case ShellCaptureStdout:
		return ShellStdout
	case ShellCaptureStderr:
		return ShellStderr
	case ShellCaptureCode:
		return ShellCode
	}
	return ShellStreamNone
}

// ShellPositionalCaptures reports which streams a positional assignment with
// numTargets targets has to capture. Positional targets are filled in
// (stdout, stderr, code) order, so the target count is the stream count - the
// exit code rides along free in the third slot.
//
// A stream that isn't captured goes to the terminal, which makes this the
// silencing dial too: `_ = $cmd` swallows stdout, `_, _ = $cmd` swallows both.
func ShellPositionalCaptures(numTargets int) (captureStdout, captureStderr bool) {
	return numTargets >= 1, numTargets >= 2
}

// ShellNamedCaptures reports which streams a named assignment has to capture.
// The exit code is always available, so only stdout/stderr matter.
func ShellNamedCaptures(targets []Node) (captureStdout, captureStderr bool) {
	for _, target := range targets {
		switch ShellTargetRootName(target) {
		case ShellCaptureStdout:
			captureStdout = true
		case ShellCaptureStderr:
			captureStderr = true
		}
	}
	return
}

// ShellCaptures reports which streams the statement has to capture, picking
// the named or positional rule as appropriate.
func ShellCaptures(targets []Node) (captureStdout, captureStderr bool) {
	if IsNamedShellAssignment(targets) {
		return ShellNamedCaptures(targets)
	}
	return ShellPositionalCaptures(len(targets))
}

// ShellTargetRootName gets the identifier name from a VarPath or Identifier
// assignment target. Returns "" for any other shape (which is a syntax error
// caught elsewhere).
func ShellTargetRootName(node Node) string {
	switch n := node.(type) {
	case *VarPath:
		if ident, ok := n.Root.(*Identifier); ok {
			return ident.Name
		}
	case *Identifier:
		return n.Name
	}
	return ""
}

// ShellAccessor is the virtual member an inline invocation reads:
// `$`cmd`.stdout`. The four names match the capture-target vocabulary above, so
// there is one thing to learn rather than two.
//
// A superset of ShellStream, and deliberately NOT a member of it. `ok` is a
// predicate over the exit code, not a fourth stream: it never affects what gets
// captured, and it has no positional slot. Folding it into ShellStream would
// silently invalidate shellPositionalOrder, which is a [3], and oldOrderStream
// in rts/check, which is a 3-cycle - neither of which the compiler would catch.
// If you are here to unify them: don't.
type ShellAccessor int

const (
	// ShellAccessorNone is the zero value on purpose: a forgotten accessor is
	// an error to be reported, never a silent default to stdout.
	ShellAccessorNone ShellAccessor = iota
	ShellAccessorStdout
	ShellAccessorStderr
	ShellAccessorCode
	ShellAccessorOk
)

// ShellAccessorOk is spelled out here rather than in the capture constants
// because it has no assignment-target counterpart - there is no `ok = $cmd`.
const ShellAccessorOkName = "ok"

// ShellAccessorNames lists the accessors in the order diagnostics should offer
// them: the two that yield output first, then the two that describe the outcome.
var ShellAccessorNames = []string{
	ShellCaptureStdout,
	ShellCaptureStderr,
	ShellCaptureCode,
	ShellAccessorOkName,
}

// ParseShellAccessor resolves an accessor name, reporting whether it is one.
func ParseShellAccessor(name string) (ShellAccessor, bool) {
	switch name {
	case ShellCaptureStdout:
		return ShellAccessorStdout, true
	case ShellCaptureStderr:
		return ShellAccessorStderr, true
	case ShellCaptureCode:
		return ShellAccessorCode, true
	case ShellAccessorOkName:
		return ShellAccessorOk, true
	default:
		return ShellAccessorNone, false
	}
}

// Name returns the accessor as written in source.
func (a ShellAccessor) Name() string {
	switch a {
	case ShellAccessorStdout:
		return ShellCaptureStdout
	case ShellAccessorStderr:
		return ShellCaptureStderr
	case ShellAccessorCode:
		return ShellCaptureCode
	case ShellAccessorOk:
		return ShellAccessorOkName
	default:
		return ""
	}
}

// ShellExprCaptures reports which streams an inline invocation must capture.
// Capture follows the accessor, exactly as it follows the target names in a
// named assignment: read stdout and stderr still reaches the terminal, read
// stderr and stdout does. Asking only about the outcome captures nothing, so
// both streams pass through - `code` as a target behaves the same way.
func ShellExprCaptures(a ShellAccessor) (captureStdout, captureStderr bool) {
	switch a {
	case ShellAccessorStdout:
		return true, false
	case ShellAccessorStderr:
		return false, true
	default:
		return false, false
	}
}

// Propagates reports whether a non-zero exit raises for this accessor.
//
// The output accessors do: you asked for what the command produced, and a
// command that failed did not produce it, so handing back a partial or empty
// string would be answering a question that has no answer.
//
// `.code` and `.ok` never do. Asking about the outcome *is* handling the
// failure - grep exiting 1 is data. Raising there would make the one check
// these accessors exist for, `if not $`which docker`.ok:`, impossible to write.
func (a ShellAccessor) Propagates() bool {
	return a == ShellAccessorStdout || a == ShellAccessorStderr
}
