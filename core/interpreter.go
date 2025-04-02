package core

import (
	"fmt"
	com "rad/core/common"
	"runtime/debug"
	"strconv"
	"strings"

	rts "github.com/amterp/rts"

	"github.com/samber/lo"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type Delimiter struct {
	Open string
}

type Interpreter struct {
	sd          *ScriptData
	env         *Env
	deferBlocks []*DeferBlock

	forWhileLoopLevel int
	breaking          bool
	continuing        bool
	// Used to track current delimiter, currently for correct delimiter escaping handling
	delimiterStack *com.Stack[Delimiter]
}

func NewInterpreter(scriptData *ScriptData) *Interpreter {
	i := &Interpreter{
		sd:             scriptData,
		delimiterStack: com.NewStack[Delimiter](),
	}
	i.env = NewEnv(i)
	return i
}

func (i *Interpreter) InitArgs(args []RslArg) {
	env := i.env

	for _, arg := range args {
		if !arg.IsDefined() {
			continue
		}
		switch coerced := arg.(type) {
		case *BoolRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *BoolArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *StringRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *StringArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *IntRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *IntArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *FloatRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *FloatArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		default:
			i.errorf(arg.GetNode(), "Unsupported arg type, cannot init: %T", arg)
		}
	}
}

func (i *Interpreter) Run() {
	node := i.sd.Tree.Root()
	i.recursivelyRun(node)
}

func (i *Interpreter) recursivelyRun(node *ts.Node) {
	if !IsTest {
		defer func() {
			if r := recover(); r != nil {
				i.errorf(node, "Bug! Panic: %v\n%s", r, debug.Stack())
			}
		}()
	}
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
		leftNodes := i.getChildren(node, F_LEFT)
		rightNodes := i.getChildren(node, F_RIGHT)
		i.assignRightsToLefts(node, leftNodes, rightNodes)
	case K_COMPOUND_ASSIGN:
		leftVarPathNode := i.getChild(node, F_LEFT)
		rightNode := i.getChild(node, F_RIGHT)
		opNode := i.getChild(node, F_OP)
		newValue := i.executeCompoundOp(node, leftVarPathNode, rightNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue)
	case K_EXPR:
		i.evaluate(i.getOnlyChild(node), NO_NUM_RETURN_VALUES_CONSTRAINT)
	case K_BREAK_STMT:
		if i.forWhileLoopLevel > 0 {
			i.breaking = true
		} else {
			i.errorf(node, "Cannot 'break' outside of a for loop")
		}
	case K_CONTINUE_STMT:
		if i.forWhileLoopLevel > 0 {
			i.continuing = true
		} else {
			i.errorf(node, "Cannot 'continue' outside of a for loop")
		}
	case K_FOR_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		stmts := i.getChildren(node, F_STMT)
		i.executeForLoop(node, func() { i.runBlock(stmts) })
	case K_WHILE_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		condNode := i.getChild(node, F_CONDITION)
		stmtNodes := i.getChildren(node, F_STMT)
		for {
			condValue := true
			if condNode != nil {
				condValue = i.evaluate(condNode, 1)[0].TruthyFalsy()
			}
			if condValue {
				i.runBlock(stmtNodes)
				if i.breaking {
					i.breaking = false
					break
				}
				i.continuing = false
			} else {
				break
			}
		}
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
	case K_SWITCH_STMT:
		leftVarPathNodes := i.getChildren(node, F_LEFT)
		discriminantNode := i.getChild(node, F_DISCRIMINANT)
		caseNodes := i.getChildren(node, F_CASE)
		defaultNode := i.getChild(node, F_DEFAULT)

		discriminantVal := i.evaluate(discriminantNode, 1)[0]

		matchedCaseNodes := make([]ts.Node, 0)
		for _, caseNode := range caseNodes {
			caseKeyNodes := i.getChildren(&caseNode, F_CASE_KEY)
			for _, caseKeyNode := range caseKeyNodes {
				caseKey := i.evaluate(&caseKeyNode, 1)[0]
				if caseKey.Equals(discriminantVal) {
					matchedCaseNodes = append(matchedCaseNodes, caseNode)
					break
				}
			}
		}

		if len(matchedCaseNodes) == 0 {
			if defaultNode != nil {
				caseValueAltNode := i.getChild(defaultNode, F_ALT)
				i.executeSwitchCase(caseValueAltNode, leftVarPathNodes)
				return
			}
			i.errorf(discriminantNode, "No matching case found for switch")
		}

		if len(matchedCaseNodes) > 1 {
			// todo fancier error msg: should point at the cases that matched, example:
			// 13 | case 1, "one": blah blah
			//           ^  ^^^^^ MATCHED        << in red
			// 14 | case 2, "two": blah blah
			// 15 | case 3, val: blah blah
			//              ^^^ MATCHED          << in red
			// 16 | case 4, "four":  blah blah
			i.errorf(discriminantNode, "Multiple matching cases found for switch")
		}

		matchedCaseNode := matchedCaseNodes[0]
		caseValueAltNode := i.getChild(&matchedCaseNode, F_ALT)
		i.executeSwitchCase(caseValueAltNode, leftVarPathNodes)
	case K_DEFER_BLOCK:
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
	if !IsTest {
		defer func() {
			if r := recover(); r != nil {
				i.errorDetailsf(node, fmt.Sprintf("%s\n%s", r, debug.Stack()), "Bug! Panic'd here")
			}
		}()
	}
	return i.unsafeEval(node, numExpectedOutputs)
}

func (i *Interpreter) unsafeEval(node *ts.Node, numExpectedOutputs int) []RslValue {
	switch node.Kind() {
	case K_EXPR, K_PRIMARY_EXPR, K_LITERAL:
		return i.evaluate(i.getOnlyChild(node), numExpectedOutputs)
	case K_PARENTHESIZED_EXPR:
		return i.evaluate(i.getChild(node, F_EXPR), numExpectedOutputs)
	case K_UNARY_EXPR:
		delegateNode := i.getChild(node, F_DELEGATE)
		if delegateNode != nil {
			return i.evaluate(delegateNode, numExpectedOutputs)
		}

		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		opNode := i.getChild(node, F_OP)
		argNode := i.getChild(node, F_ARG)
		return newRslValues(i, node, i.executeUnaryOp(node, argNode, opNode))
	case K_OR_EXPR, K_AND_EXPR, K_COMPARE_EXPR, K_ADD_EXPR, K_MULT_EXPR:
		delegateNode := i.getChild(node, F_DELEGATE)
		if delegateNode != nil {
			return i.evaluate(delegateNode, numExpectedOutputs)
		}

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
		rootNode := i.getChild(node, F_ROOT)
		indexingNodes := i.getChildren(node, F_INDEXING)
		val := i.evaluate(rootNode, 1)[0]
		if len(indexingNodes) > 0 {
			for indexIdx, indexNode := range indexingNodes {
				expectReturnVal := numExpectedOutputs != NO_NUM_RETURN_VALUES_CONSTRAINT && indexIdx < len(indexingNodes)-1
				val = i.evaluateIndexing(rootNode, indexNode, val, expectReturnVal)
			}
		}
		return newRslValues(i, node, val)
	case K_INDEXED_EXPR:
		rootNode := i.getChild(node, F_ROOT)
		indexingNodes := i.getChildren(node, F_INDEXING)
		if len(indexingNodes) > 0 {
			val := i.evaluate(rootNode, 1)[0]
			for indexIdx, index := range indexingNodes {
				expectReturnVal := numExpectedOutputs != NO_NUM_RETURN_VALUES_CONSTRAINT || indexIdx != len(indexingNodes)-1
				val = i.evaluateIndexing(rootNode, index, val, expectReturnVal)
			}
			return newRslValues(i, node, val)
		} else {
			return i.evaluate(rootNode, numExpectedOutputs)
		}
	case K_INT:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asInt, _ := rts.ParseInt(asStr) // todo unhandled err
		return newRslValues(i, node, asInt)
	case K_FLOAT:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asFloat, _ := rts.ParseFloat(asStr) // todo unhandled err
		return newRslValues(i, node, asFloat)
	case K_STRING:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		str := NewRslString("")

		contentsNode := i.getChild(node, F_CONTENTS)

		// With current TS grammar, last character of closing delimiter is always the delimiter
		// Admittedly bad, very white boxy and brittle
		endNode := i.getChild(node, F_END)
		endStr := i.sd.Src[endNode.StartByte():endNode.EndByte()]
		delimiterStr := endStr[len(endStr)-1]
		i.delimiterStack.Push(Delimiter{Open: string(delimiterStr)})

		if contentsNode != nil {
			for _, child := range contentsNode.Children(contentsNode.Walk()) {
				str = str.Concat(i.evaluate(&child, 1)[0].RequireStr(i, &child))
			}
		}

		i.delimiterStack.Pop()

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
		return newRslValues(i, node, "\\")
	case K_ESC_SINGLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "'" {
			return newRslValues(i, node, "'")
		} else {
			return newRslValues(i, node, `\'`)
		}
	case K_ESC_DOUBLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == `"` {
			return newRslValues(i, node, `"`)
		} else {
			return newRslValues(i, node, `\"`)
		}
	case K_ESC_BACKTICK:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "`" {
			return newRslValues(i, node, "`")
		} else {
			return newRslValues(i, node, "\\`")
		}
	case K_ESC_NEWLINE:
		return newRslValues(i, node, "\n")
	case K_ESC_TAB:
		return newRslValues(i, node, "\t")
	case K_ESC_OPEN_BRACKET:
		return newRslValues(i, node, "{")
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
		return i.callFunction(node, numExpectedOutputs, nil)
	case K_LIST_COMPREHENSION:
		resultExprNode := i.getChild(node, F_EXPR)
		conditionNode := i.getChild(node, F_CONDITION)

		resultList := NewRslList()
		doOneLoop := func() {
			if conditionNode == nil || i.evaluate(conditionNode, 1)[0].TruthyFalsy() {
				results := i.evaluate(resultExprNode, NO_NUM_RETURN_VALUES_CONSTRAINT)
				if len(results) > 0 {
					// note: if expr (e.g. function) returns several values, we keep only the first.
					resultList.Append(results[0])
				}
			}
		}
		i.executeForLoop(node, doOneLoop)
		return newRslValues(i, node, resultList)
	case K_TERNARY_EXPR:
		delegateNode := i.getChild(node, F_DELEGATE)
		if delegateNode != nil {
			return i.evaluate(delegateNode, numExpectedOutputs)
		}

		conditionNode := i.getChild(node, F_CONDITION)
		trueNode := i.getChild(node, F_TRUE_BRANCH)
		falseNode := i.getChild(node, F_FALSE_BRANCH)
		condition := i.evaluate(conditionNode, 1)[0].TruthyFalsy()
		return i.evaluate(lo.Ternary(condition, trueNode, falseNode), numExpectedOutputs)
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
			coloredLen := int64(com.StrLen(exprStr.String()))
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
			return fmt.Sprintf(goFmt.String(), ToPrintableQuoteStr(exprResult, false))
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
		i.errorf(node, "Expected %s, got %d", com.Pluralize(expectedOutputs, "output"), actualOutputs)
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
		runForLoopList(i, leftsNode, rightNode, coercedRight.ToRuneList(), doOneLoop)
	case *RslList:
		runForLoopList(i, leftsNode, rightNode, coercedRight, doOneLoop)
	case *RslMap:
		runForLoopMap(i, leftsNode, coercedRight, doOneLoop)
	default:
		i.errorf(rightNode, "Cannot iterate through a %s", TypeAsString(rightVal))
	}
}

func runForLoopList(i *Interpreter, leftsNode, rightNode *ts.Node, list *RslList, doOneLoop func()) {
	var idxNode *ts.Node
	itemNodes := make([]*ts.Node, 0)

	leftNodes := i.getChildren(leftsNode, F_LEFT)

	if len(leftNodes) == 0 {
		i.errorf(leftsNode, "Expected at least one variable on the left side of for loop")
	} else if len(leftNodes) == 1 {
		itemNodes = append(itemNodes, &leftNodes[0])
	} else {
		idxNode = &leftNodes[0]
		for idx := 1; idx < len(leftNodes); idx++ {
			itemNodes = append(itemNodes, &leftNodes[idx])
		}
	}

	for idx, val := range list.Values {
		if idxNode != nil {
			idxName := i.sd.Src[idxNode.StartByte():idxNode.EndByte()]
			i.env.SetVar(idxName, newRslValue(i, idxNode, int64(idx)))
		}

		if len(itemNodes) == 1 {
			itemNode := itemNodes[0]
			itemName := i.sd.Src[itemNode.StartByte():itemNode.EndByte()]
			i.env.SetVar(itemName, val)
		} else if len(itemNodes) > 1 {
			// expecting list of lists, unpacking by idx
			listInList, ok := val.TryGetList()
			if !ok {
				i.errorf(rightNode, "Expected list of lists, got element type %q", TypeAsString(val))
			}

			if listInList.LenInt() < len(itemNodes) {
				i.errorf(rightNode, "Expected at least %s in inner list, got %d",
					com.Pluralize(len(itemNodes), "value"), listInList.LenInt())
			}

			for idx, itemNode := range itemNodes {
				itemName := i.sd.Src[itemNode.StartByte():itemNode.EndByte()]
				i.env.SetVar(itemName, listInList.Values[idx])
			}
		}

		doOneLoop()
		if i.breaking {
			i.breaking = false
			break
		}
		i.continuing = false
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
		if i.breakingOrContinuing() {
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

func (i *Interpreter) evaluateIndexing(rootNode *ts.Node, index ts.Node, val RslValue, expectReturnValue bool) RslValue {
	if index.Kind() == K_CALL {
		// ufcs
		ufcsArg := &positionalArg{
			// todo 'rootNode' is not great to use, it misses indexes in between that and this call,
			//  resulting in bad error pointing. could potentially replace ts.Node with interface
			//  'Pointable' i.e. a range we can point to in an error, that's ultimately all we need (?)
			node:  rootNode,
			value: val,
		}
		if expectReturnValue {
			return i.callFunction(&index, 1, ufcsArg)[0]
		} else {
			returnVals := i.callFunction(&index, NO_NUM_RETURN_VALUES_CONSTRAINT, ufcsArg)
			if len(returnVals) == 1 {
				// todo not quite right, what if multiple returned?
				//  e.g. print("1".parse_int()), should print both return vals, but only first one passed on here
				return returnVals[0]
			}
			return val
		}
	} else {
		return val.Index(i, &index)
	}
}

func (i *Interpreter) assignRightsToLefts(parentNode *ts.Node, leftNodes, rightNodes []ts.Node) {
	numReturnValues := lo.Ternary(len(leftNodes) == 1, 1, NO_NUM_RETURN_VALUES_CONSTRAINT)
	outputs := make([]RslValue, 0)
	for _, rightNode := range rightNodes {
		if rightNode.Kind() == K_JSON_PATH {
			// json path assignment
			jsonFieldVar := NewJsonFieldVar(i, &leftNodes[len(outputs)], &rightNode) // todo index bounds error isn't user friendly
			i.env.SetJsonFieldVar(jsonFieldVar)
			outputs = append(outputs, JSON_SENTINAL)
		} else {
			outputs = append(outputs, i.evaluate(&rightNode, numReturnValues)...)
		}
	}

	if len(leftNodes) != len(outputs) {
		i.errorf(parentNode, "Cannot assign %d values to %d variables", len(outputs), len(leftNodes))
	}

	for idx, output := range outputs {
		if output == JSON_SENTINAL {
			// json path assignment, no need to assign
			continue
		}
		leftVarPathNode := &leftNodes[idx]
		i.doVarPathAssign(leftVarPathNode, output)
	}
}

func (i *Interpreter) executeSwitchCase(caseValueAltNode *ts.Node, leftVarPathNodes []ts.Node) {
	numExpectedAssigns := len(leftVarPathNodes)
	switch caseValueAltNode.Kind() {
	case K_SWITCH_CASE_EXPR:
		valueNodes := i.getChildren(caseValueAltNode, F_VALUE)
		i.assignRightsToLefts(caseValueAltNode, leftVarPathNodes, valueNodes)
	case K_SWITCH_CASE_BLOCK:
		stmtNodes := i.getChildren(caseValueAltNode, F_STMT)
		i.runBlock(stmtNodes)

		if i.breakingOrContinuing() {
			return
		}

		yieldNode := i.getChild(caseValueAltNode, F_YIELD_STMT)
		if numExpectedAssigns > 0 && yieldNode == nil {
			i.errorf(caseValueAltNode, "Cannot assign without yielding from the switch case")
		}
		if yieldNode != nil {
			valueNodes := i.getChildren(yieldNode, F_VALUE)
			if numExpectedAssigns == 0 {
				// nothing to assign, just evaluate
				for _, valueNode := range valueNodes {
					i.evaluate(&valueNode, NO_NUM_RETURN_VALUES_CONSTRAINT)
				}
			} else {
				i.assignRightsToLefts(caseValueAltNode, leftVarPathNodes, valueNodes)
			}
		}
	default:
		i.errorf(caseValueAltNode, "Bug! Unsupported switch case value node kind: %s", caseValueAltNode.Kind())
	}
}

func (i *Interpreter) breakingOrContinuing() bool {
	return i.forWhileLoopLevel > 0 && i.breaking || i.continuing
}
