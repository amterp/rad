package core

import (
	"github.com/amterp/rad/rts/rl"
	"math/rand"
	"time"
)

var RNG *rand.Rand

func init() {
	RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
}

var FuncSeedRandom = BuiltInFunc{
	Name:            FUNC_SEED_RANDOM,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  1,
	PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadIntT}}),
	NamedArgs:       NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) RadValue {
		arg := f.args[0]
		asInt := arg.value.RequireInt(f.i, arg.node)
		RNG = rand.New(rand.NewSource(asInt))
		return VOID_SENTINEL
	},
}

var FuncRand = BuiltInFunc{
	Name:            FUNC_RAND,
	ReturnValues:    ONE_RETURN_VAL,
	MinPosArgCount:  0,
	PosArgValidator: NO_POS_ARGS,
	NamedArgs:       NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) RadValue {
		return newRadValues(f.i, f.callNode, RNG.Float64())
	},
}

var FuncRandInt = BuiltInFunc{
	Name:            FUNC_RAND_INT,
	ReturnValues:    ONE_RETURN_VAL,
	MinPosArgCount:  1,
	PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadIntT}, {rl.RadIntT}}),
	NamedArgs:       NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) RadValue {
		var min, max int64

		if len(f.args) == 1 {
			arg := f.args[0]
			min = 0
			max = arg.value.RequireInt(f.i, arg.node)
		} else {
			// two args
			minArg := f.args[0]
			maxArg := f.args[1]
			min = minArg.value.RequireInt(f.i, minArg.node)
			max = maxArg.value.RequireInt(f.i, maxArg.node)
		}

		if min >= max {
			f.i.errorf(f.callNode,
				"%s() min (%d) must be less than max (%d).", FUNC_RAND_INT, min, max)
		}

		n := max - min
		return newRadValues(f.i, f.callNode, min+RNG.Int63n(n))
	},
}
