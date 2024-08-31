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

func FromGoValue(val interface{}) RslTypeEnum {
	switch val.(type) {
	case *string:
		return RslString
	case *[]string:
		return RslStringArray
	case *int:
		return RslInt
	case *[]int:
		return RslIntArray
	case *float64:
		return RslFloat
	case *[]float64:
		return RslFloatArray
	case *bool:
		return RslBool
	default:
		panic("Unknown type")
	}
}

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
