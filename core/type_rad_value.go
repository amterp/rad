package core

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

// used to internally delete things e.g. vars from env, but also empty returns. too much? subtle bugs?
var VOID_SENTINEL = RadValue{Val: 0x0}
var JSON_SENTINEL = RadValue{Val: 0x1}

type RadValue struct {
	// int64, float64, RadString, bool stored as values
	// collections (lists, maps) stored as pointers
	// lists are *RadList
	// maps are *RadMap
	// functions are RadFn
	// nulls are RadNull
	// errors are *RadError
	Val interface{}
}

func (v RadValue) Type() RadTypeEnum {
	switch v.Val.(type) {
	case int64:
		return RadIntT
	case float64:
		return RadFloatT
	case RadString:
		return RadStringT
	case bool:
		return RadBoolT
	case *RadList:
		return RadListT
	case *RadMap:
		return RadMapT
	case RadFn:
		return RadFnT // todo add to equals, hash in this file
	case RadNull:
		return RadNullT
	case *RadError:
		return RadErrorT
	default:
		panic(fmt.Sprintf("Bug! Unhandled Rad type: %T", v.Val))
	}
}

func (v RadValue) Index(i *Interpreter, idxNode *ts.Node) RadValue {
	switch coerced := v.Val.(type) {
	case RadString:
		return newRadValue(i, idxNode, coerced.Index(i, idxNode))
	case *RadList:
		return newRadValue(i, idxNode, coerced.GetIdx(i, idxNode))
	case *RadMap:
		return newRadValue(i, idxNode, coerced.GetNode(i, idxNode))
	default:
		i.errorf(idxNode, "Indexing not supported for %s", TypeAsString(v))
		panic(UNREACHABLE)
	}
}

func (v RadValue) RequireInt(i *Interpreter, node *ts.Node) int64 {
	switch coerced := v.Val.(type) {
	case int64:
		return coerced
	default:
		i.errorf(node, "Expected int, got %q: %s", TypeAsString(v), ToPrintable(v))
		panic(UNREACHABLE)
	}
}

func (v RadValue) RequireIntAllowingBool(i *Interpreter, node *ts.Node) int64 {
	switch coerced := v.Val.(type) {
	case int64:
		return coerced
	case bool:
		if coerced {
			return 1
		}
		return 0
	default:
		i.errorf(node, "Expected int, got %q: %s", TypeAsString(v), ToPrintable(v))
		panic(UNREACHABLE)
	}
}

