package core

import "fmt"

type MainInterpreter struct {
	env        *Env
	LiteralI   *LiteralInterpreter
	argBlockI  *ArgBlockInterpreter
	radBlockI  *RadBlockInterpreter
	switchI    *SwitchInterpreter
	statements []Stmt

	breaking   bool
	continuing bool
}

func NewInterpreter(statements []Stmt) *MainInterpreter {
	i := &MainInterpreter{
		statements: statements,
	}
	i.LiteralI = NewLiteralInterpreter(i)
	i.argBlockI = NewArgBlockInterpreter(i)
	i.radBlockI = NewRadBlockInterpreter(i)
	i.switchI = NewSwitchInterpreter(i)
	i.env = NewEnv(i)
	return i
}

func (i *MainInterpreter) InitArgs(args []*CobraArg) {
	for _, arg := range args {
		i.env.InitArg(*arg)
	}
}

func (i *MainInterpreter) Run() {
	for _, stmt := range i.statements {
		stmt.Accept(i)
	}
}

func (i *MainInterpreter) VisitExprLoaExpr(loa ExprLoa) interface{} {
	return loa.Value.Accept(i.LiteralI)
}

func (i *MainInterpreter) VisitArrayExprExpr(expr ArrayExpr) interface{} {
	var values []interface{} // e.g. []string
	for _, v := range expr.Values {
		val := v.Accept(i)
		values = append(values, val)
	}
	return values
}

