package core

type RslTypeEnum int

const (
	RslString RslTypeEnum = iota
	RslInt
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
