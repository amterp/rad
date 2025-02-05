package core

type Env struct {
	i         *Interpreter
	Enclosing *Env
	Vars      map[string]RslValue
}

func NewEnv(i *Interpreter) *Env {
	return &Env{
		i:         i,
		Enclosing: nil,
		Vars:      make(map[string]RslValue),
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
	e.Vars[name] = v
}
