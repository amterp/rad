package check

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadChecker interface {
	UpdateSrc(src string)
	Update(tree *rts.RadTree, src string, ast *rl.SourceFile)
	Check() (Result, error)
}

type RadCheckerImpl struct {
	parser *rts.RadParser
	tree   *rts.RadTree
	src    string
	ast    *rl.SourceFile
}

func NewChecker() (RadChecker, error) {
	parser, err := rts.NewRadParser()
	if err != nil {
		return nil, err
	}
	tree := parser.Parse("")
	return NewCheckerWithTree(tree, parser, "", nil), nil
}

func NewCheckerWithTree(tree *rts.RadTree, parser *rts.RadParser, src string, ast *rl.SourceFile) RadChecker {
	return &RadCheckerImpl{
		parser: parser,
		tree:   tree,
		src:    src,
		ast:    ast,
	}
}

func (c *RadCheckerImpl) UpdateSrc(src string) {
	if c.tree == nil {
		c.tree = c.parser.Parse(src)
	} else {
		c.tree.Update(c.parser, src)
	}
	c.src = src
	// Attempt AST conversion for AST-based checks.
	// Falls back to nil on invalid syntax (converter may panic on ERROR nodes).
	c.ast = c.tryConvertAST(src)
}

func (c *RadCheckerImpl) tryConvertAST(src string) (ast *rl.SourceFile) {
	defer func() {
		if r := recover(); r != nil {
			ast = nil
		}
	}()
	return rts.ConvertCST(c.tree.Root(), src, "<check>")
}

func (c *RadCheckerImpl) Update(tree *rts.RadTree, src string, ast *rl.SourceFile) {
	c.tree = tree
	c.src = src
	c.ast = ast
}

func (c *RadCheckerImpl) Check() (Result, error) {
	diagnostics := make([]Diagnostic, 0)
	c.addInvalidNodes(&diagnostics)
	c.addIntScientificNotationErrors(&diagnostics)
	c.addFnParamScientificNotationErrors(&diagnostics)

	// Resolve once and share across checks that need the symbol table.
	// Subsequent commits will rebuild more checks on this view.
	var resolved *Resolved
	var typeInfo *TypeInfo
	if c.ast != nil {
		resolved = Resolve(c.ast)
		typeInfo = TypeCheck(c.ast, resolved)
	}

	c.addUnknownFunctionHints(resolved, &diagnostics)
	c.addBindIssues(resolved, &diagnostics)
	c.addTypeIssues(typeInfo, &diagnostics)
	c.addBreakContinueOutsideLoopErrorsAST(&diagnostics)
	c.addReturnOutsideFunctionErrorsAST(&diagnostics)
	c.addInvalidAssignmentLHSErrorsAST(&diagnostics)
	c.addUnknownCommandCallbackWarnings(resolved, &diagnostics)
	c.addDeprecatedBlockKeywordErrors(&diagnostics)
	c.addRadOptionNoEffectWarnings(&diagnostics)
	return Result{
		Diagnostics: diagnostics,
		Resolved:    resolved,
		Types:       typeInfo,
	}, nil
}

// addUnknownFunctionHints is retained as a no-op shim. The binder
// now emits a hard RAD20028 (Undefined identifier) for every
// unresolved name including call-site fn names, with a did-you-
// mean suggestion. The old RAD40003 hint was a strictly weaker
// version of the same signal; keeping both around would double-
// surface the problem.
func (c *RadCheckerImpl) addUnknownFunctionHints(resolved *Resolved, d *[]Diagnostic) {
}

// addTypeIssues surfaces type-checker findings (type mismatches, arg
// count errors, etc.) as Diagnostics. Empty in Phase 2a since the
// type checker hasn't started emitting yet; wiring it up now means
// later sub-commits don't need to touch the orchestration code.
func (c *RadCheckerImpl) addTypeIssues(info *TypeInfo, d *[]Diagnostic) {
	if info == nil {
		return
	}
	for _, issue := range info.Issues {
		*d = append(*d, diagnosticFromBindIssue(issue, c.src))
	}
}

