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

type CtrlKind int

const (
	CtrlNormal CtrlKind = iota
	CtrlBreak
	CtrlContinue
	CtrlReturn
	CtrlYield
)

type EvalResult struct {
	Val  RadValue
	Ctrl CtrlKind
}

func NewEvalResult(val RadValue, ctrl CtrlKind) EvalResult {
	return EvalResult{
		Val:  val,
		Ctrl: ctrl,
	}
}

func NormalVal(val RadValue) EvalResult {
	return NewEvalResult(val, CtrlNormal)
}

func ReturnVal(val RadValue) EvalResult {
	return NewEvalResult(val, CtrlReturn)
}

func YieldVal(val RadValue) EvalResult {
	return NewEvalResult(val, CtrlYield)
}

var VoidNormal = NewEvalResult(VOID_SENTINEL, CtrlNormal)
var VoidBreak = NewEvalResult(VOID_SENTINEL, CtrlBreak)
var VoidContinue = NewEvalResult(VOID_SENTINEL, CtrlContinue)

type Interpreter struct {
	sd          *ScriptData
	env         *Env
	deferBlocks []*DeferBlock

	forWhileLoopLevel int
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
	res := i.safelyEvaluate(node)
	if res.Ctrl != CtrlNormal {
		i.errorf(node, "Bug? Unexpected control flow: %s", res.Ctrl)
	}
}

func (i *Interpreter) safelyEvaluate(node *ts.Node) EvalResult {
	defer func() {
		if r := recover(); r != nil {
			radPanic, ok := r.(*RadPanic)
			if ok {
				err := radPanic.Err()
				msg := err.Msg().Plain()
				if !com.IsBlank(string(err.Code)) {
					msg += fmt.Sprintf(" (code %s)", err.Code)
				}
				i.errorf(err.Node, msg)
			}
			if !IsTest {
				i.errorf(node, "Bug! Panic: %v\n%s", r, debug.Stack())
			}
		}
	}()
	return i.eval(node)
}

