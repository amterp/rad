package core

import (
	"fmt"
	com "rad/core/common"
	"strings"

	"github.com/amterp/rad/rts/rl"

	"github.com/dustin/go-humanize/english"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var (
	EMPTY []RadValue

	NO_RETURN_LIMIT       = []int{NO_NUM_RETURN_VALUES_CONSTRAINT}
	ZERO_RETURN_VALS      = []int{}
	ONE_RETURN_VAL        = []int{1}
	UP_TO_TWO_RETURN_VALS = []int{1, 2}
)

type PosArg struct {
	node  *ts.Node
	value RadValue
}

func NewPosArg(node *ts.Node, value RadValue) PosArg {
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
	value     RadValue
	nameNode  *ts.Node
	valueNode *ts.Node
}

func (i *Interpreter) callFunction(
	callNode *ts.Node,
	numExpectedOutputs int,
	ufcsArg *PosArg,
) []RadValue {
	funcNameNode := i.getChild(callNode, rl.F_FUNC)
	argNodes := i.getChildren(callNode, rl.F_ARG)
	namedArgNodes := i.getChildren(callNode, rl.F_NAMED_ARG)

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
		namedArgNameNode := i.getChild(&namedArgNode, rl.F_NAME)
		namedArgValueNode := i.getChild(&namedArgNode, rl.F_VALUE)

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
	if !exist {
		i.errorf(funcNameNode, "Cannot invoke unknown function: %s", funcName)
	}

	fn, ok := val.TryGetFn()
	if !ok {
		i.errorf(funcNameNode, "Cannot invoke '%s' as a function: it is a %s", funcName, val.Type().AsString())
	}

	return fn.Execute(NewFuncInvocationArgs(i, callNode, funcName, args, namedArgs, numExpectedOutputs))
}

func assertMinNumPosArgs(f FuncInvocationArgs, builtInFunc *BuiltInFunc) {
	if len(f.args) < builtInFunc.MinPosArgCount {
		f.i.errorf(f.callNode, "%s() requires at least %s, but got %d",
			builtInFunc.Name, com.Pluralize(builtInFunc.MinPosArgCount, "argument"), len(f.args))
	}
}

func assertCorrectNumReturnValues(f FuncInvocationArgs, builtInFunc *BuiltInFunc) {
	allowedNumReturnValues := builtInFunc.ReturnValues
	if f.numExpectedOutputs == NO_NUM_RETURN_VALUES_CONSTRAINT {
		return
	}

	if lo.Contains(allowedNumReturnValues, f.numExpectedOutputs) {
		return
	}

	if lo.Contains(allowedNumReturnValues, NO_NUM_RETURN_VALUES_CONSTRAINT) {
		return
	}

	var errMsg string
	if len(allowedNumReturnValues) == 0 {
		errMsg = fmt.Sprintf("%s() returns no values, but %s expected",
			builtInFunc.Name, com.NumIsAre(f.numExpectedOutputs))
	} else if len(allowedNumReturnValues) == 1 {
		errMsg = fmt.Sprintf("%s() returns %s, but %s expected",
			builtInFunc.Name, com.Pluralize(allowedNumReturnValues[0], "value"), com.NumIsAre(f.numExpectedOutputs))
	} else {
		// allows different numbers of return values
		stringified := lo.Map(allowedNumReturnValues, func(item int, _ int) string { return fmt.Sprintf("%d", item) })
		allowedReturnNums := strings.Join(stringified, " or ")
		errMsg = fmt.Sprintf("%s() returns %s values, but %s expected",
			builtInFunc.Name, allowedReturnNums, com.NumIsAre(f.numExpectedOutputs))
	}
	f.i.errorf(f.callNode, errMsg)
}

func assertAllowedNamedArgs(f FuncInvocationArgs, builtInFunc *BuiltInFunc) {
	allowedNamedArgs := builtInFunc.NamedArgs
	allowedNames := lo.Keys(allowedNamedArgs)

	// check for invalid names
	for actualName, arg := range f.namedArgs {
		if !lo.Contains(allowedNames, actualName) {
			f.i.errorf(arg.nameNode, "Unknown named argument %q", actualName)
		}
	}

	// check for invalid types
	for name, arg := range f.namedArgs {
		allowedTypes, _ := allowedNamedArgs[name]
		if len(allowedTypes) > 0 && !lo.Contains(allowedTypes, arg.value.Type()) {
			acceptable := english.OxfordWordSeries(
				lo.Map(allowedTypes, func(t RadTypeEnum, _ int) string { return t.AsString() }), "or")
			f.i.errorf(arg.valueNode, "%s(): Named arg %s was %s, but must be: %s",
				builtInFunc.Name, name, arg.value.Type().AsString(), acceptable)
		}
	}
}
