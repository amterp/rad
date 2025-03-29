package core

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

var NIL_SENTINAL = RslValue{Val: 0x0}
var JSON_SENTINAL = RslValue{Val: 0x1}

type RslValue struct {
	// int64, float64, RslString, bool stored as values
	// collections (lists, maps) stored as pointers
	// lists are *RslList
	// maps are *RslMap
	Val interface{}
}

func (v RslValue) Type() RslTypeEnum {
	switch v.Val.(type) {
	case int64:
		return RslIntT
	case float64:
		return RslFloatT
	case RslString:
		return RslStringT
	case bool:
		return RslBoolT
	case *RslList:
		return RslListT
	case *RslMap:
		return RslMapT
	default:
		panic(fmt.Sprintf("Bug! Unhandled RSL type: %T", v.Val))
	}
}

func (v RslValue) Index(i *Interpreter, idxNode *ts.Node) RslValue {
	switch coerced := v.Val.(type) {
	case RslString:
		return newRslValue(i, idxNode, coerced.Index(i, idxNode))
	case *RslList:
		return newRslValue(i, idxNode, coerced.GetIdx(i, idxNode))
	case *RslMap:
		return newRslValue(i, idxNode, coerced.GetNode(i, idxNode))
	default:
		i.errorf(idxNode, "Indexing not supported for %s", TypeAsString(v))
		panic(UNREACHABLE)
	}
}

func (v RslValue) RequireInt(i *Interpreter, node *ts.Node) int64 {
	switch coerced := v.Val.(type) {
	case int64:
		return coerced
	default:
		i.errorf(node, "Expected int, got %q: %s", TypeAsString(v), ToPrintable(v))
		panic(UNREACHABLE)
	}
}

