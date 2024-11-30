package core

import (
	"fmt"
)

const (
	HEADERS_NAMED_ARG = "headers"
)

// todo handle exceptions?
//   - auth?
//   - query params help?
func runHttpGet(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) RslMap {
	if len(args) != 1 {
		i.error(function, SORT_FUNC+fmt.Sprintf("() takes exactly 1 positional arg, got %d", len(args)))
	}

	validateExpectedNamedArgs(i, function, []string{HEADERS_NAMED_ARG}, namedArgs)
	parsedArgs := parseHttpGetArgs(i, function, namedArgs)

	switch coerced := args[0].(type) {
	case RslString:
		resp, err := RReq.Get(coerced.Plain(), parsedArgs.Headers)
		if err != nil {
			i.error(function, fmt.Sprintf("Error making request: %v", err))
		}
		return resp.ToRslMap(i, function)
	default:
		i.error(function, HTTP_GET+fmt.Sprintf("() takes a string, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}

func parseHttpGetArgs(i *MainInterpreter, function Token, args map[string]interface{}) HttpGetNamedArgs {
	parsedArgs := HttpGetNamedArgs{
		Headers: make(map[string]string),
	}
	if headerMap, ok := args[HEADERS_NAMED_ARG]; ok {
		if rslMap, ok := headerMap.(RslMap); ok {
			parsedArgs.Headers = rslMap.ToStringMap()
		} else {
			i.error(function, HTTP_GET+fmt.Sprintf("() %s must be a map, got %s", HEADERS_NAMED_ARG, TypeAsString(headerMap)))
		}
	}
	return parsedArgs
}

type HttpGetNamedArgs struct {
	Headers map[string]string
}
