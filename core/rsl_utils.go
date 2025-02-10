package core

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"

	ts "github.com/tree-sitter/go-tree-sitter"
)

func ToPrintable(val interface{}) string {
	return ToPrintableQuoteStr(val, true)
}

func ToPrintableQuoteStr(val interface{}, quoteStrings bool) string {
	switch coerced := val.(type) {
	case int64:
		return strconv.FormatInt(coerced, 10)
	case float64:
		// todo results many cases of printing many places due to float imprecision. Display fewer places?
		return strconv.FormatFloat(coerced, 'f', -1, 64)
	case string:
		// todo based on contents, should escape quotes, or use other quotes. python does this.
		if quoteStrings {
			return `"` + coerced + `"`
		} else {
			return coerced
		}
	case RslString:
		return ToPrintableQuoteStr(coerced.String(), quoteStrings)
	case RslValue:
		return ToPrintableQuoteStr(coerced.Val, quoteStrings)
	case bool:
		return strconv.FormatBool(coerced)
	case *[]interface{}:
		return ToPrintableQuoteStr(*coerced, quoteStrings)
	case []interface{}:
		out := "["
		for i, elem := range coerced {
			if i > 0 {
				out += ", "
			}
			out += ToPrintableQuoteStr(elem, quoteStrings)
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
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for TypeAsString: %T\n%s\n", val, debug.Stack()))
		panic(UNREACHABLE)
	}
}

// Convert a json interface{} into native RSL types
func TryConvertJsonToNativeTypes(i *Interpreter, node *ts.Node, maybeJsonStr string) (RslValue, error) {
	var m interface{}
	err := json.Unmarshal([]byte(maybeJsonStr), &m)
	if err != nil {
		return newRslValue(i, node, maybeJsonStr), err
	}
	return ConvertToNativeTypes(i, node, m), nil
}

// it was originally implemented because we might capture JSON as a list of unhandled types, but
// now we should be able to capture json and convert it entirely to native RSL types up front
func ConvertToNativeTypes(i *Interpreter, node *ts.Node, val interface{}) RslValue {
	switch coerced := val.(type) {
	// strictly speaking, I don't think ints are necessary to handle, since it seems Go unmarshalls
	// json 'ints' into floats
	case RslString, string, int64, float64, bool:
		return newRslValue(i, node, coerced)
	case []interface{}:
		list := NewRslList()
		for _, val := range coerced {
			list.Append(ConvertToNativeTypes(i, node, val))
		}
		return newRslValue(i, node, list)
	case map[string]interface{}:
		m := NewRslMap()
		sortedKeys := SortedKeys(coerced)
		for _, key := range sortedKeys {
			m.Set(newRslValue(i, node, key), ConvertToNativeTypes(i, node, coerced[key]))
		}
		return newRslValue(i, node, m)
	case nil:
		return newRslValue(i, node, "nil")
	default:
		i.errorf(node, fmt.Sprintf("Unhandled type in array: %T", val))
		panic(UNREACHABLE)
	}
}

func ConvertValuesToNativeTypes(i *Interpreter, node *ts.Node, vals []interface{}) []RslValue {
	output := make([]RslValue, len(vals))
	for idx, val := range vals {
		output[idx] = ConvertToNativeTypes(i, node, val)
	}
	return output
}

// converts an RSL data structure to a JSON-serializable structure
func RslToJsonType(arg RslValue) interface{} {
	switch coerced := arg.Val.(type) {
	case RslString:
		return coerced.Plain()
	case int64, float64, bool:
		return coerced
	case *RslList:
		var slice []interface{}
		for _, elem := range coerced.Values {
			slice = append(slice, RslToJsonType(elem))
		}
		return slice
	case *RslMap:
		mapping := make(map[string]interface{})
		for _, key := range coerced.Keys() {
			value, _ := coerced.Get(key)
			mapping[ToPrintableQuoteStr(key, false)] = RslToJsonType(value)
		}
		return mapping
	case nil:
		return nil
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for RslToJsonType: %T\n%s\n", arg.Val, debug.Stack()))
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

func ErrorRslMap(err RslError, errMsg string) *RslMap {
	m := NewRslMap()
	m.SetPrimitiveStr("code", string(err))
	m.SetPrimitiveStr("msg", errMsg)
	return m
}

func NoErrorRslMap() *RslMap {
	return NewRslMap()
}
