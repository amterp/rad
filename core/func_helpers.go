package core

import (
	"github.com/amterp/rad/rts/rl"
)

type PosArg struct {
	node  rl.Node
	value RadValue
}

func NewPosArg(node rl.Node, value RadValue) PosArg {
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
	nameNode  rl.Node
	valueNode rl.Node
}

func (i *Interpreter) callFunction(callNode *rl.Call, ufcsArg *PosArg) RadValue {
	// Resolve the function name from the Call's Func node
	funcExpr := callNode.Func
	funcName := ""
	if id, ok := funcExpr.(*rl.Identifier); ok {
		funcName = id.Name
	} else if vp, ok := funcExpr.(*rl.VarPath); ok {
		if rootId, ok := vp.Root.(*rl.Identifier); ok {
			funcName = rootId.Name
		}
	}

	var args []PosArg

	if ufcsArg != nil {
		args = append(args, *ufcsArg)
	}

	for _, argNode := range callNode.Args {
		argValue := i.eval(argNode).Val
		if argValue != VOID_SENTINEL {
			args = append(args, NewPosArg(argNode, argValue))
		}
	}

	namedArgs := make(map[string]namedArg)
	for _, na := range callNode.NamedArgs {
		argName := na.Name
		nameNode := rl.NewIdentifier(na.NameSpan, na.Name)
		argValue := i.eval(na.Value).Val

		_, exist := namedArgs[argName]
		if exist {
			i.emitErrorf(rl.ErrInvalidArgType, nameNode, "Duplicate named argument: %s", argName)
		}

		namedArgs[argName] = namedArg{
			name:      argName,
			value:     argValue,
			nameNode:  nameNode,
			valueNode: na.Value,
		}
	}

	// For UFCS or method-like calls, resolve via VarPath
	if funcName == "" {
		// Dynamic call - evaluate Func expression to get callable
		funcVal := i.eval(funcExpr).Val
		fn, ok := funcVal.TryGetFn()
		if !ok {
			i.emitErrorf(rl.ErrTypeMismatch, funcExpr, "Cannot invoke as a function: it is a %s", funcVal.Type().AsString())
		}
		out := fn.Execute(NewFnInvocation(i, callNode, "<dynamic>", args, namedArgs, fn.IsBuiltIn()))
		return out
	}

	val, exist := i.env.GetVar(funcName)
	if !exist {
		switch funcName {
		case "get_default":
			i.emitErrorWithHint(rl.ErrUnknownFunction, funcExpr,
				"Cannot invoke unknown function: get_default",
				"get_default was removed. Use: map[\"key\"] ?? default. See: https://amterp.github.io/rad/migrations/v0.8/")
		case "get_stash_dir":
			i.emitErrorWithHint(rl.ErrUnknownFunction, funcExpr,
				"Cannot invoke unknown function: get_stash_dir",
				"get_stash_dir was renamed to get_stash_path. See: https://amterp.github.io/rad/migrations/v0.9/")
		}
		i.emitErrorf(rl.ErrUnknownFunction, funcExpr, "Cannot invoke unknown function: %s", funcName)
	}

	fn, ok := val.TryGetFn()
	if !ok {
		i.emitErrorf(rl.ErrTypeMismatch, funcExpr, "Cannot invoke '%s' as a function: it is a %s", funcName, val.Type().AsString())
	}

	out := fn.Execute(NewFnInvocation(i, callNode, funcName, args, namedArgs, fn.IsBuiltIn()))
	return out
}
