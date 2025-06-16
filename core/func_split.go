package core

import (
	"regexp"
	"strings"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncSplit = BuiltInFunc{
	Name:            FUNC_SPLIT,
	ReturnValues:    ONE_RETURN_VAL,
	MinPosArgCount:  2,
	PosArgValidator: NewEnumerableArgSchema([][]rl.RadType{{rl.RadStrT}, {rl.RadStrT}}),
	NamedArgs:       NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) RadValue {
		strArg := f.args[0]
		splitterArg := f.args[1]

		str := strArg.value.RequireStr(f.i, strArg.node).Plain()
		splitter := splitterArg.value.RequireStr(f.i, splitterArg.node).Plain()

		return newRadValues(f.i, f.callNode, regexSplit(f.i, f.callNode, str, splitter))
	},
}

func regexSplit(i *Interpreter, callNode *ts.Node, input string, sep string) []RadValue {
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
