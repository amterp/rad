package check

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// --- AST-based checks ---
// These checks walk the Go-native AST and skip when ast is nil (invalid syntax).

// walkAST recursively visits all nodes in the AST, calling visit for each.
func walkAST(node rl.Node, visit func(rl.Node)) {
	if node == nil {
		return
	}
	visit(node)
	walkASTChildren(node, func(child rl.Node) {
		walkAST(child, visit)
	})
}

// walkASTChildren calls visit for each direct child of node.
func walkASTChildren(node rl.Node, visit func(rl.Node)) {
	for _, child := range node.Children() {
		visit(child)
	}
}

// Check 4: Function name shadowing (AST version)
func (c *RadCheckerImpl) addFunctionNameShadowingErrorsAST(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	if c.ast.Args == nil {
		return
	}

	argNames := make(map[string]bool)
	for _, decl := range c.ast.Args.Decls {
		argNames[decl.Name] = true
	}

	for _, stmt := range c.ast.Stmts {
		fnDef, ok := stmt.(*rl.FnDef)
		if !ok {
			continue
		}
		if argNames[fnDef.Name] {
			msg := "Hoisted function '" + fnDef.Name + "' shadows an argument with the same name"
			*d = append(*d, NewDiagnosticErrorFromSpan(fnDef.DefSpan, c.src, msg, rl.ErrHoistedFunctionShadowsArgument))
		}
	}
}

// getHoistedFunctionsAST returns names of top-level function definitions from the AST.
func (c *RadCheckerImpl) getHoistedFunctionsAST() []string {
	if c.ast == nil {
		return nil
	}
	var names []string
	for _, stmt := range c.ast.Stmts {
		if fnDef, ok := stmt.(*rl.FnDef); ok {
			names = append(names, fnDef.Name)
		}
	}
	return names
}

// Check 5: Unknown function hints (AST version)
func (c *RadCheckerImpl) addUnknownFunctionHintsAST(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	builtInFunctions := rts.GetBuiltInFunctions()

	hoistedFunctionSet := make(map[string]bool)
	for _, name := range c.getHoistedFunctionsAST() {
		hoistedFunctionSet[name] = true
	}

	walkAST(c.ast, func(node rl.Node) {
		call, ok := node.(*rl.Call)
		if !ok {
			return
		}
		ident, ok := call.Func.(*rl.Identifier)
		if !ok {
			return
		}
		fnName := ident.Name
		if builtInFunctions.Contains(fnName) || hoistedFunctionSet[fnName] {
			return
		}
		msg := "Function '" + fnName + "' may not be defined (only built-in and top-level functions are tracked)"
		*d = append(*d, NewDiagnosticHintFromSpan(ident.Span(), c.src, msg, rl.ErrUnknownFunction))
	})
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
	}

	walkASTChildren(node, func(child rl.Node) {
		c.walkASTForReturn(child, d, inFunction, inYieldContext)
	})
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
