package check

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// todo be able to check scripts with different versions of Rad?

type RadChecker interface {
	UpdateSrc(src string)
	Update(tree *rts.RadTree, src string, ast *rl.SourceFile)
	CheckDefault() (Result, error)
	Check(Opts) (Result, error)
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
		c.tree.Update(src)
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

func (c *RadCheckerImpl) CheckDefault() (Result, error) {
	return c.Check(NewOpts())
}

// todo use opts
func (c *RadCheckerImpl) Check(opts Opts) (Result, error) {
	diagnostics := make([]Diagnostic, 0)
	c.addInvalidNodes(&diagnostics)
	c.addIntScientificNotationErrors(&diagnostics)
	c.addFnParamScientificNotationErrors(&diagnostics)
	c.addFunctionNameShadowingErrorsAST(&diagnostics)
	c.addUnknownFunctionHintsAST(&diagnostics)
	c.addBreakContinueOutsideLoopErrorsAST(&diagnostics)
	c.addReturnOutsideFunctionErrorsAST(&diagnostics)
	c.addInvalidAssignmentLHSErrorsAST(&diagnostics)
	c.addUnknownCommandCallbackWarnings(&diagnostics)
	return Result{
		Diagnostics: diagnostics,
	}, nil
}

func (c *RadCheckerImpl) addInvalidNodes(d *[]Diagnostic) {
	nodes := c.tree.FindInvalidNodes()
	for _, node := range nodes {
		msg, code, suggestion := GenerateErrorMessage(node, c.src)
		*d = append(*d, NewDiagnosticErrorWithSuggestion(node, c.src, msg, code, suggestion))
	}
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

func (c *RadCheckerImpl) addUnknownCommandCallbackWarnings(d *[]Diagnostic) {
	if c.ast == nil || len(c.ast.Cmds) == 0 {
		return
	}

	builtInFunctions := rts.GetBuiltInFunctions()

	hoistedFunctionSet := make(map[string]bool)
	for _, name := range c.getHoistedFunctionsAST() {
		hoistedFunctionSet[name] = true
	}

	for _, cmd := range c.ast.Cmds {
		cb := cmd.Callback
		if cb.IsLambda || cb.IdentifierName == nil {
			continue
		}

		fnName := *cb.IdentifierName
		if builtInFunctions.Contains(fnName) || hoistedFunctionSet[fnName] {
			continue
		}

		if cb.IdentifierSpan == nil {
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
