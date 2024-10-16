package core

import "fmt"

type ScriptArg struct {
	Name             string
	ApiName          string
	DeclarationToken Token
	Flag             *string
	Type             RslArgTypeT
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
	DefaultBoolArray   *[]bool
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
		case ArgStringT:
			val := literal.(string)
			scriptArg.DefaultString = &val
		case ArgStringArrayT:
			arr, ok := literal.([]interface{})
			if !ok {
				RP.TokenErrorExit(argDecl.Identifier, "Expected array of strings as default")
			}
			var vals []string
			for _, elem := range arr {
				vals = append(vals, elem.(string))
			}
			scriptArg.DefaultStringArray = &vals
		case ArgIntT:
			val := literal.(int64)
			scriptArg.DefaultInt = &val
		case ArgIntArrayT:
			arr, ok := literal.([]interface{})
			if !ok {
				RP.TokenErrorExit(argDecl.Identifier, "Expected array of ints as default")
			}
			var vals []int64
			for _, elem := range arr {
				vals = append(vals, elem.(int64))
			}
			scriptArg.DefaultIntArray = &vals
		case ArgFloatT:
			val := literal.(float64)
			scriptArg.DefaultFloat = &val
		case ArgFloatArrayT:
			arr, ok := literal.([]interface{})
			if !ok {
				RP.TokenErrorExit(argDecl.Identifier, "Expected array of floats as default")
			}
			var vals []float64
			for _, elem := range arr {
				vals = append(vals, elem.(float64))
			}
			scriptArg.DefaultFloatArray = &vals
		case ArgBoolT:
			val := literal.(bool)
			scriptArg.DefaultBool = &val
		case ArgBoolArrayT:
			arr, ok := literal.([]interface{})
			if !ok {
				RP.TokenErrorExit(argDecl.Identifier, "Expected array of bools as default")
			}
			var vals []bool
			for _, elem := range arr {
				vals = append(vals, elem.(bool))
			}
			scriptArg.DefaultBoolArray = &vals
		default:
			l.i.error(scriptArg.DeclarationToken, fmt.Sprintf("Unknown arg type: %v", scriptArg.Type))
		}
	}

	return scriptArg
}
