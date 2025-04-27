package core

import (
	"fmt"
	com "rad/core/common"
	"strings"

	"github.com/amterp/rts/rsl"

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

type PosArg struct {
	node  *ts.Node
	value RslValue
}

func NewPosArg(node *ts.Node, value RslValue) PosArg {
	return PosArg{
		node:  node,
		value: value,
	}
}

func NewPosArgs(args ...PosArg) []PosArg {
	list := make([]PosArg, len(args))
	for i, arg := range args {
		list[i] = NewPosArg(arg.node, arg.value)
	}
	return list
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
	ufcsArg *PosArg,
) []RslValue {
	funcNameNode := i.getChild(callNode, rsl.F_FUNC)
	argNodes := i.getChildren(callNode, rsl.F_ARG)
	namedArgNodes := i.getChildren(callNode, rsl.F_NAMED_ARG)

	funcName := GetSrc(i.sd.Src, funcNameNode)

	var args []PosArg
	if ufcsArg != nil {
		args = append(args, *ufcsArg)
	}
	for _, argNode := range argNodes {
		// TODO 'expected output 1' prevents something like
		//  `print(function_that_returns_two_values())`, it should just "spread out" the args to print
		value := i.evaluate(&argNode, 1)[0]
		args = append(args, NewPosArg(&argNode, value))
	}

	namedArgs := make(map[string]namedArg)
	for _, namedArgNode := range namedArgNodes {
		namedArgNameNode := i.getChild(&namedArgNode, rsl.F_NAME)
		namedArgValueNode := i.getChild(&namedArgNode, rsl.F_VALUE)

		argName := GetSrc(i.sd.Src, namedArgNameNode)
		argValue := i.evaluate(namedArgValueNode, 1)[0]
		namedArgs[argName] = namedArg{
			name:      argName,
			value:     argValue,
			nameNode:  namedArgNameNode,
			valueNode: namedArgValueNode,
		}
	}

	val, exist := i.env.GetVar(funcName)
	if exist {
		// custom function
		fn := val.RequireFn(i, funcNameNode)
		return fn.Execute(NewFuncInvocationArgs(i, callNode, funcName, args, namedArgs, numExpectedOutputs))
	}

	f, exists := FunctionsByName[funcName] // todo replace this with variable in the environment
	if !exists {
		i.errorf(funcNameNode, "Unknown function: %s", funcName)
		panic(UNREACHABLE)
	}

	assertMinNumPosArgs(i, callNode, f, args)
	f.PosArgValidator.validate(i, callNode, f, args)
	assertAllowedNamedArgs(i, callNode, f, namedArgs)
	assertCorrectNumReturnValues(i, callNode, f, numExpectedOutputs)

	return f.Execute(NewFuncInvocationArgs(i, callNode, funcName, args, namedArgs, numExpectedOutputs))
}

func assertMinNumPosArgs(i *Interpreter, callNode *ts.Node, function Func, args []PosArg) {
	if len(args) < function.MinPosArgCount {
		i.errorf(callNode, "%s() requires at least %s, but got %d",
			function.Name, com.Pluralize(function.MinPosArgCount, "argument"), len(args))
	}
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