func (v RslValue) RequireStr(i *Interpreter, node *ts.Node) RslString {
	if str, ok := v.TryGetStr(); ok {
		return str
	}
	i.errorf(node, "Expected string, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RslValue) TryGetStr() (RslString, bool) {
	if str, ok := v.Val.(RslString); ok {
		return str, true
	}
	return NewRslString(""), false
}

func (v RslValue) RequireList(i *Interpreter, node *ts.Node) *RslList {
	if list, ok := v.TryGetList(); ok {
		return list
	}
	i.errorf(node, "Expected list, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RslValue) TryGetList() (*RslList, bool) {
	if list, ok := v.Val.(*RslList); ok {
		return list, true
	}
	return nil, false
}

func (v RslValue) RequireBool(i *Interpreter, node *ts.Node) bool {
	if b, ok := v.TryGetBool(); ok {
		return b
	}
	i.errorf(node, "Expected bool, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RslValue) TryGetBool() (bool, bool) {
	if b, ok := v.Val.(bool); ok {
		return b, true
	}
	return false, false
}

func (v RslValue) RequireMap(i *Interpreter, node *ts.Node) *RslMap {
	if b, ok := v.TryGetMap(); ok {
		return b
	}
	i.errorf(node, "Expected map, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RslValue) TryGetMap() (*RslMap, bool) {
	if m, ok := v.Val.(*RslMap); ok {
		return m, true
	}
	return nil, false
}

func (v RslValue) TryGetFloatAllowingInt() (float64, bool) {
	switch coerced := v.Val.(type) {
	case int64:
		return float64(coerced), true
	case float64:
		return coerced, true
	default:
		return 0, false
	}
}

func (v RslValue) RequireFloatAllowingInt(i *Interpreter, node *ts.Node) float64 {
	if f, ok := v.TryGetFloatAllowingInt(); ok {
		return f
	}
	i.errorf(node, "Expected float, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RslValue) ModifyIdx(i *Interpreter, idxNode *ts.Node, rightValue RslValue) {
	switch coerced := v.Val.(type) {
	case *RslList:
		coerced.ModifyIdx(i, idxNode, rightValue)
	case *RslMap:
		if idxNode.Kind() == K_IDENTIFIER {
			// dot syntax e.g. myMap.myKey
			keyName := i.sd.Src[idxNode.StartByte():idxNode.EndByte()]
			coerced.Set(newRslValueStr(keyName), rightValue)
		} else {
			// 'traditional' syntax e.g. myMap["myKey"]
			idxVal := evalMapKey(i, idxNode)
			coerced.Set(idxVal, rightValue)
		}
	default:
		i.errorf(idxNode, "Cannot modify indices for type '%s'", TypeAsString(v))
		panic(UNREACHABLE)
	}
}

func (v RslValue) Hash() string {
	switch val := v.Val.(type) {
	case RslString:
		return val.Plain() // attributes don't impact hash
	case int64, float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		panic(fmt.Sprintf("Cannot key on a %s", TypeAsString(v)))
	}
}

func (left RslValue) Equals(right RslValue) bool {
	leftT := left.Type()
	rightT := right.Type()

	if leftT != rightT {
		// todo should do bespoke float/int comparison
		return false
	}

	switch coercedLeft := left.Val.(type) {
	case RslString:
		coercedRight := right.Val.(RslString)
		return coercedLeft.Plain() == coercedRight.Plain()
	case int64, float64, bool:
		return left.Val == right.Val
	case *RslList:
		coercedRight := right.Val.(*RslList)
		return coercedLeft.Equals(coercedRight)
	case *RslMap:
		coercedRight := right.Val.(*RslMap)
		return coercedLeft.Equals(coercedRight)
	default:
		return false
	}
}

func (v RslValue) RequireType(i *Interpreter, node *ts.Node, errPrefix string, allowedTypes ...RslTypeEnum) RslValue {
	for _, allowedType := range allowedTypes {
		if v.Type() == allowedType {
			return v
		}
	}

	i.errorf(node, "%s: %s", errPrefix, TypeAsString(v))
	panic(UNREACHABLE)
}

func (v RslValue) RequireNotType(i *Interpreter, node *ts.Node, errPrefix string, disallowedTypes ...RslTypeEnum) RslValue {
	for _, disallowedType := range disallowedTypes {
		if v.Type() == disallowedType {
			i.errorf(node, "%s: %s", errPrefix, TypeAsString(v))
			panic(UNREACHABLE)
		}
	}

	return v
}

func (v RslValue) TruthyFalsy() bool {
	switch coerced := v.Val.(type) {
	case int64:
		return coerced != 0
	case float64:
		return coerced != 0
	case RslString:
		return coerced.Plain() != ""
	case bool:
		return coerced
	case *RslList:
		return coerced.Len() != 0
	case *RslMap:
		return coerced.Len() != 0
	default:
		panic(fmt.Sprintf("Bug! Unhandled type for TruthyFalsy: %T", v.Val))
	}
}

func newRslValue(i *Interpreter, node *ts.Node, value interface{}) RslValue {
	switch coerced := value.(type) {
	case RslValue:
		return coerced
	case []RslValue: // todo risky to have this? might cover up issues
		list := NewRslList()
		list.Values = coerced
		return newRslValue(i, node, list)
	case RslString:
		return RslValue{Val: coerced}
	case string:
		return RslValue{Val: NewRslString(coerced)}
	case int:
		return RslValue{Val: int64(coerced)}
	case int64, float64, bool:
		return RslValue{Val: coerced}
	case *RslList:
		return RslValue{Val: coerced}
	case RslList:
		return RslValue{Val: &coerced}
	case *RslMap:
		return RslValue{Val: coerced}
	case RslMap:
		return RslValue{Val: &coerced}
	case map[string]interface{}:
		rslMap := NewRslMap()
		for key, val := range coerced {
			rslMap.Set(newRslValue(i, node, key), newRslValue(i, node, val))
		}
		return RslValue{Val: rslMap}
	case []interface{}:
		list := NewRslListFromGeneric(i, node, coerced)
		return RslValue{Val: list}
	case []string:
		list := NewRslListFromGeneric(i, node, coerced)
		return RslValue{Val: list}
	case nil:
		return RslValue{Val: NewRslString("null")} // todo not good, can't differentiate between string null and actual null
	default:
		if i != nil && node != nil {
			i.errorf(node, "Unsupported value type: %s", TypeAsString(coerced))
			panic(UNREACHABLE)
		} else {
			panic(fmt.Sprintf("Bug! Unsafe call w/ unsupported value type: %T", coerced))
		}
	}
}

func newRslValues(i *Interpreter, node *ts.Node, value ...interface{}) []RslValue {
	values := make([]RslValue, len(value))
	for idx, v := range value {
		values[idx] = newRslValue(i, node, v)
	}
	return values
}

func newRslValueStr(str string) RslValue {
	return newRslValue(nil, nil, str)
}

func newRslValueRslStr(str RslString) RslValue {
	return newRslValue(nil, nil, str)
}

func newRslValueInt(val int) RslValue {
	return newRslValue(nil, nil, val)
}

func newRslValueInt64(val int64) RslValue {
	return newRslValue(nil, nil, val)
}

func newRslValueFloat64(val float64) RslValue {
	return newRslValue(nil, nil, val)
}

func newRslValueBool(val bool) RslValue {
	return newRslValue(nil, nil, val)
}

func newRslValueMap(val *RslMap) RslValue {
	return newRslValue(nil, nil, val)
}

func newRslValueList(val *RslList) RslValue {
	return newRslValue(nil, nil, val)
}
