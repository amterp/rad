package core

import (
	"fmt"
	com "rad/core/common"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/amterp/rad/rts"

	"github.com/amterp/rad/rts/rl"

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

func (i *Interpreter) InitBuiltIns() {
	for name, fn := range FunctionsByName {
		fnVal := newRadValueFn(NewBuiltIn(fn))
		i.env.SetVar(name, fnVal)
	}
}

func (i *Interpreter) InitArgs(args []RadArg) {
	env := i.env

	for _, arg := range args {
		if !arg.IsDefined() {
			env.SetVar(arg.GetIdentifier(), RAD_NULL_VAL)
			continue
		}
		switch coerced := arg.(type) {
		case *BoolRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *BoolArrRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *StringRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *StringArrRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *IntRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *IntArrRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *FloatRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), coerced.Value))
		case *FloatArrRadArg:
			env.SetVar(coerced.Identifier, newRadValue(i, arg.GetNode(), NewRadListFromGeneric(i, arg.GetNode(), coerced.Value)))
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
	case rl.K_SOURCE_FILE:
		children := node.Children(node.Walk())
		// first define custom named functions
		for _, child := range children {
			if child.Kind() == rl.K_FN_NAMED {
				i.defineCustomNamedFunction(child)
			}
		}
		for _, child := range children {
			i.recursivelyRun(&child)
		}
	case rl.K_COMMENT, rl.K_SHEBANG, rl.K_FILE_HEADER, rl.K_ARG_BLOCK:
		return
	case rl.K_ERROR:
		i.errorf(node, "Bug! Error pre-check should've prevented running into this node")
	case rl.K_ASSIGN:
		leftNodes := i.getChildren(node, rl.F_LEFT)
		rightNodes := i.getChildren(node, rl.F_RIGHT)
		i.assignRightsToLefts(node, leftNodes, rightNodes)
	case rl.K_COMPOUND_ASSIGN:
		leftVarPathNode := i.getChild(node, rl.F_LEFT)
		rightNode := i.getChild(node, rl.F_RIGHT)
		opNode := i.getChild(node, rl.F_OP)
		newValue := i.executeCompoundOp(node, leftVarPathNode, rightNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue, true)
	case rl.K_EXPR:
		i.evaluate(i.getOnlyChild(node), NO_NUM_RETURN_VALUES_CONSTRAINT)
	case rl.K_PASS:
		// no-op
	case rl.K_BREAK_STMT:
		if i.forWhileLoopLevel > 0 {
			i.breaking = true
		} else {
			i.errorf(node, "Cannot 'break' outside of a for loop")
		}
	case rl.K_CONTINUE_STMT:
		if i.forWhileLoopLevel > 0 {
			i.continuing = true
		} else {
			i.errorf(node, "Cannot 'continue' outside of a for loop")
		}
	case rl.K_FOR_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		stmts := i.getChildren(node, rl.F_STMT)
		i.executeForLoop(node, func() { i.runBlock(stmts) })
	case rl.K_WHILE_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		condNode := i.getChild(node, rl.F_CONDITION)
		stmtNodes := i.getChildren(node, rl.F_STMT)
		for {
			condValue := true
			if condNode != nil {
				condValue = i.evaluate(condNode, 1).TruthyFalsy()
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
	case rl.K_IF_STMT:
		altNodes := i.getChildren(node, rl.F_ALT)
		for _, altNode := range altNodes {
			condNode := i.getChild(&altNode, rl.F_CONDITION)

			shouldExecute := true
			if condNode != nil {
				condResult := i.evaluate(condNode, 1).TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := i.getChildren(&altNode, rl.F_STMT)
				i.runBlock(stmtNodes)
				break
			}
		}
	case rl.K_SWITCH_STMT:
		leftVarPathNodes := i.getChildren(node, rl.F_LEFT)
		discriminantNode := i.getChild(node, rl.F_DISCRIMINANT)
		caseNodes := i.getChildren(node, rl.F_CASE)
		defaultNode := i.getChild(node, rl.F_DEFAULT)

		discriminantVal := i.evaluate(discriminantNode, 1)

		matchedCaseNodes := make([]ts.Node, 0)
		for _, caseNode := range caseNodes {
			caseKeyNodes := i.getChildren(&caseNode, rl.F_CASE_KEY)
			for _, caseKeyNode := range caseKeyNodes {
				caseKey := i.evaluate(&caseKeyNode, 1)
				if caseKey.Equals(discriminantVal) {
					matchedCaseNodes = append(matchedCaseNodes, caseNode)
					break
				}
			}
		}

		if len(matchedCaseNodes) == 0 {
			if defaultNode != nil {
				caseValueAltNode := i.getChild(defaultNode, rl.F_ALT)
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
		caseValueAltNode := i.getChild(&matchedCaseNode, rl.F_ALT)
		i.executeSwitchCase(caseValueAltNode, leftVarPathNodes)
	case rl.K_FN_NAMED:
		if node.Parent().Kind() == rl.K_SOURCE_FILE {
			// these are already defined initially
			return
		}
		i.defineCustomNamedFunction(*node)
	case rl.K_DEFER_BLOCK:
		keywordNode := i.getChild(node, rl.F_KEYWORD)
		stmtNodes := i.getChildren(node, rl.F_STMT)
		i.deferBlocks = append(i.deferBlocks, NewDeferBlock(i, keywordNode, stmtNodes))
	case rl.K_SHELL_STMT:
		i.executeShellStmt(node)
	case rl.K_DEL_STMT:
		rightVarPathNodes := i.getChildren(node, rl.F_RIGHT)
		for _, rightVarPathNode := range rightVarPathNodes {
			i.doVarPathAssign(&rightVarPathNode, VOID_SENTINEL, true)
		}
	case rl.K_RAD_BLOCK:
		i.runRadBlock(node)
	case rl.K_INCR_DECR:
		leftVarPathNode := i.getChild(node, rl.F_LEFT)
		opNode := i.getChild(node, rl.F_OP)
		newValue := i.executeUnaryOp(node, leftVarPathNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue, true)
	default:
		i.errorf(node, "Unsupported node kind: %s", node.Kind())
	}
}

func (i *Interpreter) evaluate(node *ts.Node, evalCtx EvalCtx) RadValue {
	if !IsTest {
		defer func() {
			if r := recover(); r != nil {
				i.errorDetailsf(node, fmt.Sprintf("%s\n%s", r, debug.Stack()), "Bug! Panicked here")
			}
		}()
	}
	return i.unsafeEval(node, evalCtx)
}

func (i *Interpreter) unsafeEval(node *ts.Node, evalCtx EvalCtx) RadValue {
	switch node.Kind() {
	case rl.K_EXPR:
		out := i.evaluate(i.getOnlyChild(node), evalCtx)

		if out.IsError() {
			err := out.RequireError(i, node)
			catchNode := i.getChild(node, rl.F_CATCH)
			err.ShouldPropagate = catchNode == nil
		}

		return out
	case rl.K_PRIMARY_EXPR, rl.K_LITERAL:
		return i.evaluate(i.getOnlyChild(node), evalCtx)
	case rl.K_PARENTHESIZED_EXPR:
		return i.evaluate(i.getChild(node, rl.F_EXPR), evalCtx)
	case rl.K_UNARY_EXPR:
		delegateNode := i.getChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.evaluate(delegateNode, evalCtx)
		}

		i.assertExpectedNumOutputs(node, evalCtx, One)
		opNode := i.getChild(node, rl.F_OP)
		argNode := i.getChild(node, rl.F_ARG)
		return newRadValues(i, node, i.executeUnaryOp(node, argNode, opNode))
	case rl.K_OR_EXPR, rl.K_AND_EXPR, rl.K_COMPARE_EXPR, rl.K_ADD_EXPR, rl.K_MULT_EXPR:
		delegateNode := i.getChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.evaluate(delegateNode, evalCtx)
		}

		i.assertExpectedNumOutputs(node, evalCtx, One)
		left := i.getChild(node, rl.F_LEFT)
		op := i.getChild(node, rl.F_OP)
		right := i.getChild(node, rl.F_RIGHT)
		return newRadValues(i, node, i.executeBinary(node, left, right, op))

	// LEAF NODES
	case rl.K_IDENTIFIER:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		identifier := i.sd.Src[node.StartByte():node.EndByte()]
		val, ok := i.env.GetVar(identifier)
		if !ok {
			i.errorf(node, "Undefined variable: %s", identifier)
		}
		return newRadValues(i, node, val)
	case rl.K_VAR_PATH:
		rootNode := i.getChild(node, rl.F_ROOT)
		indexingNodes := i.getChildren(node, rl.F_INDEXING)
		val := i.evaluate(rootNode, 1)
		if len(indexingNodes) > 0 {
			for indexIdx, indexNode := range indexingNodes {
				expectReturnVal := numExpectedOutputs != NO_NUM_RETURN_VALUES_CONSTRAINT && indexIdx < len(indexingNodes)-1
				val = i.evaluateIndexing(rootNode, indexNode, val, expectReturnVal)
			}
		}
		return newRadValues(i, node, val)
	case rl.K_INDEXED_EXPR:
		rootNode := i.getChild(node, rl.F_ROOT)
		indexingNodes := i.getChildren(node, rl.F_INDEXING)
		if len(indexingNodes) > 0 {
			val := i.evaluate(rootNode, 1)
			for indexIdx, index := range indexingNodes {
				expectReturnVal := numExpectedOutputs != NO_NUM_RETURN_VALUES_CONSTRAINT || indexIdx != len(indexingNodes)-1
				val = i.evaluateIndexing(rootNode, index, val, expectReturnVal)
			}
			return newRadValues(i, node, val)
		} else {
			return i.evaluate(rootNode, evalCtx)
		}
	case rl.K_INT:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asInt, _ := rts.ParseInt(asStr) // todo unhandled err
		return newRadValues(i, node, asInt)
	case rl.K_FLOAT:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asFloat, _ := rts.ParseFloat(asStr) // todo unhandled err
		return newRadValues(i, node, asFloat)
	case rl.K_STRING:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		str := NewRadString("")

		contentsNode := i.getChild(node, rl.F_CONTENTS)

		// With current TS grammar, last character of closing delimiter is always the delimiter
		// Admittedly bad, very white boxy and brittle
		endNode := i.getChild(node, rl.F_END)
		endStr := i.sd.Src[endNode.StartByte():endNode.EndByte()]
		delimiterStr := endStr[len(endStr)-1]
		i.delimiterStack.Push(Delimiter{Open: string(delimiterStr)})

		if contentsNode != nil {
			for _, child := range contentsNode.Children(contentsNode.Walk()) {
				str = str.Concat(i.evaluate(&child, 1).RequireStr(i, &child))
			}
		}

		i.delimiterStack.Pop()

		return newRadValues(i, node, str)
	case rl.K_BOOL:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asBool, _ := strconv.ParseBool(asStr)
		return newRadValues(i, node, asBool)
	case rl.K_NULL:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		return newRadValues(i, node, nil)
	case rl.K_STRING_CONTENT:
		src := i.sd.Src[node.StartByte():node.EndByte()]
		return newRadValues(i, node, src)
	case rl.K_BACKSLASH:
		return newRadValues(i, node, "\\")
	case rl.K_ESC_SINGLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "'" {
			return newRadValues(i, node, "'")
		} else {
			return newRadValues(i, node, `\'`)
		}
	case rl.K_ESC_DOUBLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == `"` {
			return newRadValues(i, node, `"`)
		} else {
			return newRadValues(i, node, `\"`)
		}
	case rl.K_ESC_BACKTICK:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "`" {
			return newRadValues(i, node, "`")
		} else {
			return newRadValues(i, node, "\\`")
		}
	case rl.K_ESC_NEWLINE:
		return newRadValues(i, node, "\n")
	case rl.K_ESC_TAB:
		return newRadValues(i, node, "\t")
	case rl.K_ESC_OPEN_BRACKET:
		return newRadValues(i, node, "{")
	case rl.K_INTERPOLATION:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		exprResult := evaluateInterpolation(i, node)
		return newRadValues(i, node, exprResult)
	case rl.K_ESC_BACKSLASH:
		return newRadValues(i, node, "\\")
	case rl.K_LIST:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		entries := i.getChildren(node, rl.F_LIST_ENTRY)
		list := NewRadList()
		for _, entry := range entries {
			out := i.evaluate(&entry, EXPECT_ONE_OUTPUT)
			if out.IsErrorToPropagate() {
				return out
			}
			list.Append(out)
		}
		return newRadValues(i, node, list)
	case rl.K_MAP:
		i.assertExpectedNumOutputs(node, evalCtx, One)
		radMap := NewRadMap()
		entryNodes := i.getChildren(node, rl.F_MAP_ENTRY)
		for _, entryNode := range entryNodes {
			keyNode := i.getChild(&entryNode, rl.F_KEY)
			valueNode := i.getChild(&entryNode, rl.F_VALUE)
			key := evalMapKey(i, keyNode)
			radMap.Set(key, i.evaluate(valueNode, 1))
		}
		return newRadValues(i, node, radMap)
	case rl.K_CALL:
		return i.callFunction(node, evalCtx, nil)
	case rl.K_FN_LAMBDA:
		return newRadValues(i, node, NewLambda(i, node))
	case rl.K_LIST_COMPREHENSION:
		resultExprNode := i.getChild(node, rl.F_EXPR)
		conditionNode := i.getChild(node, rl.F_CONDITION)

		var errorToPropagate *RadValue
		resultList := NewRadList()
		doOneLoop := func() {
			if conditionNode == nil || i.evaluate(conditionNode, 1).TruthyFalsy() {
				out := i.evaluate(resultExprNode, 1)
				if out.IsErrorToPropagate() {
					i.breaking = true
					errorToPropagate = &out
				} else {
					resultList.Append(out)
				}
			}
		}
		i.executeForLoop(node, doOneLoop)

		if errorToPropagate != nil {
			return *errorToPropagate
		}

		return newRadValues(i, node, resultList)
	case rl.K_TERNARY_EXPR:
		delegateNode := i.getChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.evaluate(delegateNode, evalCtx)
		}

		conditionNode := i.getChild(node, rl.F_CONDITION)
		trueNode := i.getChild(node, rl.F_TRUE_BRANCH)
		falseNode := i.getChild(node, rl.F_FALSE_BRANCH)
		condition := i.evaluate(conditionNode, 1).TruthyFalsy()
		return i.evaluate(lo.Ternary(condition, trueNode, falseNode), evalCtx)
	default:
		i.errorf(node, "Unsupported expr node kind: %s", node.Kind())
		panic(UNREACHABLE)
	}
}

