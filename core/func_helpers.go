package core

import (
	"fmt"
	com "rad/core/common"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var (
	EMPTY []RslValue

	NO_RETURN_LIMIT       = []int{NO_NUM_RETURN_VALUES_CONSTRAINT}
	ZERO_RETURN_VALS      = []int{}
	ONE_RETURN_VAL        = []int{1}
	UP_TO_TWO_RETURN_VALS = []int{1, 2}
)

type positionalArg struct {
	node  *ts.Node
	value RslValue
}

type namedArg struct {
	name      string
	value     RslValue
	nameNode  *ts.Node
	valueNode *ts.Node
}

func (i *Interpreter) callFunction(
	callNode *ts.Node,
	numExpectedOutputs int,
) []RslValue {
	funcNameNode := i.getChild(callNode, F_FUNC)
	argNodes := i.getChildren(callNode, F_ARG)
	namedArgNodes := i.getChildren(callNode, F_NAMED_ARG)

	funcName := i.sd.Src[funcNameNode.StartByte():funcNameNode.EndByte()]

	var args []positionalArg
	for _, argNode := range argNodes {
		// TODO 'expected output 1' prevents something like
		//  `print(function_that_returns_two_values())`, it should just "spread out" the args to print
		value := i.evaluate(&argNode, 1)[0]
		args = append(args, positionalArg{node: &argNode, value: value})
	}

	namedArgs := make(map[string]namedArg)
	for _, namedArgNode := range namedArgNodes {
		namedArgNameNode := i.getChild(&namedArgNode, F_NAME)
		namedArgValueNode := i.getChild(&namedArgNode, F_VALUE)

		argName := i.sd.Src[namedArgNameNode.StartByte():namedArgNameNode.EndByte()]
		argValue := i.evaluate(namedArgValueNode, 1)[0]
		namedArgs[argName] = namedArg{
			name:      argName,
			value:     argValue,
			nameNode:  namedArgNameNode,
			valueNode: namedArgValueNode,
		}
	}

	f, exists := FunctionsByName[funcName]
	if !exists {
		i.errorf(funcNameNode, "Unknown function: %s", funcName)
		panic(UNREACHABLE)
	}

	assertCorrectNumReturnValues(i, callNode, f, numExpectedOutputs)
	assertCorrectPositionalArgs(i, callNode, f, args)
	assertAllowedNamedArgs(i, callNode, f, namedArgs)

	return f.Execute(NewFuncInvocationArgs(i, callNode, args, namedArgs, numExpectedOutputs))
}

func assertCorrectNumReturnValues(i *Interpreter, callNode *ts.Node, function Func, numExpectedReturnValues int) {
	allowedNumReturnValues := function.ReturnValues
	if numExpectedReturnValues == NO_NUM_RETURN_VALUES_CONSTRAINT {
		return
	}

	if lo.Contains(allowedNumReturnValues, numExpectedReturnValues) {
		return
	}

	if lo.Contains(allowedNumReturnValues, NO_NUM_RETURN_VALUES_CONSTRAINT) {
		return
	}

	var errMsg string
	if len(allowedNumReturnValues) == 0 {
		errMsg = fmt.Sprintf("%s() returns no values, but %s expected",
			function.Name, com.NumIsAre(numExpectedReturnValues))
	} else if len(allowedNumReturnValues) == 1 {
		errMsg = fmt.Sprintf("%s() returns %s, but %s expected",
			function.Name, com.Pluralize(allowedNumReturnValues[0], "value"), com.NumIsAre(numExpectedReturnValues))
	} else {
		// allows different numbers of return values
		stringified := lo.Map(allowedNumReturnValues, func(item int, _ int) string { return fmt.Sprintf("%d", item) })
		allowedReturnNums := strings.Join(stringified, " or ")
		errMsg = fmt.Sprintf("%s() returns %s values, but %s expected",
			function.Name, allowedReturnNums, com.NumIsAre(numExpectedReturnValues))
	}
	i.errorf(callNode, errMsg)
}

func assertCorrectPositionalArgs(i *Interpreter, callNode *ts.Node, function Func, args []positionalArg) {
	if len(args) < function.MinPosArgCount {
		i.errorf(callNode, "%s() requires at least %s, but got %d",
			function.Name, com.Pluralize(function.MinPosArgCount, "argument"), len(args))
	}

	maxAcceptableArgs := len(function.PosArgTypes)
	if len(args) > maxAcceptableArgs {
		i.errorf(callNode, "%s() requires at most %s, but got %d",
			function.Name, com.Pluralize(maxAcceptableArgs, "argument"), len(args))
	}

	for idx, acceptableTypes := range function.PosArgTypes {
		if len(acceptableTypes) == 0 {
			// there are no type constraints
			continue
		}

		if idx >= len(args) {
			// rest of the args are optional and not supplied
			break
		}

		arg := args[idx]
		if !lo.Contains(acceptableTypes, arg.value.Type()) {
			acceptable := english.OxfordWordSeries(
				lo.Map(acceptableTypes, func(t RslTypeEnum, _ int) string { return t.AsString() }), "or")
			i.errorf(arg.node, "Got %q as the %s argument of %s(), but must be: %s",
				arg.value.Type().AsString(), humanize.Ordinal(idx+1), function.Name, acceptable)
		}
	}
}

func assertAllowedNamedArgs(i *Interpreter, callNode *ts.Node, function Func, namedArgs map[string]namedArg) {
	allowedNamedArgs := function.NamedArgs
	allowedNames := lo.Keys(allowedNamedArgs)

	// check for invalid names
	for actualName, arg := range namedArgs {
		if !lo.Contains(allowedNames, actualName) {
			i.errorf(arg.nameNode, "Unknown named argument %q", actualName)
		}
	}

	// check for invalid types
	for name, arg := range namedArgs {
		allowedTypes, _ := allowedNamedArgs[name]
		if len(allowedTypes) > 0 && !lo.Contains(allowedTypes, arg.value.Type()) {
			acceptable := english.OxfordWordSeries(
				lo.Map(allowedTypes, func(t RslTypeEnum, _ int) string { return t.AsString() }), "or")
			i.errorf(arg.valueNode, "%s(): Named arg %s was %s, but must be: %s",
				function.Name, name, arg.value.Type().AsString(), acceptable)
		}
	}
}
