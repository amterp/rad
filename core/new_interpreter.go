package core

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type Interpreter struct {
	sd          *ScriptData
	env         *Env
	deferBlocks []*DeferBlock

	forLoopLevel int
	breaking     bool
	continuing   bool
}

func NewInterpreter(scriptData *ScriptData) *Interpreter {
	i := &Interpreter{
		sd: scriptData,
	}
	i.env = NewEnv(i)
	return i
}

func (i *Interpreter) Run() {
	node := i.sd.Tree.Root()
	i.recursivelyRun(node)
}

func (i *Interpreter) recursivelyRun(node *ts.Node) {
	defer func() {
		if r := recover(); r != nil {
			i.errorf(node, "Bug! Panic: %v\n%s", r, debug.Stack())
		}
	}()
	i.unsafeRecurse(node)
}

func (i *Interpreter) unsafeRecurse(node *ts.Node) {
	switch node.Kind() {
	// no-ops
	case K_SOURCE_FILE:
		children := node.Children(node.Walk())
		for _, child := range children {
			i.recursivelyRun(&child)
		}
	case K_COMMENT, K_SHEBANG, K_FILE_HEADER, K_ARG_BLOCK:
		return
	case K_ERROR:
		i.errorf(node, "Bug! Error pre-check should've prevented running into this node")
	case K_ASSIGN:
		rightNode := i.getChild(node, F_RIGHT)
		if rightNode.Kind() == K_JSON_PATH {
			// json path assignment
			jsonFieldVar := NewJsonFieldVar(i, node, rightNode)
			i.env.SetJsonFieldVar(jsonFieldVar)
		} else {
			// regular expr assignment
			leftVarPathNodes := i.getChildren(node, F_LEFT)
			numExpectedOutputs := len(leftVarPathNodes)
			values := i.evaluate(rightNode, numExpectedOutputs)
			for idx, leftVarPathNode := range leftVarPathNodes {
				i.doVarPathAssign(&leftVarPathNode, values[idx])
			}
		}
	case K_COMPOUND_ASSIGN:
		leftVarPathNode := i.getChild(node, F_LEFT)
		rightNode := i.getChild(node, F_RIGHT)
		opNode := i.getChild(node, F_OP)
		newValue := i.executeCompoundOp(node, leftVarPathNode, rightNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue)
	case K_EXPR_STMT:
		i.evaluate(i.getOnlyChild(node), NO_NUM_RETURN_VALUES_CONSTRAINT)
	case K_BREAK_STMT:
		if i.forLoopLevel > 0 {
			i.breaking = true
		} else {
			i.errorf(node, "Cannot 'break' outside of a for loop")
		}
	case K_CONTINUE_STMT:
		if i.forLoopLevel > 0 {
			i.continuing = true
		} else {
			i.errorf(node, "Cannot 'continue' outside of a for loop")
		}
	case K_FOR_LOOP:
		i.forLoopLevel++
		defer func() {
			i.forLoopLevel--
		}()
		stmts := i.getChildren(node, F_STMT)
		i.executeForLoop(node, func() { i.runBlock(stmts) })
	case K_IF_STMT:
		altNodes := i.getChildren(node, F_ALT)
		for _, altNode := range altNodes {
			condNode := i.getChild(&altNode, F_CONDITION)

			shouldExecute := true
			if condNode != nil {
				condResult := i.evaluate(condNode, 1)[0].TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := i.getChildren(&altNode, F_STMT)
				i.runBlock(stmtNodes)
				break
			}
		}
	case K_DEFER_BLOCK, K_ERRDEFER_BLOCK:
		keywordNode := i.getChild(node, F_KEYWORD)
		stmtNodes := i.getChildren(node, F_STMT)
		i.deferBlocks = append(i.deferBlocks, NewDeferBlock(i, keywordNode, stmtNodes))
	case K_SHELL_STMT:
		i.executeShellStmt(node)
	case K_DEL_STMT:
		rightVarPathNodes := i.getChildren(node, F_RIGHT)
		for _, rightVarPathNode := range rightVarPathNodes {
			i.doVarPathAssign(&rightVarPathNode, NIL_SENTINAL)
		}
	case K_RAD_BLOCK:
		i.runRadBlock(node)
	case K_INCR_DECR:
		leftVarPathNode := i.getChild(node, F_LEFT)
		opNode := i.getChild(node, F_OP)
		newValue := i.executeUnaryOp(node, leftVarPathNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue)
	default:
		i.errorf(node, "Unsupported node kind: %s", node.Kind())
	}
}

