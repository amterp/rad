package core

import (
	"strings"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// ReplSession is one interactive session: a live interpreter, the parser and
// tree it reuses each turn, and the input history.
type ReplSession struct {
	interpreter *Interpreter
	reader      replReader
	parser      *rts.RadParser
	tree        *rts.RadTree
	history     []string
}

// NewReplSession wires up a session. The parser and tree are held for the whole
// session and reparsed in place each turn - building either per turn leaks
// C-heap memory, since tree-sitter trees are not garbage collected.
func NewReplSession() (*ReplSession, error) {
	parser, err := rts.NewRadParser()
	if err != nil {
		return nil, err
	}

	interpreter := NewInterpreter(InterpreterInput{ScriptName: replScriptName})
	interpreter.InitBuiltIns()
	interpreter.RegisterWithExit()

	return &ReplSession{
		interpreter: interpreter,
		reader:      newReplReader(),
		parser:      parser,
	}, nil
}

// Run is the read-eval-print loop.
func (s *ReplSession) Run() error {
	// The session hosts execution rather than being it, so a fatal error
	// unwinds to the prompt instead of ending the process.
	RExit.SetUnwinding(true)
	defer RExit.SetUnwinding(false)

	// Signals only reach user code once dispatch is running; without this
	// Ctrl+C during a long statement would kill the session outright.
	s.interpreter.signals.Start()
	defer s.interpreter.signals.Stop()

	printReplBanner()

	for {
		src, outcome := s.reader.Read(s.history)
		switch outcome {
		case replEOF:
			RP.Print("\n")
			return nil
		case replDiscarded:
			continue
		}

		src = ReplTrimTrailingBlankLines(src)
		if strings.TrimSpace(src) == "" {
			continue
		}
		s.history = append(s.history, src)

		if isReplMeta(src) {
			if !s.runReplMeta(src) {
				return nil
			}
			continue
		}

		if !s.runSource(src) {
			return nil
		}
	}
}

// runSource parses and evaluates one turn. It returns false when the turn asked
// the session to end.
func (s *ReplSession) runSource(src string) (keepGoing bool) {
	// Each turn starts with the exit latch clear, or the first error would
	// leave every later turn skipping its deferred blocks.
	RExit.ResetExiting()

	root, ok := s.parse(src)
	if !ok {
		return true
	}

	value, abort := s.interpreter.EvalReplTurn(src, root, s.tree)
	if abort != nil {
		return s.handleAbort(abort)
	}

	if value != VOID_SENTINEL {
		RP.Printf("%s\n", ToPrintable(value))
	}
	return true
}

// parse reparses the session's tree in place and converts it, reporting syntax
// errors rather than letting the converter panic on them.
func (s *ReplSession) parse(src string) (*rl.SourceFile, bool) {
	if s.tree == nil {
		s.tree = s.parser.Parse(src)
	} else {
		s.tree.Update(s.parser, src)
	}

	if s.tree.HasInvalidNodes() {
		s.reportSyntaxErrors(src)
		return nil, false
	}
	return rts.ConvertCST(s.tree.Root(), src, replScriptName), true
}

// reportSyntaxErrors renders the same diagnostics a script would get for the
// same typo.
//
// It borrows them from the checker, but keeps only the syntax band. The rest of
// the checker is unusable here: it rebuilds its scope from the buffer alone, so
// every variable defined in an earlier turn would come back undefined. Syntax
// diagnostics don't depend on scope, so they carry over intact - and they name
// the offending text, which raw invalid-node spans do not.
func (s *ReplSession) reportSyntaxErrors(src string) {
	renderer := NewDiagnosticRenderer(RIo.StdErr)

	checker := check.NewCheckerWithTree(s.tree, s.parser, src, nil)
	if result, err := checker.Check(); err == nil {
		rendered := false
		for _, diag := range result.Diagnostics {
			if diag.Severity == check.Error && diag.Code != nil && strings.HasPrefix(string(*diag.Code), "1") {
				renderer.Render(NewDiagnosticFromCheck(diag, replScriptName))
				rendered = true
			}
		}
		if rendered {
			return
		}
	}

	// The checker had nothing to say about a tree we know is broken. Fall back
	// to pointing at the spans themselves rather than failing silently.
	for _, span := range s.tree.FindInvalidNodeSpans(replScriptName) {
		renderer.Render(NewDiagnostic(SeverityError, rl.ErrInvalidSyntax, "Invalid syntax", src, span))
	}
}

// handleAbort decides what an early-ended turn means for the session. An error
// has already printed itself, so the prompt simply comes back; an interrupt
// says so, because a statement vanishing silently reads like a crash.
func (s *ReplSession) handleAbort(abort *RadAbort) (keepGoing bool) {
	switch abort.Reason {
	case ExitError:
		return true
	case ExitSignal:
		RP.RadStderrf("Interrupted.\n")
		return true
	default:
		// exit() - the session is over, with the code the caller asked for.
		RExit.SetUnwinding(false)
		RExit.ResetExiting()
		RExit.Exit(abort.Code)
		return false
	}
}

// Shutdown releases the tree-sitter resources the session held open.
func (s *ReplSession) Shutdown() error {
	if s.tree != nil {
		s.tree.Close()
	}
	if s.parser != nil {
		s.parser.Close()
	}
	return s.reader.Close()
}
