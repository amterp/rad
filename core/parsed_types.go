package core

type RslTypeEnum int

const (
	RslString RslTypeEnum = iota
	RslStringArray
	RslInt
	RslIntArray
	RslFloat
	RslFloatArray
	RslBool
)

func (r *RslTypeEnum) NonArrayType() *RslTypeEnum {
	if r == nil {
		return nil
	}

	var output RslTypeEnum
	if *r == RslStringArray {
		output = RslString
	}
	if *r == RslIntArray {
		output = RslInt
	}
	if *r == RslFloatArray {
		output = RslFloat
	}
	return &output
}

func (r *RslTypeEnum) IsArray() bool {
	if r == nil {
		return false
	}

	return *r == RslStringArray || *r == RslIntArray || *r == RslFloatArray
}

type RslType struct {
	Token Token
	Type  RslTypeEnum
}

type JsonPath struct {
	elements []JsonPathElement
}

type JsonPathElement struct {
	token      Token
	arrayToken *Token
}
