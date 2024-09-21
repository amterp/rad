package core

import "fmt"

type ScriptArg struct {
	Name             string
	ApiName          string
	DeclarationToken Token
	Flag             *string
	Type             RslTypeEnum
	Description      *string
	IsOptional       bool
	// first check the Type and IsOptional, then get the value
	// todo I think just make these non-pointers, and have a separate flag to indicate the arg is set
	DefaultString      *string
	DefaultStringArray *[]string
	DefaultInt         *int64
	DefaultIntArray    *[]int64
	DefaultFloat       *float64
	DefaultFloatArray  *[]float64
	DefaultBool        *bool
}

func FromArgDecl(l *LiteralInterpreter, argDecl *ArgDeclaration) *ScriptArg {
	name := argDecl.Identifier.GetLexeme()
	apiName := name
	rename := argDecl.Rename
	if NotNil(rename, func() Token { return nil }) {
		apiName = (*rename).(*StringLiteralToken).Literal
	}

	var flag *string
	flagToken := argDecl.Flag
	if NotNil(flagToken, func() Token { return nil }) {
		lexeme := (*flagToken).GetLexeme()
		if len(lexeme) != 1 {
			l.i.error(*flagToken, fmt.Sprintf("Flag %q must be a single character", lexeme))
		}
		flag = &lexeme
	}

	var comment *string
	if argDecl.Comment != nil {
		comment = argDecl.Comment.Literal
	}

	scriptArg := &ScriptArg{
		Name:             name,
		ApiName:          apiName,
		DeclarationToken: argDecl.Identifier,
		Flag:             flag,
		Type:             argDecl.ArgType.Type,
		Description:      comment,
		IsOptional:       argDecl.IsOptional,
	}

	defaultVal := argDecl.Default
	if NotNil(defaultVal, func() LiteralOrArray { return nil }) {
		literal := (*defaultVal).Accept(l)
		switch scriptArg.Type {
		case RslString:
			val := literal.(string)
			scriptArg.DefaultString = &val
		case RslStringArray:
			if _, isEmptyArray := literal.([]interface{}); isEmptyArray {
				var val []string
				scriptArg.DefaultStringArray = &val
			} else {
				val := literal.([]string)
				scriptArg.DefaultStringArray = &val
			}
		case RslInt:
			val := literal.(int64)
			scriptArg.DefaultInt = &val
		case RslIntArray:
			if _, isEmptyArray := literal.([]interface{}); isEmptyArray {
				var val []int64
				scriptArg.DefaultIntArray = &val
			} else {
				val := literal.([]int64)
				scriptArg.DefaultIntArray = &val
			}
		case RslFloat:
			val := literal.(float64)
			scriptArg.DefaultFloat = &val
		case RslFloatArray:
			if _, isEmptyArray := literal.([]interface{}); isEmptyArray {
				var val []float64
				scriptArg.DefaultFloatArray = &val
			} else {
				val := literal.([]float64)
				scriptArg.DefaultFloatArray = &val
			}
		case RslBool:
			val := literal.(bool)
			scriptArg.DefaultBool = &val
		default:
			l.i.error(scriptArg.DeclarationToken, fmt.Sprintf("Unknown arg type: %v", scriptArg.Type))
		}
	}

	return scriptArg
}
