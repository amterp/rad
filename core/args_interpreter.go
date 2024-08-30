package core

type ArgBlockInterpreter struct {
	main *MainInterpreter
}

func NewArgBlockInterpreter(m *MainInterpreter) *ArgBlockInterpreter {
	return &ArgBlockInterpreter{m}
}

func (i ArgBlockInterpreter) VisitArgDeclarationArgStmt(declaration ArgDeclaration) {
	// todo activate the the arg in the arg block, ready for execution
}
