package core

type ArgBlockInterpreter struct {
	i *MainInterpreter
}

func NewArgBlockInterpreter(i *MainInterpreter) *ArgBlockInterpreter {
	return &ArgBlockInterpreter{i}
}

func (a ArgBlockInterpreter) VisitArgDeclarationArgStmt(declaration ArgDeclaration) {
	// arg declarations already initialized in env, nothing to do on visit here, just pass
}

func (a ArgBlockInterpreter) VisitArgEnumArgStmt(enum ArgEnum) {
	// arg enum constraints are applied prior to running the script, nothing to do on visit here, just pass
}

func (a ArgBlockInterpreter) Run(block ArgBlock) {
	for _, stmt := range block.Stmts {
		stmt.Accept(a)
	}
	// todo would need to build up some 'state' and then 'execute' including constraints, etc
}
