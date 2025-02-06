package core

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func ToPrintable(val interface{}) string {
	switch coerced := val.(type) {
	case int64:
		return strconv.FormatInt(coerced, 10)
	case float64:
		// todo results many cases of printing many places due to float imprecision. Display fewer places?
		return strconv.FormatFloat(coerced, 'f', -1, 64)
	case string:
		// todo based on contents, should escape quotes, or use other quotes. python does this.
		return `"` + coerced + `"`
	case RslString:
		return ToPrintable(coerced.String())
	case RslValue:
		return ToPrintable(coerced.Val)
	case bool:
		return strconv.FormatBool(coerced)
	case *[]interface{}:
		return ToPrintable(*coerced)
	case []interface{}:
		out := "["
		for i, elem := range coerced {
			if i > 0 {
				out += ", "
			}
			out += ToPrintable(elem)
		}
		return out + "]"
	case *RslList:
		return coerced.ToString()
	case RslList:
		return coerced.ToString()
	case *RslMap:
		return coerced.ToString()
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
	switch coerced := val.(type) {
	case RslValue:
		return TypeAsString(coerced.Val)
	case int64:
		return "int"
	case float64:
		return "float"
	case RslString, string:
		return "string"
	case bool:
		return "bool"
	case []interface{}, *[]interface{}, RslList, *RslList:
		return "list"
	case RslMap, *RslMap:
		return "map"
	case nil:
		return "null"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for as-string: %T", val))
		panic(UNREACHABLE)
	}
}

// Convert a json interface{} into native RSL types
func TryConvertJsonToNativeTypes(i *MainInterpreter, function Token, maybeJsonStr string) (interface{}, error) {
	var m interface{}
	err := json.Unmarshal([]byte(maybeJsonStr), &m)
	if err != nil {
		return NewRslString(maybeJsonStr), err
	}
	return ConvertToNativeTypes(i, function, m), nil
}

// it was originally implemented because we might capture JSON as a list of unhandled types, but
// now we should be able to capture json and convert it entirely to native RSL types up front
func ConvertToNativeTypes(interp *MainInterpreter, token Token, val interface{}) interface{} {
	switch coerced := val.(type) {
	// strictly speaking, I don't think ints are necessary to handle, since it seems Go unmarshalls
	// json 'ints' into floats
	case string:
		return NewRslString(coerced)
	case RslString, int64, float64, bool:
		return coerced
	case int:
		return int64(coerced)
	case []interface{}:
		output := make([]interface{}, len(coerced))
		for i, val := range coerced {
			output[i] = ConvertToNativeTypes(interp, token, val)
		}
		return output
	case map[string]interface{}:
		m := NewOldRslMap()
		sortedKeys := SortedKeys(coerced)
		for _, key := range sortedKeys {
			m.SetStr(key, ConvertToNativeTypes(interp, token, coerced[key]))
		}
		return *m
	case RslMapOld:
		return coerced
	case nil:
		return nil
	default:
		interp.error(token, fmt.Sprintf("Unhandled type in array: %T", val))
		panic(UNREACHABLE)
	}
}

func ConvertValuesToNativeTypes(interp *MainInterpreter, token Token, vals []interface{}) []interface{} {
	output := make([]interface{}, len(vals))
	for i, val := range vals {
		output[i] = ConvertToNativeTypes(interp, token, val)
	}
	return output
}

// converts an RSL data structure to a JSON-serializable structure
func RslToJsonType(arg interface{}) interface{} {
	switch coerced := arg.(type) {
	case RslString:
		return coerced.Plain()
	case int64, float64, bool:
		return coerced
	case []interface{}:
		var slice []interface{}
		for _, elem := range coerced {
			slice = append(slice, RslToJsonType(elem))
		}
		return slice
	case RslMapOld:
		mapping := make(map[string]interface{})
		for _, key := range coerced.Keys() {
			value, _ := coerced.GetStr(key)
			mapping[key] = RslToJsonType(value)
		}
		return mapping
	case nil:
		return nil
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for RslToJsonType: %T", arg))
		panic(UNREACHABLE)
	}
}

func JsonToString(jsonVal interface{}) string {
	jsonBytes, err := json.Marshal(jsonVal)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		RP.RadErrorExit(fmt.Sprintf("Bug! Non-marshallable json object passed to JsonToString (%T): %v", jsonVal, jsonVal))
	}

	return string(jsonBytes)
}

func AbsInt(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func AbsFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TruthyFalsy(val interface{}) bool {
	switch coerced := val.(type) {
	case int64:
		return coerced != 0
	case float64:
		return coerced != 0
	case RslString:
		return coerced.Plain() != ""
	case bool:
		return coerced
	case []interface{}:
		return len(coerced) != 0
	case RslMapOld:
		return len(coerced.Keys()) != 0
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for TruthyFalsy: %T", val))
		panic(UNREACHABLE)
	}
}

func ErrorRslMap(err RslError, errMsg string) RslMapOld {
	m := NewOldRslMap()
	m.SetStr("code", NewRslString(string(err)))
	m.SetStr("msg", NewRslString(errMsg))
	return *m
}

func NoErrorRslMap() RslMapOld {
	m := NewOldRslMap()
	return *m
}

func (e StringLiteral) FullString() string {
	return e.Value[len(e.Value)-1].FullStringLiteral
}
