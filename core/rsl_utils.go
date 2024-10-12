package core

import (
	"fmt"
	"strconv"
)

func ToPrintable(val interface{}) string {
	switch v := val.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		out := "["
		for i, elem := range v {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(elem)
		}
		return out + "]"
	default:
		RP.RadErrorExit(fmt.Sprintf("unknown type: %T", val))
		panic(UNREACHABLE)
	}
}
