package core

import "fmt"

type ScriptArg struct {
	Name             string
	DeclarationToken Token
	Flag             *string
	Type             RslTypeEnum
	Description      *string
	IsOptional       bool
	// first check the Type and IsOptional, then get the value
	// todo I think just make these non-pointers, and have a separate flag to indicate the arg is set
	DefaultString      *string
	DefaultStringArray *[]string
	DefaultInt         *int
	DefaultIntArray    *[]int
	DefaultFloat       *float64
	DefaultFloatArray  *[]float64
	DefaultBool        *bool
}

func FromArgDecl(i *LiteralInterpreter, argDecl *ArgDeclaration) *ScriptArg {
	var name string
	rename := argDecl.Rename
	if NotNil(rename, func() Token { return nil }) {
		name = (*rename).GetLexeme()
	} else {
		name = argDecl.Identifier.GetLexeme()
	}

	var flag *string
	flagToken := argDecl.Flag
	if NotNil(flagToken, func() Token { return nil }) {
		lexeme := (*flagToken).GetLexeme()
		flag = &lexeme
	}

	scriptArg := &ScriptArg{
		Name:             name,
		DeclarationToken: argDecl.Identifier,
		Flag:             flag,
		Type:             argDecl.ArgType.Type,
		Description:      argDecl.Comment.Literal,
		IsOptional:       argDecl.IsOptional,
	}

	defaultVal := argDecl.Default
	if NotNil(defaultVal, func() LiteralOrArray { return nil }) {
		literal := (*defaultVal).Accept(i)
		switch scriptArg.Type {
		case RslString:
			scriptArg.DefaultString = literal.(*string)
		case RslStringArray:
			scriptArg.DefaultStringArray = literal.(*[]string)
		case RslInt:
			scriptArg.DefaultInt = literal.(*int)
		case RslIntArray:
			scriptArg.DefaultIntArray = literal.(*[]int)
		case RslFloat:
			scriptArg.DefaultFloat = literal.(*float64)
		case RslFloatArray:
			scriptArg.DefaultFloatArray = literal.(*[]float64)
		case RslBool:
			scriptArg.DefaultBool = literal.(*bool)
		default:
			// todo improve
			panic(fmt.Sprintf("Unknown arg type: %v", scriptArg.Type))
		}
	}

	return scriptArg
}
