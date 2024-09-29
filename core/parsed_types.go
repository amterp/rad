package core

type RslTypeEnum int

const (
	RslStringT RslTypeEnum = iota
	RslIntT
	RslFloatT
	RslBoolT
	RslArrayT
	RslStringArrayT
	RslIntArrayT
	RslFloatArrayT
	RslBoolArrayT
)

func (r *RslTypeEnum) IsArray() bool {
	if r == nil {
		return false
	}

	return *r == RslArrayT
}

type RslType struct {
	Token Token
	Type  RslTypeEnum
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
