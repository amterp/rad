package core

import "strings"

func RunJoin(i *MainInterpreter, function Token, values []interface{}) interface{} {
	if len(values) < 2 {
		i.error(function, "join() takes at least two arguments")
	}

	prefix := ""
	suffix := ""
	if len(values) == 3 {
		prefix = ToPrintable(values[2])
	} else if len(values) == 4 {
		prefix = ToPrintable(values[2])
		suffix = ToPrintable(values[3])
	}

	var arr []string
	switch values[0].(type) {
	case []interface{}:
		elements := values[0].([]interface{})
		for _, v := range elements {
			arr = append(arr, ToPrintable(v))
		}
	default:
		i.error(function, "join() takes an array as the first argument")
	}

	separator := ToPrintable(values[1])

	return prefix + strings.Join(arr, separator) + suffix
}
