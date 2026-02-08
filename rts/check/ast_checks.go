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
	switch n := node.(type) {
	case *rl.SourceFile:
		for _, s := range n.Stmts {
			visit(s)
		}
	case *rl.Assign:
		for _, t := range n.Targets {
			visit(t)
		}
		for _, v := range n.Values {
			visit(v)
		}
		if n.Catch != nil {
			for _, s := range n.Catch.Stmts {
				visit(s)
			}
		}
	case *rl.ExprStmt:
		visit(n.Expr)
		if n.Catch != nil {
			for _, s := range n.Catch.Stmts {
				visit(s)
			}
		}
	case *rl.If:
		for _, b := range n.Branches {
			if b.Condition != nil {
				visit(b.Condition)
			}
			for _, s := range b.Body {
				visit(s)
			}
		}
	case *rl.Switch:
		visit(n.Discriminant)
		for _, c := range n.Cases {
			for _, k := range c.Keys {
				visit(k)
			}
			visit(c.Alt)
		}
		if n.Default != nil {
			visit(n.Default.Alt)
		}
	case *rl.SwitchCaseExpr:
		for _, v := range n.Values {
			visit(v)
		}
	case *rl.SwitchCaseBlock:
		for _, s := range n.Stmts {
			visit(s)
		}
	case *rl.ForLoop:
		visit(n.Iter)
		for _, s := range n.Body {
			visit(s)
		}
	case *rl.WhileLoop:
		if n.Condition != nil {
			visit(n.Condition)
		}
		for _, s := range n.Body {
			visit(s)
		}
	case *rl.Shell:
		for _, t := range n.Targets {
			visit(t)
		}
		visit(n.Cmd)
		if n.Catch != nil {
			for _, s := range n.Catch.Stmts {
				visit(s)
			}
		}
	case *rl.Del:
		for _, t := range n.Targets {
			visit(t)
		}
	case *rl.Defer:
		for _, s := range n.Body {
			visit(s)
		}
	case *rl.Return:
		for _, v := range n.Values {
			visit(v)
		}
	case *rl.Yield:
		for _, v := range n.Values {
			visit(v)
		}
	case *rl.FnDef:
		if n.Typing != nil {
			for _, param := range n.Typing.Params {
				if param.DefaultAST != nil && param.DefaultAST.Node != nil {
					visit(param.DefaultAST.Node)
				}
			}
		}
		for _, s := range n.Body {
			visit(s)
		}
	case *rl.OpBinary:
		visit(n.Left)
		visit(n.Right)
	case *rl.OpUnary:
		visit(n.Operand)
	case *rl.Ternary:
		visit(n.Condition)
		visit(n.True)
		visit(n.False)
	case *rl.Fallback:
		visit(n.Left)
		visit(n.Right)
	case *rl.Call:
		visit(n.Func)
		for _, a := range n.Args {
			visit(a)
		}
		for _, na := range n.NamedArgs {
			visit(na.Value)
		}
	case *rl.VarPath:
		visit(n.Root)
		for _, seg := range n.Segments {
			if seg.Index != nil {
				visit(seg.Index)
			}
			if seg.Start != nil {
				visit(seg.Start)
			}
			if seg.End != nil {
				visit(seg.End)
			}
		}
	case *rl.Lambda:
		if n.Typing != nil {
			for _, param := range n.Typing.Params {
				if param.DefaultAST != nil && param.DefaultAST.Node != nil {
					visit(param.DefaultAST.Node)
				}
			}
		}
		for _, s := range n.Body {
			visit(s)
		}
	case *rl.LitString:
		if !n.Simple {
			for _, seg := range n.Segments {
				if !seg.IsLiteral && seg.Expr != nil {
					visit(seg.Expr)
				}
				if seg.Format != nil {
					if seg.Format.Padding != nil {
						visit(seg.Format.Padding)
					}
					if seg.Format.Precision != nil {
						visit(seg.Format.Precision)
					}
				}
			}
		}
	case *rl.LitList:
		for _, e := range n.Elements {
			visit(e)
		}
	case *rl.LitMap:
		for _, e := range n.Entries {
			visit(e.Key)
			visit(e.Value)
		}
	case *rl.ListComp:
		visit(n.Expr)
		visit(n.Iter)
		if n.Condition != nil {
			visit(n.Condition)
		}
	case *rl.RadBlock:
		visit(n.Source)
		for _, s := range n.Stmts {
			visit(s)
		}
	case *rl.RadField:
		for _, id := range n.Identifiers {
			visit(id)
		}
	case *rl.RadFieldMod:
		for _, f := range n.Fields {
			visit(f)
		}
		for _, a := range n.Args {
			visit(a)
		}
	case *rl.RadIf:
		for _, b := range n.Branches {
			if b.Condition != nil {
				visit(b.Condition)
			}
			for _, s := range b.Body {
				visit(s)
			}
		}
	case *rl.JsonPath:
		for _, seg := range n.Segments {
			for _, idx := range seg.Indexes {
				if idx.Expr != nil {
					visit(idx.Expr)
				}
			}
		}
	case *rl.RadSort:
		// leaf - no children
	case *rl.Identifier:
		// leaf
	case *rl.LitInt:
		// leaf
	case *rl.LitFloat:
		// leaf
	case *rl.LitBool:
		// leaf
	case *rl.LitNull:
		// leaf
	case *rl.Break:
		// leaf
	case *rl.Continue:
		// leaf
	case *rl.Pass:
		// leaf
	}
}

// Check 4: Function name shadowing (AST version)
func (c *RadCheckerImpl) addFunctionNameShadowingErrorsAST(d *[]Diagnostic) {
	if c.ast == nil {
		return
	}

	argBlock, ok := c.tree.FindArgBlock()
	if !ok {
		return
	}

	argNames := make(map[string]bool)
	for _, arg := range argBlock.Args {
		argNames[arg.Name.Name] = true
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

func (c *RadCheckerImpl) validateAssignmentTargetAST(node rl.Node, d *[]Diagnostic) {
	if node == nil {
		return
	}

	switch node.(type) {
	case *rl.Identifier, *rl.VarPath:
		// Valid assignment targets
		return
	case *rl.LitInt, *rl.LitFloat, *rl.LitString, *rl.LitBool, *rl.LitNull:
		content := c.src[node.Span().StartByte:node.Span().EndByte]
		msg := "Cannot assign to literal '" + truncate(content, 20) + "'"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	case *rl.Call:
		msg := "Cannot assign to function call result"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	case *rl.OpBinary, *rl.Ternary:
		msg := "Cannot assign to expression"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	default:
		content := c.src[node.Span().StartByte:node.Span().EndByte]
		msg := "Cannot assign to '" + truncate(content, 20) + "'"
		*d = append(*d, NewDiagnosticErrorFromSpan(node.Span(), c.src, msg, rl.ErrInvalidAssignmentTarget))
	}
}
