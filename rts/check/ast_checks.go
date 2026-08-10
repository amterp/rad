package check

import (
	"github.com/amterp/rad/rts/rl"
)

// --- AST-based checks ---
// These checks walk the Go-native AST and skip when ast is nil (invalid syntax).

// walkAST is a thin alias for rl.Walk. Kept as a package-local
// name so existing check-package callers read naturally. The
// shared walker lives in rl so the LSP analysis layer gets the
// same traversal semantics.
func walkAST(node rl.Node, visit func(rl.Node)) {
	rl.Walk(node, visit)
}

// walkASTChildren calls visit for each direct child of node.
func walkASTChildren(node rl.Node, visit func(rl.Node)) {
	for _, child := range node.Children() {
		visit(child)
	}
}

// Check 7: Break/continue outside loop (AST version)
func (c *RadCheckerImpl) addBreakContinueOutsideLoopErrorsAST(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}
	c.walkASTForBreakContinue(c.ast, d, 0)
}

func (c *RadCheckerImpl) walkASTForBreakContinue(node rl.Node, d *[]Diagnostic, loopDepth int) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *rl.Break:
		if loopDepth == 0 {
			msg := "'break' can only be used inside a loop"
			*d = append(*d, NewDiagnosticErrorFromSpan(n.Span(), c.src, msg, rl.ErrBreakOutsideLoop))
		}
		return
	case *rl.Continue:
		if loopDepth == 0 {
			msg := "'continue' can only be used inside a loop"
			*d = append(*d, NewDiagnosticErrorFromSpan(n.Span(), c.src, msg, rl.ErrContinueOutsideLoop))
		}
		return
	case *rl.ForLoop:
		loopDepth++
	case *rl.WhileLoop:
		loopDepth++
	case *rl.FnDef, *rl.Lambda:
		// break/continue don't cross function boundaries
		loopDepth = 0
	}

	walkASTChildren(node, func(child rl.Node) {
		c.walkASTForBreakContinue(child, d, loopDepth)
	})
}

// Check 8: Return/yield outside function (AST version)
func (c *RadCheckerImpl) addReturnOutsideFunctionErrorsAST(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}
	c.walkASTForReturn(c.ast, d, false, false)
}

func (c *RadCheckerImpl) walkASTForReturn(node rl.Node, d *[]Diagnostic, inFunction, inYieldContext bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *rl.Return:
		if !inFunction {
			msg := "'return' can only be used inside a function"
			*d = append(*d, NewDiagnosticErrorFromSpan(n.Span(), c.src, msg, rl.ErrReturnOutsideFunction))
		}
		return
	case *rl.Yield:
		if !inFunction && !inYieldContext {
			msg := "'yield' can only be used inside a function or switch expression"
			*d = append(*d, NewDiagnosticErrorFromSpan(n.Span(), c.src, msg, rl.ErrYieldOutsideFunction))
		}
		return
	case *rl.FnDef:
		inFunction = true
	case *rl.Lambda:
		inFunction = true
	case *rl.Switch:
		inYieldContext = true
	case *rl.Assign:
		// Assign.Children() flattens the catch block's stmts in
		// alongside the value-side children, which loses the
		// "inside a catch body" signal. Handle Assign explicitly so
		// the catch stmts get inYieldContext = true and the
		// canonical `x = foo() catch: yield 0` shape doesn't fire
		// RAD40005 spuriously.
		c.walkAssignForReturn(n, d, inFunction, inYieldContext)
		return
	case *rl.ExprStmt:
		// Same as Assign: ExprStmt also carries a CatchBlock that
		// gets flattened in Children().
		c.walkExprStmtForReturn(n, d, inFunction, inYieldContext)
		return
	case *rl.Shell:
		// Same as Assign: Shell also carries a CatchBlock.
		c.walkShellForReturn(n, d, inFunction, inYieldContext)
		return
	}

	walkASTChildren(node, func(child rl.Node) {
		c.walkASTForReturn(child, d, inFunction, inYieldContext)
	})
}

