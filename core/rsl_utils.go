package core

import "strconv"

func ToPrintable(val interface{}) string {
	switch v := val.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case []int:
		out := "["
		for i, elem := range v {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(elem)
		}
		return out + "]"
	case []float64:
		out := "["
		for i, elem := range v {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(elem)
		}
		return out + "]"
	case []string:
		out := "["
		for i, elem := range v {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(elem)
		}
		return out + "]"
	default:
		panic("Unknown type")
	}
}