func (i *Interpreter) eval(node *ts.Node) (out EvalResult) {
	out = VoidNormal

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
			i.eval(&child)
		}
	case rl.K_COMMENT, rl.K_SHEBANG, rl.K_FILE_HEADER, rl.K_ARG_BLOCK:
		// no-op
	case rl.K_ERROR:
		i.errorf(node, "Bug! Error pre-check should've prevented running into this node")
	case rl.K_ASSIGN:
		rightNodes := i.getChildren(node, rl.F_RIGHT)
		leftNodes := i.getChildren(node, rl.F_LEFT)
		if len(leftNodes) > 0 {
			return i.assignRightsToLefts(leftNodes, rightNodes, false)
		} else {
			leftsNodes := i.getChildren(node, rl.F_LEFTS)
			return i.assignRightsToLefts(leftsNodes, rightNodes, true)
		}
	case rl.K_COMPOUND_ASSIGN:
		leftVarPathNode := i.getChild(node, rl.F_LEFT)
		rightNode := i.getChild(node, rl.F_RIGHT)
		opNode := i.getChild(node, rl.F_OP)
		newValue := i.executeCompoundOp(node, leftVarPathNode, rightNode, opNode)
		i.doVarPathAssign(leftVarPathNode, newValue, true)
	case rl.K_EXPR:
		catchNode := i.getChild(node, rl.F_CATCH)
		if catchNode != nil {
			defer func() {
				if r := recover(); r != nil {
					if radPanic, ok := r.(*RadPanic); ok {
						out = NormalVal(radPanic.ErrV)
					} else {
						panic(r)
					}
				}
			}()
		}
		out = i.eval(i.getChild(node, rl.F_DELEGATE))
	case rl.K_PASS:
		// no-op
	case rl.K_RETURN_STMT:
		return ReturnVal(i.evalRights(node))
	case rl.K_YIELD_STMT:
		return YieldVal(i.evalRights(node))
	case rl.K_BREAK_STMT:
		if i.forWhileLoopLevel > 0 {
			return VoidBreak
		}
		i.errorf(node, "Cannot 'break' outside of a for loop")
	case rl.K_CONTINUE_STMT:
		if i.forWhileLoopLevel > 0 {
			return VoidContinue
		}
		i.errorf(node, "Cannot 'continue' outside of a for loop")
	case rl.K_FOR_LOOP:
		i.forWhileLoopLevel++
		defer func() {
			i.forWhileLoopLevel--
		}()
		stmts := i.getChildren(node, rl.F_STMT)
		i.executeForLoop(node, func() EvalResult { return i.runBlock(stmts) })
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
				condValue = i.eval(condNode).Val.TruthyFalsy()
			}
			if condValue {
				res := i.runBlock(stmtNodes)
				switch res.Ctrl {
				case CtrlBreak:
					return VoidNormal
				case CtrlReturn, CtrlYield:
					return res
				}
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
				condResult := i.eval(condNode).Val.TruthyFalsy()
				shouldExecute = condResult
			}

			if shouldExecute {
				stmtNodes := i.getChildren(&altNode, rl.F_STMT)
				return i.runBlock(stmtNodes)
			}
		}
	case rl.K_SWITCH_STMT:
		discriminantNode := i.getChild(node, rl.F_DISCRIMINANT)
		caseNodes := i.getChildren(node, rl.F_CASE)
		defaultNode := i.getChild(node, rl.F_DEFAULT)

		discriminantVal := i.eval(discriminantNode).Val

		matchedCaseNodes := make([]ts.Node, 0)
		for _, caseNode := range caseNodes {
			caseKeyNodes := i.getChildren(&caseNode, rl.F_CASE_KEY)
			for _, caseKeyNode := range caseKeyNodes {
				caseKey := i.eval(&caseKeyNode).Val
				if caseKey.Equals(discriminantVal) {
					matchedCaseNodes = append(matchedCaseNodes, caseNode)
					break
				}
			}
		}

		if len(matchedCaseNodes) == 0 {
			if defaultNode != nil {
				caseValueAltNode := i.getChild(defaultNode, rl.F_ALT)
				return i.executeSwitchCase(caseValueAltNode)
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
		return i.executeSwitchCase(caseValueAltNode)
	case rl.K_FN_NAMED:
		// do not redefine top-level functions as they're already defined
		if node.Parent().Kind() != rl.K_SOURCE_FILE {
			i.defineCustomNamedFunction(*node)
		}
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
	case rl.K_PRIMARY_EXPR, rl.K_LITERAL:
		return i.eval(i.getOnlyChild(node))
	case rl.K_PARENTHESIZED_EXPR:
		return i.eval(i.getChild(node, rl.F_EXPR))
	case rl.K_UNARY_EXPR:
		delegateNode := i.getChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		opNode := i.getChild(node, rl.F_OP)
		argNode := i.getChild(node, rl.F_ARG)
		return NormalVal(newRadValues(i, node, i.executeUnaryOp(node, argNode, opNode)))
	case rl.K_OR_EXPR, rl.K_AND_EXPR, rl.K_COMPARE_EXPR, rl.K_ADD_EXPR, rl.K_MULT_EXPR:
		delegateNode := i.getChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		left := i.getChild(node, rl.F_LEFT)
		op := i.getChild(node, rl.F_OP)
		right := i.getChild(node, rl.F_RIGHT)
		return NormalVal(newRadValues(i, node, i.executeBinary(node, left, right, op)))

	// LEAF NODES
	case rl.K_IDENTIFIER:
		identifier := i.sd.Src[node.StartByte():node.EndByte()]
		val, ok := i.env.GetVar(identifier)
		if !ok {
			i.errorf(node, "Undefined variable: %s", identifier)
		}
		return NormalVal(newRadValues(i, node, val))
	case rl.K_VAR_PATH:
		rootNode := i.getChild(node, rl.F_ROOT)
		indexingNodes := i.getChildren(node, rl.F_INDEXING)
		val := i.eval(rootNode).Val
		for _, indexNode := range indexingNodes {
			val = i.evaluateIndexing(rootNode, indexNode, val)
		}
		return NormalVal(newRadValues(i, node, val))
	case rl.K_INDEXED_EXPR:
		rootNode := i.getChild(node, rl.F_ROOT)
		indexingNodes := i.getChildren(node, rl.F_INDEXING)
		if len(indexingNodes) > 0 {
			val := i.eval(rootNode).Val
			for _, index := range indexingNodes {
				val = i.evaluateIndexing(rootNode, index, val)
			}
			return NormalVal(newRadValues(i, node, val))
		} else {
			return i.eval(rootNode)
		}
	case rl.K_INT:
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asInt, _ := rts.ParseInt(asStr) // todo unhandled err
		return NormalVal(newRadValues(i, node, asInt))
	case rl.K_FLOAT:
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asFloat, _ := rts.ParseFloat(asStr) // todo unhandled err
		return NormalVal(newRadValues(i, node, asFloat))
	case rl.K_STRING:
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
				str = str.Concat(i.eval(&child).Val.RequireStr(i, &child))
			}
		}

		i.delimiterStack.Pop()

		return NormalVal(newRadValues(i, node, str))
	case rl.K_BOOL:
		asStr := i.sd.Src[node.StartByte():node.EndByte()]
		asBool, _ := strconv.ParseBool(asStr)
		return NormalVal(newRadValues(i, node, asBool))
	case rl.K_NULL:
		return NormalVal(newRadValues(i, node, nil))
	case rl.K_STRING_CONTENT:
		src := i.sd.Src[node.StartByte():node.EndByte()]
		return NormalVal(newRadValues(i, node, src))
	case rl.K_BACKSLASH:
		return NormalVal(newRadValues(i, node, "\\"))
	case rl.K_ESC_SINGLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "'" {
			return NormalVal(newRadValues(i, node, "'"))
		} else {
			return NormalVal(newRadValues(i, node, `\'`))
		}
	case rl.K_ESC_DOUBLE_QUOTE:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == `"` {
			return NormalVal(newRadValues(i, node, `"`))
		} else {
			return NormalVal(newRadValues(i, node, `\"`))
		}
	case rl.K_ESC_BACKTICK:
		if delim, ok := i.delimiterStack.Peek(); ok && delim.Open == "`" {
			return NormalVal(newRadValues(i, node, "`"))
		} else {
			return NormalVal(newRadValues(i, node, "\\`"))
		}
	case rl.K_ESC_NEWLINE:
		return NormalVal(newRadValues(i, node, "\n"))
	case rl.K_ESC_TAB:
		return NormalVal(newRadValues(i, node, "\t"))
	case rl.K_ESC_OPEN_BRACKET:
		return NormalVal(newRadValues(i, node, "{"))
	case rl.K_INTERPOLATION:
		exprResult := evaluateInterpolation(i, node)
		return NormalVal(newRadValues(i, node, exprResult))
	case rl.K_ESC_BACKSLASH:
		return NormalVal(newRadValues(i, node, "\\"))
	case rl.K_LIST:
		entries := i.getChildren(node, rl.F_LIST_ENTRY)
		list := NewRadList()
		for _, entry := range entries {
			list.Append(i.eval(&entry).Val)
		}
		return NormalVal(newRadValues(i, node, list))
	case rl.K_MAP:
		radMap := NewRadMap()
		entryNodes := i.getChildren(node, rl.F_MAP_ENTRY)
		for _, entryNode := range entryNodes {
			keyNode := i.getChild(&entryNode, rl.F_KEY)
			valueNode := i.getChild(&entryNode, rl.F_VALUE)
			key := evalMapKey(i, keyNode)
			radMap.Set(key, i.eval(valueNode).Val)
		}
		return NormalVal(newRadValues(i, node, radMap))
	case rl.K_CALL:
		return NormalVal(i.callFunction(node, nil))
	case rl.K_FN_LAMBDA:
		return NormalVal(newRadValues(i, node, NewLambda(i, node)))
	case rl.K_LIST_COMPREHENSION:
		resultExprNode := i.getChild(node, rl.F_EXPR)
		conditionNode := i.getChild(node, rl.F_CONDITION)

		resultList := NewRadList()
		doOneLoop := func() EvalResult {
			if conditionNode == nil || i.eval(conditionNode).Val.TruthyFalsy() {
				resultList.Append(i.eval(resultExprNode).Val)
			}
			return VoidNormal
		}
		i.executeForLoop(node, doOneLoop)

		return NormalVal(newRadValues(i, node, resultList))
	case rl.K_TERNARY_EXPR:
		delegateNode := i.getChild(node, rl.F_DELEGATE)
		if delegateNode != nil {
			return i.eval(delegateNode)
		}

		conditionNode := i.getChild(node, rl.F_CONDITION)
		trueNode := i.getChild(node, rl.F_TRUE_BRANCH)
		falseNode := i.getChild(node, rl.F_FALSE_BRANCH)
		condition := i.eval(conditionNode).Val.TruthyFalsy()
		return i.eval(lo.Ternary(condition, trueNode, falseNode))
	default:
		i.errorf(node, "Unsupported node kind: %s", node.Kind())
	}
	return
}

