package core

import (
	"fmt"
	"rad/cmd"
)

type Env struct {
	i    *MainInterpreter
	Vars map[string]RuntimeLiteral
	//jsonFields map[string]*JsonField
	Enclosing *Env
}

func NewEnv(i *MainInterpreter) *Env {
	return &Env{
		i:    i,
		Vars: make(map[string]RuntimeLiteral),
		//jsonFields: make(map[string]*JsonField),
		Enclosing: nil,
	}
}

func (e *Env) InitArg(arg cmd.CobraArg) {
	argType := arg.Arg.Type
	switch argType {
	case RslString:
		e.Vars[arg.Arg.Name] = *NewRuntimeString(arg.GetString())
	case RslStringArray:
		e.Vars[arg.Arg.Name] = *NewRuntimeStringArray(arg.GetStringArray())
	case RslInt:
		e.Vars[arg.Arg.Name] = *NewRuntimeInt(arg.GetInt())
	case RslIntArray:
		e.Vars[arg.Arg.Name] = *NewRuntimeIntArray(arg.GetIntArray())
	case RslFloat:
		e.Vars[arg.Arg.Name] = *NewRuntimeFloat(arg.GetFloat())
	case RslFloatArray:
		e.Vars[arg.Arg.Name] = *NewRuntimeFloatArray(arg.GetFloatArray())
	case RslBool:
		e.Vars[arg.Arg.Name] = *NewRuntimeBool(arg.GetBool())
	default:
		e.i.error(arg.Arg.DeclarationToken, fmt.Sprintf("Unknown arg type, cannot init: %v", argType))
	}
}

// Set 'value' expected to be a pointer e.g. *string
func (e *Env) Set(varNameToken Token, value interface{}) {
	// todo could make the literal interpreter return LiteralOrArray instead of Go values, making this translation better

	varName := varNameToken.GetLexeme()
	switch value.(type) {
	case *string:
		e.Vars[varName] = *NewRuntimeString(value.(*string))
	case *[]string:
		e.Vars[varName] = *NewRuntimeStringArray(value.(*[]string))
	case *int:
		e.Vars[varName] = *NewRuntimeInt(value.(*int))
	case *[]int:
		e.Vars[varName] = *NewRuntimeIntArray(value.(*[]int))
	case *float64:
		e.Vars[varName] = *NewRuntimeFloat(value.(*float64))
	case *[]float64:
		e.Vars[varName] = *NewRuntimeFloatArray(value.(*[]float64))
	case *bool:
		e.Vars[varName] = *NewRuntimeBool(value.(*bool))
	default:
		e.i.error(varNameToken, fmt.Sprintf("Unknown type, cannot set: %v = %v", varName, value))
	}
}

func (e *Env) Get(varNameToken Token, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	varName := varNameToken.GetLexeme()
	val, ok := e.Vars[varName]
	if !ok {
		e.i.error(varNameToken, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}

	for _, acceptableType := range acceptableTypes {
		if val.Type == acceptableType {
			return val
		}
	}
	e.i.error(varNameToken, fmt.Sprintf("Variable type mismatch: %v, expected: %v", varName, acceptableTypes))
	panic(UNREACHABLE)
}
