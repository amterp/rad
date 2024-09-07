package core

import "fmt"

// todo
//  toStringArray(array), toString(non-string)

// RunRslNonVoidFunction returns pointers to values e.g. *string
func RunRslNonVoidFunction(i *MainInterpreter, function Token, values []interface{}) interface{} {
	functionName := function.GetLexeme()
	switch functionName {
	case "len":
		return runLen(i, function, values)
	default:
		i.error(function, fmt.Sprintf("Unknown function: %v", functionName))
		panic(UNREACHABLE)
	}
}

func RunRslFunction(i *MainInterpreter, function Token, values []interface{}) {
	functionName := function.GetLexeme()
	switch functionName {
	case "print": // todo would be nice to make this a reference to a var that GoLand can find
		runPrint(i, values)
	default:
		RunRslNonVoidFunction(i, function, values)
	}
}

func runPrint(i *MainInterpreter, values []interface{}) {
	if len(values) == 0 {
		fmt.Println()
		return
	}

	output := ""
	for _, v := range values {
		output += ToPrintable(v) + " "
	}
	output = output[:len(output)-1] // remove last space
	fmt.Println(output)
}

func runLen(i *MainInterpreter, function Token, values []interface{}) interface{} {
	if len(values) != 1 {
		i.error(function, "len() takes exactly one argument")
	}
	switch v := values[0].(type) {
	case string:
		return len(v)
	case []string:
		return len(v)
	case []int:
		return len(v)
	case []float64:
		return len(v)
	default:
		i.error(function, "len() takes a string or array")
		panic(UNREACHABLE)
	}
}