func evaluateInterpolation(i *Interpreter, interpNode *ts.Node) RadValue {
	exprNode := i.getChild(interpNode, rl.F_EXPR)
	formatNode := i.getChild(interpNode, rl.F_FORMAT)

	exprResult := i.evaluate(exprNode, 1)
	resultType := exprResult.Type()

	if formatNode == nil {
		switch resultType {
		case RadStringT:
			// to maintain attributes
			return exprResult
		case RadErrorT:
			return newRadValue(i, exprNode, exprResult.RequireError(i, interpNode).Msg())
		default:
			return newRadValue(i, exprNode, NewRadString(ToPrintable(exprResult)))
		}
	}

	alignmentNode := i.getChild(formatNode, rl.F_ALIGNMENT)
	paddingNode := i.getChild(formatNode, rl.F_PADDING)
	precisionNode := i.getChild(formatNode, rl.F_PRECISION)

	var goFmt strings.Builder
	goFmt.WriteString("%")

	if alignmentNode != nil {
		alignment := i.sd.Src[alignmentNode.StartByte():alignmentNode.EndByte()]
		if alignment == "<" {
			goFmt.WriteString("-")
		}
	}

	if paddingNode != nil {
		padding := i.evaluate(paddingNode, 1).RequireInt(i, paddingNode)
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
		precision := i.evaluate(precisionNode, 1).RequireInt(i, precisionNode)

		if resultType != RadIntT && resultType != RadFloatT {
			precisionStr := "." + i.sd.Src[precisionNode.StartByte():precisionNode.EndByte()]
			i.errorf(interpNode, "Cannot format %s with a precision %q", TypeAsString(exprResult), precisionStr)
		}

		goFmt.WriteString(fmt.Sprintf(".%d", precision))
	}

	formatted := func() string {
		switch resultType {
		case RadIntT:
			if precisionNode == nil {
				goFmt.WriteString("d")
				return fmt.Sprintf(goFmt.String(), int(exprResult.Val.(int64)))
			} else {
				goFmt.WriteString("f")
				return fmt.Sprintf(goFmt.String(), float64(exprResult.Val.(int64)))
			}
		case RadFloatT:
			goFmt.WriteString("f")
			return fmt.Sprintf(goFmt.String(), exprResult.Val)
		default:
			goFmt.WriteString("s")
			return fmt.Sprintf(goFmt.String(), ToPrintableQuoteStr(exprResult, false))
		}
	}()

	return newRadValue(i, interpNode, formatted)
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

func (i *Interpreter) assertExpectedNumOutputs(node *ts.Node, evalCtx EvalCtx, actual ExpectedOutput) {
	if evalCtx.ExpectedOutput == NoConstraint {
		return
	}

	if evalCtx.ExpectedOutput == actual {
		return
	}

	i.errorf(node, "Expected %s, got %s", evalCtx.ExpectedOutput, actual)
}

func (i *Interpreter) errorf(node *ts.Node, oneLinerFmt string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.sd.Src, node, fmt.Sprintf(oneLinerFmt, args...), ""))
}

