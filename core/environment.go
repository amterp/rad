package core

import (
	"encoding/json"
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

func (e *Env) NewChildEnv() Env {
	return Env{
		i:          e.i,
		Vars:       make(map[string]RuntimeLiteral),
		jsonFields: make(map[string]JsonFieldVar),
		Enclosing:  e,
	}
}

func (e *Env) InitArg(arg CobraArg) {
	if arg.IsNull {
		return
	}

	argType := arg.Arg.Type
	switch argType {
	case RslStringT:
		e.Vars[arg.Arg.Name] = NewRuntimeString(arg.GetString())
	case RslStringArrayT:
		e.Vars[arg.Arg.Name] = NewRuntimeStringArray(arg.GetStringArray())
	case RslIntT:
		e.Vars[arg.Arg.Name] = NewRuntimeInt(arg.GetInt())
	case RslIntArrayT:
		e.Vars[arg.Arg.Name] = NewRuntimeIntArray(arg.GetIntArray())
	case RslFloatT:
		e.Vars[arg.Arg.Name] = NewRuntimeFloat(arg.GetFloat())
	case RslFloatArrayT:
		e.Vars[arg.Arg.Name] = NewRuntimeFloatArray(arg.GetFloatArray())
	case RslBoolT:
		e.Vars[arg.Arg.Name] = NewRuntimeBool(arg.GetBool())
	case RslBoolArrayT:
		e.Vars[arg.Arg.Name] = NewRuntimeBoolArray(arg.GetBoolArray())
	case RslArrayT:
		e.Vars[arg.Arg.Name] = NewRuntimeMixedArray(arg.GetMixedArray())
	default:
		e.i.error(arg.Arg.DeclarationToken, fmt.Sprintf("Unsupported arg type, cannot init: %v", argType))
	}
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *Env) SetAndImplyType(varNameToken Token, value interface{}) {
	// todo could make the literal interpreter return LiteralOrArray instead of Go values, making this translation better

	varName := varNameToken.GetLexeme()

	if e.Enclosing != nil {
		_, ok := e.Enclosing.get(varName, varNameToken)
		if ok {
			e.Enclosing.SetAndImplyType(varNameToken, value)
		}
	}

	switch value.(type) {
	case string:
		e.Vars[varName] = NewRuntimeString(value.(string))
	case []string:
		e.Vars[varName] = NewRuntimeStringArray(value.([]string))
	case int64:
		e.Vars[varName] = NewRuntimeInt(value.(int64))
	case []int64:
		e.Vars[varName] = NewRuntimeIntArray(value.([]int64))
	case float64:
		e.Vars[varName] = NewRuntimeFloat(value.(float64))
	case []float64:
		e.Vars[varName] = NewRuntimeFloatArray(value.([]float64))
	case bool:
		e.Vars[varName] = NewRuntimeBool(value.(bool))
	case []bool:
		e.Vars[varName] = NewRuntimeBoolArray(value.([]bool))
	case []interface{}:
		converted := e.recursivelyConvertTypes(varNameToken, value.([]interface{}))
		e.Vars[varName] = NewRuntimeMixedArray(converted.([]interface{}))
	default:
		e.i.error(varNameToken, fmt.Sprintf("Unknown type, cannot set: '%T' %q = %q", value, varName, value))
	}
}

// SetAndExpectType 'value' expected to not be a pointer, should be e.g. string
func (e *Env) SetAndExpectType(varNameToken Token, expectedType *RslTypeEnum, value interface{}) {
	varName := varNameToken.GetLexeme()

	if e.Enclosing != nil {
		_, ok := e.Enclosing.get(varName, varNameToken)
		if ok {
			e.Enclosing.SetAndExpectType(varNameToken, expectedType, value)
		}
	}

	if expectedType != nil {
		expectedTypeVal := *expectedType
		switch expectedTypeVal {
		case RslStringT:
			val, ok := value.(string)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected string: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeString(val)
			}
		case RslStringArrayT:
			switch coerced := value.(type) {
			case []string:
				e.Vars[varName] = NewRuntimeStringArray(coerced)
			case []interface{}:
				strings, ok := AsStringArray(coerced)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected string array: %v", value))
				} else {
					e.Vars[varName] = NewRuntimeStringArray(strings)
				}
			}
		case RslIntT:
			val, ok := value.(int64)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected int: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeInt(val)
			}
		case RslIntArrayT:
			switch coerced := value.(type) {
			case []int64:
				e.Vars[varName] = NewRuntimeIntArray(coerced)
			case []interface{}:
				strings, ok := AsIntArray(coerced)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected int array: %v", value))
				} else {
					e.Vars[varName] = NewRuntimeIntArray(strings)
				}
			}
		case RslFloatT:
			val, ok := value.(float64)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected float: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeFloat(val)
			}
		case RslFloatArrayT:
			switch coerced := value.(type) {
			case []float64:
				e.Vars[varName] = NewRuntimeFloatArray(coerced)
			case []interface{}:
				floats, ok := AsFloatArray(coerced)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected float array: %v", value))
				} else {
					e.Vars[varName] = NewRuntimeFloatArray(floats)
				}
			}
		case RslBoolT:
			val, ok := value.(bool)
			if !ok {
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected bool: %v", value))
			} else {
				e.Vars[varName] = NewRuntimeBool(val)
			}
		case RslBoolArrayT:
			switch coerced := value.(type) {
			case []bool:
				e.Vars[varName] = NewRuntimeBoolArray(coerced)
			case []interface{}:
				bools, ok := AsBoolArray(coerced)
				if !ok {
					e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected bool array: %v", value))
				} else {
					e.Vars[varName] = NewRuntimeBoolArray(bools)
				}
			}
		case RslArrayT:
			switch coerced := value.(type) {
			case []interface{}:
				e.Vars[varName] = NewRuntimeMixedArray(coerced)
			case []string:
				array, _ := AsMixedArray(coerced)
				e.Vars[varName] = NewRuntimeMixedArray(array)
			case []int64:
				array, _ := AsMixedArray(coerced)
				e.Vars[varName] = NewRuntimeMixedArray(array)
			case []float64:
				array, _ := AsMixedArray(coerced)
				e.Vars[varName] = NewRuntimeMixedArray(array)
			case []bool:
				array, _ := AsMixedArray(coerced)
				e.Vars[varName] = NewRuntimeMixedArray(array)
			default:
				e.i.error(varNameToken, fmt.Sprintf("Type mismatch, expected mixed array: %v", value))
			}
		default:
			e.i.error(varNameToken, fmt.Sprintf("Unknown type, cannot set: %v = %v", varName, value))
		}
	}
}

