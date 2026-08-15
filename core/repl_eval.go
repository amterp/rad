package core

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// replScriptName is what diagnostics call the buffer you just typed.
const replScriptName = "<repl>"

// EvalReplTurn runs one turn's worth of source against the live session.
//
// A turn is the REPL's unit of execution - its mini-script - so it goes through
// the same top-level path a real script does rather than walking statements
// itself. That is what gives it function hoisting (without which a named fn
// typed at the prompt never registers), a signal checkpoint between statements
// (without which Ctrl+C cannot interrupt anything), and deferred blocks.
//
// It returns a non-nil abort when the turn ended early: an error already
// rendered, a signal, or an explicit exit(). The caller decides which of those
// ends the session.
func (i *Interpreter) EvalReplTurn(src string, root *rl.SourceFile, tree *rts.RadTree) (out RadValue, abort *RadAbort) {
	// Diagnostics render the source they point into, so the interpreter has to
	// be looking at this turn's text while the turn runs.
	original := i.sd
	i.sd = &ScriptData{
		ScriptName:        replScriptName,
		Tree:              tree,
		Src:               src,
		DisableGlobalOpts: true,
		DisableArgsBlock:  true,
	}
	defer func() { i.sd = original }()

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		a, ok := r.(*RadAbort)
		if !ok {
			panic(r)
		}
		abort = a
	}()

	if diag := i.replRejectScriptOnlyBlocks(root); diag {
		return VOID_SENTINEL, nil
	}

	res := i.safelyExecuteTopLevel(root)
	i.runTurnDefers(0)

	switch res.Ctrl {
	case CtrlNormal:
		return res.Val, nil
	case CtrlReturn:
		i.emitError(rl.ErrReturnOutsideFunction, root, "Cannot 'return' outside of a function")
	case CtrlYield:
		i.emitError(rl.ErrYieldOutsideFunction, root, "Cannot 'yield' outside of a function")
	}
	return VOID_SENTINEL, nil
}

// runTurnDefers runs and clears the blocks this turn registered. Scoping them to
// the turn rather than the session is what makes `defer` demonstrable at a
// prompt: deferring to session exit would mean a defer you type does nothing
// observable for as long as you keep the session open.
func (i *Interpreter) runTurnDefers(errCode int) {
	if len(i.deferBlocks) == 0 {
		return
	}
	i.executeDeferBlocks(errCode)
	i.deferBlocks = nil
}

// replRejectScriptOnlyBlocks refuses the constructs that describe a script's
// command-line interface. They are fields on the AST root rather than
// statements, so evaluation never reaches them - meaning that without this they
// are silently ignored, which is a worse answer than an error.
func (i *Interpreter) replRejectScriptOnlyBlocks(root *rl.SourceFile) bool {
	var node rl.Node
	var what string

	switch {
	case root.Args != nil:
		node, what = root.Args, "An args block"
	case len(root.Cmds) > 0:
		node, what = root.Cmds[0], "A command block"
	case root.Header != nil:
		node, what = root.Header, "A file header"
	default:
		return false
	}

	i.emitErrorWithHint(
		rl.ErrReplUnsupportedConstruct,
		node,
		what+" describes a script's interface, which the REPL has no use for.",
		"Put it in a file and load that with ':load <file>'.",
	)
	return true
}
