package core

import "fmt"

// todo
//  toStringArray(array), toString(non-string)

// RunRslNonVoidFunction returns pointers to values e.g. *string
func RunRslNonVoidFunction(i *MainInterpreter, function Token, values []interface{}) interface{} {
	functionName := function.GetLexeme()
	switch functionName {
	// todo add functions here
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
