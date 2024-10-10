package core

import (
	"encoding/json"
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

// todo brute-forced this implementation, it's not good, can almost definitely be improved
func runPrettyPrint(i *MainInterpreter, function Token, values []interface{}) {
	if len(values) == 0 {
		RP.Print("\n")
	}

	output := ""
	arg := values[0]
	switch coerced := arg.(type) {
	case string:
		output = prettify(i, function, coerced)
	case []interface{}:
		var items []interface{}
		for _, item := range coerced {
			switch coercedItem := item.(type) {
			case string:
				var jsonData interface{}
				if err := json.Unmarshal([]byte(coercedItem), &jsonData); err != nil {
					i.error(function, fmt.Sprintf("Error unmarshalling JSON: %v", err))
				}
				items = append(items, jsonData)
			default:
				items = append(items, coercedItem)
			}
		}
		if marshalled, err := json.Marshal(items); err != nil {
			i.error(function, fmt.Sprintf("Error marshalling JSON: %v", err))
		} else {
			output = prettify(i, function, string(marshalled))
		}
	default:
		unformatted, err := jsoncolor.Marshal(arg)
		if err != nil {
			RP.RadDebug(fmt.Sprintf("Failed to marshall %v", arg))
			i.error(function, fmt.Sprintf("Error marshalling JSON: %v", err))
		}
		output = prettify(i, function, string(unformatted))
	}
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

func prettify(i *MainInterpreter, function Token, unformatted string) string {
	bytes := []byte(unformatted)
	if json.Valid(bytes) {
		var unmarshalled interface{}
		if err := json.Unmarshal(bytes, &unmarshalled); err != nil {
			i.error(function, fmt.Sprintf("Error unmarshalling JSON: %v", err))
		}
		f := jsoncolor.NewFormatter()
		// todo could add coloring here on formatter
		out, err := jsoncolor.MarshalIndentWithFormatter(unmarshalled, "", "  ", f)
		if err != nil {
			i.error(function, fmt.Sprintf("Error marshalling JSON: %v", err))
		}
		return string(out)
	} else {
		return unformatted
	}
}
