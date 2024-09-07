package core

import (
	"fmt"
)

type MainInterpreter struct {
	env        *Env
	LiteralI   *LiteralInterpreter
	argBlockI  *ArgBlockInterpreter
	radBlockI  *RadBlockInterpreter
	switchI    *SwitchInterpreter
	statements []Stmt
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
	literal := i.env.GetByToken(access.Array, RslStringArray, RslIntArray, RslFloatArray)
	index := access.Index.Accept(i)

	switch literal.Type {
	case RslStringArray:
		arr := literal.GetStringArray()
		return arr[index.(int)]
	case RslIntArray:
		arr := literal.GetIntArray()
		return arr[index.(int)]
	case RslFloatArray:
		arr := literal.GetFloatArray()
		return arr[index.(int)]
	default:
		i.error(access.Array, "Bug! Should've failed earlier")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) VisitFunctionCallExpr(call FunctionCall) interface{} {
	var values []interface{}
	for _, v := range call.Args {
		val := v.Accept(i)
		values = append(values, val)
	}
	return RunRslNonVoidFunction(i, call.Function, values)
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
		case NOT:
			return !valBool
		default:
			i.error(unary.Operator, "Invalid logical operator, only 'not' is allowed")
		}
	}

	var multiplier int
	switch unary.Operator.GetType() {
	case MINUS:
		multiplier = -1
	case PLUS:
		multiplier = 1
	default:
		i.error(unary.Operator, "Invalid number unary operation, only + and - are allowed")
	}

	valInt, ok := value.(int)
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
	if assign.VarType != nil {
		varType := &(*assign.VarType).Type
		i.env.SetAndExpectType(assign.Name, varType, value)
	} else {
		i.env.SetAndImplyType(assign.Name, value)
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
	panic("Bug! IfCase should not be visited directly")
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

	// very duplicated :( but generic non-duped doesn't look super simple?
	switch rangeValue.(type) {
	case []string:
		arr := rangeValue.([]string)
		for idx, val := range arr {
			i.runWithChildEnv(func() {
				i.env.SetAndImplyType(valIdentifier, val)
				if idxIdentifier != nil {
					i.env.SetAndImplyType(*idxIdentifier, idx)
				}
				for _, s := range stmt.Body.Stmts {
					s.Accept(i)
				}
			})
		}
	case []int:
		arr := rangeValue.([]int)
		for idx, val := range arr {
			i.runWithChildEnv(func() {
				i.env.SetAndImplyType(valIdentifier, val)
				if idxIdentifier != nil {
					i.env.SetAndImplyType(*idxIdentifier, idx)
				}
				for _, s := range stmt.Body.Stmts {
					s.Accept(i)
				}
			})
		}
	case []float64:
		arr := rangeValue.([]float64)
		for idx, val := range arr {
			i.runWithChildEnv(func() {
				i.env.SetAndImplyType(valIdentifier, val)
				if idxIdentifier != nil {
					i.env.SetAndImplyType(*idxIdentifier, idx)
				}
				for _, s := range stmt.Body.Stmts {
					s.Accept(i)
				}
			})
		}
	default:
		i.error(stmt.ForToken, "For loop range must be an array")
	}
}

func (i *MainInterpreter) error(token Token, message string) {
	if token == nil {
		panic(message) // todo not good
	} else {
		panic(fmt.Sprintf("Error at L%d/%d on '%s': %s",
			token.GetLine(), token.GetCharLineStart(), token.GetLexeme(), message))
	}
}

func (i *MainInterpreter) runWithChildEnv(runnable func()) {
	originalEnv := i.env
	env := originalEnv.NewChildEnv()
	i.env = &env
	runnable()
	i.env = originalEnv
}
