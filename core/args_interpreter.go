package core

type ArgBlockInterpreter struct {
	env *Env
}

func NewArgBlockInterpreter(env *Env) *ArgBlockInterpreter {
	return &ArgBlockInterpreter{env: env}
}

func (i *ArgBlockInterpreter) VisitArgDeclarationArgStmt(declaration *ArgDeclaration) {
	//TODO activate args in env
	panic("implement me")
}
