package core

import "github.com/amterp/rad/rts/rl"

var FuncParseDuration = BuiltInFunc{
	Name: FUNC_PARSE_DURATION,
	Execute: func(f FuncInvocation) RadValue {
		durStr := f.GetStr("_duration").Plain()

		nanos, err := ParseDurationString(durStr)
		if err != nil {
			return f.ReturnErrf(rl.ErrParseDuration, "Failed to parse duration %q: %s", durStr, err)
		}

		return f.Return(NewDurationMap(nanos))
	},
}