func (i *MainInterpreter) VisitArrayAccessExpr(access ArrayAccess) interface{} {
	array := access.Array.Accept(i)
	index := access.Index.Accept(i)

	switch coerced := array.(type) {
	case []string:
		return coerced[index.(int64)]
	case []int64:
		return coerced[index.(int64)]
	case []float64:
		return coerced[index.(int64)]
	case []bool:
		return coerced[index.(int64)]
	case []interface{}:
		return coerced[index.(int64)]
	default:
		i.error(access.OpenBracketToken, "Bug! Should've failed earlier")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) VisitFunctionCallExpr(call FunctionCall) interface{} {
	var args []interface{}
	for _, v := range call.Args {
		val := v.Accept(i)
		args = append(args, val)
	}
	return RunRslNonVoidFunction(i, call.Function, call.NumExpectedReturnValues, args)
}

func (i *MainInterpreter) VisitFunctionStmtStmt(functionStmt FunctionStmt) {
	var values []interface{}
	for _, v := range functionStmt.Call.Args {
		val := v.Accept(i)
		values = append(values, val)
	}
	RunRslFunction(i, functionStmt.Call.Function, values)
}

func (i *MainInterpreter) VisitVariableExpr(variable Variable) interface{} {
	return i.env.GetByToken(variable.Name).value
}

func (i *MainInterpreter) VisitLogicalExpr(logical Logical) interface{} {
	left := logical.Left.Accept(i).(bool)
	right := logical.Right.Accept(i).(bool)

	switch logical.Operator.GetType() {
	case AND:
		return left && right
	case OR:
		return left || right
	default:
		i.error(logical.Operator, "Bug! Non-and/or logical operator should've not passed the parser")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) VisitGroupingExpr(grouping Grouping) interface{} {
	return grouping.Value.Accept(i)
}

func (i *MainInterpreter) VisitUnaryExpr(unary Unary) interface{} {
	value := unary.Right.Accept(i)

	valBool, ok := value.(bool)
	if ok {
		switch unary.Operator.GetType() {
		case EXCLAMATION:
			return !valBool
		default:
			i.error(unary.Operator, "Invalid logical operator, only 'not' is allowed")
		}
	}

	var multiplier int64
	switch unary.Operator.GetType() {
	case MINUS:
		multiplier = -1
	case PLUS:
		multiplier = 1
	default:
		i.error(unary.Operator, "Invalid number unary operation, only + and - are allowed")
	}

	valInt, ok := value.(int64)
	if ok {
		return valInt * multiplier
	}

	valFloat, ok := value.(float64)
	if ok {
		return valFloat * float64(multiplier)
	}

	i.error(unary.Operator, "Invalid unary operands, only bool, float, or int is allowed")
	panic(UNREACHABLE)
}

func (i *MainInterpreter) VisitExpressionStmt(expression Expr) {
	expression.Accept(i)
}

func (i *MainInterpreter) VisitPrimaryAssignStmt(assign PrimaryAssign) {
	value := assign.Initializer.Accept(i)

	if len(assign.Identifiers) == 1 {
		value = []interface{}{value}
	}

	switch values := value.(type) {
	case []interface{}:
		handleMultiAssignment(i, values, assign.Identifiers, assign.VarTypes) // todo does this handle all arrays?
	default:
		i.error(assign.Identifiers[0], "Expected multiple values, got one")
	}
}

func handleMultiAssignment[T any](i *MainInterpreter, values []T, identifiers []Token, varTypes []*RslType) {
	if len(values) != len(identifiers) {
		i.error(identifiers[0], fmt.Sprintf("Expected %d values, got %d", len(identifiers), len(values)))
	}
	for idx, val := range values {
		if varTypes[idx] != nil {
			varType := &(*varTypes[idx]).Type
			i.env.SetAndExpectType(identifiers[idx], varType, val)
		} else {
			i.env.SetAndImplyType(identifiers[idx], val)
		}
	}
}

func (i *MainInterpreter) VisitFileHeaderStmt(header FileHeader) {
	// ignore from interpretation
	// file header statements will be extracted
	// and processed separately before script runs
}

func (i *MainInterpreter) VisitEmptyStmt(Empty) {
	// nothing to do
}

func (i *MainInterpreter) VisitArgBlockStmt(block ArgBlock) {
	i.argBlockI.Run(block)
}

func (i *MainInterpreter) VisitRadBlockStmt(block RadBlock) {
	i.radBlockI.Run(block)
}

func (i *MainInterpreter) VisitJsonPathAssignStmt(assign JsonPathAssign) {
	i.env.AssignJsonField(assign.Identifier, assign.Path)
}

func (i *MainInterpreter) VisitExprStmtStmt(stmt ExprStmt) {
	stmt.Expression.Accept(i)
}

func (i *MainInterpreter) VisitSwitchBlockStmtStmt(block SwitchBlockStmt) {
	i.switchI.RunBlock(block.Block)
}

func (i *MainInterpreter) VisitSwitchAssignmentStmt(assignment SwitchAssignment) {
	i.switchI.RunAssignment(assignment)
}

func (i *MainInterpreter) VisitBlockStmt(block Block) {
	i.runWithChildEnv(func() {
		for _, stmt := range block.Stmts {
			stmt.Accept(i)
			if i.breaking {
				break
			}
			if i.continuing {
				break
			}
		}
	})
}

func (i *MainInterpreter) VisitIfStmtStmt(stmt IfStmt) {
	cases := stmt.Cases
	for _, c := range cases {
		conditionResult := c.Condition.Accept(i)
		bval, ok := conditionResult.(bool)
		if !ok {
			i.error(c.IfToken, "If condition must resolve to a bool")
		}
		if bval {
			c.Body.Accept(i)
			return
		}
	}
	if stmt.ElseBlock != nil {
		stmt.ElseBlock.Accept(i)
	}
}

func (i *MainInterpreter) VisitIfCaseStmt(ifCase IfCase) {
	RP.RadErrorExit("Bug! IfCase should not be visited directly\n")
}

func (i *MainInterpreter) VisitForStmtStmt(stmt ForStmt) {
	rangeValue := stmt.Range.Accept(i)
	var valIdentifier Token
	var idxIdentifier *Token
	if stmt.Identifier2 != nil {
		idxIdentifier = &stmt.Identifier1
		valIdentifier = *stmt.Identifier2
	} else {
		valIdentifier = stmt.Identifier1
	}

	switch rangeValue.(type) {
	case []string:
		arr := rangeValue.([]string)
		i.runWithChildEnv(runForLoop(i, stmt, arr, idxIdentifier, valIdentifier))
	case []int64:
		arr := rangeValue.([]int64)
		i.runWithChildEnv(runForLoop(i, stmt, arr, idxIdentifier, valIdentifier))
	case []float64:
		arr := rangeValue.([]float64)
		i.runWithChildEnv(runForLoop(i, stmt, arr, idxIdentifier, valIdentifier))
	case []bool:
		arr := rangeValue.([]bool)
		i.runWithChildEnv(runForLoop(i, stmt, arr, idxIdentifier, valIdentifier))
	case []interface{}:
		arr := rangeValue.([]interface{})
		i.runWithChildEnv(runForLoop(i, stmt, arr, idxIdentifier, valIdentifier))
	default:
		i.error(stmt.ForToken, "For loop range must be an array")
	}
}

func runForLoop[T any](i *MainInterpreter, stmt ForStmt, arr []T, idxIdentifier *Token, valIdentifier Token) func() {
	return func() {
		for idx, val := range arr {
			i.env.SetAndImplyType(valIdentifier, val)
			if idxIdentifier != nil {
				i.env.SetAndImplyType(*idxIdentifier, int64(idx))
			}
			stmt.Body.Accept(i)
			if i.breaking {
				i.breaking = false
				break
			}
			if i.continuing {
				i.continuing = false
				continue
			}
		}
	}
}

func (i *MainInterpreter) VisitBreakStmtStmt(stmt BreakStmt) {
	i.breaking = true
}

func (i *MainInterpreter) VisitContinueStmtStmt(stmt ContinueStmt) {
	i.continuing = true
}

func (i *MainInterpreter) error(token Token, message string) {
	RP.TokenErrorExit(token, message+"\n")
}

func (i *MainInterpreter) runWithChildEnv(runnable func()) {
	originalEnv := i.env
	env := originalEnv.NewChildEnv()
	i.env = &env
	runnable()
	i.env = originalEnv
}
