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
	return Result{
		Diagnostics: diagnostics,
	}, nil
}

func (c *RadCheckerImpl) addInvalidNodes(d *[]Diagnostic) {
	nodes := c.tree.FindInvalidNodes()
	for _, node := range nodes {
		*d = append(*d, NewDiagnosticError(node, c.src, "Invalid syntax", rl.ErrInvalidSyntax))
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
