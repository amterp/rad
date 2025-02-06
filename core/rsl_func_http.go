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
//   - generic http for other/all methods?
func runHttpGet(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) RslMapOld {
	if len(args) != 1 {
		i.error(function, HTTP_GET+fmt.Sprintf("() takes exactly 1 positional arg, got %d", len(args)))
	}

	validateExpectedNamedArgs(i, function, []string{HEADERS_NAMED_ARG}, namedArgs)
	parsedArgs := parseHttpReqArgs(i, function, namedArgs)

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

func runHttpPost(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) RslMapOld {
	return runHttpPutOrPost(i, function, args, namedArgs, HTTP_POST, "POST")
}

func runHttpPut(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) RslMapOld {
	return runHttpPutOrPost(i, function, args, namedArgs, HTTP_PUT, "PUT")
}

func runHttpPutOrPost(i *MainInterpreter,
	function Token,
	args []interface{},
	namedArgs map[string]interface{},
	funcName string,
	method string,
) RslMapOld {
	if len(args) < 1 || len(args) > 2 {
		i.error(function, funcName+fmt.Sprintf("() takes 1 or 2 positional arguments, got %d", len(args)))
	}

	validateExpectedNamedArgs(i, function, []string{HEADERS_NAMED_ARG}, namedArgs)
	parsedArgs := parseHttpReqArgs(i, function, namedArgs)

	url, ok := args[0].(RslString)
	if !ok {
		i.error(function, funcName+fmt.Sprintf("() takes a string as the first argument, got %s", TypeAsString(args[0])))
	}

	body := ""
	if len(args) == 2 {
		jsonObj := RslToJsonType(args[1])
		body = JsonToString(jsonObj)
	}

	resp, err := RReq.PutOrPost(method, url.Plain(), body, parsedArgs.Headers)
	if err != nil {
		i.error(function, fmt.Sprintf("Error making request: %v", err))
	}
	return resp.ToRslMap(i, function)
}

func parseHttpReqArgs(i *MainInterpreter, function Token, args map[string]interface{}) HttpGetNamedArgs {
	parsedArgs := HttpGetNamedArgs{
		Headers: make(map[string]string),
	}
	if headerMap, ok := args[HEADERS_NAMED_ARG]; ok {
		if rslMap, ok := headerMap.(RslMapOld); ok {
			parsedArgs.Headers = rslMap.ToStringMap()
		} else {
			i.error(function, function.GetLexeme()+fmt.Sprintf("() %s must be a map, got %s", HEADERS_NAMED_ARG, TypeAsString(headerMap)))
		}
	}
	return parsedArgs
}

type HttpGetNamedArgs struct {
	Headers map[string]string
}
