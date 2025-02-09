package core

import (
	"math/rand"
	"time"

	ts "github.com/tree-sitter/go-tree-sitter"
)

var RNG *rand.Rand

func init() {
	RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
}

var FuncSeedRandom = Func{
	Name:             FUNC_SEED_RANDOM,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 1,
	ArgTypes:         [][]RslTypeEnum{{RslIntT}},
	NamedArgs:        NO_NAMED_ARGS,
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
		switch coerced := args[0].value.Val.(type) {
		case int64:
			RNG = rand.New(rand.NewSource(coerced))
			return EMPTY
		default:
			bugIncorrectTypes(FUNC_SEED_RANDOM)
			panic(UNREACHABLE)
		}
	},
}

var FuncRand = Func{
	Name:             FUNC_RAND,
	ReturnValues:     ONE_RETURN_VAL,
	RequiredArgCount: 0,
	ArgTypes:         NO_POS_ARGS,
	NamedArgs:        NO_NAMED_ARGS,
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
		return newRslValues(i, callNode, RNG.Float64())
	},
}

var FuncRandInt = Func{
	Name:             FUNC_RAND_INT,
	ReturnValues:     ONE_RETURN_VAL,
	RequiredArgCount: 1,
	ArgTypes:         [][]RslTypeEnum{{RslIntT}, {RslIntT}},
	NamedArgs:        NO_NAMED_ARGS,
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
		var min, max int64

		if len(args) == 1 {
			arg := args[0]
			min = 0
			max = arg.value.RequireInt(i, arg.node)
		} else {
			// two args
			minArg := args[0]
			maxArg := args[1]
			min = minArg.value.RequireInt(i, minArg.node)
			max = maxArg.value.RequireInt(i, maxArg.node)
		}

		if min > max {
			i.errorf(callNode, "%s() min (%d) must be less than or equal to max (%d).", FUNC_RAND_INT, min, max)
		}

		n := max - min
		return newRslValues(i, callNode, min+RNG.Int63n(n))
	},
}
