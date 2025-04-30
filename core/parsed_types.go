package core

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslTypeEnum int

const (
	RslStringT RslTypeEnum = iota
	RslIntT
	RslFloatT
	RslBoolT
	RslListT
	RslMapT
	RslFnT
	RslNullT
)

func (r RslTypeEnum) AsString() string {
	switch r {
	case RslStringT:
		return "string"
	case RslIntT:
		return "int"
	case RslFloatT:
		return "float"
	case RslBoolT:
		return "bool"
	case RslListT:
		return "list"
	case RslMapT:
		return "map"
	case RslFnT:
		return "function"
	case RslNullT:
		return "null"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled RSL type: %v", r))
		panic(UNREACHABLE)
	}
}

type RslArgTypeT int

const (
	ArgStringT RslArgTypeT = iota
	ArgIntT
	ArgFloatT
	ArgBoolT
	ArgStringArrayT
	ArgIntArrayT
	ArgFloatArrayT
	ArgBoolArrayT
)

func ToRslArgTypeT(str string) RslArgTypeT {
	switch str {
	case "string":
		return ArgStringT
	case "int":
		return ArgIntT
	case "float":
		return ArgFloatT
	case "bool":
		return ArgBoolT
	case "string[]":
		return ArgStringArrayT
	case "int[]":
		return ArgIntArrayT
	case "float[]":
		return ArgFloatArrayT
	case "bool[]":
		return ArgBoolArrayT
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled RSL type: %v", str))
		panic(UNREACHABLE)
	}
}

func (r *RslArgTypeT) AsString() string {
	switch *r {
	case ArgStringT:
		return "string"
	case ArgIntT:
		return "int"
	case ArgFloatT:
		return "float"
	case ArgBoolT:
		return "bool"
	case ArgStringArrayT:
		return "string list"
	case ArgIntArrayT:
		return "int list"
	case ArgFloatArrayT:
		return "float list"
	case ArgBoolArrayT:
		return "bool list"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled RSL type: %v", *r))
		panic(UNREACHABLE)
	}
}

type SortDir int

const (
	Asc SortDir = iota
	Desc
)

type RadBlockType string

const (
	Rad     RadBlockType = "rad"
	Request RadBlockType = "request"
	Display RadBlockType = "display"
)

type Lambda struct { // todo delete, replace with RslFn
	Node     *ts.Node
	Args     []string
	ExprNode *ts.Node
}

type OpType int

const ( // todo replace with Node Kinds?
	OP_PLUS OpType = iota
	OP_MINUS
	OP_MULTIPLY
	OP_DIVIDE
	OP_MODULO
	OP_EQUAL
	OP_NOT_EQUAL
	OP_IN
	OP_NOT_IN
	OP_GREATER
	OP_GREATER_EQUAL
	OP_LESS
	OP_LESS_EQUAL
	OP_AND
	OP_OR
	//OpPow?
)
