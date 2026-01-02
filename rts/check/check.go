package check

import (
	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// todo be able to check scripts with different versions of Rad?

type RadChecker interface {
	UpdateSrc(src string)
	CheckDefault() (Result, error)
	Check(Opts) (Result, error)
}

type RadCheckerImpl struct {
	parser *rts.RadParser
	tree   *rts.RadTree
	src    string
}

func NewChecker() (RadChecker, error) {
	parser, err := rts.NewRadParser()
	if err != nil {
		return nil, err
	}
	tree := parser.Parse("")
	return NewCheckerWithTree(tree, parser, ""), nil
}

func NewCheckerWithTree(tree *rts.RadTree, parser *rts.RadParser, src string) RadChecker {
	return &RadCheckerImpl{
		parser: parser,
		tree:   tree,
		src:    src,
	}
}

func (c *RadCheckerImpl) UpdateSrc(src string) {
	if c.tree == nil {
		c.tree = c.parser.Parse(src)
	} else {
		c.tree.Update(src)
	}
	c.src = src
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
	c.addFunctionNameShadowingErrors(&diagnostics)
	c.addUnknownFunctionHints(&diagnostics)
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
	// Use RadTree API to find arg block
	argBlock, ok := c.tree.FindArgBlock()
	if !ok {
		return
	}

	// Iterate through structured arg declarations
	for _, arg := range argBlock.Args {
		// Check if this is an int type argument
		if arg.Type.Type != rl.T_INT {
			continue
		}

		// Check if it has a default value
		if arg.Default == nil {
			continue
		}

		// Check if the default value node contains scientific notation
		valueNode := arg.Default.Node().ChildByFieldName(rl.F_VALUE)
		if valueNode == nil {
			continue
		}
		if valueNode.Kind() != rl.K_SCIENTIFIC_NUMBER {
			continue
		}

		c.validateScientificNumberAsInt(valueNode, d)
	}
}

func (c *RadCheckerImpl) addFnParamScientificNotationErrors(d *[]Diagnostic) {
	// Walk the tree to find all function definitions
	root := c.tree.Root()
	c.walkForFunctions(root, d)
}

func (c *RadCheckerImpl) walkForFunctions(node *ts.Node, d *[]Diagnostic) {
	if node == nil {
		return
	}

	// Check if this node is a function definition
	if node.Kind() == rl.K_FN_NAMED || node.Kind() == rl.K_FN_LAMBDA {
		c.checkFunctionParams(node, d)
	}

	// Recursively walk children
	for i := uint(0); i < node.ChildCount(); i++ {
		c.walkForFunctions(node.Child(i), d)
	}
}

func (c *RadCheckerImpl) checkFunctionParams(fnNode *ts.Node, d *[]Diagnostic) {
	// Get all parameter nodes
	normalParams := fnNode.ChildrenByFieldName(rl.F_NORMAL_PARAM, fnNode.Walk())
	namedOnlyParams := fnNode.ChildrenByFieldName(rl.F_NAMED_ONLY_PARAM, fnNode.Walk())
	varargParams := fnNode.ChildrenByFieldName(rl.F_VARARG_PARAM, fnNode.Walk())

	allParams := append(append(normalParams, namedOnlyParams...), varargParams...)

	for _, param := range allParams {
		// Check if parameter has int type
		typeNode := param.ChildByFieldName(rl.F_TYPE)
		if typeNode == nil {
			continue
		}

		// Navigate to the actual type node (could be nested in union/optional)
		leafTypeNode := typeNode.ChildByFieldName(rl.F_LEAF_TYPE)
		if leafTypeNode != nil {
			typeNode = leafTypeNode.ChildByFieldName(rl.F_TYPE)
			if typeNode == nil {
				continue
			}
		}

		if typeNode.Kind() != rl.K_INT_TYPE {
			continue
		}

		// Check if parameter has a default value
		defaultNode := param.ChildByFieldName(rl.F_DEFAULT)
		if defaultNode == nil {
			continue
		}

		// Find scientific_number in the default expression
		c.checkExprForScientificNumber(defaultNode, d)
	}
}

func (c *RadCheckerImpl) checkExprForScientificNumber(exprNode *ts.Node, d *[]Diagnostic) {
	if exprNode == nil {
		return
	}

	// If this is a scientific_number node, validate it
	if exprNode.Kind() == rl.K_SCIENTIFIC_NUMBER {
		c.validateScientificNumberAsInt(exprNode, d)
		return
	}

	// Recursively check children
	for i := uint(0); i < exprNode.ChildCount(); i++ {
		c.checkExprForScientificNumber(exprNode.Child(i), d)
	}
}

func (c *RadCheckerImpl) validateScientificNumberAsInt(node *ts.Node, d *[]Diagnostic) {
	valueStr := c.src[node.StartByte():node.EndByte()]
	floatVal, err := rts.ParseFloat(valueStr)
	if err != nil {
		return // parsing error will be caught elsewhere
	}

	if floatVal != float64(int64(floatVal)) {
		msg := "Scientific notation value does not evaluate to a whole number"
		*d = append(*d, NewDiagnosticError(node, c.src, msg, rl.ErrScientificNotationNotWholeNumber))
	}
}

func (c *RadCheckerImpl) addFunctionNameShadowingErrors(d *[]Diagnostic) {
	// Get argument names from the args block
	argBlock, ok := c.tree.FindArgBlock()
	if !ok {
		// No args block, so no shadowing is possible
		return
	}

	// Collect all argument names into a map for quick lookup
	argNames := make(map[string]bool)
	for _, arg := range argBlock.Args {
		argNames[arg.Name.Name] = true
	}

	hoistedFunctions := c.getHoistedFunctions()
	for _, fnName := range hoistedFunctions {
		if argNames[fnName] {
			// Find the node for the error reporting
			root := c.tree.Root()
			for i := uint(0); i < root.ChildCount(); i++ {
				child := root.Child(i)
				if child.Kind() == rl.K_FN_NAMED {
					nameNode := child.ChildByFieldName(rl.F_NAME)
					if nameNode != nil {
						name := c.src[nameNode.StartByte():nameNode.EndByte()]
						if name == fnName {
							msg := "Hoisted function '" + fnName + "' shadows an argument with the same name"
							*d = append(*d, NewDiagnosticError(nameNode, c.src, msg, rl.ErrHoistedFunctionShadowsArgument))
							break
						}
					}
				}
			}
		}
	}
}

func (c *RadCheckerImpl) getHoistedFunctions() []string {
	hoistedFunctions := make([]string, 0)

	// Only check top-level functions (direct children of source file)
	// This matches the interpreter's hoisting behavior
	root := c.tree.Root()
	for i := uint(0); i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child.Kind() == rl.K_FN_NAMED {
			nameNode := child.ChildByFieldName(rl.F_NAME)
			if nameNode != nil {
				fnName := c.src[nameNode.StartByte():nameNode.EndByte()]
				hoistedFunctions = append(hoistedFunctions, fnName)
			}
		}
	}

	return hoistedFunctions
}