func (i *Interpreter) errorDetailsf(node *ts.Node, details string, oneLinerFmt string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.sd.Src, node, fmt.Sprintf(oneLinerFmt, args...), details))
}

func (i *Interpreter) doVarPathAssign(varPathNode *ts.Node, rightValue RadValue, updateEnclosing bool) {
	rootIdentifier := i.getChild(varPathNode, rl.F_ROOT) // identifier required by grammar
	rootIdentifierName := GetSrc(i.sd.Src, rootIdentifier)
	indexings := i.getChildren(varPathNode, rl.F_INDEXING)
	val, ok := i.env.GetVar(rootIdentifierName)

	if len(indexings) == 0 {
		// simple assignment, no collection lookups
		i.env.SetVarUpdatingEnclosing(rootIdentifierName, rightValue, updateEnclosing)
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
	leftsNode := i.getChild(node, rl.F_LEFTS)
	rightNode := i.getChild(node, rl.F_RIGHT)

	rightVal := i.evaluate(rightNode, 1)
	switch coercedRight := rightVal.Val.(type) {
	case RadString:
		runForLoopList(i, leftsNode, rightNode, coercedRight.ToRuneList(), doOneLoop)
	case *RadList:
		runForLoopList(i, leftsNode, rightNode, coercedRight, doOneLoop)
	case *RadMap:
		runForLoopMap(i, leftsNode, coercedRight, doOneLoop)
	default:
		i.errorf(rightNode, "Cannot iterate through a %s", TypeAsString(rightVal))
	}
}

func runForLoopList(i *Interpreter, leftsNode, rightNode *ts.Node, list *RadList, doOneLoop func()) {
	var idxNode *ts.Node
	itemNodes := make([]*ts.Node, 0)

	leftNodes := i.getChildren(leftsNode, rl.F_LEFT)

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
			i.env.SetVar(idxName, newRadValue(i, idxNode, int64(idx)))
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

func runForLoopMap(i *Interpreter, leftsNode *ts.Node, radMap *RadMap, doOneLoop func()) {
	var keyNode *ts.Node
	var valueNode *ts.Node

	leftNodes := i.getChildren(leftsNode, rl.F_LEFT)
	numLefts := len(leftNodes)

	if numLefts == 0 || numLefts > 2 {
		i.errorf(leftsNode, "Expected 1 or 2 variables on left side of for loop")
	}

	keyNode = &leftNodes[0]
	if numLefts == 2 {
		valueNode = &leftNodes[1]
	}

	for _, key := range radMap.Keys() {
		keyName := i.sd.Src[keyNode.StartByte():keyNode.EndByte()]
		i.env.SetVar(keyName, key)

		if valueNode != nil {
			valueName := i.sd.Src[valueNode.StartByte():valueNode.EndByte()]
			value, _ := radMap.Get(key)
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

func (i *Interpreter) evaluateIndexing(rootNode *ts.Node, index ts.Node, val RadValue, expectReturnValue bool) RadValue {
	if index.Kind() == rl.K_CALL {
		// ufcs
		ufcsArg := &PosArg{
			// todo 'rootNode' is not great to use, it misses indexes in between that and this call,
			//  resulting in bad error pointing. could potentially replace ts.Node with interface
			//  'Pointable' i.e. a range we can point to in an error, that's ultimately all we need (?)
			node:  rootNode,
			value: val,
		}
		if expectReturnValue {
			return i.callFunction(&index, 1, ufcsArg)
		} else {
			return i.callFunction(&index, NO_NUM_RETURN_VALUES_CONSTRAINT, ufcsArg)
		}
	} else {
		return val.Index(i, &index)
	}
}

func (i *Interpreter) assignRightsToLefts(parentNode *ts.Node, leftNodes, rightNodes []ts.Node) {
	outputs := make([]RadValue, 0)
	for _, rightNode := range rightNodes {
		if rightNode.Kind() == rl.K_JSON_PATH {
			// json path assignment
			jsonFieldVar := NewJsonFieldVar(i, &leftNodes[len(outputs)], &rightNode) // todo index bounds error isn't user friendly
			i.env.SetJsonFieldVar(jsonFieldVar)
			outputs = append(outputs, JSON_SENTINEL)
		} else {
			out := i.evaluate(&rightNode, 1)
			if out.IsErrorToPropagate() {

			}
			outputs = append(outputs, out)
		}
	}

	if len(leftNodes) != len(outputs) {
		i.errorf(parentNode, "Cannot assign %d values to %d variables", len(outputs), len(leftNodes))
	}

	for idx, output := range outputs {
		if output == JSON_SENTINEL {
			// json path assignment, no need to assign
			continue
		}
		leftVarPathNode := &leftNodes[idx]
		i.doVarPathAssign(leftVarPathNode, output, false)
	}
}

func (i *Interpreter) defineCustomNamedFunction(fnNamedNode ts.Node) {
	nameNode := i.getChild(&fnNamedNode, rl.F_NAME)
	name := GetSrc(i.sd.Src, nameNode)
	lambda := NewLambda(i, &fnNamedNode)
	i.env.SetVar(name, newRadValueFn(lambda))
}

func (i *Interpreter) executeSwitchCase(caseValueAltNode *ts.Node, leftVarPathNodes []ts.Node) {
	numExpectedAssigns := len(leftVarPathNodes)
	switch caseValueAltNode.Kind() {
	case rl.K_SWITCH_CASE_EXPR:
		valueNodes := i.getChildren(caseValueAltNode, rl.F_VALUE)
		i.assignRightsToLefts(caseValueAltNode, leftVarPathNodes, valueNodes)
	case rl.K_SWITCH_CASE_BLOCK:
		stmtNodes := i.getChildren(caseValueAltNode, rl.F_STMT)
		i.runBlock(stmtNodes)

		if i.breakingOrContinuing() {
			return
		}

		yieldNode := i.getChild(caseValueAltNode, rl.F_YIELD_STMT)
		if numExpectedAssigns > 0 && yieldNode == nil {
			i.errorf(caseValueAltNode, "Cannot assign without yielding from the switch case")
		}
		if yieldNode != nil {
			valueNodes := i.getChildren(yieldNode, rl.F_VALUE)
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