func (c *RadCheckerImpl) walkAssignForReturn(a *rl.Assign, d *[]Diagnostic, inFunction, inYieldContext bool) {
	for _, target := range a.Targets {
		c.walkASTForReturn(target, d, inFunction, inYieldContext)
	}
	for _, value := range a.Values {
		c.walkASTForReturn(value, d, inFunction, inYieldContext)
	}
	if a.Catch != nil {
		for _, stmt := range a.Catch.Stmts {
			c.walkASTForReturn(stmt, d, inFunction, true)
		}
	}
}

func (c *RadCheckerImpl) walkExprStmtForReturn(e *rl.ExprStmt, d *[]Diagnostic, inFunction, inYieldContext bool) {
	c.walkASTForReturn(e.Expr, d, inFunction, inYieldContext)
	if e.Catch != nil {
		for _, stmt := range e.Catch.Stmts {
			c.walkASTForReturn(stmt, d, inFunction, true)
		}
	}
}

func (c *RadCheckerImpl) walkShellForReturn(s *rl.Shell, d *[]Diagnostic, inFunction, inYieldContext bool) {
	for _, target := range s.Targets {
		c.walkASTForReturn(target, d, inFunction, inYieldContext)
	}
	c.walkASTForReturn(s.Cmd, d, inFunction, inYieldContext)
	if s.Catch != nil {
		for _, stmt := range s.Catch.Stmts {
			c.walkASTForReturn(stmt, d, inFunction, true)
		}
	}
}

// Check 9: Invalid assignment LHS (AST version)
func (c *RadCheckerImpl) addInvalidAssignmentLHSErrorsAST(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		assign, ok := node.(*rl.Assign)
		if !ok {
			return
		}
		for _, target := range assign.Targets {
			c.validateAssignmentTargetAST(target, d)
		}
	})
}

// Check 10: Deprecated block keywords ('request', 'display')
func (c *RadCheckerImpl) addDeprecatedBlockKeywordErrors(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		radBlock, ok := node.(*rl.RadBlock)
		if !ok {
			return
		}
		if radBlock.Keyword == rl.KEYWORD_REQUEST || radBlock.Keyword == rl.KEYWORD_DISPLAY {
			msg := "'" + radBlock.Keyword + "' blocks have been removed. Use 'rad' instead. See https://amterp.dev/rad/migrations/v0.9/"
			*d = append(*d, NewDiagnosticErrorFromSpan(radBlock.KeywordSpan, c.src, msg, rl.ErrDeprecatedBlockKeyword))
		}
	})
}

// Check 12: Hand-written quotes around an interpolation in a shell command.
//
// Rad quotes interpolated values itself, so quotes the script adds around one
// are no longer doing the job they were added for - they now land in the
// argument as literal characters. This fires on the pre-quoting idiom so a
// migrating script gets told rather than discovering it in output.
//
// We only look at command *literals*. A `$cmd` built up elsewhere is the raw
// form by design, and we have no interpolation to point at anyway.
func (c *RadCheckerImpl) addShellInterpolationQuoteErrors(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		shell, ok := node.(*rl.Shell)
		if !ok {
			return
		}
		lit, ok := shell.Cmd.(*rl.LitString)
		if !ok || lit.Simple {
			return
		}

		for _, seg := range quotedInterpolations(lit) {
			msg := "This value is inside quotes, but Rad already quotes interpolated values in " +
				"shell commands - these quotes will end up in the argument itself. Remove them. " +
				"If the argument is literal text plus a value, build it as a string first and " +
				"interpolate that."
			*d = append(*d, NewDiagnosticErrorFromSpan(seg.Span(), c.src, msg, rl.ErrShellInterpolationQuoted))
		}
	})
}

