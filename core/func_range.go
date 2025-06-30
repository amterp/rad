package core

import (
	"github.com/amterp/rad/rts/rl"
)

// todo
//   - somehow improve implementation to be a generator, rather than eagerly created list? chugs at e.g. 100_000

var FuncRange = BuiltInFunc{
	Name: FUNC_RANGE,
	Execute: func(f FuncInvocation) RadValue {
		useFloats := false

		arg1 := f.GetArg("_arg1")
		arg2 := f.GetArg("_arg2")
		step := f.GetArg("_step")

		for _, arg := range []RadValue{arg1, arg2, step} {
			switch arg.Type() {
			case rl.RadFloatT:
				useFloats = true
			case rl.RadIntT, rl.RadNullT:
			default:
				bugIncorrectTypes(FUNC_RANGE)
			}
		}

		if useFloats {
			return runFloatRange(f, arg1, arg2, step)
		} else {
			return runIntRange(f, arg1, arg2, step)
		}
	},
}

func runFloatRange(f FuncInvocation, arg1, arg2, stepArg RadValue) RadValue {
	var start, end, step float64

	if arg2.IsNull() {
		start = 0.0
		end = arg1.RequireFloatAllowingInt(f.i, f.callNode)
	} else {
		start = arg1.RequireFloatAllowingInt(f.i, f.callNode)
		end = arg2.RequireFloatAllowingInt(f.i, f.callNode)
	}
	step = stepArg.RequireFloatAllowingInt(f.i, f.callNode)

	if step == 0 {
		return f.ReturnErrf(rl.ErrNumInvalidRange, "Step argument cannot be zero")
	}

	if start > end && step > 0 {
		return f.ReturnErrf(rl.ErrArgsContradict, "Start %f cannot be greater than end %f with positive step %f",
			start, end, step)
	}

	if start < end && step < 0 {
		return f.ReturnErrf(rl.ErrArgsContradict, "Start %f cannot be less than end %f with negative step %f",
			start, end, step)
	}

	var result []RadValue

	if step < 0 {
		for i := start; i > end; i += step {
			result = append(result, newRadValueFloat64(i))
		}
	} else {
		for i := start; i < end; i += step {
			result = append(result, newRadValueFloat64(i))
		}
	}

	return f.Return(result)
}

func runIntRange(f FuncInvocation, arg1, arg2, stepArg RadValue) RadValue {
	var start, end, step int64

	if arg2.IsNull() {
		start = 0
		end = arg1.RequireInt(f.i, f.callNode)
	} else {
		start = arg1.RequireInt(f.i, f.callNode)
		end = arg2.RequireInt(f.i, f.callNode)
	}
	step = stepArg.RequireInt(f.i, f.callNode)

	if step == 0 {
		return f.ReturnErrf(
			rl.ErrNumInvalidRange,
			"Step argument cannot be zero")
	}

	if start > end && step > 0 {
		return f.ReturnErrf(
			rl.ErrArgsContradict,
			"Start %d cannot be greater than end %d with positive step %d",
			start, end, step)
	}
	if start < end && step < 0 {
		return f.ReturnErrf(
			rl.ErrArgsContradict,
			"Start %d cannot be less than end %d with negative step %d",
			start, end, step)
	}

	var result []RadValue
	if step < 0 {
		for i := start; i > end; i += step {
			result = append(result, newRadValueInt64(i))
		}
	} else {
		for i := start; i < end; i += step {
			result = append(result, newRadValueInt64(i))
		}
	}

	return f.Return(result)
}
