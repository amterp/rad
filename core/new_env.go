package core

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

func (e *Env) GetVar(name string) (RslValue, bool) {
	if val, exists := e.Vars[name]; exists {
		return val, true
	}
	if e.Enclosing != nil {
		return e.Enclosing.GetVar(name)
	}
	return RslValue{}, false
}

func (e *Env) SetVar(name string, v RslValue) {
	if v == NIL_SENTINAL {
		delete(e.Vars, name)
	} else {
		e.Vars[name] = v
	}
}

func (e *Env) SetJsonFieldVar(jsonFieldVar *JsonFieldVar) {
	e.JsonFieldVars[jsonFieldVar.Name] = jsonFieldVar
	// define empty list for json field
	e.SetVar(jsonFieldVar.Name, newRslValue(e.i, jsonFieldVar.Node, NewRslList()))
}
