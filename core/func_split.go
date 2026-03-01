package core

import (
	"regexp"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

var FuncSplit = BuiltInFunc{
	Name: FUNC_SPLIT,
	Execute: func(f FuncInvocation) RadValue {
		toSplit := f.GetStr("_val").Plain()
		splitter := f.GetStr("_sep").Plain()

		limitArg := f.GetArg("limit")
		limit := -1
		if !limitArg.IsNull() {
			limitVal := limitArg.RequireInt(f.i, f.callNode)
			if limitVal < 1 {
				return f.Return(NewErrorStrf("limit must be at least 1, got %d", limitVal))
			}
			// limit counts splits, but Go's SplitN counts parts, so +1
			limit = int(limitVal) + 1
		}

		return f.Return(regexSplit(f.i, f.callNode, toSplit, splitter, limit))
	},
}

func regexSplit(i *Interpreter, callNode rl.Node, input string, sep string, limit int) []RadValue {
	re, err := regexp.Compile(sep)

	var parts []string
	if err == nil {
		parts = re.Split(input, limit)
	} else {
		if limit < 0 {
			parts = strings.Split(input, sep)
		} else {
			parts = strings.SplitN(input, sep, limit)
		}
	}

	result := make([]RadValue, 0, len(parts))
	for _, part := range parts {
		result = append(result, newRadValue(i, callNode, part))
	}

	return result
}
