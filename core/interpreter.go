package core

type Interpreter struct {
	env *Env
}

func NewInterpreter() *Interpreter {
	return &Interpreter{env: NewEnv()}
}

func (i *Interpreter) Interpret(statements []Stmt) {
	for _, stmt := range statements {
		stmt.Accept(i)
	}
}

func (i *Interpreter) VisitExpressionStmt(expression *Expression) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *Interpreter) VisitPrimaryAssignStmt(assign *PrimaryAssign) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *Interpreter) VisitFileHeaderStmt(header *FileHeader) {
	// ignore from interpretation
	// file header statements will be extracted
	// and processed separately before script runs
}

func (i *Interpreter) VisitEmptyStmt(empty *Empty) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *Interpreter) VisitArgBlockStmt(block *ArgBlock) {
	argBlockInterpreter := NewArgBlockInterpreter(i.env)
	for _, stmt := range block.argStmts {
		stmt.Accept(argBlockInterpreter)
	}
}

func (i *Interpreter) VisitRadBlockStmt(block *RadBlock) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}

func (i *Interpreter) VisitJsonPathAssignStmt(assign *JsonPathAssign) {
	//TODO implement me
	panic(NOT_IMPLEMENTED)
}
