package core

import (
	"fmt"
)

type OldEnv struct {
	i          *MainInterpreter
	Vars       map[string]interface{} // values are NOT pointers, they're the actual value
	jsonFields map[string]JsonFieldVarOld
	Enclosing  *OldEnv
}

func NewOldEnv(i *MainInterpreter) *OldEnv {
	return &OldEnv{
		i:          i,
		Vars:       make(map[string]interface{}),
		jsonFields: make(map[string]JsonFieldVarOld),
		Enclosing:  nil,
	}
}

func (e *OldEnv) NewChildEnv() OldEnv {
	return OldEnv{
		i:          e.i,
		Vars:       make(map[string]interface{}),
		jsonFields: make(map[string]JsonFieldVarOld),
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

func (e *OldEnv) AssignJsonField(name Token, path JsonPathOld) {
	e.jsonFields[name.GetLexeme()] = JsonFieldVarOld{
		Name: name,
		Path: path,
	}
}

func (e *OldEnv) GetJsonField(nameToken Token) JsonFieldVarOld {
	return e.GetJsonFieldWithToken(nameToken, nameToken.GetLexeme())
}

func (e *OldEnv) GetJsonFieldWithToken(token Token, name string) JsonFieldVarOld {
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
	// DELETE
}

// SetAndImplyType 'value' expected to not be a pointer, should be e.g. string
func (e *OldEnv) setAndImplyTypeWithToken(token Token, varName string, value interface{}, modifyEnclosing bool) {
	// DELETE
}

func (e *OldEnv) getOrError(varName string, token Token, acceptableTypes ...RslTypeEnum) interface{} {
	val, ok := e.get(token, varName, acceptableTypes...)
	if !ok {
		e.i.error(token, fmt.Sprintf("Undefined variable referenced: %v", varName))
	}
	return val
}

func (e *OldEnv) get(token Token, varName string, acceptableTypes ...RslTypeEnum) (interface{}, bool) {
	// DELETE
	return nil, false
}

func convertToInterfaceArr[T any](i []T) []interface{} {
	converted := make([]interface{}, len(i))
	for j, v := range i {
		converted[j] = v
	}
	return converted
}