// addBindIssues surfaces structural findings the binder collected
// (duplicate params, fn/arg shadowing) as Diagnostics. Each issue
// carries its own severity so the binder, not the checker, decides
// how loudly to flag a problem.
func (c *RadCheckerImpl) addBindIssues(resolved *Resolved, d *[]Diagnostic) {
	if resolved == nil {
		return
	}
	for _, issue := range resolved.Issues {
		*d = append(*d, diagnosticFromBindIssue(issue, c.src))
	}
}

// diagnosticFromBindIssue is the single conversion point from the
// binder/type-checker's BindIssue value-type into a Diagnostic that
// the checker layer can hand to the renderer. Centralized so the
// Suggestion plumbing only lives in one place.
func diagnosticFromBindIssue(issue BindIssue, src string) Diagnostic {
	diag := NewDiagnosticFromSpan(issue.Span, src, issueSeverityToCheck(issue.Severity), issue.Message, codePtr(issue.Code))
	if issue.Suggestion != "" {
		s := issue.Suggestion
		diag.Suggestion = &s
	}
	return diag
}

// issueSeverityToCheck maps the binder's local IssueSeverity onto the
// checker's wider Severity scale. Kept in check.go (not resolve.go) so
// the binder doesn't drag the wider Severity enum into its imports.
func issueSeverityToCheck(s IssueSeverity) Severity {
	switch s {
	case IssueWarning:
		return Warning
	case IssueHint:
		return Hint
	default:
		return Error
	}
}

func codePtr(c rl.Error) *rl.Error { return &c }

func (c *RadCheckerImpl) addInvalidNodes(d *[]Diagnostic) {
	nodes := c.tree.FindInvalidNodes()
	// When the parser bails on a misordered-default switch it emits a
	// cascade of ERROR nodes for the same misshape. We collapse the
	// cascade to just the outermost one so users see a single helpful
	// hint instead of three near-identical RAD10001s. The outermost
	// is identified by its byte range strictly containing the inner
	// ones.
	suppressed := computeMisorderedSwitchSuppressionRanges(nodes, c.src)
	for _, node := range nodes {
		if isSuppressedByMisorderedSwitch(node, suppressed) {
			continue
		}
		msg, code, suggestion := GenerateErrorMessage(node, c.src)
		*d = append(*d, NewDiagnosticErrorWithSuggestion(node, c.src, msg, code, suggestion))
	}
}

// computeMisorderedSwitchSuppressionRanges returns the byte ranges
// of misordered-switch errors that should swallow nested cascade
// diagnostics. Only the outermost qualifying ERROR contributes a
// range; any narrower ERRORs nested inside fall under it.
func computeMisorderedSwitchSuppressionRanges(nodes []*ts.Node, src string) []byteRange {
	var ranges []byteRange
	for _, n := range nodes {
		if !n.IsError() {
			continue
		}
		if !errorIsTopLevel(n) {
			continue
		}
		body := nodeSrc(n, src)
		if !containsMisorderedSwitch(body) {
			continue
		}
		ranges = append(ranges, byteRange{n.StartByte(), n.EndByte()})
	}
	return ranges
}

// isSuppressedByMisorderedSwitch reports whether `node` lies
// strictly inside any of the misordered-switch suppression ranges -
// meaning its diagnostic would be cascade noise.
func isSuppressedByMisorderedSwitch(node *ts.Node, ranges []byteRange) bool {
	start, end := node.StartByte(), node.EndByte()
	for _, r := range ranges {
		if start == r.start && end == r.end {
			continue // the outermost itself - keep, that's the canonical diagnostic
		}
		if start >= r.start && end <= r.end {
			return true
		}
	}
	return false
}

type byteRange struct {
	start, end uint
}

