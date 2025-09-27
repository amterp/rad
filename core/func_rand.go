package core

import (
	"math/rand"

	"github.com/amterp/rad/rts/rl"
)

var FuncSeedRandom = BuiltInFunc{
	Name: FUNC_SEED_RANDOM,
	Execute: func(f FuncInvocation) RadValue {
		RNG = rand.New(rand.NewSource(f.GetInt("_seed")))
		return VOID_SENTINEL
	},
}

var FuncRand = BuiltInFunc{
	Name: FUNC_RAND,
	Execute: func(f FuncInvocation) RadValue {
		return f.Return(RNG.Float64())
	},
}

var FuncRandInt = BuiltInFunc{
	Name: FUNC_RAND_INT,
	Execute: func(f FuncInvocation) RadValue {
		arg1 := f.GetInt("_arg1")
		arg2 := f.GetArg("_arg2")

		var min, max int64

		if arg2.IsNull() {
			min = 0
			max = arg1
		} else {
			min = arg1
			max = arg2.RequireInt(f.i, f.callNode)
		}

		if min >= max {
			return f.ReturnErrf(rl.ErrArgsContradict, "min (%d) must be less than max (%d).", min, max)
		}

		n := max - min
		return newRadValues(f.i, f.callNode, min+RNG.Int63n(n))
	},
}
