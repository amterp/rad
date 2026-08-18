package check

import (
	"fmt"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

// Shell invocations used as values.
//
// `$` binds only the command, so anything following it reads the *result*. The
// grammar deliberately accepts postfix it can't make sense of, because these
// shapes used to mean something else and a syntax error could neither say so
// nor offer the fix. Two diagnostics sort them out:
//
//   - RAD40025: an invocation reached expression position with no accessor.
//   - RAD40026: postfix was written, but it isn't one of the four accessors.
//
// Both are errors rather than warnings: every one of them is either code that
// silently did the wrong thing before, or code with no meaning at all.

func (c *RadCheckerImpl) addShellExprAccessorErrors(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		shellExpr, ok := node.(*rl.ShellExpr)
		if !ok || shellExpr.Accessor != rl.ShellAccessorNone {
			return
		}
		if shellExpr.BadAccessor == nil {
			*d = append(*d, c.bareShellExprDiagnostic(shellExpr))
			return
		}
		*d = append(*d, c.badAccessorDiagnostic(shellExpr))
	})
}

// bareShellExprDiagnostic covers `if $`cmd`:`, `f($cmd)`, `return $cmd` - an
// invocation used where a value is wanted, with nothing said about which value.
//
// No quick fix: there are four right answers and no way to tell which was
// meant. Naming all four in the message beats offering four competing edits.
func (c *RadCheckerImpl) bareShellExprDiagnostic(n *rl.ShellExpr) Diagnostic {
	msg := "A shell command has no value on its own - say which result you want"
	suggestion := "Add one of: " + accessorList() + "."
	return NewDiagnosticErrorFromSpanWithSuggestion(
		n.Span(), c.src, msg, rl.ErrShellExprNoAccessor, suggestion)
}

// badAccessorDiagnostic covers the three ways postfix can follow an invocation
// and not be an accessor. They need different explanations, because only one of
// them used to do something.
func (c *RadCheckerImpl) badAccessorDiagnostic(n *rl.ShellExpr) Diagnostic {
	seg := n.BadAccessor

	switch {
	case seg.Field != nil && !seg.IsUFCS:
		// `$`cmd`.stdout2` - right shape, wrong name.
		field := *seg.Field
		msg := fmt.Sprintf("'%s' is not a shell result. Use one of: %s", field, accessorList())
		if didYouMean := formatDidYouMean(findSimilarInSet(rl.ShellAccessorNames, field, 3)); didYouMean != "" {
			return NewDiagnosticErrorFromSpanWithSuggestion(
				seg.Span(), c.src, msg, rl.ErrShellPostfixNotAccessor, didYouMean)
		}
		return NewDiagnosticErrorFromSpan(seg.Span(), c.src, msg, rl.ErrShellPostfixNotAccessor)

	case seg.IsUFCS:
		// `$`echo hi`.upper()` - this ran ECHO HI, because the call built the
		// command rather than transforming its output.
		msg := "A method here applies to the command text, not its output"
		suggestion := "Insert an accessor: `$`cmd`.stdout.upper()` rather than `$`cmd`.upper()`."
		return NewDiagnosticErrorFromSpanWithSuggestion(
			seg.Span(), c.src, msg, rl.ErrShellPostfixNotAccessor, suggestion)

	default:
		// `$cmds[1]`, `$cmds[1:2]` - the migration case. `$` used to swallow the
		// whole expression, so this indexed `cmds` and ran whatever came back.
		msg := "`$` takes only the command, so this index reads its result, not the command"
		diag := NewDiagnosticErrorFromSpan(n.Span(), c.src, msg, rl.ErrShellPostfixNotAccessor)
		suggestion := fmt.Sprintf(
			"Wrap the command in parentheses: `%s`. See https://amterp.dev/rad/migrations/v0.12/",
			ParenthesizedCommandFix(diag.RangedSrc))
		diag.Suggestion = &suggestion
		return diag
	}
}

// ParenthesizedCommandFix rewrites `$cmds[1]` into `$(cmds[1])`, preserving any
// modifiers. Exported so the language server can offer the same edit as a quick
// fix rather than reconstructing it from the message prose.
//
// Falls back to the generic shape when the source doesn't look the way we
// expect, so a surprising span degrades to advice rather than nonsense.
func ParenthesizedCommandFix(src string) string {
	dollar := strings.Index(src, "$")
	if dollar < 0 {
		return "$(<command>)"
	}
	return fmt.Sprintf("%s$(%s)", src[:dollar], src[dollar+1:])
}

// accessorList renders the four accessors as an Oxford-or list, matching
// formatDidYouMean so the two read alike when they appear together.
func accessorList() string {
	quoted := make([]string, 0, len(rl.ShellAccessorNames))
	for _, name := range rl.ShellAccessorNames {
		quoted = append(quoted, "`."+name+"`")
	}
	return strings.Join(quoted[:len(quoted)-1], ", ") + ", or " + quoted[len(quoted)-1]
}
