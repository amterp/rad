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

type RslType struct {
	Token Token
	Type  RslTypeEnum
}

type JsonPath struct {
	identifier Token
	elements   []JsonPathElement
}

type JsonPathElement struct {
	token      Token
	arrayToken *Token
}
