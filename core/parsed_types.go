package core

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadTypeEnum int

const (
	RadStringT RadTypeEnum = iota
	RadIntT
	RadFloatT
	RadBoolT
	RadListT
	RadMapT
	RadFnT
	RadNullT
	RadErrorT
)

func (r RadTypeEnum) AsString() string {
	switch r {
	case RadStringT:
		return "string"
	case RadIntT:
		return "int"
	case RadFloatT:
		return "float"
	case RadBoolT:
		return "bool"
	case RadListT:
		return "list"
	case RadMapT:
		return "map"
	case RadFnT:
		return "function"
	case RadNullT:
		return "null"
	case RadErrorT:
		return "error"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled Rad type: %v", r))
		panic(UNREACHABLE)
	}
}

type RadArgTypeT int

const (
	ArgStringT RadArgTypeT = iota
	ArgIntT
	ArgFloatT
	ArgBoolT
	ArgStringArrayT
	ArgIntArrayT
	ArgFloatArrayT
	ArgBoolArrayT
)

func ToRadArgTypeT(str string) RadArgTypeT {
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
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled Rad type: %v", str))
		panic(UNREACHABLE)
	}
}

func (r *RadArgTypeT) AsString() string {
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
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled Rad type: %v", *r))
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
	RadBlock     RadBlockType = "rad"
	RequestBlock RadBlockType = "request"
	DisplayBlock RadBlockType = "display"
)

type Lambda struct { // todo delete, replace with RadFn
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
