package core

import (
	"fmt"
	"github.com/nwidger/jsoncolor"
)

func runPrint(values []interface{}) {
	output := resolveOutputString(values)
	RP.Print(output)
}

func runDebug(values []interface{}) {
	output := resolveOutputString(values)
	RP.ScriptDebug(output)
}

func runPrettyPrint(i *MainInterpreter, function Token, values []interface{}) {
	if len(values) == 0 {
		RP.Print("\n")
	}

	arg := values[0]
	jsonStruct := jsonify(arg)
	output := prettify(i, function, jsonStruct)
	RP.Print(output + "\n")
}

func resolveOutputString(values []interface{}) string {
	if len(values) == 0 {
		return "\n"
	}

	output := ""
	for _, v := range values {
		output += ToPrintable(v) + " "
	}
	output = output[:len(output)-1] // remove last space
	output = output + "\n"
	return output
}

func jsonify(arg interface{}) interface{} {
	switch coerced := arg.(type) {
	case string:
		return coerced
	case int64:
		return coerced
	case float64:
		return coerced
	case bool:
		return coerced
	case []interface{}:
		var slice []interface{}
		for _, elem := range coerced {
			slice = append(slice, jsonify(elem))
		}
		return slice
	case RslMap:
		mapping := make(map[string]interface{})
		for _, key := range coerced.Keys() {
			value, _ := coerced.Get(key)
			mapping[key] = jsonify(value)
		}
		return mapping
	case nil:
		return nil
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for jsonify: %T", arg))
		panic(UNREACHABLE)
	}
}

func prettify(i *MainInterpreter, function Token, jsonStruct interface{}) string {
	f := jsoncolor.NewFormatter()
	// todo could add coloring here on formatter
	out, err := jsoncolor.MarshalIndentWithFormatter(jsonStruct, "", "  ", f)
	if err != nil {
		i.error(function, fmt.Sprintf("Error marshalling JSON: %v", err))
	}
	return string(out)
}
