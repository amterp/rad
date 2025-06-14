package core

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"

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
		return rl.T_STR
	case RadIntT:
		return rl.T_INT
	case RadFloatT:
		return rl.T_FLOAT
	case RadBoolT:
		return rl.T_BOOL
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
	ArgStrListT
	ArgIntListT
	ArgFloatListT
	ArgBoolListT
)

func ToRadArgTypeT(str string) RadArgTypeT {
	switch str {
	case rl.T_STR:
		return ArgStringT
	case rl.T_INT:
		return ArgIntT
	case rl.T_FLOAT:
		return ArgFloatT
	case rl.T_BOOL:
		return ArgBoolT
	case rl.T_STR_LIST:
		return ArgStrListT
	case rl.T_INT_LIST:
		return ArgIntListT
	case rl.T_FLOAT_LIST:
		return ArgFloatListT
	case rl.T_BOOL_LIST:
		return ArgBoolListT
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled Rad type: %v", str))
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
	// OpPow?
)
