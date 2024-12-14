package core

import "fmt"

type RslTypeEnum int

const (
	RslStringT RslTypeEnum = iota
	RslIntT
	RslFloatT
	RslBoolT
	RslArrayT
	RslMapT
)

type RslArgTypeT int

const (
	ArgStringT RslArgTypeT = iota
	ArgIntT
	ArgFloatT
	ArgBoolT
	ArgMixedArrayT
	ArgStringArrayT
	ArgIntArrayT
	ArgFloatArrayT
	ArgBoolArrayT
)

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
	case ArgMixedArrayT:
		return "mixed array"
	case ArgStringArrayT:
		return "string array"
	case ArgIntArrayT:
		return "int array"
	case ArgFloatArrayT:
		return "float array"
	case ArgBoolArrayT:
		return "bool array"
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled RSL type: %v", *r))
		panic(UNREACHABLE)
	}
}

func (r *RslTypeEnum) MatchesValue(val interface{}) bool {
	if r == nil {
		return false
	}

	switch *r {
	case RslStringT:
		_, ok := val.(RslString)
		return ok
	case RslIntT:
		_, ok := val.(int64)
		return ok
	case RslFloatT:
		_, ok := val.(float64)
		return ok
	case RslBoolT:
		_, ok := val.(bool)
		return ok
	case RslArrayT:
		_, ok := val.([]interface{})
		return ok
	case RslMapT:
		_, ok := val.(RslMap)
		return ok
	default:
		RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled RSL type: %v", *r))
	}

	return false
}

func (r *RslTypeEnum) IsArray() bool {
	if r == nil {
		return false
	}

	return *r == RslArrayT
}

type RslArgType struct {
	Token Token
	Type  RslArgTypeT
}

type JsonPath struct {
	Elements []JsonPathElement
}

type JsonPathElement struct {
	Identifier Token
	ArrElems   []JsonPathElementArr
}

type JsonPathElementArr struct {
	ArrayToken *Token // e.g. json.names[]
	Index      *Expr  // e.g. json.names[0]
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

type InlineExpr struct {
	Expression Expr
	Formatting *InlineExprFormat
}

type InlineExprFormat struct {
	GoFormat      string // does not contain 's' or 'f' at the end; need to add at runtime depending on the given type
	RslFormat     string
	IsFloatFormat bool // i.e. something like '.2' has been specified, requiring decimal places
}

type IfCase struct {
	IfToken   Token
	Condition Expr
	Body      Block
}

type RadIfCase struct {
	IfToken   Token
	Condition Expr
	Body      []RadStmt
}

type NamedArg struct {
	Arg   Token
	Value Expr
}

type Lambda struct {
	Args []Token
	Op   Expr
}

type CollectionKey struct {
	Opener  Token
	IsSlice bool
	Start   *Expr
	End     *Expr
}

type OpType int

const (
	OP_PLUS OpType = iota
	OP_MINUS
	OP_MULTIPLY
	OP_DIVIDE
	OP_EQUAL
	OP_NOT_EQUAL
	OP_IN
	OP_NOT_IN
	OP_GREATER
	OP_GREATER_EQUAL
	OP_LESS
	OP_LESS_EQUAL
	//OpMod
	//OpPow?
)

var (
	TKN_TYPE_TO_OP_MAP = map[TokenType]OpType{
		PLUS:        OP_PLUS,
		PLUS_EQUAL:  OP_PLUS,
		MINUS:       OP_MINUS,
		MINUS_EQUAL: OP_MINUS,
		STAR:        OP_MULTIPLY,
		STAR_EQUAL:  OP_MULTIPLY,
		SLASH:       OP_DIVIDE,
		SLASH_EQUAL: OP_DIVIDE,

		EQUAL_EQUAL: OP_EQUAL,
		NOT_EQUAL:   OP_NOT_EQUAL,

		IN:     OP_IN,
		NOT_IN: OP_NOT_IN,

		GREATER:       OP_GREATER,
		GREATER_EQUAL: OP_GREATER_EQUAL,
		LESS:          OP_LESS,
		LESS_EQUAL:    OP_LESS_EQUAL,
	}
)
