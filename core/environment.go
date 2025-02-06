package core

import (
	"fmt"
)

type OldEnv struct {
	i          *MainInterpreter
	Vars       map[string]interface{} // values are NOT pointers, they're the actual value
	jsonFields map[string]JsonFieldVar
	Enclosing  *OldEnv
}

func NewOldEnv(i *MainInterpreter) *OldEnv {
	return &OldEnv{
		i:          i,
		Vars:       make(map[string]interface{}),
		jsonFields: make(map[string]JsonFieldVar),
		Enclosing:  nil,
	}
}

func (e *OldEnv) NewChildEnv() OldEnv {
	return OldEnv{
		i:          e.i,
		Vars:       make(map[string]interface{}),
		jsonFields: make(map[string]JsonFieldVar),
		Enclosing:  e,
	}
}

func (e *OldEnv) InitArg(arg RslArg) {
	//if arg.IsNull {
	//	return // todo re-assess this optional thing
	//}

	switch coerced := arg.(type) {
	case *BoolRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *BoolArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	case *StringRslArg:
		e.Vars[coerced.Identifier] = NewRslString(coerced.Value)
	case *StringArrRslArg:
		converted := make([]interface{}, len(coerced.Value))
		for j, v := range coerced.Value {
			converted[j] = NewRslString(v)
		}
		e.Vars[coerced.Identifier] = converted
	case *IntRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *IntArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	case *FloatRslArg:
		e.Vars[coerced.Identifier] = coerced.Value
	case *FloatArrRslArg:
		e.Vars[coerced.Identifier] = convertToInterfaceArr(coerced.Value)
	default:
		e.i.errorNode(arg.GetNode(), fmt.Sprintf("Unsupported arg type, cannot init: %T", arg))
	}
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *OldEnv) SetAndImplyType(varNameToken Token, value interface{}) {
	e.SetAndImplyTypeWithToken(varNameToken, varNameToken.GetLexeme(), value)
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *OldEnv) SetAndImplyTypeWithToken(token Token, varName string, value interface{}) {
	e.setAndImplyTypeWithToken(token, varName, value, true)
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *OldEnv) SetAndImplyTypeWithTokenIgnoringEnclosing(token Token, varName string, value interface{}) {
	e.setAndImplyTypeWithToken(token, varName, value, false)
}

func (e *OldEnv) Exists(name string) bool {
	_, ok := e.get(nil, name)
	return ok
}

func (e *OldEnv) Delete(name string) {
	delete(e.Vars, name)
}

func (e *OldEnv) GetByToken(varNameToken Token, acceptableTypes ...RslTypeEnum) interface{} {
	return e.getOrError(varNameToken.GetLexeme(), varNameToken, acceptableTypes...)
}

func (e *OldEnv) GetByName(token Token, varName string, acceptableTypes ...RslTypeEnum) interface{} {
	return e.getOrError(varName, token, acceptableTypes...)
}

func (e *OldEnv) AssignJsonField(name Token, path JsonPath) {
	e.jsonFields[name.GetLexeme()] = JsonFieldVar{
		Name: name,
		Path: path,
	}
}

func (e *OldEnv) GetJsonField(nameToken Token) JsonFieldVar {
	return e.GetJsonFieldWithToken(nameToken, nameToken.GetLexeme())
}

func (e *OldEnv) GetJsonFieldWithToken(token Token, name string) JsonFieldVar {
	field, ok := e.jsonFields[name]
	if !ok {
		if e.Enclosing != nil {
			return e.Enclosing.GetJsonFieldWithToken(token, name)
		}
		e.i.error(token, fmt.Sprintf("Undefined json field referenced: %v", name))
	}
	return field
}

func (e *OldEnv) PrintShellExports() {
	keys := SortedKeys(e.Vars)
	for _, varName := range keys {
		val := e.Vars[varName]
		// todo handle different data types specifically
		// todo avoid *dangerous* exports like PATH!!
		RP.PrintForShellEval(fmt.Sprintf("export %s=\"%v\"\n", varName, ToPrintable(val)))
	}
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *OldEnv) setAndImplyTypeWithToken(token Token, varName string, value interface{}, modifyEnclosing bool) {
	// todo could make the literal interpreter return LiteralOrArray instead of Go values, making this translation better

	if modifyEnclosing && e.Enclosing != nil {
		_, ok := e.Enclosing.get(token, varName)
		if ok {
			e.Enclosing.SetAndImplyType(token, value)
			return
		}
	}

	switch coerced := value.(type) {
	case RslString, int64, float64, bool:
		e.Vars[varName] = coerced
	case string:
		e.Vars[varName] = NewRslString(coerced)
	case []interface{}:
		converted := ConvertToNativeTypes(e.i, token, coerced)
		e.Vars[varName] = converted.([]interface{})
	case RslMapOld:
		converted := ConvertToNativeTypes(e.i, token, coerced)
		e.Vars[varName] = converted.(RslMapOld)
	case map[string]interface{}:
		converted := ConvertToNativeTypes(e.i, token, coerced)
		e.Vars[varName] = converted.(RslMapOld)
	default:
		e.i.error(token, fmt.Sprintf("Unknown type, cannot set: '%T' %q = %q", value, varName, value))
	}
}

func (e *OldEnv) getOrError(varName string, token Token, acceptableTypes ...RslTypeEnum) interface{} {
	val, ok := e.get(token, varName, acceptableTypes...)
	if !ok {
		e.i.error(token, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}
	return val
}

func (e *OldEnv) get(token Token, varName string, acceptableTypes ...RslTypeEnum) (interface{}, bool) {
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
