package core

import ts "github.com/tree-sitter/go-tree-sitter"

type Env struct {
	i             *Interpreter
	Enclosing     *Env
	Vars          map[string]RslValue
	JsonFieldVars map[string]*JsonFieldVar // not pointer?
}

func NewEnv(i *Interpreter) *Env {
	return &Env{
		i:             i,
		Enclosing:     nil,
		Vars:          make(map[string]RslValue),
		JsonFieldVars: make(map[string]*JsonFieldVar),
	}
}

func (e *Env) NewChildEnv() Env {
	return Env{
		i:             e.i,
		Enclosing:     e,
		Vars:          make(map[string]RslValue),
		JsonFieldVars: make(map[string]*JsonFieldVar),
	}
}

func (e *Env) SetVar(name string, v RslValue) {
	e.setVar(name, v, false)
}

func (e *Env) SetVarIgnoringEnclosing(name string, v RslValue) {
	e.setVar(name, v, true)
}

func (e *Env) GetVar(name string) (RslValue, bool) {
	if val, exists := e.Vars[name]; exists {
		return val, true
	}
	if e.Enclosing != nil {
		return e.Enclosing.GetVar(name)
	}
	return RslValue{}, false
}

func (e *Env) GetVarElseBug(i *Interpreter, node *ts.Node, name string) RslValue {
	if val, exists := e.GetVar(name); exists {
		return val
	}
	i.errorf(node, "Bug! Expected variable but didn't find: "+name)
	panic(UNREACHABLE)
}

func (e *Env) SetJsonFieldVar(jsonFieldVar *JsonFieldVar) {
	e.JsonFieldVars[jsonFieldVar.Name] = jsonFieldVar
	// define empty list for json field
	e.SetVar(jsonFieldVar.Name, newRslValue(e.i, jsonFieldVar.Node, NewRslList()))
}

func (e *Env) GetJsonFieldVar(name string) (*JsonFieldVar, bool) {
	if val, exists := e.JsonFieldVars[name]; exists {
		return val, true
	}
	if e.Enclosing != nil {
		return e.Enclosing.GetJsonFieldVar(name)
	}
	return nil, false
}

func (e *Env) setVar(name string, v RslValue, ignoreEnclosing bool) {
	if !ignoreEnclosing && e.Enclosing != nil {
		if _, exists := e.Enclosing.GetVar(name); exists {
			e.Enclosing.SetVar(name, v)
			return
		}
	}

	if v == NIL_SENTINAL {
		delete(e.Vars, name)
	} else {
		e.Vars[name] = v
	}
}
