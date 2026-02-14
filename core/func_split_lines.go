package core

import (
	"regexp"

	"github.com/amterp/rad/rts/rl"
)

var splitLinesRe = regexp.MustCompile(`\r\n|\r|\n`)

var FuncSplitLines = BuiltInFunc{
	Name: FUNC_SPLIT_LINES,
	Execute: func(f FuncInvocation) RadValue {
		toSplit := f.GetStr("_val").Plain()
		return f.Return(splitLines(f.i, f.callNode, toSplit))
	},
}

func splitLines(i *Interpreter, callNode rl.Node, input string) []RadValue {
	parts := splitLinesRe.Split(input, -1)

	result := make([]RadValue, 0, len(parts))
	for _, part := range parts {
		result = append(result, newRadValue(i, callNode, part))
	}

	return result
}