func (i *Interpreter) evaluate(node *ts.Node, numExpectedOutputs int) []RslValue {
	defer func() {
		if r := recover(); r != nil {
			i.errorDetailsf(node, fmt.Sprintf("%s\n%s", r, debug.Stack()), "Bug! Panic'd here")
		}
	}()
	return i.unsafeEval(node, numExpectedOutputs)
}

func (i *Interpreter) unsafeEval(node *ts.Node, numExpectedOutputs int) []RslValue {
	switch node.Kind() {
	case K_EXPR:
		return i.evaluate(i.getOnlyChild(node), numExpectedOutputs)
	case K_PRIMARY_EXPR:
		return i.evaluate(i.getOnlyChild(node), numExpectedOutputs)
	case K_PARENTHESIZED_EXPR:
		return i.evaluate(i.getChild(node, F_EXPR), numExpectedOutputs)
	case K_LITERAL:
		return i.evaluate(i.getOnlyChild(node), numExpectedOutputs)
	case K_UNARY_OP, K_NOT_OP:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		opNode := i.getChild(node, F_OP)
		argNode := i.getChild(node, F_ARG)
		return newRslValues(i, node, i.executeUnaryOp(node, argNode, opNode))
	case K_BINARY_OP, K_COMPARISON_OP, K_BOOL_OP:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		left := i.getChild(node, F_LEFT)
		op := i.getChild(node, F_OP)
		right := i.getChild(node, F_RIGHT)
		return newRslValues(i, node, i.executeBinary(node, left, right, op))

	// LEAF NODES
	case K_IDENTIFIER:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		identifier := i.sd.Src[node.StartByte():node.EndByte()]
		val, ok := i.env.GetVar(identifier)
		if !ok {
			i.errorf(node, "Undefined variable: %s", identifier)
		}
		return newRslValues(i, node, val)
	case K_VAR_PATH:
		rootIdentifier := i.getChild(node, F_ROOT) // identifier required by grammar
		indexings := i.getChildren(node, F_INDEXING)
		val := i.evaluate(rootIdentifier, 1)[0]
		if len(indexings) > 0 {
			for _, index := range indexings {
				val = val.Index(i, &index)
			}
		}
		return newRslValues(i, node, val)
	case K_INT:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asInt, _ := strconv.ParseInt(asStr, 10, 64) // todo unhandled err
		return newRslValues(i, node, asInt)
	case K_STRING:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		str := NewRslString("")

		contentsNode := i.getChild(node, F_CONTENTS)
		if contentsNode != nil {
			for _, child := range contentsNode.Children(contentsNode.Walk()) {
				str = str.Concat(i.evaluate(&child, 1)[0].RequireStr(i, &child))
			}
		}

		return newRslValues(i, node, str)
	case K_BOOL:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asBool, _ := strconv.ParseBool(asStr)
		return newRslValues(i, node, asBool)
	case K_STRING_CONTENT:
		src := i.sd.Src[node.StartByte():node.EndByte()]
		return newRslValues(i, node, src)
	case K_BACKSLASH:
		// todo potentially divisive - there are 3 options for escaping of 'insignificant' characters
		//  1. print the backslash and char as-are
		//  2. 'absorb' the backslash and print the char as-is
		//  3. error node (tree sitter should no allow it)
		//  this implementation is 2. may change. Go does 3. python & others do 1 (seems popular)
		return newRslValues(i, node, "")
	case K_ESC_SINGLE_QUOTE:
		return newRslValues(i, node, "'")
	case K_ESC_DOUBLE_QUOTE:
		return newRslValues(i, node, `"`)
	case K_ESC_BACKTICK:
		return newRslValues(i, node, "`")
	case K_ESC_NEWLINE:
		return newRslValues(i, node, "\n")
	case K_ESC_TAB:
		return newRslValues(i, node, "\t")
	case K_INTERPOLATION:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		exprResult := evaluateInterpolation(i, node)
		return newRslValues(i, node, exprResult)
	case K_ESC_BACKSLASH:
		return newRslValues(i, node, "\\")
	case K_LIST:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		entries := i.getChildren(node, F_LIST_ENTRY)
		list := NewRslList()
		for _, entry := range entries {
			list.Append(i.evaluate(&entry, 1)[0])
		}
		return newRslValues(i, node, list)
	case K_MAP:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		rslMap := NewRslMap()
		entryNodes := i.getChildren(node, F_MAP_ENTRY)
		for _, entryNode := range entryNodes {
			keyNode := i.getChild(&entryNode, F_KEY)
			valueNode := i.getChild(&entryNode, F_VALUE)
			key := evalMapKey(i, keyNode)
			rslMap.Set(key, i.evaluate(valueNode, 1)[0])
		}
		return newRslValues(i, node, rslMap)
	case K_CALL:
		funcName := i.getChild(node, F_FUNC)
		args := i.getChildren(node, F_ARG)
		var argValues []RslValue
		for _, arg := range args {
			argValues = append(argValues, i.evaluate(&arg, 1)[0])
		}
		return i.callFunction(node, funcName, argValues, numExpectedOutputs)
	case K_LIST_COMPREHENSION:
		resultExprNode := i.getChild(node, F_EXPR)
		conditionNode := i.getChild(node, F_CONDITION)

		resultList := NewRslList()
		doOneLoop := func() {
			if conditionNode == nil || i.evaluate(conditionNode, 1)[0].TruthyFalsy() {
				result := i.evaluate(resultExprNode, 1)[0]
				resultList.Append(result)
			}
		}
		i.executeForLoop(node, doOneLoop)
		return newRslValues(i, node, resultList)
	default:
		i.errorf(node, "Unsupported expr node kind: %s", node.Kind())
		panic(UNREACHABLE)
	}
}

