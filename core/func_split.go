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
		useRegex := f.GetBool("regex")

		limitArg := f.GetArg("limit")
		limit := -1
		if !limitArg.IsNull() {
			limitVal := limitArg.RequireInt(f.i, f.callNode)
			if limitVal < 1 {
				return f.Return(NewErrorStrf("limit must be at least 1, got %d", limitVal).SetCode(rl.ErrNumInvalidRange))
			}
			// limit counts splits, but Go's SplitN counts parts, so +1
			limit = int(limitVal) + 1
		}

		var parts []string
		if useRegex {
			re, err := regexp.Compile(splitter)
			if err != nil {
				return f.ReturnErrf(rl.ErrInvalidRegex, "Error compiling regex pattern: %s", err)
			}
			parts = re.Split(toSplit, limit)
		} else if limit < 0 {
			parts = strings.Split(toSplit, splitter)
		} else {
			parts = strings.SplitN(toSplit, splitter, limit)
		}

		result := make([]RadValue, 0, len(parts))
		for _, part := range parts {
			result = append(result, newRadValue(f.i, f.callNode, part))
		}

		return f.Return(result)
	},
}
