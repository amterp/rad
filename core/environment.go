package core

type Env struct {
	vars map[string]RuntimeLiteral
	//args       map[string]ArgValue
	//jsonFields map[string]*JsonField
	enclosing *Env
}

func NewEnv() *Env {
	return &Env{
		vars: make(map[string]RuntimeLiteral),
		//args:       make(map[string]ArgValue),
		//jsonFields: make(map[string]*JsonField),
		enclosing: nil,
	}
}
