package core

import (
	"math/rand"
	"time"
)

var RNG *rand.Rand

func init() {
	RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func runRand(i *MainInterpreter, randToken Token, args []interface{}) float64 {
	if len(args) > 0 {
		i.error(randToken, RAND+"() does not take arguments.")
	}
	return RNG.Float64()
}

func runRandInt(i *MainInterpreter, randToken Token, args []interface{}) int64 {
	if len(args) == 0 || len(args) > 2 {
		i.error(randToken, RAND_INT+"() takes 1 or 2 arguments.")
	}

	var min, max int64
	switch coerced := args[0].(type) {
	case int64:
		max = coerced
	default:
		i.error(randToken, RAND_INT+"() takes an int, got "+TypeAsString(args[0]))
	}

	if len(args) == 2 {
		switch coerced := args[1].(type) {
		case int64:
			min = max
			max = coerced
		default:
			i.error(randToken, RAND_INT+"() takes an int, got "+TypeAsString(args[1]))
		}
	} else {
		min = 0
	}

	if min > max {
		i.error(randToken, RAND_INT+"() min must be less than or equal to max.")
	}

	n := max - min
	return min + RNG.Int63n(n)
}

func runSeedRandom(i *MainInterpreter, randToken Token, args []interface{}) {
	if len(args) != 1 {
		i.error(randToken, SEED_RANDOM+"() takes exactly 1 argument.")
	}
	switch coerced := args[0].(type) {
	case int64:
		RNG = rand.New(rand.NewSource(coerced))
	default:
		i.error(randToken, SEED_RANDOM+"() takes an int, got "+TypeAsString(args[0]))
	}
}
