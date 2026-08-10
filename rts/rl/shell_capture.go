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
