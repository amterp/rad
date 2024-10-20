package core

import (
	"fmt"
	"strconv"
)

func ToPrintable(val interface{}) string {
	switch coerced := val.(type) {
	case int64:
		return strconv.FormatInt(coerced, 10)
	case float64:
		return strconv.FormatFloat(coerced, 'f', -1, 64)
	case string:
		return coerced
	case bool:
		return strconv.FormatBool(coerced)
	case []interface{}:
		out := "["
		for i, elem := range coerced {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(elem)
		}
		return out + "]"
	case RslMap:
		return coerced.ToString()
	case nil:
		return "null"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for printable: %T", val))
		panic(UNREACHABLE)
	}
}

func TypeAsString(val interface{}) string {
	switch val.(type) {
	case int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case bool:
		return "bool"
	case []interface{}:
		return "array"
	case RslMap:
		return "map"
	case nil:
		return "null"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for as-string: %T", val))
		panic(UNREACHABLE)
	}
}
