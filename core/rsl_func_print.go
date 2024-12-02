package core

import (
	"bytes"
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
	jsonStruct := RslToJsonType(arg)
	output := prettify(i, function, jsonStruct)
	RP.Print(output)
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

func prettify(i *MainInterpreter, function Token, jsonStruct interface{}) string {
	f := jsoncolor.NewFormatter()
	// todo could add coloring here on formatter

	buf := &bytes.Buffer{}

	enc := jsoncolor.NewEncoderWithFormatter(buf, f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	err := enc.Encode(jsonStruct)

	if err != nil {
		i.error(function, fmt.Sprintf("Error marshalling JSON: %v", err))
	}

	return buf.String()
}
