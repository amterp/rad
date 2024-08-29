package interpreters

import "rad/core"

type ArgBlockInterpreter struct {
	main *MainInterpreter
}

func NewArgBlockInterpreter(m *MainInterpreter) *ArgBlockInterpreter {
	return &ArgBlockInterpreter{m}
}

func (i *ArgBlockInterpreter) VisitArgDeclarationArgStmt(declaration *core.ArgDeclaration) {
	arg := core.FromArgDecl(i.main.literalI, declaration)
	i.main.env.Args[arg.Name] = *arg
}
