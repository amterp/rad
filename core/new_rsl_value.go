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
	// todo handle identifier dot syntax
	// todo handle slice nodes

	switch coerced := v.Val.(type) {
	case *RslList:
		return newRslValue(i, idxNode, coerced.GetIdx(i, idxNode))
	//case *RslMap: todo
	//	idxStr := idx.RequireString(i, idxNode)
	//	return newRslValue(i, idxNode, (*coerced).Get(idxStr))
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

func (v RslValue) ModifyIdx(i *Interpreter, idxNode *ts.Node, rightValue RslValue) {
	// todo handle identifier dot syntax
	// todo handle slice nodes

	switch coerced := v.Val.(type) {
	case *RslList:
		coerced.ModifyIdx(i, idxNode, rightValue)
	//case *RslMap: todo
	//	idxStr := idx.RequireString(i, idxNode)
	//	return newRslValue(i, idxNode, (*coerced).Get(idxStr))
	default:
		i.errorf(idxNode, "Indexing not supported for %s", TypeAsString(v))
		panic(UNREACHABLE)
	}
}

func newRslValue(i *Interpreter, node *ts.Node, value interface{}) RslValue {
	switch coerced := value.(type) {
	case RslValue:
		return coerced
	case string:
		return RslValue{Val: NewRslString(coerced)}
	case int64, float64, bool:
		return RslValue{Val: coerced}
	case RslList:
		return RslValue{Val: &coerced}
	case *RslList:
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
