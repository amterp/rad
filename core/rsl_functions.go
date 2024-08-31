package core

import "fmt"

func RunRslFunction(i *MainInterpreter, function Token, values []interface{}) {
	functionName := function.GetLexeme()
	switch functionName {
	case "print": // todo would be nice to make this a reference to a var that GoLand can find
		runPrint(i, values)
	}
}

func runPrint(i *MainInterpreter, values []interface{}) {
	output := ""
	for _, v := range values {
		output += fmt.Sprintf("%v", v) // todo is %v right?
	}
	output = output[:len(output)-1] // remove last space
	fmt.Println(output)
}
