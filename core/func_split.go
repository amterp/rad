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

		return f.Return(regexSplit(f.i, f.callNode, toSplit, splitter))
	},
}

func regexSplit(i *Interpreter, callNode rl.Node, input string, sep string) []RadValue {
	re, err := regexp.Compile(sep)

	var parts []string
	if err == nil {
		parts = re.Split(input, -1)
	} else {
		parts = strings.Split(input, sep)
	}

	result := make([]RadValue, 0, len(parts))
	for _, part := range parts {
		result = append(result, newRadValue(i, callNode, part))
	}

	return result
}
