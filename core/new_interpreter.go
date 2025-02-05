package core

import (
	"fmt"
	"runtime/debug"
	"strconv"

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
			i.errorf(node, "Bug! Panic: %v\n%s", r, debug.Stack())
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

	// LEAF NODES
	case K_VAR_PATH:
		rootIdentifier := i.getChild(node, F_ROOT) // identifier required by grammar
		rootIdentifierName := i.sd.Src[rootIdentifier.StartByte():rootIdentifier.EndByte()]
		indexings := i.getChildren(node, F_INDEXING)
		val, ok := i.env.GetVar(rootIdentifierName)
		if !ok {
			i.errorf(rootIdentifier, "Undefined variable: %s", rootIdentifierName)
		}
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
	case K_LIST:
		i.assertExpectedNumOutputs(node, numExpectedOutputs, 1)
		entries := i.getChildren(node, F_LIST_ENTRY)
		list := NewRslList()
		for _, entry := range entries {
			list.Append(i.evaluate(&entry, 1)[0])
		}
		return newRslValues(i, node, list)
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

func (i *Interpreter) errorf(node *ts.Node, format string, args ...interface{}) {
	RP.CtxErrorExit(NewCtx(i.sd.Src, node), fmt.Sprintf(format, args...))
}
