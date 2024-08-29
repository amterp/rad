package core

import "rad/core/interpreters"

type ScriptArg struct {
	Name        string
	Flag        *string
	Type        RslTypeEnum
	Description *string
	IsOptional  bool
	// first check the Type and IsOptional, then get the value
	DefaultString      *string
	DefaultStringArray *[]string
	DefaultInt         *int
	DefaultIntArray    *[]int
	DefaultBool        *bool
}

func FromArgDecl(i *interpreters.LiteralInterpreter, argDecl *ArgDeclaration) *ScriptArg {
	return &ScriptArg{
		Name:        argDecl.Identifier.GetLexeme(),
		Flag:        nil,
		Type:        argDecl.ArgType.Type,
		Description: nil,
		IsOptional:  argDecl.IsOptional,
	}
}
