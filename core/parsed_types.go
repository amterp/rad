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