func (v RadValue) RequireStr(i *Interpreter, node *ts.Node) RadString {
	if str, ok := v.TryGetStr(); ok {
		return str
	}
	i.errorf(node, "Expected string, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) TryGetStr() (RadString, bool) {
	if str, ok := v.Val.(RadString); ok {
		return str, true
	}
	return NewRadString(""), false
}

func (v RadValue) RequireList(i *Interpreter, node *ts.Node) *RadList {
	if list, ok := v.TryGetList(); ok {
		return list
	}
	i.errorf(node, "Expected list, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) TryGetList() (*RadList, bool) {
	if list, ok := v.Val.(*RadList); ok {
		return list, true
	}
	return nil, false
}

func (v RadValue) RequireBool(i *Interpreter, node *ts.Node) bool {
	if b, ok := v.TryGetBool(); ok {
		return b
	}
	i.errorf(node, "Expected bool, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) TryGetBool() (bool, bool) {
	if b, ok := v.Val.(bool); ok {
		return b, true
	}
	return false, false
}

func (v RadValue) RequireMap(i *Interpreter, node *ts.Node) *RadMap {
	if b, ok := v.TryGetMap(); ok {
		return b
	}
	i.errorf(node, "Expected map, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) TryGetMap() (*RadMap, bool) {
	if m, ok := v.Val.(*RadMap); ok {
		return m, true
	}
	return nil, false
}

func (v RadValue) RequireFn(i *Interpreter, node *ts.Node) RadFn {
	if fn, ok := v.TryGetFn(); ok {
		return fn
	}
	i.errorf(node, "Expected function, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) TryGetFn() (RadFn, bool) {
	if fn, ok := v.Val.(RadFn); ok {
		return fn, true
	}
	var zero RadFn
	return zero, false
}

func (v RadValue) RequireError(i *Interpreter, node *ts.Node) *RadError {
	if err, ok := v.TryGetError(); ok {
		return err
	}
	i.errorf(node, "Expected error, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) TryGetError() (*RadError, bool) {
	if err, ok := v.Val.(*RadError); ok {
		return err, true
	}
	return nil, false
}

func (v RadValue) TryGetFloatAllowingInt() (float64, bool) {
	switch coerced := v.Val.(type) {
	case int64:
		return float64(coerced), true
	case float64:
		return coerced, true
	default:
		return 0, false
	}
}

func (v RadValue) RequireFloatAllowingInt(i *Interpreter, node *ts.Node) float64 {
	if f, ok := v.TryGetFloatAllowingInt(); ok {
		return f
	}
	i.errorf(node, "Expected float, got %q: %s", TypeAsString(v), ToPrintable(v))
	panic(UNREACHABLE)
}

func (v RadValue) IsError() bool {
	_, ok := v.Val.(*RadError)
	return ok
}

func (v RadValue) ModifyIdx(i *Interpreter, idxNode *ts.Node, rightValue RadValue) {
	switch coerced := v.Val.(type) {
	case *RadList:
		coerced.ModifyIdx(i, idxNode, rightValue)
	case *RadMap:
		if idxNode.Kind() == rl.K_IDENTIFIER {
			// dot syntax e.g. myMap.myKey
			keyName := i.sd.Src[idxNode.StartByte():idxNode.EndByte()]
			coerced.Set(newRadValueStr(keyName), rightValue)
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

func (v RadValue) Hash() string {
	switch val := v.Val.(type) {
	case RadString:
		return val.Plain() // attributes don't impact hash
	case int64, float64, bool:
		return fmt.Sprintf("%v", val)
	case *RadError:
		return val.Hash()
	default:
		panic(fmt.Sprintf("Cannot key on a %s", TypeAsString(v)))
	}
}

func (left RadValue) Equals(right RadValue) bool {
	leftT := left.Type()
	rightT := right.Type()

	if leftT != rightT {
		// todo should do bespoke float/int comparison
		return false
	}

	switch coercedLeft := left.Val.(type) {
	case RadString:
		coercedRight := right.Val.(RadString)
		return coercedLeft.Plain() == coercedRight.Plain()
	case int64, float64, bool:
		return left.Val == right.Val
	case *RadList:
		coercedRight := right.Val.(*RadList)
		return coercedLeft.Equals(coercedRight)
	case *RadMap:
		coercedRight := right.Val.(*RadMap)
		return coercedLeft.Equals(coercedRight)
	case RadNull:
		// we know they're both null, so true
		return true
	case *RadError:
		coercedRight := right.Val.(*RadError)
		return coercedLeft.Equals(coercedRight)
	default:
		return false
	}
}

func (v RadValue) RequireType(i *Interpreter, node *ts.Node, errPrefix string, allowedTypes ...RadTypeEnum) RadValue {
	for _, allowedType := range allowedTypes {
		if v.Type() == allowedType {
			return v
		}
	}

	i.errorf(node, "%s: %s", errPrefix, TypeAsString(v))
	panic(UNREACHABLE)
}

func (v RadValue) RequireNotType(
	i *Interpreter,
	node *ts.Node,
	errPrefix string,
	disallowedTypes ...RadTypeEnum,
) RadValue {
	for _, disallowedType := range disallowedTypes {
		if v.Type() == disallowedType {
			i.errorf(node, "%s: %s", errPrefix, TypeAsString(v))
			panic(UNREACHABLE)
		}
	}

	return v
}

func (v RadValue) TruthyFalsy() bool {
	if v == VOID_SENTINEL {
		// should we even error?
		return false
	}

	out := false
	NewTypeVisitorUnsafe().ForInt(func(v RadValue, i int64) {
		out = i != 0
	}).ForFloat(func(v RadValue, f float64) {
		out = f != 0
	}).ForString(func(v RadValue, s RadString) {
		out = s.Plain() != ""
	}).ForBool(func(v RadValue, b bool) {
		out = b
	}).ForList(func(v RadValue, l *RadList) {
		out = l.Len() != 0
	}).ForMap(func(v RadValue, m *RadMap) {
		out = m.Len() != 0
	}).ForNull(func(v RadValue, n RadNull) {
		out = false
	}).Visit(v)
	return out
}

func (v RadValue) Accept(visitor *RadTypeVisitor) {
	switch coerced := v.Val.(type) {
	case bool:
		if visitor.visitBool != nil {
			visitor.visitBool(v, coerced)
			return
		}
	case int64:
		if visitor.visitInt != nil {
			visitor.visitInt(v, coerced)
			return
		}
	case float64:
		if visitor.visitFloat != nil {
			visitor.visitFloat(v, coerced)
			return
		}
	case RadString:
		if visitor.visitString != nil {
			visitor.visitString(v, coerced)
			return
		}
	case *RadList:
		if visitor.visitList != nil {
			visitor.visitList(v, coerced)
			return
		}
	case *RadMap:
		if visitor.visitMap != nil {
			visitor.visitMap(v, coerced)
			return
		}
	case RadFn:
		if visitor.visitFn != nil {
			visitor.visitFn(v, coerced)
			return
		}
	case RadNull:
		if visitor.visitNull != nil {
			visitor.visitNull(v, coerced)
			return
		}
	case *RadError:
		if visitor.visitError != nil {
			visitor.visitError(v, coerced)
			return
		}
	}
	if visitor.defaultVisit != nil {
		visitor.defaultVisit(v)
		return
	}
	visitor.UnhandledTypeError(v)
}

func newRadValue(i *Interpreter, node *ts.Node, value interface{}) RadValue {
	switch coerced := value.(type) {
	case RadValue:
		return coerced
	case []RadValue: // todo risky to have this? might cover up issues
		list := NewRadList()
		list.Values = coerced
		return newRadValue(i, node, list)
	case RadString:
		return RadValue{Val: coerced}
	case string:
		return RadValue{Val: NewRadString(coerced)}
	case int:
		return RadValue{Val: int64(coerced)}
	case int64, float64, bool:
		return RadValue{Val: coerced}
	case *RadList:
		return RadValue{Val: coerced}
	case RadList:
		return RadValue{Val: &coerced}
	case *RadMap:
		return RadValue{Val: coerced}
	case RadMap:
		return RadValue{Val: &coerced}
	case RadFn:
		return RadValue{Val: coerced}
	case *RadError:
		return RadValue{Val: coerced}
	case map[string]interface{}:
		radMap := NewRadMap()
		for key, val := range coerced {
			radMap.Set(newRadValue(i, node, key), newRadValue(i, node, val))
		}
		return RadValue{Val: radMap}
	case []interface{}:
		list := NewRadListFromGeneric(i, node, coerced)
		return RadValue{Val: list}
	case []string:
		list := NewRadListFromGeneric(i, node, coerced)
		return RadValue{Val: list}
	case RadNull, nil:
		return RadValue{Val: RAD_NULL}
	default:
		if i != nil && node != nil {
			i.errorf(node, "Unsupported value type: %s", TypeAsString(coerced))
			panic(UNREACHABLE)
		} else {
			panic(fmt.Sprintf("Bug! Unsafe call w/ unsupported value type: %T", coerced))
		}
	}
}

func newRadValues(i *Interpreter, node *ts.Node, value ...interface{}) RadValue {
	if len(value) == 0 {
		return RAD_NULL_VAL
	}

	if len(value) == 1 {
		val := value[0]
		err, ok := val.(*RadError)
		if ok && err.Node == nil {
			err.SetNode(node)
		}
		return newRadValue(i, node, val)
	}

	list := NewRadList()
	for _, v := range value {
		list.Append(newRadValue(i, node, v))
	}
	return newRadValue(i, node, list)
}

func newRadValueStr(str string) RadValue {
	return newRadValue(nil, nil, str)
}

func newRadValueRadStr(str RadString) RadValue {
	return newRadValue(nil, nil, str)
}

func newRadValueInt(val int) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueInt64(val int64) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueFloat64(val float64) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueBool(val bool) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueMap(val *RadMap) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueList(val *RadList) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueFn(val RadFn) RadValue {
	return newRadValue(nil, nil, val)
}

func newRadValueError(val *RadError) RadValue {
	return newRadValue(nil, nil, val)
}
