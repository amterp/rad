package core

func AsStringArray(v []interface{}) ([]string, bool) {
	output := make([]string, len(v))
	for i, val := range v {
		coerced, ok := val.(RslString)
		if !ok {
			return nil, false
		}
		output[i] = coerced.Plain()
	}
	return output, true
}

func AsIntArray(v []interface{}) ([]int64, bool) {
	output := make([]int64, len(v))
	for i, val := range v {
		coerced, ok := val.(int64)
		if !ok {
			return nil, false
		}
		output[i] = coerced
	}
	return output, true
}

func AsFloatArray(v []interface{}) ([]float64, bool) {
	output := make([]float64, len(v))
	for i, val := range v {
		coerced, ok := val.(float64)
		if !ok {
			if coerced, ok := val.(int64); ok {
				output[i] = float64(coerced)
				continue
			}
			return nil, false
		}
		output[i] = coerced
	}
	return output, true
}

func AsBoolArray(v []interface{}) ([]bool, bool) {
	output := make([]bool, len(v))
	for i, val := range v {
		coerced, ok := val.(bool)
		if !ok {
			return nil, false
		}
		output[i] = coerced
	}
	return output, true
}

func AsMixedArray[T any](v []T) ([]interface{}, bool) {
	output := make([]interface{}, len(v))
	for i, val := range v {
		output[i] = val
	}
	return output, true
}

func ToStringArray[T any](v []T) []string {
	return ToStringArrayQuoteStr(v, true)
}

func ToStringArrayQuoteStr[T any](v []T, quoteStrings bool) []string {
	output := make([]string, len(v))
	for i, val := range v {
		output[i] = ToPrintableQuoteStr(val, quoteStrings)
	}
	return output
}
