package core

import (
	"fmt"
	"strconv"
)

func runParseInt(i *MainInterpreter, function Token, numExpectedReturnValues int, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, PARSE_INT+fmt.Sprintf("() takes 1 argument, got %d", len(args)))
	}

	switch coerced := args[0].(type) {
	case RslString:
		str := coerced.Plain()
		parsed, err := strconv.Atoi(str)

		if err != nil {
			errMsg := PARSE_INT + fmt.Sprintf("() could not parse %q as an integer", str)
			if numExpectedReturnValues == 1 {
				i.error(function, errMsg)
				panic(UNREACHABLE)
			} else {
				return []interface{}{int64(0), ErrorRslMap(PARSE_INT_FAILED, errMsg)}
			}
		} else {
			if numExpectedReturnValues == 1 {
				return int64(parsed)
			} else {
				return []interface{}{int64(parsed), NoErrorRslMap()}
			}
		}
	default:
		i.error(function, PARSE_INT+fmt.Sprintf("() takes a string, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}

func runParseFloat(i *MainInterpreter, function Token, numExpectedReturnValues int, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, PARSE_FLOAT+fmt.Sprintf("() takes 1 argument, got %d", len(args)))
	}

	switch coerced := args[0].(type) {
	case RslString:
		str := coerced.Plain()
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			errMsg := PARSE_FLOAT + fmt.Sprintf("() could not parse %q as an float", str)
			if numExpectedReturnValues == 1 {
				i.error(function, errMsg)
				panic(UNREACHABLE)
			} else {
				return []interface{}{0.0, ErrorRslMap(PARSE_FLOAT_FAILED, errMsg)}
			}
		} else {
			if numExpectedReturnValues == 1 {
				return parsed
			} else {
				return []interface{}{parsed, NoErrorRslMap()}
			}
		}
	default:
		i.error(function, PARSE_FLOAT+fmt.Sprintf("() takes a string, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}
