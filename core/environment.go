package core

type Env struct {
	Vars map[string]RuntimeLiteral
	Args map[string]ScriptArg
	//jsonFields map[string]*JsonField
	Enclosing *Env
}

func NewEnv() *Env {
	return &Env{
		Vars: make(map[string]RuntimeLiteral),
		Args: make(map[string]ScriptArg),
		//jsonFields: make(map[string]*JsonField),
		Enclosing: nil,
	}
}