func (e *Env) Exists(name string) bool {
	_, ok := e.get(name, nil)
	return ok
}

func (e *Env) GetByToken(varNameToken Token, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	return e.getOrError(varNameToken.GetLexeme(), varNameToken, acceptableTypes...)
}

func (e *Env) GetByName(varName string, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	return e.getOrError(varName, nil, acceptableTypes...)
}

func (e *Env) AssignJsonField(name Token, path JsonPath) {
	isArray := false
	for _, element := range path.elements {
		if element.token.IsArray || element.token.GetLexeme() == WILDCARD {
			isArray = true
			break
		}
	}
	e.jsonFields[name.GetLexeme()] = JsonFieldVar{
		Name:    name,
		Path:    path,
		IsArray: isArray,
		env:     e,
	}
	if isArray {
		e.SetAndImplyType(name, []interface{}{})
	}
}

func (e *Env) GetJsonField(name Token) JsonFieldVar {
	field, ok := e.jsonFields[name.GetLexeme()]
	if !ok {
		if e.Enclosing != nil {
			return e.Enclosing.GetJsonField(name)
		}
		e.i.error(name, fmt.Sprintf("Undefined json field referenced: %v", name))
	}
	return field
}

func (e *Env) getOrError(varName string, varNameToken Token, acceptableTypes ...RslTypeEnum) RuntimeLiteral {
	val, ok := e.get(varName, varNameToken, acceptableTypes...)
	if !ok {
		e.i.error(varNameToken, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}
	return val
}

func (e *Env) get(varName string, varNameToken Token, acceptableTypes ...RslTypeEnum) (RuntimeLiteral, bool) {
	val, ok := e.Vars[varName]
	if !ok {
		if e.Enclosing != nil {
			return e.Enclosing.get(varName, varNameToken, acceptableTypes...)
		}
		return RuntimeLiteral{}, false
	}

	if len(acceptableTypes) == 0 {
		return val, true
	}

	for _, acceptableType := range acceptableTypes {
		if val.Type == acceptableType {
			return val, true
		}
	}
	e.i.error(varNameToken, fmt.Sprintf("Variable type mismatch: %v, expected: %v", varName, acceptableTypes))
	panic(UNREACHABLE)
}

func (e *Env) recursivelyConvertTypes(token Token, arr interface{}) interface{} {
	switch coerced := arr.(type) {
	// strictly speaking, I don't think ints are necessary to handle, since it seems Go unmarshalls
	// json 'ints' into floats
	case string, int64, float64, bool:
		return coerced
	case int:
		return int64(coerced)
	case []interface{}:
		output := make([]interface{}, len(coerced))
		for i, val := range coerced {
			output[i] = e.recursivelyConvertTypes(token, val)
		}
		return output
	case map[string]interface{}:
		jsonData, err := json.Marshal(coerced)
		if err != nil {
			e.i.error(token, fmt.Sprintf("Error marshalling json: %v", err))
		}
		return string(jsonData)
	case nil:
		// todo this is me wanting to avoid nulls in RSL.... but definitely surprising?
		return "null"
	default:
		e.i.error(token, "Unsupported type in array")
		panic(UNREACHABLE)
	}
}
