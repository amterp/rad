package core

type MainInterpreter struct {
	env      *Env
	literalI *LiteralInterpreter
}

func NewInterpreter() *MainInterpreter {
	return &MainInterpreter{env: NewEnv()}
}

func (i *MainInterpreter) Interpret(statements []Stmt) {
	for _, stmt := range statements {
		stmt.Accept(i)
	}
}

func (i *MainInterpreter) VisitExpressionStmt(expression Expr) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitPrimaryAssignStmt(assign PrimaryAssign) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitFileHeaderStmt(header FileHeader) {
	// ignore from interpretation
	// file header statements will be extracted
	// and processed separately before script runs
}

func (i *MainInterpreter) VisitEmptyStmt(empty Empty) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitArgBlockStmt(block ArgBlock) {
	argBlockInterpreter := NewArgBlockInterpreter(i)
	for _, stmt := range block.ArgStmts {
		stmt.Accept(argBlockInterpreter)
	}
}

func (i *MainInterpreter) VisitRadBlockStmt(block RadBlock) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitJsonPathAssignStmt(assign JsonPathAssign) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *MainInterpreter) VisitExprStmtStmt(stmt ExprStmt) {
	//TODO implement me
	panic("implement me")
}
