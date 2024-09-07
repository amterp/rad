package core

import (
	"fmt"
)

type Env struct {
	i          *MainInterpreter
	Vars       map[string]RuntimeLiteral
	jsonFields map[string]JsonFieldVar
	Enclosing  *Env
}

func NewEnv(i *MainInterpreter) *Env {
	return &Env{
		i:          i,
		Vars:       make(map[string]RuntimeLiteral),
		jsonFields: make(map[string]JsonFieldVar),
		Enclosing:  nil,
	}
}

func (e *Env) InitArg(arg CobraArg) {
	if arg.IsNull {
		return
	}

	argType := arg.Arg.Type
	switch argType {
	case RslString:
		e.Vars[arg.Arg.Name] = NewRuntimeString(arg.GetString())
	case RslStringArray:
		e.Vars[arg.Arg.Name] = NewRuntimeStringArray(arg.GetStringArray())
	case RslInt:
		e.Vars[arg.Arg.Name] = NewRuntimeInt(arg.GetInt())
	case RslIntArray:
		e.Vars[arg.Arg.Name] = NewRuntimeIntArray(arg.GetIntArray())
	case RslFloat:
		e.Vars[arg.Arg.Name] = NewRuntimeFloat(arg.GetFloat())
	case RslFloatArray:
		e.Vars[arg.Arg.Name] = NewRuntimeFloatArray(arg.GetFloatArray())
	case RslBool:
		e.Vars[arg.Arg.Name] = NewRuntimeBool(arg.GetBool())
	default:
		e.i.error(arg.Arg.DeclarationToken, fmt.Sprintf("Unknown arg type, cannot init: %v", argType))
	}
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *Env) SetAndImplyType(varNameToken Token, value interface{}) {
	// todo could make the literal interpreter return LiteralOrArray instead of Go values, making this translation better

	varName := varNameToken.GetLexeme()
	switch value.(type) {
	case string:
		e.Vars[varName] = NewRuntimeString(value.(string))
	case []string:
		e.Vars[varName] = NewRuntimeStringArray(value.([]string))
	case int:
		e.Vars[varName] = NewRuntimeInt(value.(int))
	case []int:
		e.Vars[varName] = NewRuntimeIntArray(value.([]int))
	case float64:
		e.Vars[varName] = NewRuntimeFloat(value.(float64))
	case []float64:
		e.Vars[varName] = NewRuntimeFloatArray(value.([]float64))
	case bool:
		e.Vars[varName] = NewRuntimeBool(value.(bool))
	default:
		e.i.error(varNameToken, fmt.Sprintf("Unknown type, cannot set: '%T' %q = %q", value, varName, value))
	}
}

// SetAndExpectType 'value' expected to not be a pointer, should be e.g. string
func (e *Env) SetAndExpectType(varNameToken Token, expectedType *RslTypeEnum, value interface{}) {
	varName := varNameToken.GetLexeme()
	if expectedType != nil {
		expectedTypeVal := *expectedType
		switch expectedTypeVal {
		case RslString:
			val, ok := value.(string)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected string: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeString(val)
			}
		case RslStringArray:
			if _, isEmptyArray := value.([]interface{}); isEmptyArray {
				e.Vars[varName] = NewRuntimeStringArray([]string{})
			} else {
				val, ok := value.([]string)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected string array: %v", value))
				}
				e.Vars[varName] = NewRuntimeStringArray(val)
			}
		case RslInt:
			val, ok := value.(int)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected int: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeInt(val)
			}
		case RslIntArray:
			if _, isEmptyArray := value.([]interface{}); isEmptyArray {
				e.Vars[varName] = NewRuntimeIntArray([]int{})
			} else {
				val, ok := value.([]int)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected int array: %v", value))
				}
				e.Vars[varName] = NewRuntimeIntArray(val)
			}
		case RslFloat:
			val, ok := value.(float64)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected float: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeFloat(val)
			}
		case RslFloatArray:
			if _, isEmptyArray := value.([]interface{}); isEmptyArray {
				e.Vars[varName] = NewRuntimeFloatArray([]float64{})
			} else {
				val, ok := value.([]float64)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected float array: %v", value))
				}
				e.Vars[varName] = NewRuntimeFloatArray(val)
			}
		case RslBool:
			val, ok := value.(bool)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected bool: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeBool(val)
			}
		default:
			e.i.error(varNameToken, fmt.Sprintf("Unknown type, cannot set: %v = %v", varName, value))
		}
	}
}

func (e *Env) Exists(name string) bool {
	_, ok := e.Vars[name]
	return ok
}

func (e *Env) GetByToken(varNameToken Token, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	return e.get(varNameToken.GetLexeme(), varNameToken, acceptableTypes...)
}

func (e *Env) GetByName(varName string, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	return e.get(varName, nil, acceptableTypes...)
}

func (e *Env) AssignJsonField(name Token, path JsonPath) {
	e.jsonFields[name.GetLexeme()] = JsonFieldVar{
		Name: name,
		Path: path,
		env:  e,
	}
	e.SetAndImplyType(name, []string{})
}

func (e *Env) GetJsonField(name Token) JsonFieldVar {
	field, ok := e.jsonFields[name.GetLexeme()]
	if !ok {
		e.i.error(name, fmt.Sprintf("Undefined json field referenced: %v", name.GetLexeme()))
	}
	return field
}

func (e *Env) get(varName string, varNameToken Token, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	val, ok := e.Vars[varName]
	if !ok {
		e.i.error(varNameToken, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}

	if len(acceptableTypes) == 0 {
		return val
	}

	for _, acceptableType := range acceptableTypes {
		if val.Type == acceptableType {
			return val
		}
	}
	e.i.error(varNameToken, fmt.Sprintf("Variable type mismatch: %v, expected: %v", varName, acceptableTypes))
	panic(UNREACHABLE)
}