func evaluateInterpolation(i *Interpreter, interpNode *ts.Node) RslValue {
	exprNode := i.getChild(interpNode, F_EXPR)
	formatNode := i.getChild(interpNode, F_FORMAT)

	exprResult := i.evaluate(exprNode, 1)[0]
	resultType := exprResult.Type()

	if formatNode == nil {
		if rslStr, ok := exprResult.TryGetStr(); ok {
			// maintain RslString attributes
			return newRslValue(i, exprNode, rslStr)
		} else {
			return newRslValue(i, exprNode, NewRslString(ToPrintable(exprResult)))
		}
	}

	alignmentNode := i.getChild(formatNode, F_ALIGNMENT)
	paddingNode := i.getChild(formatNode, F_PADDING)
	precisionNode := i.getChild(formatNode, F_PRECISION)

	var goFmt strings.Builder
	goFmt.WriteString("%")

	if alignmentNode != nil {
		alignment := i.sd.Src[alignmentNode.StartByte():alignmentNode.EndByte()]
		if alignment == "<" {
			goFmt.WriteString("-")
		}
	}

	if paddingNode != nil {
		padding := i.evaluate(paddingNode, 1)[0].RequireInt(i, paddingNode)
		if exprStr, ok := exprResult.TryGetStr(); ok {
			// is string, need to account for color chars (increase padding len if present)
			plainLen := exprStr.Len()
			coloredLen := int64(StrLen(exprStr.String()))
			diff := coloredLen - plainLen
			padding += diff
		}

		goFmt.WriteString(fmt.Sprint(padding))
	}

	if precisionNode != nil {
		precision := i.evaluate(precisionNode, 1)[0].RequireInt(i, precisionNode)

		if resultType != RslIntT && resultType != RslFloatT {
			precisionStr := "." + i.sd.Src[precisionNode.StartByte():precisionNode.EndByte()]
			i.errorf(interpNode, "Cannot format %s with a precision %q", TypeAsString(exprResult), precisionStr)
		}

		goFmt.WriteString(fmt.Sprintf(".%d", precision))
	}

	formatted := func() string {
		switch resultType {
		case RslIntT:
			if precisionNode == nil {
				goFmt.WriteString("d")
				return fmt.Sprintf(goFmt.String(), int(exprResult.Val.(int64)))
			} else {
				goFmt.WriteString("f")
				return fmt.Sprintf(goFmt.String(), float64(exprResult.Val.(int64)))
			}
		case RslFloatT:
			goFmt.WriteString("f")
			return fmt.Sprintf(goFmt.String(), exprResult.Val)
		default:
			goFmt.WriteString("s")
			return fmt.Sprintf(goFmt.String(), ToPrintable(exprResult))
		}
	}()

	return newRslValue(i, interpNode, formatted)
}

