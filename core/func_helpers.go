package core

import (
	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
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

func (i *Interpreter) callFunction(callNode *ts.Node, ufcsArg *PosArg) RadValue {
	funcNameNode := rl.GetChild(callNode, rl.F_FUNC)
	argNodes := rl.GetChildren(callNode, rl.F_ARG)
	namedArgNodes := rl.GetChildren(callNode, rl.F_NAMED_ARG)

	funcName := i.GetSrcForNode(funcNameNode)

	var args []PosArg

	if ufcsArg != nil {
		args = append(args, *ufcsArg)
	}

	for _, argNode := range argNodes {
		argValue := i.eval(&argNode).Val
		if argValue != VOID_SENTINEL {
			args = append(args, NewPosArg(&argNode, argValue))
		}
	}

	namedArgs := make(map[string]namedArg)
	for _, namedArgNode := range namedArgNodes {
		namedArgNameNode := rl.GetChild(&namedArgNode, rl.F_NAME)
		namedArgValueNode := rl.GetChild(&namedArgNode, rl.F_VALUE)

		argName := i.GetSrcForNode(namedArgNameNode)
		argValue := i.eval(namedArgValueNode).Val

		_, exist := namedArgs[argName]
		if exist {
			i.emitErrorf(rl.ErrInvalidArgType, namedArgNameNode, "Duplicate named argument: %s", argName)
		}

		namedArgs[argName] = namedArg{
			name:      argName,
			value:     argValue,
			nameNode:  namedArgNameNode,
			valueNode: namedArgValueNode,
		}
	}

	val, exist := i.env.GetVar(funcName)
	if !exist {
		if funcName == "get_default" {
			i.emitErrorWithHint(rl.ErrUnknownFunction, funcNameNode,
				"Cannot invoke unknown function: get_default",
				"get_default was removed. Use: map[\"key\"] ?? default. See: https://amterp.github.io/rad/migrations/v0.8/")
		}
		i.emitErrorf(rl.ErrUnknownFunction, funcNameNode, "Cannot invoke unknown function: %s", funcName)
	}

	fn, ok := val.TryGetFn()
	if !ok {
		i.emitErrorf(rl.ErrTypeMismatch, funcNameNode, "Cannot invoke '%s' as a function: it is a %s", funcName, val.Type().AsString())
	}

	out := fn.Execute(NewFnInvocation(i, callNode, funcName, args, namedArgs, fn.IsBuiltIn()))
	return out
}
