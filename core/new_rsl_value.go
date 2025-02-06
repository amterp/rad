package core

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

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
	// todo handle slice nodes
	switch coerced := v.Val.(type) {
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
		i.errorf(node, "Expected int, got %s", TypeAsString(v))
		panic(UNREACHABLE)
	}
}

func (v RslValue) RequireStr(i *Interpreter, node *ts.Node) RslString {
	if str, ok := v.TryGetStr(); ok {
		return str
	}
	i.errorf(node, "Expected string, got %s", TypeAsString(v))
	panic(UNREACHABLE)
}

func (v RslValue) TryGetStr() (RslString, bool) {
	if str, ok := v.Val.(RslString); ok {
		return str, true
	}
	return NewRslString(""), false
}

func (v RslValue) ModifyIdx(i *Interpreter, idxNode *ts.Node, rightValue RslValue) {
	// todo handle slice nodes

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
		i.errorf(idxNode, "Indexing not supported for %s", TypeAsString(v))
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
	case RslString:
		return RslValue{Val: coerced}
	case string:
		return RslValue{Val: NewRslString(coerced)}
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
	default:
		i.errorf(node, "Unsupported value type: %s", TypeAsString(coerced))
		panic(UNREACHABLE)
	}
}

func newRslValues(i *Interpreter, node *ts.Node, value interface{}) []RslValue {
	return []RslValue{newRslValue(i, node, value)}
}

func newRslValueStr(str string) RslValue {
	return newRslValue(nil, nil, str)
}

func newRslValueRslStr(str RslString) RslValue {
	return newRslValue(nil, nil, str)
}
