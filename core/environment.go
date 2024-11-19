package core

import (
	"fmt"
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

func (e *Env) InitArg(arg RslArg) {
	//if arg.IsNull {
	//	return // todo re-assess this optional thing
	//}

	switch coerced := arg.(type) {
	case *BoolRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *BoolArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	case *StringRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *StringArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	case *IntRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *IntArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	case *FloatRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *FloatArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	default:
		e.i.error(arg.GetToken(), fmt.Sprintf("Unsupported arg type, cannot init: %T", arg))
	}
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *Env) SetAndImplyType(varNameToken Token, value interface{}) {
	e.SetAndImplyTypeWithToken(varNameToken, varNameToken.GetLexeme(), value)
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *Env) SetAndImplyTypeWithToken(token Token, varName string, value interface{}) {
	// todo could make the literal interpreter return LiteralOrArray instead of Go values, making this translation better

	if e.Enclosing != nil {
		_, ok := e.Enclosing.get(token, varName)
		if ok {
			e.Enclosing.SetAndImplyType(token, value)
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
		converted := ConvertToNativeTypes(e.i, token, coerced)
		e.Vars[varName] = converted.([]interface{})
	case RslMap:
		converted := ConvertToNativeTypes(e.i, token, coerced)
		e.Vars[varName] = converted.(RslMap)
	case map[string]interface{}:
		converted := ConvertToNativeTypes(e.i, token, coerced)
		e.Vars[varName] = converted.(RslMap)
	default:
		e.i.error(token, fmt.Sprintf("Unknown type, cannot set: '%T' %q = %q", value, varName, value))
	}
}

func (e *Env) Exists(name string) bool {
	_, ok := e.get(nil, name)
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
	e.jsonFields[name.GetLexeme()] = JsonFieldVar{
		Name: name,
		Path: path,
	}
}

func (e *Env) GetJsonField(nameToken Token) JsonFieldVar {
	return e.GetJsonFieldWithToken(nameToken, nameToken.GetLexeme())
}

func (e *Env) GetJsonFieldWithToken(token Token, name string) JsonFieldVar {
	field, ok := e.jsonFields[name]
	if !ok {
		if e.Enclosing != nil {
			return e.Enclosing.GetJsonFieldWithToken(token, name)
		}
		e.i.error(token, fmt.Sprintf("Undefined json field referenced: %v", name))
	}
	return field
}

func (e *Env) PrintShellExports() {
	keys := SortedKeys(e.Vars)
	for _, varName := range keys {
		val := e.Vars[varName]
		// todo handle different data types specifically
		// todo avoid *dangerous* exports like PATH!!
		RP.PrintForShellEval(fmt.Sprintf("export %s=\"%v\"\n", varName, val))
	}
}

func (e *Env) getOrError(varName string, token Token, acceptableTypes ...RslTypeEnum) interface{} {
	val, ok := e.get(token, varName, acceptableTypes...)
	if !ok {
		e.i.error(token, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}
	return val
}

func (e *Env) get(token Token, varName string, acceptableTypes ...RslTypeEnum) (interface{}, bool) {
	val, ok := e.Vars[varName]
	if !ok {
		if e.Enclosing != nil {
			return e.Enclosing.get(token, varName, acceptableTypes...)
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

func convertToInterfaceArr[T any](i []T) []interface{} {
	converted := make([]interface{}, len(i))
	for j, v := range i {
		converted[j] = v
	}
	return converted
}
