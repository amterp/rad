package core

type RslTypeEnum int

const (
	RslStringT RslTypeEnum = iota
	RslStringArrayT
	RslIntT
	RslIntArrayT
	RslFloatT
	RslFloatArrayT
	RslBoolT
)

func (r *RslTypeEnum) NonArrayType() *RslTypeEnum {
	if r == nil {
		return nil
	}

	var output RslTypeEnum
	if *r == RslStringArrayT {
		output = RslStringT
	}
	if *r == RslIntArrayT {
		output = RslIntT
	}
	if *r == RslFloatArrayT {
		output = RslFloatT
	}
	return &output
}

func (r *RslTypeEnum) IsArray() bool {
	if r == nil {
		return false
	}

	return *r == RslStringArrayT || *r == RslIntArrayT || *r == RslFloatArrayT
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
