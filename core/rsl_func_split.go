package core

import (
	"fmt"
	"regexp"
	"strings"
)

func runSplit(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 2 {
		i.error(function, SPLIT+fmt.Sprintf("() takes 2 arguments, got %d", len(args)))
	}

	switch str := args[0].(type) {
	case RslString:
		switch sep := args[1].(type) {
		case RslString:
			return regexSplit(str.Plain(), sep.Plain())
		default:
			i.error(function, SPLIT+fmt.Sprintf("() takes strings as args, got %s", TypeAsString(args[1])))
			panic(UNREACHABLE)
		}
	default:
		i.error(function, SPLIT+fmt.Sprintf("() takes strings as args, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}

func regexSplit(input string, sep string) []interface{} {
	re, err := regexp.Compile(sep)

	var parts []string
	if err == nil {
		parts = re.Split(input, -1)
	} else {
		parts = strings.Split(input, sep)
	}

	result := make([]interface{}, 0, len(parts))
	for _, part := range parts {
		result = append(result, NewRslString(part))
	}

	return result
}