// quotedInterpolations returns the interpolation segments of a command literal
// that sit inside a shell quote the script wrote itself.
//
// This tracks quote state across the literal text rather than only checking the
// characters either side, so it also catches `"some text {x}"`, where the value
// is quoted along with a literal prefix.
func quotedInterpolations(lit *rl.LitString) []rl.StringSegment {
	var found []rl.StringSegment
	var quote byte // 0, '\'' or '"'

	for _, seg := range lit.Segments {
		if !seg.IsLiteral {
			if quote != 0 {
				found = append(found, seg)
			}
			continue
		}

		for k := 0; k < len(seg.Text); k++ {
			ch := seg.Text[k]
			switch {
			case quote == '\'':
				// Nothing is special inside single quotes except the closer.
				if ch == '\'' {
					quote = 0
				}
			case ch == '\\':
				// Escapes the next character, in double quotes and bare alike.
				k++
			case quote == '"':
				if ch == '"' {
					quote = 0
				}
			case ch == '\'' || ch == '"':
				quote = ch
			}
		}
	}
	return found
}

// Check 11: Rad block options that have no effect in certain contexts.
//   - 'insecure' and 'quiet' only apply to URL sources (string); they have no
//     effect on list/map or no-source rad blocks.
//   - 'noprint' on a no-source rad block is a no-op because the save/restore
//     pattern undoes all mutations when the block returns.
//
// We can only statically detect the no-source case (Source == nil).
// When a source expression exists, we can't know at compile time whether
// it resolves to a URL or list/map, and both code paths are legitimate,
// so we don't warn in that case.
func (c *RadCheckerImpl) addRadOptionNoEffectWarnings(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		radBlock, ok := node.(*rl.RadBlock)
		if !ok || radBlock.Source != nil {
			return
		}

		for _, stmt := range radBlock.Stmts {
			opt, ok := stmt.(*rl.RadOption)
			if !ok {
				continue
			}
			switch opt.Keyword {
			case rl.KEYWORD_INSECURE:
				msg := "'insecure' has no effect without a URL source"
				*d = append(*d, NewDiagnosticWarnFromSpan(opt.Span(), c.src, msg, rl.ErrRadOptionNoEffect))
			case rl.KEYWORD_QUIET:
				msg := "'quiet' has no effect without a URL source"
				*d = append(*d, NewDiagnosticWarnFromSpan(opt.Span(), c.src, msg, rl.ErrRadOptionNoEffect))
			case rl.KEYWORD_NOPRINT:
				msg := "'noprint' has no effect without a source (mutations are not preserved)"
				*d = append(*d, NewDiagnosticWarnFromSpan(opt.Span(), c.src, msg, rl.ErrRadOptionNoEffect))
			}
		}
	})
}

// Check 13: Interpolation of a constant literal.
//
// `{4}` interpolates the integer literal 4, which is exactly what the
// author would have written anyway - so the interpolation carries no
// information. The reason this earns a warning rather than a style hint
// is `"\d{4}"`: a regex quantifier is a valid interpolation, so it
// silently becomes `\d4` with no error anywhere. Multi-part quantifiers
// (`{2,3}`) already fail to parse, so flagging the single-token form
// closes the last silent hole in that family.
//
// The rule is deliberately syntactic: we fire when the expression *node*
// is a literal, not when it merely evaluates to a constant. `{x}` with
// `x = 4` stays quiet. Anything wider needs const-folding, which buys
// false positives in a diagnostic tier that is already fighting for
// trust.
func (c *RadCheckerImpl) addConstantInterpolationWarnings(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	walkAST(c.ast, func(node rl.Node) {
		str, ok := node.(*rl.LitString)
		if !ok || str.Simple {
			return
		}

		// Iterate Segments rather than Children(): the child list also
		// carries Format.Padding and Format.Precision, which are int
		// literals on every padded format spec ("{v:<12}").
		for _, seg := range str.Segments {
			if seg.IsLiteral || !isConstantScalarLiteral(seg.Expr) {
				continue
			}
			literal := safeSlice(c.src, seg.Expr.Span().StartByte, seg.Expr.Span().EndByte)
			msg := "Interpolating the literal " + truncate(literal, 20) + " has no effect"
			suggestion := "Write the value directly. For a literal brace, escape it as '\\{' or use a raw string: r\"...\""
			*d = append(*d, NewDiagnosticWarnFromSpanWithSuggestion(
				seg.Span(), c.src, msg, rl.ErrConstantInterpolation, suggestion))
		}
	})
}

