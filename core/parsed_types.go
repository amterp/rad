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
		_, ok := val.(string)
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
	elements []JsonPathElement
}

type JsonPathElement struct {
	token      JsonPathElementToken
	arrayToken *Token
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