func (c *RadCheckerImpl) addUnknownFunctionHints(d *[]Diagnostic) {
	// Get built-in functions from singleton
	builtInFunctions := rts.GetBuiltInFunctions()

	// Get hoisted functions
	hoistedFunctions := c.getHoistedFunctions()
	hoistedFunctionSet := make(map[string]bool)
	for _, fnName := range hoistedFunctions {
		hoistedFunctionSet[fnName] = true
	}

	calls := c.tree.FindCalls()
	for _, call := range calls {
		fnName := call.Name

		if builtInFunctions.Contains(fnName) || hoistedFunctionSet[fnName] {
			continue
		}

		// Function is not defined - create a hint
		msg := "Function '" + fnName + "' may not be defined (only built-in and top-level functions are tracked)"
		*d = append(*d, NewDiagnosticHint(call.NameNode, c.src, msg, rl.ErrUnknownFunction))
	}
}

func (c *RadCheckerImpl) addUnknownCommandCallbackWarnings(d *[]Diagnostic) {
	builtInFunctions := rts.GetBuiltInFunctions()

	hoistedFunctions := c.getHoistedFunctions()
	hoistedFunctionSet := make(map[string]bool)
	for _, fnName := range hoistedFunctions {
		hoistedFunctionSet[fnName] = true
	}

	cmdBlocks, ok := c.tree.FindCmdBlocks()
	if !ok {
		return
	}

	for _, cmdBlock := range cmdBlocks {
		callback := cmdBlock.Callback

		// Only check identifier callbacks (not inline lambdas)
		if callback.Type != rts.CallbackIdentifier {
			continue
		}

		if callback.IdentifierName == nil {
			continue
		}

		fnName := *callback.IdentifierName

		if builtInFunctions.Contains(fnName) || hoistedFunctionSet[fnName] {
			continue
		}

		identifierNode := callback.Node().ChildByFieldName(rl.F_CALLBACK_IDENTIFIER)
		if identifierNode == nil {
			continue
		}

		msg := "Function '" + fnName + "' may not be defined (only built-in and top-level functions are tracked)"
		*d = append(*d, NewDiagnosticWarn(identifierNode, c.src, msg, rl.ErrUnknownFunction))
	}
}