// isConstantScalarLiteral reports whether node is a literal whose value
// is fixed at parse time. Collections are excluded: their elements need
// not be constant. An interpolated string is excluded for the same
// reason - only `Simple` strings have a parse-time value.
//
// The switch-case analysis has two cousins of this test that carry
// per-kind payloads rather than a bare bool: caseLiteralKey and
// caseKeyDisplayText, both in type_check.go.
func isConstantScalarLiteral(node rl.Node) bool {
	switch v := node.(type) {
	case *rl.LitInt, *rl.LitFloat, *rl.LitBool:
		return true
	case *rl.LitString:
		return v.Simple
	}
	return false
}

func (c *RadCheckerImpl) validateAssignmentTargetAST(node rl.Node, d *[]Diagnostic) {
	if node == nil {
		return
	}

	switch node.(type) {
	case *rl.Identifier, *rl.VarPath:
		// Valid assignment targets
		return
	case *rl.LitInt, *rl.LitFloat, *rl.LitString, *rl.LitBool, *rl.LitNull:
		content := safeSlice(c.src, node.Span().StartByte, node.Span().EndByte)
		msg := "Cannot assign to literal '" + truncate(content, 20) + "'"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	case *rl.Call:
		msg := "Cannot assign to function call result"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	case *rl.OpBinary, *rl.Ternary:
		msg := "Cannot assign to expression"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	default:
		content := safeSlice(c.src, node.Span().StartByte, node.Span().EndByte)
		msg := "Cannot assign to '" + truncate(content, 20) + "'"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	}
}

// Check 13: command block coherence.
//
// The grammar deliberately accepts shapes that mean nothing - a command with no
// `calls` and no sub-commands, one with both, `default` on a namespace, two
// defaults among siblings - so that these can be reported here with a span and
// a message rather than as "Invalid syntax" at a dedent. Forgetting `calls` is
// the likeliest first mistake anyone writing a command makes, and the parser
// has nothing useful to say about it.
//
// Sibling context is what makes this a level-aware recursion rather than a
// WalkCmds: whether a `default` is a duplicate cannot be told from the block
// itself.
func (c *RadCheckerImpl) addCommandBlockErrors(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	var checkLevel func(cmds []*rl.CmdBlock)
	checkLevel = func(cmds []*rl.CmdBlock) {
		var firstDefault *rl.CmdBlock
		for _, cmd := range cmds {
			switch {
			case cmd.IsNamespace() && cmd.Callback != nil:
				*d = append(*d, NewDiagnosticErrorFromSpan(cmd.Callback.Span_, c.src,
					"'"+cmd.Name+"' contains sub-commands, so it routes to them rather than running itself. Remove 'calls', or move it to a sub-command.",
					rl.ErrNamespaceHasCallback))
			case !cmd.IsNamespace() && cmd.Callback == nil:
				*d = append(*d, NewDiagnosticErrorFromSpan(cmd.Span(), c.src,
					"Command '"+cmd.Name+"' does nothing: it needs a 'calls' line naming what to run, or nested command blocks to route to.",
					rl.ErrCommandMissingCallback))
			}

			if cmd.IsDefault() {
				switch {
				case cmd.IsNamespace():
					*d = append(*d, NewDiagnosticErrorFromSpan(*cmd.DefaultSpan, c.src,
						"'"+cmd.Name+"' contains sub-commands, so it cannot be the default. Mark one of its sub-commands instead.",
						rl.ErrDefaultOnNamespace))
				case firstDefault != nil:
					*d = append(*d, NewDiagnosticErrorFromSpan(*cmd.DefaultSpan, c.src,
						"'"+firstDefault.Name+"' is already the default at this level, so '"+cmd.Name+"' cannot be. Only one command per level runs when none is named.",
						rl.ErrMultipleDefaultCommands))
				default:
					firstDefault = cmd
				}
			}

			checkLevel(cmd.SubCmds)
		}
	}
	checkLevel(c.ast.Cmds)
}