func (i *Interpreter) evalRights(node *ts.Node) RadValue {
	rightNodes := i.getChildren(node, rl.F_RIGHT)
	if len(rightNodes) == 0 {
		return VOID_SENTINEL
	}
	if len(rightNodes) == 1 {
		return i.eval(&rightNodes[0]).Val
	}
	list := NewRadList()
	for _, rightNode := range rightNodes {
		list.Append(i.eval(&rightNode).Val)
	}
	return newRadValue(i, node, list)
}

func evaluateInterpolation(i *Interpreter, interpNode *ts.Node) RadValue {
	exprNode := i.getChild(interpNode, rl.F_EXPR)
	formatNode := i.getChild(interpNode, rl.F_FORMAT)

	exprResult := i.eval(exprNode).Val
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
		padding := i.eval(paddingNode).Val.RequireInt(i, paddingNode)
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
		precision := i.eval(precisionNode).Val.RequireInt(i, precisionNode)

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

func (i *Interpreter) executeForLoop(node *ts.Node, doOneLoop func() EvalResult) EvalResult {
	leftsNode := i.getChild(node, rl.F_LEFTS)
	rightNode := i.getChild(node, rl.F_RIGHT)

	res := i.eval(rightNode)
	switch coercedRight := res.Val.Val.(type) {
	case RadString:
		return runForLoopList(i, leftsNode, rightNode, coercedRight.ToRuneList(), doOneLoop)
	case *RadList:
		return runForLoopList(i, leftsNode, rightNode, coercedRight, doOneLoop)
	case *RadMap:
		return runForLoopMap(i, leftsNode, coercedRight, doOneLoop)
	default:
		i.errorf(rightNode, "Cannot iterate through a %s", TypeAsString(res.Val))
		panic(UNREACHABLE)
	}
}

func runForLoopList(
	i *Interpreter,
	leftsNode, rightNode *ts.Node,
	list *RadList,
	doOneLoop func() EvalResult,
) EvalResult {
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

Loop:
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

		res := doOneLoop()
		switch res.Ctrl {
		case CtrlBreak:
			break Loop
		case CtrlReturn, CtrlYield:
			return res
		}
	}
	return VoidNormal
}