func errorIsTopLevel(n *ts.Node) bool {
	p := n.Parent()
	return p != nil && p.Kind() == "source_file"
}

func nodeSrc(n *ts.Node, src string) string {
	start := int(n.StartByte())
	end := int(n.EndByte())
	if start < 0 || end > len(src) || start > end {
		return ""
	}
	return src[start:end]
}

func (c *RadCheckerImpl) addIntScientificNotationErrors(d *[]Diagnostic) {
	if c.ast == nil || c.ast.Args == nil {
		return
	}
	for _, decl := range c.ast.Args.Decls {
		if decl.TypeName != rl.T_INT || decl.Default == nil {
			continue
		}
		c.checkExprForNonWholeFloat(decl.Default, d)
	}
}

func (c *RadCheckerImpl) addFnParamScientificNotationErrors(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}
	walkAST(c.ast, func(node rl.Node) {
		var params []rl.TypingFnParam
		switch n := node.(type) {
		case *rl.FnDef:
			if n.Typing != nil {
				params = n.Typing.Params
			}
		case *rl.Lambda:
			if n.Typing != nil {
				params = n.Typing.Params
			}
		default:
			return
		}
		for _, param := range params {
			if !isIntType(param.Type) {
				continue
			}
			if param.DefaultAST == nil || param.DefaultAST.Node == nil {
				continue
			}
			c.checkExprForNonWholeFloat(param.DefaultAST.Node, d)
		}
	})
}

// checkExprForNonWholeFloat walks an AST expression for LitFloat nodes whose
// values aren't whole numbers. This catches scientific notation like 1.5e2 used
// in int-typed defaults - the converter turns whole-number scientific notation
// into LitInt, so any remaining LitFloat IS the error case.
func (c *RadCheckerImpl) checkExprForNonWholeFloat(node rl.Node, d *[]Diagnostic) {
	walkAST(node, func(n rl.Node) {
		litFloat, ok := n.(*rl.LitFloat)
		if !ok {
			return
		}
		if litFloat.Value != float64(int64(litFloat.Value)) {
			msg := "Scientific notation value does not evaluate to a whole number"
			*d = append(*d, NewDiagnosticErrorFromSpan(litFloat.Span(), c.src, msg, rl.ErrScientificNotationNotWholeNumber))
		}
	})
}

// isIntType reports whether a typing is the simple int type.
// Returns false for union types, optional types, etc.
func isIntType(t *rl.TypingT) bool {
	if t == nil {
		return false
	}
	_, ok := (*t).(*rl.TypingIntT)
	return ok
}

// addUnknownCommandCallbackWarnings flags `calls <name>` callbacks
// whose target isn't visible at file scope or as an ambient builtin.
//
// The file-scope check goes through the resolved view so it stays
// consistent with the rest of the checker on what's bound where. The
// builtin check goes through the runtime function set directly:
// Builtin Symbols are synthesized lazily on first reference, so a
// script that uses `print` ONLY as a cmd callback never triggers the
// synthesis and a Resolved-only check would emit a false positive.
func (c *RadCheckerImpl) addUnknownCommandCallbackWarnings(resolved *Resolved, d *[]Diagnostic) {
	if c.ast == nil || len(c.ast.Cmds) == 0 || resolved == nil {
		return
	}
	builtins := rts.GetBuiltInFunctions()
	for _, cmd := range c.ast.Cmds {
		cb := cmd.Callback
		if cb.IsLambda || cb.IdentifierName == nil || cb.IdentifierSpan == nil {
			continue
		}
		fnName := *cb.IdentifierName
		if resolved.File.Lookup(fnName) != nil || builtins.Contains(fnName) {
			continue
		}
		msg := "Function '" + fnName + "' may not be defined (only built-in and top-level functions are tracked)"
		*d = append(*d, NewDiagnosticWarnFromSpan(*cb.IdentifierSpan, c.src, msg, rl.ErrUnknownFunction))
	}
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