func (i *Interpreter) getChildren(node *ts.Node, fieldName string) []ts.Node {
	return node.ChildrenByFieldName(fieldName, node.Walk())
}

func (i *Interpreter) getChild(node *ts.Node, fieldName string) *ts.Node {
	return node.ChildByFieldName(fieldName)
}

func (i *Interpreter) getOnlyChild(node *ts.Node) *ts.Node {
	count := node.ChildCount()
	if count != 1 {
		i.errorf(node, "Bug? Expected exactly one child, got %d", count)
	}
	return node.Child(0)
}

func (i *Interpreter) assertExpectedNumOutputs(node *ts.Node, expectedOutputs int, actualOutputs int) {
	if expectedOutputs == NO_NUM_RETURN_VALUES_CONSTRAINT {
		return
	}

	if expectedOutputs != actualOutputs {
		i.errorf(node, "Expected %d outputs, got %d", expectedOutputs, actualOutputs)
	}
}

func (i *Interpreter) errorf(node *ts.Node, oneLinerFmt string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.sd.Src, node, fmt.Sprintf(oneLinerFmt, args...), ""))
}

func (i *Interpreter) errorDetailsf(node *ts.Node, details string, oneLinerFmt string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.sd.Src, node, fmt.Sprintf(oneLinerFmt, args...), details))
}

func (i *Interpreter) doVarPathAssign(varPathNode *ts.Node, rightValue RslValue) {
	rootIdentifier := i.getChild(varPathNode, F_ROOT) // identifier required by grammar
	rootIdentifierName := i.sd.Src[rootIdentifier.StartByte():rootIdentifier.EndByte()]
	indexings := i.getChildren(varPathNode, F_INDEXING)
	val, ok := i.env.GetVar(rootIdentifierName)

	if len(indexings) == 0 {
		// simple assignment, no collection lookups
		i.env.SetVar(rootIdentifierName, rightValue)
		return
	}

	// modifying collection
	if !ok {
		// modifying collection must exist
		i.errorf(rootIdentifier, "Undefined variable: %s", rootIdentifierName)
	}
	for _, index := range indexings[:len(indexings)-1] {
		val = val.Index(i, &index)
	}
	// val is now the collection to modify, using the last index
	lastIndex := indexings[len(indexings)-1]
	val.ModifyIdx(i, &lastIndex, rightValue)
}

func (i *Interpreter) executeForLoop(node *ts.Node, doOneLoop func()) {
	leftsNode := i.getChild(node, F_LEFTS)
	rightNode := i.getChild(node, F_RIGHT)

	rightVal := i.evaluate(rightNode, 1)[0]
	switch coercedRight := rightVal.Val.(type) {
	case RslString:
		runForLoopList(i, leftsNode, coercedRight.ToRuneList(), doOneLoop)
	case *RslList:
		runForLoopList(i, leftsNode, coercedRight, doOneLoop)
	case *RslMap:
		runForLoopMap(i, leftsNode, coercedRight, doOneLoop)
	default:
		i.errorf(rightNode, "Cannot iterate through a %s", TypeAsString(rightVal))
	}
}