func runForLoopMap(i *Interpreter, leftsNode *ts.Node, radMap *RadMap, doOneLoop func() EvalResult) EvalResult {
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

		res := doOneLoop()
		switch res.Ctrl {
		case CtrlBreak:
			break
		case CtrlReturn, CtrlYield:
			return res
		}
	}
	return VoidNormal
}

// if stmts, for loops
func (i *Interpreter) runBlock(stmtNodes []ts.Node) EvalResult {
	for _, stmtNode := range stmtNodes {
		res := i.eval(&stmtNode)
		if res.Ctrl != CtrlNormal {
			return res
		}
	}
	return VoidNormal
}

func (i *Interpreter) runWithChildEnv(runnable func()) {
	originalEnv := i.env
	env := originalEnv.NewChildEnv()
	i.env = &env
	runnable()
	i.env = originalEnv
}

func (i *Interpreter) evaluateIndexing(rootNode *ts.Node, index ts.Node, val RadValue) RadValue {
	if index.Kind() == rl.K_CALL {
		// ufcs
		ufcsArg := &PosArg{
			// todo 'rootNode' is not great to use, it misses indexes in between that and this call,
			//  resulting in bad error pointing. could potentially replace ts.Node with interface
			//  'Pointable' i.e. a range we can point to in an error, that's ultimately all we need (?)
			node:  rootNode,
			value: val,
		}
		return i.callFunction(&index, ufcsArg)
	} else {
		return val.Index(i, &index)
	}
}

