package interpreters

import "rad/core"

type MainInterpreter struct {
	env      *core.Env
	literalI *LiteralInterpreter
}

func NewInterpreter() *MainInterpreter {
	return &MainInterpreter{env: core.NewEnv()}
}

func (i *MainInterpreter) Interpret(statements []core.Stmt) {
	for _, stmt := range statements {
		stmt.Accept(i)
	}
}

func (i *MainInterpreter) VisitExpressionStmt(expression *core.Expression) {
	//TODO implement me
	panic(core.NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitPrimaryAssignStmt(assign *core.PrimaryAssign) {
	//TODO implement me
	panic(core.NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitFileHeaderStmt(header *core.FileHeader) {
	// ignore from interpretation
	// file header statements will be extracted
	// and processed separately before script runs
}

func (i *MainInterpreter) VisitEmptyStmt(empty *core.Empty) {
	//TODO implement me
	panic(core.NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitArgBlockStmt(block *core.ArgBlock) {
	argBlockInterpreter := NewArgBlockInterpreter(i)
	for _, stmt := range block.ArgStmts {
		stmt.Accept(argBlockInterpreter)
	}
}

func (i *MainInterpreter) VisitRadBlockStmt(block *core.RadBlock) {
	//TODO implement me
	panic(core.NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitJsonPathAssignStmt(assign *core.JsonPathAssign) {
	//TODO implement me
	panic(core.NOT_IMPLEMENTED)
}