func runForLoopList(i *Interpreter, leftsNode *ts.Node, list *RslList, doOneLoop func()) {
	var idxNode *ts.Node
	var itemNode *ts.Node

	leftNodes := i.getChildren(leftsNode, F_LEFT)

	if len(leftNodes) == 1 {
		itemNode = &leftNodes[0]
	} else if len(leftNodes) == 2 {
		idxNode = &leftNodes[0]
		itemNode = &leftNodes[1]
	} else {
		i.errorf(leftsNode, "Expected 1 or 2 variables on left side of for loop")
	}

	for idx, val := range list.Values {
		if idxNode != nil {
			idxName := i.sd.Src[idxNode.StartByte():idxNode.EndByte()]
			i.env.SetVar(idxName, newRslValue(i, idxNode, int64(idx)))
		}

		itemName := i.sd.Src[itemNode.StartByte():itemNode.EndByte()]
		i.env.SetVar(itemName, val)

		doOneLoop()
		if i.breaking {
			i.breaking = false
			break
		}
	}
}

func runForLoopMap(i *Interpreter, leftsNode *ts.Node, rslMap *RslMap, doOneLoop func()) {
	var keyNode *ts.Node
	var valueNode *ts.Node

	leftNodes := i.getChildren(leftsNode, F_LEFT)
	numLefts := len(leftNodes)

	if numLefts == 0 || numLefts > 2 {
		i.errorf(leftsNode, "Expected 1 or 2 variables on left side of for loop")
	}

	keyNode = &leftNodes[0]
	if numLefts == 2 {
		valueNode = &leftNodes[1]
	}

	for _, key := range rslMap.Keys() {
		keyName := i.sd.Src[keyNode.StartByte():keyNode.EndByte()]
		i.env.SetVar(keyName, key)

		if valueNode != nil {
			valueName := i.sd.Src[valueNode.StartByte():valueNode.EndByte()]
			value, _ := rslMap.Get(key)
			i.env.SetVar(valueName, value)
		}

		doOneLoop()
		if i.breaking {
			i.breaking = false
			break
		}
		i.continuing = false
	}
}

// if stmts, for loops
func (i *Interpreter) runBlock(stmtNodes []ts.Node) {
	for _, stmtNode := range stmtNodes {
		i.recursivelyRun(&stmtNode)
		if i.forLoopLevel > 0 && i.breaking || i.continuing {
			break
		}
	}
}

func (i *Interpreter) runWithChildEnv(runnable func()) {
	originalEnv := i.env
	env := originalEnv.NewChildEnv()
	i.env = &env
	runnable()
	i.env = originalEnv
}

type DeferBlock struct {
	DeferNode  *ts.Node
	StmtNodes  []ts.Node
	IsErrDefer bool
}

func NewDeferBlock(i *Interpreter, deferKeywordNode *ts.Node, stmtNodes []ts.Node) *DeferBlock {
	deferKeywordStr := i.sd.Src[deferKeywordNode.StartByte():deferKeywordNode.EndByte()]
	return &DeferBlock{
		DeferNode:  deferKeywordNode,
		StmtNodes:  stmtNodes,
		IsErrDefer: deferKeywordStr == "errdefer",
	}
}

func (i *Interpreter) RegisterWithExit() {
	existing := RExit
	exiting := false
	codeToExitWith := 0
	RExit = func(code int) {
		if exiting {
			// we're already exiting. if we're here again, it's probably because one of the deferred
			// statements is calling exit again (perhaps because it failed). we should keep running
			// all the deferred statements, however, and *then* exit.
			// therefore, we panic here in order to send the stack back up to where the deferred statement is being
			// invoked in the interpreter, which should be wrapped in a recover() block to catch, maybe log, and move on.
			if codeToExitWith == 0 {
				codeToExitWith = code
			}
			panic(code)
		}
		exiting = true
		codeToExitWith = code
		// todo gets executed *after* any error is printed (if error), should delay error print until after (i think?)
		i.executeDeferBlocks(code)
		existing(codeToExitWith)
	}
}

func (i *Interpreter) executeDeferBlocks(errCode int) {
	// execute backwards (LIFO)
	for j := len(i.deferBlocks) - 1; j >= 0; j-- {
		deferBlock := i.deferBlocks[j]

		if deferBlock.IsErrDefer && errCode == 0 {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					// we only debug log. we expect the error that occurred to already have been logged.
					// we might also be here only because a deferred statement invoked a clean exit, for example, so
					// this is arguably also sometimes just standard flow.
					RP.RadDebugf("Recovered from panic in deferred statement: %v", r)
				}
			}()
			i.runBlock(deferBlock.StmtNodes)
		}()
	}
}
