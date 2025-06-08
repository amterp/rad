package core

import (
	"encoding/json"
	"fmt"
	com "rad/core/common"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/amterp/rad/rts/raderr"

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
	case RadString:
		return ToPrintableQuoteStr(coerced.String(), quoteStrings)
	case RadValue:
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
	case *RadList:
		return coerced.ToString()
	case RadList:
		return coerced.ToString()
	case *RadMap:
		return coerced.ToString()
	case RadMap:
		return coerced.ToString()
	case RadFn:
		return coerced.ToString()
	case RadNull:
		return "null"
	case RadError:
		return ToPrintableQuoteStr(coerced.Msg().String(), quoteStrings)
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for printable: %T", val))
		panic(UNREACHABLE)
	}
}

func TypeAsString(val interface{}) string {
	switch coerced := val.(type) {
	case RadValue:
		return TypeAsString(coerced.Val)
	case int64:
		return "int"
	case float64:
		return "float"
	case RadString, string:
		return "string"
	case bool:
		return "bool"
	case []interface{}, *[]interface{}, RadList, *RadList:
		return "list"
	case RadMap, *RadMap:
		return "map"
	case RadFn:
		return "function"
	case RadNull:
		return "null"
	case RadError:
		return "error"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for TypeAsString: %T\n%s\n", val, debug.Stack()))
		panic(UNREACHABLE)
	}
}

// Convert a json interface{} into native Rad types
func TryConvertJsonToNativeTypes(i *Interpreter, node *ts.Node, maybeJsonStr string) (RadValue, error) {
	var m interface{}
	decoder := json.NewDecoder(strings.NewReader(maybeJsonStr))
	decoder.UseNumber()
	err := decoder.Decode(&m)
	if err != nil {
		return newRadValue(i, node, maybeJsonStr), err
	}
	return ConvertToNativeTypes(i, node, m), nil
}

// it was originally implemented because we might capture JSON as a list of unhandled types, but
// now we should be able to capture json and convert it entirely to native Rad types up front
func ConvertToNativeTypes(i *Interpreter, node *ts.Node, val interface{}) RadValue {
	switch coerced := val.(type) {
	// strictly speaking, ints are unnecessary as Go unmarshalls them either as float64 or json.Number
	case RadString, string, int64, float64, bool:
		return newRadValue(i, node, coerced)
	case json.Number:
		s := string(coerced)
		if !strings.Contains(s, ".") {
			// try parsing as int64 (for better precision preservation)
			if iVal, err := coerced.Int64(); err == nil {
				return newRadValue(i, node, iVal)
			}
		}
		// fallback: treat as float64
		if fVal, err := coerced.Float64(); err == nil {
			return newRadValue(i, node, fVal)
		}
		i.errorf(node, fmt.Sprintf("Invalid number: %v", s))
		panic("UNREACHABLE")
	case []interface{}:
		list := NewRadList()
		for _, val := range coerced {
			list.Append(ConvertToNativeTypes(i, node, val))
		}
		return newRadValue(i, node, list)
	case map[string]interface{}:
		m := NewRadMap()
		sortedKeys := com.SortedKeys(coerced)
		for _, key := range sortedKeys {
			m.Set(newRadValue(i, node, key), ConvertToNativeTypes(i, node, coerced[key]))
		}
		return newRadValue(i, node, m)
	case nil:
		return newRadValue(i, node, nil)
	default:
		i.errorf(node, fmt.Sprintf("Unhandled type in array: %T", val))
		panic(UNREACHABLE)
	}
}

func ConvertValuesToNativeTypes(i *Interpreter, node *ts.Node, vals []interface{}) []RadValue {
	output := make([]RadValue, len(vals))
	for idx, val := range vals {
		output[idx] = ConvertToNativeTypes(i, node, val)
	}
	return output
}

// converts an Rad data structure to a JSON-schema-adhering structure.
func RadToJsonType(arg RadValue) interface{} {
	switch coerced := arg.Val.(type) {
	case RadString:
		return coerced.Plain()
	case int64, float64, bool:
		return coerced
	case *RadList:
		slice := make([]interface{}, 0)
		for _, elem := range coerced.Values {
			slice = append(slice, RadToJsonType(elem))
		}
		return slice
	case *RadMap:
		mapping := make(map[string]interface{})
		for _, key := range coerced.Keys() {
			value, _ := coerced.Get(key)
			mapping[ToPrintableQuoteStr(key, false)] = RadToJsonType(value)
		}
		return mapping
	case RadNull:
		return nil
	case RadError:
		return coerced.Msg().Plain()
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for RadToJsonType: %T\n%s\n", arg.Val, debug.Stack()))
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

func ErrorRadMap(err raderr.Error, errMsg string) *RadMap {
	m := NewRadMap()
	m.SetPrimitiveStr(constCode, string(err))
	m.SetPrimitiveStr(constMsg, errMsg)
	return m
}

func GetSrc(src string, node *ts.Node) string {
	return src[node.StartByte():node.EndByte()]
}
