package core

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type Interpreter struct {
	sd  *ScriptData
	env *Env
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
	case K_COMMENT, K_SHEBANG, K_FILE_HEADER, K_ARG_BLOCK:
		return

	case K_SOURCE_FILE:
		children := node.Children(node.Walk())
		for _, child := range children {
			i.recursivelyRun(&child)
		}
	case K_ASSIGN:
		leftVarPaths := i.getChildren(node, F_LEFT)
		right := i.getChild(node, F_RIGHT)
		numExpectedOutputs := len(leftVarPaths)
		values := i.evaluate(right, numExpectedOutputs)
		for idx, leftVarPath := range leftVarPaths {
			rightValue := values[idx]

			rootIdentifier := i.getChild(&leftVarPath, F_ROOT) // identifier required by grammar
			rootIdentifierName := i.sd.Src[rootIdentifier.StartByte():rootIdentifier.EndByte()]
			indexings := i.getChildren(&leftVarPath, F_INDEXING)
			val, ok := i.env.GetVar(rootIdentifierName)

			if len(indexings) == 0 {
				// simple assignment, no collection lookups
				i.env.SetVar(rootIdentifierName, rightValue)
				continue
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
	case K_EXPR_STMT:
		i.evaluate(i.getOnlyChild(node), NO_NUM_RETURN_VALUES_CONSTRAINT)
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
	case K_LITERAL:
		return i.evaluate(i.getOnlyChild(node), numExpectedOutputs)
	case K_BINARY_OP, K_COMPARISON_OP:
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
		contentsNode := i.getChild(node, F_CONTENTS)
		str := NewRslString("")
		for _, child := range contentsNode.Children(contentsNode.Walk()) {
			str = str.Concat(i.evaluate(&child, 1)[0].RequireStr(i, &child))
		}
		return newRslValues(i, node, str)
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
		entries := i.getChildren(node, F_MAP_ENTRY)
		for _, entry := range entries {
			keyNode := i.getChild(&entry, F_KEY)
			valueNode := i.getChild(&entry, F_VALUE)
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
		i.errorf(node, "Expected exactly one child, got %d", count)
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