func (i *Interpreter) assignRightsToLefts(leftNodes []ts.Node, rightNodes []ts.Node, destructure bool) EvalResult {
	if destructure {
		if len(rightNodes) == 1 {
			rightNode := rightNodes[0]
			if rightNode.Kind() == rl.K_JSON_PATH {
				jsonFieldVar := NewJsonFieldVar(i, &leftNodes[0], &rightNode)
				i.env.SetJsonFieldVar(jsonFieldVar)
			} else {
				val := i.eval(&rightNode).Val
				list, ok := val.TryGetList()
				if ok {
					for idx, leftNode := range leftNodes {
						if len(list.Values) > idx {
							val := list.Values[idx]
							i.doVarPathAssign(&leftNode, val, false)
						} else {
							i.doVarPathAssign(&leftNode, RAD_NULL_VAL, false)
						}
					}
					return VoidNormal
				} else {
					i.doVarPathAssign(&leftNodes[0], val, false)
				}
			}

			for _, leftNode := range leftNodes[1:] {
				i.doVarPathAssign(&leftNode, RAD_NULL_VAL, false)
			}
			return VoidNormal
		}

		for idx, leftNode := range leftNodes {
			if len(rightNodes) > idx {
				rightNode := rightNodes[idx]
				if rightNode.Kind() == rl.K_JSON_PATH {
					jsonFieldVar := NewJsonFieldVar(i, &leftNode, &rightNode)
					i.env.SetJsonFieldVar(jsonFieldVar)
				} else {
					val := i.eval(&rightNode).Val
					i.doVarPathAssign(&leftNode, val, false)
				}
			} else {
				i.doVarPathAssign(&leftNode, RAD_NULL_VAL, false)
			}
		}
		return VoidNormal
	}

	// not destructuring, means exactly 1 left node

	if len(rightNodes) == 1 {
		rightNode := rightNodes[0]
		if rightNode.Kind() == rl.K_JSON_PATH {
			jsonFieldVar := NewJsonFieldVar(i, &leftNodes[0], &rightNode)
			i.env.SetJsonFieldVar(jsonFieldVar)
		} else {
			res := i.eval(&rightNodes[0])
			if res.Ctrl != CtrlNormal {
				return res
			}
			if res.Val == VOID_SENTINEL {
				i.errorf(&rightNode, "Cannot assign to a void value")
			}
			i.doVarPathAssign(&leftNodes[0], res.Val, false)
		}
		return VoidNormal
	}

	// not destructuring (so 1 left) & not 1 right node;
	// means at least 2 right nodes -> pack into list and assign to 1 left

	list := NewRadList()
	for _, rightNode := range rightNodes {
		val := i.eval(&rightNode).Val
		list.Append(val)
	}
	i.doVarPathAssign(&leftNodes[0], newRadValueList(list), false)
	return VoidNormal
}

func (i *Interpreter) defineCustomNamedFunction(fnNamedNode ts.Node) {
	nameNode := i.getChild(&fnNamedNode, rl.F_NAME)
	name := GetSrc(i.sd.Src, nameNode)
	lambda := NewLambda(i, &fnNamedNode)
	i.env.SetVar(name, newRadValueFn(lambda))
}

func (i *Interpreter) executeSwitchCase(caseValueAltNode *ts.Node) EvalResult {
	switch caseValueAltNode.Kind() {
	case rl.K_SWITCH_CASE_EXPR:
		return NormalVal(i.evalRights(caseValueAltNode))
	case rl.K_SWITCH_CASE_BLOCK:
		stmtNodes := i.getChildren(caseValueAltNode, rl.F_STMT)
		res := i.runBlock(stmtNodes)
		switch res.Ctrl {
		case CtrlNormal, CtrlBreak, CtrlContinue, CtrlReturn:
			return res
		case CtrlYield:
			return NormalVal(res.Val)
		}
	default:
		i.errorf(caseValueAltNode, "Bug! Unsupported switch case value node kind: %s", caseValueAltNode.Kind())
	}
	return VoidNormal
}
