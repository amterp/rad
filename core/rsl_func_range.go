package core

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// todo
//   - implement steps?
//   - somehow improve implementation to be a generator, rather than eagerly created list? chugs at e.g. 100_000

var FuncRange = Func{
	Name:             FUNC_RANGE,
	ReturnValues:     ONE_RETURN_VAL,
	RequiredArgCount: 1,
	ArgTypes:         [][]RslTypeEnum{{RslIntT, RslFloatT}, {RslIntT, RslFloatT}, {RslIntT, RslFloatT}},
	NamedArgs:        NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) []RslValue {
		useFloats := false
		for _, arg := range f.args {
			switch arg.value.Type() {
			case RslFloatT:
				useFloats = true
			case RslIntT:
			default:
				bugIncorrectTypes(FUNC_RANGE)
			}
		}

		if useFloats {
			return newRslValues(f.i, f.callNode, runFloatRange(f.i, f.callNode, f.args))
		} else {
			return newRslValues(f.i, f.callNode, runIntRange(f.i, f.callNode, f.args))
		}
	},
}

func runFloatRange(interp *Interpreter, callNode *ts.Node, args []positionalArg) []RslValue {
	var start, end, step float64

	firstArg := args[0]
	secondArg := tryGetArg(1, args)
	thirdArg := tryGetArg(2, args)

	if thirdArg != nil {
		start = firstArg.value.RequireFloatAllowingInt(interp, firstArg.node)
		end = secondArg.value.RequireFloatAllowingInt(interp, secondArg.node)
		step = thirdArg.value.RequireFloatAllowingInt(interp, thirdArg.node)
	} else if secondArg != nil {
		start = firstArg.value.RequireFloatAllowingInt(interp, firstArg.node)
		end = secondArg.value.RequireFloatAllowingInt(interp, secondArg.node)
		step = 1
	} else {
		start = 0
		end = firstArg.value.RequireFloatAllowingInt(interp, firstArg.node)
		step = 1
	}

	if step == 0 {
		// third node must be present if step is zero
		interp.errorf(thirdArg.node,
			"%s() step argument cannot be zero", FUNC_RANGE)
	}

	if start > end && step > 0 {
		interp.errorf(callNode,
			"%s() start %f cannot be greater than end %f with positive step %f", FUNC_RANGE, start, end, step)
	}

	if start < end && step < 0 {
		interp.errorf(callNode,
			"%s() start %f cannot be less than end %f with negative step %f", FUNC_RANGE, start, end, step)
	}

	var result []RslValue

	if step < 0 {
		for i := start; i > end; i += step {
			result = append(result, newRslValue(interp, callNode, i))
		}
	} else {
		for i := start; i < end; i += step {
			result = append(result, newRslValue(interp, callNode, i))
		}
	}

	return result
}

func runIntRange(interp *Interpreter, callNode *ts.Node, args []positionalArg) []RslValue {
	var start, end, step int64

	firstArg := args[0]
	secondArg := tryGetArg(1, args)
	thirdArg := tryGetArg(2, args)

	if thirdArg != nil {
		start = firstArg.value.RequireInt(interp, firstArg.node)
		end = secondArg.value.RequireInt(interp, secondArg.node)
		step = thirdArg.value.RequireInt(interp, thirdArg.node)
	} else if secondArg != nil {
		start = firstArg.value.RequireInt(interp, firstArg.node)
		end = secondArg.value.RequireInt(interp, secondArg.node)
		step = 1
	} else {
		start = 0
		end = firstArg.value.RequireInt(interp, firstArg.node)
		step = 1
	}

	if step == 0 {
		// third node must be present if step is zero
		interp.errorf(thirdArg.node,
			"%s() step argument cannot be zero", FUNC_RANGE)
	}

	if start > end && step > 0 {
		interp.errorf(callNode,
			"%s() start %d cannot be greater than end %d with positive step %d", FUNC_RANGE, start, end, step)
	}

	if start < end && step < 0 {
		interp.errorf(callNode,
			"%s() start %d cannot be less than end %d with negative step %d", FUNC_RANGE, start, end, step)
	}

	var result []RslValue

	if step < 0 {
		for i := start; i > end; i += step {
			result = append(result, newRslValue(interp, callNode, i))
		}
	} else {
		for i := start; i < end; i += step {
			result = append(result, newRslValue(interp, callNode, i))
		}
	}

	return result
}
