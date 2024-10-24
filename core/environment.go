package core

import (
	"fmt"
	"sort"
)

type Env struct {
	i          *MainInterpreter
	Vars       map[string]interface{} // values are NOT pointers, they're the actual value
	jsonFields map[string]JsonFieldVar
	Enclosing  *Env
}

func NewEnv(i *MainInterpreter) *Env {
	return &Env{
		i:          i,
		Vars:       make(map[string]interface{}),
		jsonFields: make(map[string]JsonFieldVar),
		Enclosing:  nil,
	}
}

func (e *Env) NewChildEnv() Env {
	return Env{
		i:          e.i,
		Vars:       make(map[string]interface{}),
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
	case ArgStringT:
		e.Vars[arg.Arg.Name] = arg.GetString()
	case ArgIntT:
		e.Vars[arg.Arg.Name] = arg.GetInt()
	case ArgFloatT:
		e.Vars[arg.Arg.Name] = arg.GetFloat()
	case ArgBoolT:
		e.Vars[arg.Arg.Name] = arg.GetBool()
	case ArgStringArrayT, ArgIntArrayT, ArgFloatArrayT, ArgBoolArrayT, ArgMixedArrayT:
		e.Vars[arg.Arg.Name] = arg.GetArray()
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

	switch coerced := value.(type) {
	case string:
		e.Vars[varName] = coerced
	case int64:
		e.Vars[varName] = coerced
	case float64:
		e.Vars[varName] = coerced
	case bool:
		e.Vars[varName] = coerced
	case []interface{}:
		converted := e.recursivelyConvertTypes(varNameToken, coerced)
		e.Vars[varName] = converted.([]interface{})
	case RslMap:
		converted := e.recursivelyConvertTypes(varNameToken, coerced)
		e.Vars[varName] = converted.(RslMap)
	case map[string]interface{}:
		converted := e.recursivelyConvertTypes(varNameToken, coerced)
		e.Vars[varName] = converted.(RslMap)
	default:
		e.i.error(varNameToken, fmt.Sprintf("Unknown type, cannot set: '%T' %q = %q", value, varName, value))
	}
}

func (e *Env) Exists(name string) bool {
	_, ok := e.get(name, nil)
	return ok
}

func (e *Env) Delete(name string) {
	delete(e.Vars, name)
}

func (e *Env) GetByToken(varNameToken Token, acceptableTypes ...RslTypeEnum) interface{} {
	return e.getOrError(varNameToken.GetLexeme(), varNameToken, acceptableTypes...)
}

func (e *Env) GetByName(token Token, varName string, acceptableTypes ...RslTypeEnum) interface{} {
	return e.getOrError(varName, token, acceptableTypes...)
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

func (e *Env) getOrError(varName string, token Token, acceptableTypes ...RslTypeEnum) interface{} {
	val, ok := e.get(varName, token, acceptableTypes...)
	if !ok {
		e.i.error(token, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}
	return val
}

func (e *Env) get(varName string, token Token, acceptableTypes ...RslTypeEnum) (interface{}, bool) {
	val, ok := e.Vars[varName]
	if !ok {
		if e.Enclosing != nil {
			return e.Enclosing.get(varName, token, acceptableTypes...)
		}
		return nil, false
	}

	if len(acceptableTypes) == 0 {
		return val, true
	}

	for _, acceptableType := range acceptableTypes {
		if acceptableType.MatchesValue(val) {
			return val, true
		}
	}
	e.i.error(token, fmt.Sprintf("Variable type mismatch: %v, expected: %v", varName, acceptableTypes))
	panic(UNREACHABLE)
}

// todo since supporting maps, I think I can get rid of this
// it was originally implemented because we might capture JSON as a list of unhandled types, but
// now we should be able to capture json and convert it entirely to native RSL types up front
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
		m := NewRslMap()
		sortedKeys := make([]string, 0, len(coerced))
		for k := range coerced {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		for _, key := range sortedKeys {
			m.Set(key, e.recursivelyConvertTypes(token, coerced[key]))
		}
		return *m
	case RslMap:
		return coerced
	case nil:
		return nil
	default:
		e.i.error(token, fmt.Sprintf("Unhandled type in array: %T", arr))
		panic(UNREACHABLE)
	}
}
