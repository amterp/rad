package core

import "strconv"

func ToPrintable(val interface{}) string {
	switch val.(type) {
	case int:
		return strconv.Itoa(val.(int))
	case float64:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64)
	case string:
		return val.(string)
	case bool:
		return strconv.FormatBool(val.(bool))
	case []int, []float64, []string:
		out := "["
		for i, v := range val.([]interface{}) {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(v)
		}
		return out + "]"
	default:
		panic("Unknown type")
	}
}
