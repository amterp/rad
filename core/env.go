package core

import (
	"fmt"
	com "rad/core/common"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type Env struct {
	i             *Interpreter
	Enclosing     *Env
	Vars          map[string]RadValue
	JsonFieldVars map[string]*JsonFieldVar // not pointer?
}

func NewEnv(i *Interpreter) *Env {
	return &Env{
		i:             i,
		Enclosing:     nil,
		Vars:          make(map[string]RadValue),
		JsonFieldVars: make(map[string]*JsonFieldVar),
	}
}

func (e *Env) NewChildEnv() Env {
	return Env{
		i:             e.i,
		Enclosing:     e,
		Vars:          make(map[string]RadValue),
		JsonFieldVars: make(map[string]*JsonFieldVar),
	}
}

func (e *Env) SetVar(name string, v RadValue) {
	e.SetVarUpdatingEnclosing(name, v, false)
}

func (e *Env) SetVarUpdatingEnclosing(name string, v RadValue, updateEnclosing bool) {
	e.setVar(name, v, updateEnclosing)
}

func (e *Env) GetVar(name string) (RadValue, bool) {
	if val, exists := e.Vars[name]; exists {
		return val, true
	}
	if e.Enclosing != nil {
		return e.Enclosing.GetVar(name)
	}
	return RadValue{}, false
}

func (e *Env) GetVarElseBug(i *Interpreter, node *ts.Node, name string) RadValue {
	if val, exists := e.GetVar(name); exists {
		return val
	}
	i.errorf(node, "Bug! Expected variable but didn't find: "+name)
	panic(UNREACHABLE)
}

func (e *Env) SetJsonFieldVar(jsonFieldVar *JsonFieldVar) {
	e.JsonFieldVars[jsonFieldVar.Name] = jsonFieldVar
	// define empty list for json field
	e.SetVar(jsonFieldVar.Name, newRadValue(e.i, jsonFieldVar.Node, NewRadList()))
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

func (e *Env) setVar(name string, v RadValue, updateEnclosing bool) {
	if e.Enclosing != nil && updateEnclosing {
		if _, exists := e.Enclosing.Vars[name]; exists {
			e.Enclosing.setVar(name, v, updateEnclosing)
			return
		}
	}

	if v == VOID_SENTINEL {
		delete(e.Vars, name)
	} else {
		e.Vars[name] = v
	}
}

// todo avoid *dangerous* exports like PATH!!
func (e *Env) PrintShellExports() {
	keys := com.SortedKeys(e.Vars)

	printFunc := func(varName, value string) {
		RP.PrintForShellEval(fmt.Sprintf("%s=%s\n", varName, value))
	}

	for _, varName := range keys {
		val := e.Vars[varName]
		// type visitor takes a *ts.Node which isn't super applicable here...
		switch coerced := val.Val.(type) {
		case RadString, int64, float64, bool:
			printFunc(varName, ToPrintable(val))
		case *RadList:
			printFunc(varName, "("+strings.Join(coerced.AsStringList(true), " ")+")")
		case *RadMap:
			// todo can do some stuff with declare -A ?
			printFunc(varName, "'"+coerced.ToString()+"'")
		case RadFn:
			// skip, doesn't make sense
		case RadNull:
			// skip, implies undefined
		default:
			RP.RadErrorExit(fmt.Sprintf("Bug! Unhandled type for shell export: %T", val.Val))
		}
	}
}
