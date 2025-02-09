package core

func runPrint(values []interface{}) {
	output := resolveOutputString(values)
	RP.Print(output)
}

func runDebug(values []interface{}) {
	output := resolveOutputString(values)
	RP.ScriptDebug(output)
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
