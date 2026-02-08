package rts

import (
	"fmt"
	"math"
	"strconv"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// ConvertCST converts a tree-sitter CST into a Go-native AST.
// The input CST must be valid (no ERROR/MISSING nodes) - if the
// converter encounters an unexpected node kind, it panics.
func ConvertCST(root *ts.Node, src string, file string) *rl.SourceFile {
	c := &converter{src: src, file: file}
	return c.convertSourceFile(root)
}

// ConvertLambda converts a single lambda CST node into an AST Lambda.
// Used for converting command callback lambdas at the runner level.
func ConvertLambda(lambdaNode *ts.Node, src string, file string) *rl.Lambda {
	c := &converter{src: src, file: file}
	node := c.convertFnLambda(lambdaNode)
	return node
}

type converter struct {
	src  string
	file string
}

func (c *converter) makeSpan(node *ts.Node) rl.Span {
	return rl.Span{
		File:      c.file,
		StartByte: int(node.StartByte()),
		EndByte:   int(node.EndByte()),
		StartRow:  int(node.StartPosition().Row),
		StartCol:  int(node.StartPosition().Column),
		EndRow:    int(node.EndPosition().Row),
		EndCol:    int(node.EndPosition().Column),
	}
}

func (c *converter) getSrc(node *ts.Node) string {
	return c.src[node.StartByte():node.EndByte()]
}

// --- Source File ---

func (c *converter) convertSourceFile(node *ts.Node) *rl.SourceFile {
	children := node.Children(node.Walk())
	stmts := make([]rl.Node, 0, len(children))
	for _, child := range children {
		kind := child.Kind()
		// Skip non-executable top-level nodes
		if kind == rl.K_COMMENT || kind == rl.K_SHEBANG ||
			kind == rl.K_FILE_HEADER || kind == rl.K_ARG_BLOCK || kind == rl.K_CMD_BLOCK {
			continue
		}
		stmts = append(stmts, c.convertStmt(&child))
	}
	return rl.NewSourceFile(c.makeSpan(node), stmts)
}

// --- Statements ---

func (c *converter) convertStmt(node *ts.Node) rl.Node {
	switch node.Kind() {
	case rl.K_ASSIGN:
		return c.convertAssign(node)
	case rl.K_COMPOUND_ASSIGN:
		return c.convertCompoundAssign(node)
	case rl.K_INCR_DECR:
		return c.convertIncrDecr(node)
	case rl.K_EXPR_STMT:
		return c.convertExprStmt(node)
	case rl.K_IF_STMT:
		return c.convertIf(node)
	case rl.K_SWITCH_STMT:
		return c.convertSwitch(node)
	case rl.K_FOR_LOOP:
		return c.convertForLoop(node)
	case rl.K_WHILE_LOOP:
		return c.convertWhileLoop(node)
	case rl.K_SHELL_STMT:
		return c.convertShellStmt(node)
	case rl.K_DEL_STMT:
		return c.convertDel(node)
	case rl.K_DEFER_BLOCK:
		return c.convertDefer(node)
	case rl.K_BREAK_STMT:
		return rl.NewBreak(c.makeSpan(node))
	case rl.K_CONTINUE_STMT:
		return rl.NewContinue(c.makeSpan(node))
	case rl.K_RETURN_STMT:
		return c.convertReturn(node)
	case rl.K_YIELD_STMT:
		return c.convertYield(node)
	case rl.K_PASS:
		return rl.NewPass(c.makeSpan(node))
	case rl.K_FN_NAMED:
		return c.convertFnDef(node)
	case rl.K_RAD_BLOCK:
		return c.convertRadBlock(node)

	// Expression node kinds that can appear as statements
	case rl.K_EXPR:
		return c.convertExpr(node)

	default:
		panic(fmt.Sprintf("converter: unexpected statement node kind: %s", node.Kind()))
	}
}

func (c *converter) convertAssign(node *ts.Node) *rl.Assign {
	catchNode := rl.GetChild(node, rl.F_CATCH)
	var catch *rl.CatchBlock
	if catchNode != nil {
		catch = c.convertCatchBlock(catchNode)
	}

	rightNodes := rl.GetChildren(node, rl.F_RIGHT)
	values := c.convertExprs(rightNodes)

	leftNodes := rl.GetChildren(node, rl.F_LEFT)
	if len(leftNodes) > 0 {
		targets := c.convertExprs(leftNodes)
		return rl.NewAssign(c.makeSpan(node), targets, values, false, catch)
	}

	leftsNodes := rl.GetChildren(node, rl.F_LEFTS)
	targets := c.convertExprs(leftsNodes)
	return rl.NewAssign(c.makeSpan(node), targets, values, true, catch)
}

// convertCompoundAssign desugars `x += 3` into `Assign(x, OpBinary(x, +, 3))`.
func (c *converter) convertCompoundAssign(node *ts.Node) *rl.Assign {
	leftNode := rl.GetChild(node, rl.F_LEFT)
	rightNode := rl.GetChild(node, rl.F_RIGHT)
	opNode := rl.GetChild(node, rl.F_OP)

	op := c.resolveCompoundOp(opNode)
	target := c.convertExpr(leftNode)
	rightVal := c.convertExpr(rightNode)

	// Create a synthetic binary op: target op rightVal
	binOp := rl.NewOpBinary(c.makeSpan(node), op, target, rightVal)

	return rl.NewAssign(c.makeSpan(node), []rl.Node{target}, []rl.Node{binOp}, false, nil)
}

// convertIncrDecr desugars `x++` into `Assign(x, OpBinary(x, +, 1))`.
func (c *converter) convertIncrDecr(node *ts.Node) *rl.Assign {
	leftNode := rl.GetChild(node, rl.F_LEFT)
	opNode := rl.GetChild(node, rl.F_OP)

	target := c.convertExpr(leftNode)
	span := c.makeSpan(node)
	one := rl.NewLitInt(span, 1)

	var op rl.Operator
	switch opNode.Kind() {
	case rl.K_PLUS_PLUS:
		op = rl.OpAdd
	case rl.K_MINUS_MINUS:
		op = rl.OpSub
	default:
		panic(fmt.Sprintf("converter: unexpected incr/decr op: %s", opNode.Kind()))
	}

	binOp := rl.NewOpBinary(span, op, target, one)
	return rl.NewAssign(span, []rl.Node{target}, []rl.Node{binOp}, false, nil)
}

func (c *converter) convertExprStmt(node *ts.Node) *rl.ExprStmt {
	exprNode := rl.GetChild(node, rl.F_EXPR)
	catchNode := rl.GetChild(node, rl.F_CATCH)
	var catch *rl.CatchBlock
	if catchNode != nil {
		catch = c.convertCatchBlock(catchNode)
	}
	return rl.NewExprStmt(c.makeSpan(node), c.convertExpr(exprNode), catch)
}

func (c *converter) convertCatchBlock(node *ts.Node) *rl.CatchBlock {
	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	stmts := make([]rl.Node, 0, len(stmtNodes))
	for _, stmtNode := range stmtNodes {
		stmts = append(stmts, c.convertStmt(&stmtNode))
	}
	return rl.NewCatchBlock(c.makeSpan(node), stmts)
}

func (c *converter) convertIf(node *ts.Node) *rl.If {
	altNodes := rl.GetChildren(node, rl.F_ALT)
	branches := make([]rl.IfBranch, 0, len(altNodes))
	for _, altNode := range altNodes {
		condNode := rl.GetChild(&altNode, rl.F_CONDITION)
		var condition rl.Node
		if condNode != nil {
			condition = c.convertExpr(condNode)
		}
		stmtNodes := rl.GetChildren(&altNode, rl.F_STMT)
		body := c.convertStmts(stmtNodes)
		branches = append(branches, rl.IfBranch{Condition: condition, Body: body})
	}
	return rl.NewIf(c.makeSpan(node), branches)
}

func (c *converter) convertSwitch(node *ts.Node) *rl.Switch {
	discriminantNode := rl.GetChild(node, rl.F_DISCRIMINANT)
	discriminant := c.convertExpr(discriminantNode)

	caseNodes := rl.GetChildren(node, rl.F_CASE)
	cases := make([]rl.SwitchCase, 0, len(caseNodes))
	for _, caseNode := range caseNodes {
		caseKeyNodes := rl.GetChildren(&caseNode, rl.F_CASE_KEY)
		keys := c.convertExprs(caseKeyNodes)
		altNode := rl.GetChild(&caseNode, rl.F_ALT)
		alt := c.convertSwitchAlt(altNode)
		cases = append(cases, rl.SwitchCase{Keys: keys, Alt: alt})
	}

	defaultNode := rl.GetChild(node, rl.F_DEFAULT)
	var dflt *rl.SwitchDefault
	if defaultNode != nil {
		altNode := rl.GetChild(defaultNode, rl.F_ALT)
		alt := c.convertSwitchAlt(altNode)
		dflt = &rl.SwitchDefault{Alt: alt}
	}

	return rl.NewSwitch(c.makeSpan(node), discriminant, cases, dflt)
}

func (c *converter) convertSwitchAlt(node *ts.Node) rl.Node {
	switch node.Kind() {
	case rl.K_SWITCH_CASE_EXPR:
		rightNodes := rl.GetChildren(node, rl.F_RIGHT)
		values := c.convertExprs(rightNodes)
		return rl.NewSwitchCaseExpr(c.makeSpan(node), values)
	case rl.K_SWITCH_CASE_BLOCK:
		stmtNodes := rl.GetChildren(node, rl.F_STMT)
		stmts := c.convertStmts(stmtNodes)
		return rl.NewSwitchCaseBlock(c.makeSpan(node), stmts)
	default:
		panic(fmt.Sprintf("converter: unexpected switch alt kind: %s", node.Kind()))
	}
}

func (c *converter) convertForLoop(node *ts.Node) *rl.ForLoop {
	leftsNode := rl.GetChild(node, rl.F_LEFTS)
	rightNode := rl.GetChild(node, rl.F_RIGHT)
	contextNode := rl.GetChild(node, rl.F_CONTEXT)

	leftNodes := rl.GetChildren(leftsNode, rl.F_LEFT)
	vars := make([]string, 0, len(leftNodes))
	for _, leftNode := range leftNodes {
		vars = append(vars, c.getSrc(&leftNode))
	}

	iter := c.convertExpr(rightNode)

	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	body := c.convertStmts(stmtNodes)

	var context *string
	if contextNode != nil {
		ctx := c.getSrc(contextNode)
		context = &ctx
	}

	return rl.NewForLoop(c.makeSpan(node), vars, iter, body, context)
}

func (c *converter) convertWhileLoop(node *ts.Node) *rl.WhileLoop {
	condNode := rl.GetChild(node, rl.F_CONDITION)
	var condition rl.Node
	if condNode != nil {
		condition = c.convertExpr(condNode)
	}

	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	body := c.convertStmts(stmtNodes)

	return rl.NewWhileLoop(c.makeSpan(node), condition, body)
}

func (c *converter) convertShellStmt(node *ts.Node) *rl.Shell {
	leftNode := rl.GetChildren(node, rl.F_LEFT)
	leftNodes := rl.GetChildren(node, rl.F_LEFTS)
	leftNodes = append(leftNode, leftNodes...)

	targets := c.convertExprs(leftNodes)

	shellCmdNode := rl.GetChild(node, rl.F_SHELL_CMD)

	// Extract modifiers
	modifierNodes := rl.GetChildren(shellCmdNode, rl.F_MODIFIER)
	var isQuiet, isConfirm bool
	for _, modNode := range modifierNodes {
		modText := c.getSrc(&modNode)
		switch modText {
		case "quiet":
			isQuiet = true
		case "confirm":
			isConfirm = true
		}
	}

	cmdNode := rl.GetChild(shellCmdNode, rl.F_COMMAND)
	cmd := c.convertExpr(cmdNode)

	catchNode := rl.GetChild(node, rl.F_CATCH)
	var catch *rl.CatchBlock
	if catchNode != nil {
		catch = c.convertCatchBlock(catchNode)
	}

	return rl.NewShell(c.makeSpan(node), targets, cmd, catch, isQuiet, isConfirm)
}

func (c *converter) convertDel(node *ts.Node) *rl.Del {
	rightNodes := rl.GetChildren(node, rl.F_RIGHT)
	targets := c.convertExprs(rightNodes)
	return rl.NewDel(c.makeSpan(node), targets)
}

func (c *converter) convertDefer(node *ts.Node) *rl.Defer {
	keywordNode := rl.GetChild(node, rl.F_KEYWORD)
	keywordStr := c.getSrc(keywordNode)
	isErrDefer := keywordStr == "errdefer"

	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	body := c.convertStmts(stmtNodes)

	return rl.NewDefer(c.makeSpan(node), isErrDefer, body)
}

func (c *converter) convertReturn(node *ts.Node) *rl.Return {
	rightNodes := rl.GetChildren(node, rl.F_RIGHT)
	values := c.convertExprs(rightNodes)
	return rl.NewReturn(c.makeSpan(node), values)
}

func (c *converter) convertYield(node *ts.Node) *rl.Yield {
	rightNodes := rl.GetChildren(node, rl.F_RIGHT)
	values := c.convertExprs(rightNodes)
	return rl.NewYield(c.makeSpan(node), values)
}

func (c *converter) convertFnDef(node *ts.Node) *rl.FnDef {
	nameNode := rl.GetChild(node, rl.F_NAME)
	name := c.getSrc(nameNode)

	typing := rl.NewTypingFnT(node, c.src)

	// Set DefaultAST on params that have defaults
	c.convertParamDefaults(typing, node)

	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	body := c.convertStmts(stmtNodes)

	isBlock := rl.GetChild(node, rl.F_BLOCK_COLON) != nil

	defSpan := c.makeSpan(node)
	if isBlock {
		keywordNode := rl.GetChild(node, rl.F_KEYWORD)
		if keywordNode != nil {
			defSpan = c.makeSpan(keywordNode)
		}
	}

	return rl.NewFnDef(c.makeSpan(node), name, typing, body, isBlock, defSpan)
}

func (c *converter) convertFnLambda(node *ts.Node) *rl.Lambda {
	typing := rl.NewTypingFnT(node, c.src)

	// Set DefaultAST on params that have defaults
	c.convertParamDefaults(typing, node)

	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	body := c.convertStmts(stmtNodes)

	isBlock := rl.GetChild(node, rl.F_BLOCK_COLON) != nil

	defSpan := c.makeSpan(node)
	if isBlock {
		keywordNode := rl.GetChild(node, rl.F_KEYWORD)
		if keywordNode != nil {
			defSpan = c.makeSpan(keywordNode)
		}
	}

	return rl.NewLambda(c.makeSpan(node), typing, body, isBlock, defSpan)
}

// convertParamDefaults sets DefaultAST on TypingFnParams that have CST-based defaults.
func (c *converter) convertParamDefaults(typing *rl.TypingFnT, fnNode *ts.Node) {
	for i := range typing.Params {
		param := &typing.Params[i]
		if param.Default != nil {
			astNode := c.convertExpr(param.Default.Node)
			param.DefaultAST = &rl.ASTDefault{
				Node: astNode,
				Src:  param.Default.Src,
			}
		}
	}
}

// --- Expressions ---

func (c *converter) convertExpr(node *ts.Node) rl.Node {
	// Collapse delegate chains: if the node has a delegate child, skip to it
	if delegate := rl.GetChild(node, rl.F_DELEGATE); delegate != nil {
		return c.convertExpr(delegate)
	}

	switch node.Kind() {
	// Delegate-capable expression wrappers (handled by delegate check above,
	// but may also have actual content when delegate is nil)
	case rl.K_EXPR:
		// Should always have a delegate, but be defensive
		return c.convertExpr(rl.GetChild(node, rl.F_DELEGATE))

	case rl.K_TERNARY_EXPR:
		return c.convertTernary(node)
	case rl.K_OR_EXPR, rl.K_AND_EXPR, rl.K_COMPARE_EXPR, rl.K_ADD_EXPR, rl.K_MULT_EXPR:
		return c.convertBinaryExpr(node)
	case rl.K_UNARY_EXPR:
		return c.convertUnaryExpr(node)
	case rl.K_FALLBACK_EXPR:
		return c.convertFallback(node)
	case rl.K_INDEXED_EXPR:
		return c.convertIndexedExpr(node)

	// Structural wrappers (collapsed by the converter)
	case rl.K_PRIMARY_EXPR, rl.K_LITERAL:
		return c.convertExpr(c.getOnlyChild(node))
	case rl.K_PARENTHESIZED_EXPR:
		return c.convertExpr(rl.GetChild(node, rl.F_EXPR))

	// Leaf expressions
	case rl.K_IDENTIFIER:
		return rl.NewIdentifier(c.makeSpan(node), c.getSrc(node))
	case rl.K_VAR_PATH:
		return c.convertVarPath(node)
	case rl.K_INT:
		return c.convertInt(node)
	case rl.K_FLOAT:
		return c.convertFloat(node)
	case rl.K_SCIENTIFIC_NUMBER:
		return c.convertScientificNumber(node)
	case rl.K_BOOL:
		return c.convertBool(node)
	case rl.K_NULL:
		return rl.NewLitNull(c.makeSpan(node))
	case rl.K_STRING:
		return c.convertString(node)
	case rl.K_LIST:
		return c.convertList(node)
	case rl.K_MAP:
		return c.convertMap(node)
	case rl.K_CALL:
		return c.convertCall(node)
	case rl.K_FN_LAMBDA:
		return c.convertFnLambda(node)
	case rl.K_LIST_COMPREHENSION:
		return c.convertListComp(node)
	case rl.K_JSON_PATH:
		// JSON paths stay as identifiers - they're resolved at runtime
		return rl.NewIdentifier(c.makeSpan(node), c.getSrc(node))

	default:
		panic(fmt.Sprintf("converter: unexpected expression node kind: %s", node.Kind()))
	}
}

func (c *converter) convertTernary(node *ts.Node) rl.Node {
	condNode := rl.GetChild(node, rl.F_CONDITION)
	trueNode := rl.GetChild(node, rl.F_TRUE_BRANCH)
	falseNode := rl.GetChild(node, rl.F_FALSE_BRANCH)
	return rl.NewTernary(c.makeSpan(node),
		c.convertExpr(condNode),
		c.convertExpr(trueNode),
		c.convertExpr(falseNode))
}

func (c *converter) convertBinaryExpr(node *ts.Node) rl.Node {
	leftNode := rl.GetChild(node, rl.F_LEFT)
	rightNode := rl.GetChild(node, rl.F_RIGHT)
	opNode := rl.GetChild(node, rl.F_OP)

	op := c.resolveBinaryOp(opNode)
	return rl.NewOpBinary(c.makeSpan(node),
		op,
		c.convertExpr(leftNode),
		c.convertExpr(rightNode))
}

func (c *converter) convertUnaryExpr(node *ts.Node) rl.Node {
	opNode := rl.GetChild(node, rl.F_OP)
	argNode := rl.GetChild(node, rl.F_ARG)

	op := c.resolveUnaryOp(opNode)
	return rl.NewOpUnary(c.makeSpan(node), op, c.convertExpr(argNode))
}

func (c *converter) convertFallback(node *ts.Node) rl.Node {
	leftNode := rl.GetChild(node, rl.F_LEFT)
	rightNode := rl.GetChild(node, rl.F_RIGHT)
	return rl.NewFallback(c.makeSpan(node),
		c.convertExpr(leftNode),
		c.convertExpr(rightNode))
}

func (c *converter) convertIndexedExpr(node *ts.Node) rl.Node {
	rootNode := rl.GetChild(node, rl.F_ROOT)
	indexingNodes := rl.GetChildren(node, rl.F_INDEXING)

	if len(indexingNodes) == 0 {
		return c.convertExpr(rootNode)
	}

	// Build a VarPath with segments
	root := c.convertExpr(rootNode)
	segments := make([]rl.PathSegment, 0, len(indexingNodes))
	for _, indexNode := range indexingNodes {
		seg := c.convertPathSegment(&indexNode)
		segments = append(segments, seg)
	}

	return rl.NewVarPath(c.makeSpan(node), root, segments)
}

func (c *converter) convertVarPath(node *ts.Node) rl.Node {
	rootNode := rl.GetChild(node, rl.F_ROOT)
	indexingNodes := rl.GetChildren(node, rl.F_INDEXING)

	root := c.convertExpr(rootNode)

	if len(indexingNodes) == 0 {
		return root
	}

	segments := make([]rl.PathSegment, 0, len(indexingNodes))
	for _, indexNode := range indexingNodes {
		seg := c.convertPathSegment(&indexNode)
		segments = append(segments, seg)
	}

	return rl.NewVarPath(c.makeSpan(node), root, segments)
}

func (c *converter) convertPathSegment(node *ts.Node) rl.PathSegment {
	span := c.makeSpan(node)

	// Check if this is a call (UFCS)
	if node.Kind() == rl.K_CALL {
		// For UFCS calls, we store the entire call as an index expression
		call := c.convertCall(node)
		return rl.NewPathSegmentIndex(span, call)
	}

	// Check for slice syntax
	if node.Kind() == rl.K_SLICE {
		startNode := rl.GetChild(node, rl.F_START)
		endNode := rl.GetChild(node, rl.F_END)
		var start, end rl.Node
		if startNode != nil {
			start = c.convertExpr(startNode)
		}
		if endNode != nil {
			end = c.convertExpr(endNode)
		}
		return rl.NewPathSegmentSlice(span, start, end)
	}

	// Check for dot access (identifier child)
	if node.Kind() == rl.K_IDENTIFIER {
		fieldName := c.getSrc(node)
		return rl.NewPathSegmentField(span, fieldName)
	}

	// Bracket index access
	return rl.NewPathSegmentIndex(span, c.convertExpr(node))
}

func (c *converter) convertCall(node *ts.Node) *rl.Call {
	funcNode := rl.GetChild(node, rl.F_FUNC)
	argNodes := rl.GetChildren(node, rl.F_ARG)
	namedArgNodes := rl.GetChildren(node, rl.F_NAMED_ARG)

	fn := c.convertExpr(funcNode)
	args := c.convertExprs(argNodes)

	namedArgs := make([]rl.CallNamedArg, 0, len(namedArgNodes))
	for _, namedArgNode := range namedArgNodes {
		nameNode := rl.GetChild(&namedArgNode, rl.F_NAME)
		valueNode := rl.GetChild(&namedArgNode, rl.F_VALUE)
		namedArgs = append(namedArgs, rl.CallNamedArg{
			Name:      c.getSrc(nameNode),
			NameSpan:  c.makeSpan(nameNode),
			Value:     c.convertExpr(valueNode),
			ValueSpan: c.makeSpan(valueNode),
		})
	}

	return rl.NewCall(c.makeSpan(node), fn, args, namedArgs)
}

// --- Literals ---

func (c *converter) convertInt(node *ts.Node) rl.Node {
	src := c.getSrc(node)
	val, err := ParseInt(src)
	if err != nil {
		panic(fmt.Sprintf("converter: failed to parse int %q: %v", src, err))
	}
	return rl.NewLitInt(c.makeSpan(node), val)
}

func (c *converter) convertFloat(node *ts.Node) rl.Node {
	src := c.getSrc(node)
	val, err := ParseFloat(src)
	if err != nil {
		panic(fmt.Sprintf("converter: failed to parse float %q: %v", src, err))
	}
	return rl.NewLitFloat(c.makeSpan(node), val)
}

func (c *converter) convertScientificNumber(node *ts.Node) rl.Node {
	src := c.getSrc(node)
	val, err := ParseFloat(src)
	if err != nil {
		panic(fmt.Sprintf("converter: failed to parse scientific number %q: %v", src, err))
	}
	// Evaluate as int if it's a whole number that fits in int64, float otherwise.
	// The int64 range check prevents silent overflow for large values like 9.2e18.
	if val == float64(int64(val)) && !math.IsInf(val, 0) &&
		val >= math.MinInt64 && val <= math.MaxInt64 {
		return rl.NewLitInt(c.makeSpan(node), int64(val))
	}
	return rl.NewLitFloat(c.makeSpan(node), val)
}

func (c *converter) convertBool(node *ts.Node) rl.Node {
	src := c.getSrc(node)
	val, err := strconv.ParseBool(src)
	if err != nil {
		panic(fmt.Sprintf("converter: failed to parse bool %q: %v", src, err))
	}
	return rl.NewLitBool(c.makeSpan(node), val)
}

func (c *converter) convertString(node *ts.Node) rl.Node {
	contentsNode := rl.GetChild(node, rl.F_CONTENTS)
	span := c.makeSpan(node)

	// Determine the delimiter character from the end node
	endNode := rl.GetChild(node, rl.F_END)
	endStr := c.getSrc(endNode)
	delimiter := endStr[len(endStr)-1]

	if contentsNode == nil {
		// Empty string
		return rl.NewLitStringSimple(span, "")
	}

	children := contentsNode.Children(contentsNode.Walk())

	// Check if this is a simple string (no interpolation, just literal text)
	hasInterpolation := false
	for _, child := range children {
		if child.Kind() == rl.K_INTERPOLATION {
			hasInterpolation = true
			break
		}
	}

	if !hasInterpolation {
		// Simple string - resolve all escape sequences and concatenate
		var result string
		for _, child := range children {
			result += c.resolveStringPart(&child, delimiter)
		}
		return rl.NewLitStringSimple(span, result)
	}

	// Interpolated string - build segments
	segments := make([]rl.StringSegment, 0, len(children))

	// Accumulate consecutive literal parts into one segment
	var literalBuf string
	flushLiteral := func() {
		if literalBuf != "" {
			segments = append(segments, rl.StringSegment{IsLiteral: true, Text: literalBuf})
			literalBuf = ""
		}
	}

	for _, child := range children {
		if child.Kind() == rl.K_INTERPOLATION {
			flushLiteral()
			seg := c.convertInterpolation(&child)
			segments = append(segments, seg)
		} else {
			literalBuf += c.resolveStringPart(&child, delimiter)
		}
	}
	flushLiteral()

	return rl.NewLitStringInterpolated(span, segments)
}

// resolveStringPart resolves a single string content child (escape sequence or literal text).
func (c *converter) resolveStringPart(node *ts.Node, delimiter byte) string {
	switch node.Kind() {
	case rl.K_STRING_CONTENT:
		return c.getSrc(node)
	case rl.K_BACKSLASH:
		return "\\"
	case rl.K_ESC_BACKSLASH:
		return "\\"
	case rl.K_ESC_SINGLE_QUOTE:
		if delimiter == '\'' {
			return "'"
		}
		return `\'`
	case rl.K_ESC_DOUBLE_QUOTE:
		if delimiter == '"' {
			return `"`
		}
		return `\"`
	case rl.K_ESC_BACKTICK:
		if delimiter == '`' {
			return "`"
		}
		return "\\`"
	case rl.K_ESC_NEWLINE:
		return "\n"
	case rl.K_ESC_TAB:
		return "\t"
	case rl.K_ESC_OPEN_BRACKET:
		return "{"
	default:
		panic(fmt.Sprintf("converter: unexpected string part kind: %s", node.Kind()))
	}
}

func (c *converter) convertInterpolation(node *ts.Node) rl.StringSegment {
	exprNode := rl.GetChild(node, rl.F_EXPR)
	formatNode := rl.GetChild(node, rl.F_FORMAT)

	var format *rl.InterpolationFormat
	if formatNode != nil {
		format = c.convertInterpolationFormat(formatNode)
	}

	return rl.StringSegment{
		IsLiteral: false,
		Expr:      c.convertExpr(exprNode),
		Format:    format,
	}
}

func (c *converter) convertInterpolationFormat(node *ts.Node) *rl.InterpolationFormat {
	thousandsSepNode := rl.GetChild(node, rl.F_THOUSANDS_SEPARATOR)
	alignmentNode := rl.GetChild(node, rl.F_ALIGNMENT)
	paddingNode := rl.GetChild(node, rl.F_PADDING)
	precisionNode := rl.GetChild(node, rl.F_PRECISION)

	format := &rl.InterpolationFormat{
		ThousandsSeparator: thousandsSepNode != nil,
	}

	if alignmentNode != nil {
		format.Alignment = c.getSrc(alignmentNode)
	}

	if paddingNode != nil {
		format.Padding = c.convertExpr(paddingNode)
	}

	if precisionNode != nil {
		format.Precision = c.convertExpr(precisionNode)
	}

	return format
}

func (c *converter) convertList(node *ts.Node) rl.Node {
	entryNodes := rl.GetChildren(node, rl.F_LIST_ENTRY)
	elements := c.convertExprs(entryNodes)
	return rl.NewLitList(c.makeSpan(node), elements)
}

func (c *converter) convertMap(node *ts.Node) rl.Node {
	entryNodes := rl.GetChildren(node, rl.F_MAP_ENTRY)
	entries := make([]rl.MapEntry, 0, len(entryNodes))
	for _, entryNode := range entryNodes {
		keyNode := rl.GetChild(&entryNode, rl.F_KEY)
		valueNode := rl.GetChild(&entryNode, rl.F_VALUE)
		entries = append(entries, rl.MapEntry{
			Key:   c.convertExpr(keyNode),
			Value: c.convertExpr(valueNode),
		})
	}
	return rl.NewLitMap(c.makeSpan(node), entries)
}

func (c *converter) convertListComp(node *ts.Node) rl.Node {
	exprNode := rl.GetChild(node, rl.F_EXPR)
	conditionNode := rl.GetChild(node, rl.F_CONDITION)
	contextNode := rl.GetChild(node, rl.F_CONTEXT)

	// Extract loop vars from the for-loop structure embedded in the list comp
	leftsNode := rl.GetChild(node, rl.F_LEFTS)
	rightNode := rl.GetChild(node, rl.F_RIGHT)

	leftNodes := rl.GetChildren(leftsNode, rl.F_LEFT)
	vars := make([]string, 0, len(leftNodes))
	for _, leftNode := range leftNodes {
		vars = append(vars, c.getSrc(&leftNode))
	}

	iter := c.convertExpr(rightNode)

	var condition rl.Node
	if conditionNode != nil {
		condition = c.convertExpr(conditionNode)
	}

	expr := c.convertExpr(exprNode)

	var context *string
	if contextNode != nil {
		ctx := c.getSrc(contextNode)
		context = &ctx
	}

	return rl.NewListComp(c.makeSpan(node), expr, vars, iter, condition, context)
}

// --- Rad block ---

func (c *converter) convertRadBlock(node *ts.Node) rl.Node {
	srcNode := rl.GetChild(node, rl.F_SOURCE)
	radTypeNode := rl.GetChild(node, rl.F_RAD_TYPE)
	typeStr := c.getSrc(radTypeNode)

	var source rl.Node
	if srcNode != nil {
		source = c.convertExpr(srcNode)
	}

	stmtNodes := rl.GetChildren(node, rl.F_STMT)
	stmts := make([]rl.Node, 0, len(stmtNodes))
	for _, stmtNode := range stmtNodes {
		stmts = append(stmts, c.convertRadStmt(&stmtNode))
	}

	return rl.NewRadBlock(c.makeSpan(node), typeStr, source, stmts)
}

func (c *converter) convertRadStmt(node *ts.Node) rl.Node {
	switch node.Kind() {
	case rl.K_RAD_FIELD_STMT:
		identifierNodes := rl.GetChildren(node, rl.F_IDENTIFIER)
		ids := make([]rl.Node, 0, len(identifierNodes))
		for _, idNode := range identifierNodes {
			ids = append(ids, rl.NewIdentifier(c.makeSpan(&idNode), c.getSrc(&idNode)))
		}
		return rl.NewRadField(c.makeSpan(node), ids)

	case rl.K_RAD_SORT_STMT:
		specifierNodes := rl.GetChildren(node, rl.F_SPECIFIER)
		specifiers := make([]rl.RadSortSpecifier, 0, len(specifierNodes))
		for _, specNode := range specifierNodes {
			spec := c.convertRadSortSpecifier(&specNode)
			specifiers = append(specifiers, spec)
		}
		return rl.NewRadSort(c.makeSpan(node), specifiers)

	case rl.K_RAD_FIELD_MODIFIER_STMT:
		identifierNodes := rl.GetChildren(node, rl.F_IDENTIFIER)
		ids := make([]rl.Node, 0, len(identifierNodes))
		for _, idNode := range identifierNodes {
			ids = append(ids, rl.NewIdentifier(c.makeSpan(&idNode), c.getSrc(&idNode)))
		}
		stmtNodes := rl.GetChildren(node, rl.F_MOD_STMT)
		mods := make([]rl.Node, 0, len(stmtNodes))
		for _, stmtNode := range stmtNodes {
			mods = append(mods, c.convertRadModStmt(&stmtNode))
		}
		return rl.NewRadFieldMod(c.makeSpan(node), ids, "", mods)

	case rl.K_RAD_IF_STMT:
		altNodes := rl.GetChildren(node, rl.F_ALT)
		branches := make([]rl.IfBranch, 0, len(altNodes))
		for _, altNode := range altNodes {
			condNode := rl.GetChild(&altNode, rl.F_CONDITION)
			var condition rl.Node
			if condNode != nil {
				condition = c.convertExpr(condNode)
			}
			stmtNodes := rl.GetChildren(&altNode, rl.F_STMT)
			body := make([]rl.Node, 0, len(stmtNodes))
			for _, stmtNode := range stmtNodes {
				body = append(body, c.convertRadStmt(&stmtNode))
			}
			branches = append(branches, rl.IfBranch{Condition: condition, Body: body})
		}
		return rl.NewRadIf(c.makeSpan(node), branches)

	default:
		panic(fmt.Sprintf("converter: unexpected rad stmt kind: %s", node.Kind()))
	}
}

func (c *converter) convertRadSortSpecifier(node *ts.Node) rl.RadSortSpecifier {
	firstNode := rl.GetChild(node, rl.F_FIRST)
	secondNode := rl.GetChild(node, rl.F_SECOND)

	if secondNode == nil {
		firstSrc := c.getSrc(firstNode)
		if firstSrc == rl.KEYWORD_ASC || firstSrc == rl.KEYWORD_DESC {
			return rl.RadSortSpecifier{
				Field:     "",
				Ascending: firstSrc == rl.KEYWORD_ASC,
			}
		}
		return rl.RadSortSpecifier{
			Field:     firstSrc,
			Ascending: true,
		}
	}

	ascending := true
	if secondNode.Kind() == rl.K_DESC {
		ascending = false
	}

	return rl.RadSortSpecifier{
		Field:     c.getSrc(firstNode),
		Ascending: ascending,
	}
}

func (c *converter) convertRadModStmt(node *ts.Node) rl.Node {
	switch node.Kind() {
	case rl.K_RAD_FIELD_MOD_COLOR:
		clrNode := rl.GetChild(node, rl.F_COLOR)
		regexNode := rl.GetChild(node, rl.F_REGEX)
		return rl.NewRadFieldMod(c.makeSpan(node), nil, "color",
			[]rl.Node{c.convertExpr(clrNode), c.convertExpr(regexNode)})
	case rl.K_RAD_FIELD_MOD_MAP:
		lambdaNode := rl.GetChild(node, rl.F_LAMBDA)
		return rl.NewRadFieldMod(c.makeSpan(node), nil, "map",
			[]rl.Node{c.convertExpr(lambdaNode)})
	case rl.K_RAD_FIELD_MOD_FILTER:
		lambdaNode := rl.GetChild(node, rl.F_LAMBDA)
		return rl.NewRadFieldMod(c.makeSpan(node), nil, "filter",
			[]rl.Node{c.convertExpr(lambdaNode)})
	default:
		panic(fmt.Sprintf("converter: unexpected rad mod stmt kind: %s", node.Kind()))
	}
}

// --- Operator resolution ---

func (c *converter) resolveBinaryOp(opNode *ts.Node) rl.Operator {
	src := c.getSrc(opNode)
	switch src {
	case "+":
		return rl.OpAdd
	case "-":
		return rl.OpSub
	case "*":
		return rl.OpMul
	case "/":
		return rl.OpDiv
	case "%":
		return rl.OpMod
	case "==":
		return rl.OpEq
	case "!=":
		return rl.OpNeq
	case "<":
		return rl.OpLt
	case "<=":
		return rl.OpLte
	case ">":
		return rl.OpGt
	case ">=":
		return rl.OpGte
	case "and":
		return rl.OpAnd
	case "or":
		return rl.OpOr
	default:
		panic(fmt.Sprintf("converter: unexpected binary operator: %q", src))
	}
}

func (c *converter) resolveCompoundOp(opNode *ts.Node) rl.Operator {
	switch opNode.Kind() {
	case rl.K_PLUS_EQUAL:
		return rl.OpAdd
	case rl.K_MINUS_EQUAL:
		return rl.OpSub
	case rl.K_STAR_EQUAL:
		return rl.OpMul
	case rl.K_SLASH_EQUAL:
		return rl.OpDiv
	case rl.K_PERCENT_EQUAL:
		return rl.OpMod
	default:
		panic(fmt.Sprintf("converter: unexpected compound operator kind: %s", opNode.Kind()))
	}
}

func (c *converter) resolveUnaryOp(opNode *ts.Node) rl.Operator {
	switch opNode.Kind() {
	case rl.K_MINUS:
		return rl.OpNeg
	case rl.K_NOT:
		return rl.OpNot
	case rl.K_PLUS:
		return rl.OpAdd // unary + is identity
	default:
		panic(fmt.Sprintf("converter: unexpected unary operator kind: %s", opNode.Kind()))
	}
}

// --- Helpers ---

func (c *converter) convertExprs(nodes []ts.Node) []rl.Node {
	result := make([]rl.Node, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, c.convertExpr(&node))
	}
	return result
}

func (c *converter) convertStmts(nodes []ts.Node) []rl.Node {
	result := make([]rl.Node, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, c.convertStmt(&node))
	}
	return result
}

func (c *converter) getOnlyChild(node *ts.Node) *ts.Node {
	count := node.ChildCount()
	if count != 1 {
		panic(fmt.Sprintf("converter: expected exactly one child, got %d (node kind: %s)", count, node.Kind()))
	}
	return node.Child(0)
}
