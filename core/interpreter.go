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

func (i *MainInterpreter) VisitMapExprExpr(expr MapExpr) interface{} {
	rslMap := NewRslMap()
	for idx, key := range expr.Keys {
		keyVal := key.Accept(i)
		keyValString, ok := keyVal.(string)
		if !ok {
			i.error(expr.OpenBraceToken, fmt.Sprintf("Map keys must be strings; key #%d was %v (%T)", idx, keyVal, keyVal))
		}

		value := expr.Values[idx].Accept(i)
		rslMap.Set(keyValString, value)
	}
	return *rslMap
}

func (i *MainInterpreter) VisitCollectionAccessExpr(access CollectionAccess) interface{} {
	collection := access.Collection.Accept(i)
	key := access.Key.Accept(i)

	switch coerced := collection.(type) {
	case []interface{}:
		switch coercedKey := key.(type) {
		case int64:
			return coerced[coercedKey]
		default:
			i.error(access.OpenBracketToken, "Array access key must be an int")
			panic(UNREACHABLE)
		}
	case RslMap:
		switch coercedKey := key.(type) {
		case string:
			val, exists := coerced.Get(coercedKey)
			if !exists {
				i.error(access.OpenBracketToken, fmt.Sprintf("Key '%s' not found in map", coercedKey))
				panic(UNREACHABLE)
			}
			return val
		default:
			i.error(access.OpenBracketToken, fmt.Sprintf("Map access key must be a string, was %v (%T)", key, key))
			panic(UNREACHABLE)
		}
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
	return i.env.GetByToken(variable.Name)
}

func (i *MainInterpreter) VisitLogicalExpr(logical Logical) interface{} {
	left := logical.Left.Accept(i).(bool)
	right := logical.Right.Accept(i).(bool)

	operator, ok := GLOBAL_KEYWORDS[logical.Operator.GetLexeme()]
	if !ok || (operator != OR && operator != AND) {
		i.error(logical.Operator, "Bug! Non-and/or logical operator should've not passed the parser")
		panic(UNREACHABLE)
	}

	if operator == OR {
		return left || right
	}
	return left && right
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
		handleMultiAssignment(i, values, assign.Identifiers) // todo does this handle all arrays?
	default:
		i.error(assign.Identifiers[0], "Expected multiple values, got one")
	}
}

func handleMultiAssignment[T any](i *MainInterpreter, values []T, identifiers []Token) {
	if len(values) != len(identifiers) {
		i.error(identifiers[0], fmt.Sprintf("Expected %d values, got %d", len(identifiers), len(values)))
	}
	for idx, val := range values {
		i.env.SetAndImplyType(identifiers[idx], val)
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
	case []interface{}:
		arr := rangeValue.([]interface{})
		i.runWithChildEnv(runForLoop(i, stmt, arr, idxIdentifier, valIdentifier))
	// todo allow iterating through map
	default:
		i.error(stmt.ForToken, "For loop range must be an array")
	}
}

func (i *MainInterpreter) VisitListComprehensionExpr(comp ListComprehension) interface{} {
	rangeVals := comp.Range.Accept(i)
	var valIdent Token
	var idxIdent *Token
	if comp.Identifier2 != nil {
		idxIdent = &comp.Identifier1
		valIdent = *comp.Identifier2
	} else {
		valIdent = comp.Identifier1
	}
	switch coerced := rangeVals.(type) {
	case []interface{}:
		return i.computeWithChildEnv(runListComprehensionLoop(i, comp.For, coerced, idxIdent, valIdent, comp.Expression, comp.Condition))
	default:
		i.error(comp.For, "List comprehension range must be an array")
		panic(UNREACHABLE)
	}
}

func runForLoop[T any](i *MainInterpreter, stmt ForStmt, rangeArr []T, idxIdentifier *Token, valIdentifier Token) func() {
	return func() {
		for idx, val := range rangeArr {
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

func runListComprehensionLoop[T any](
	i *MainInterpreter,
	forToken Token,
	rangeArr []T,
	idxIdentifier *Token,
	valIdentifier Token,
	expression Expr,
	condition *Expr,
) func() interface{} {
	return func() interface{} {
		var output []interface{}
		for idx, val := range rangeArr {
			i.env.SetAndImplyType(valIdentifier, val)
			if idxIdentifier != nil {
				i.env.SetAndImplyType(*idxIdentifier, int64(idx))
			}
			if condition != nil {
				conditionResult := (*condition).Accept(i)
				bval, ok := conditionResult.(bool)
				if !ok {
					i.error(forToken, "List comprehension condition must resolve to a bool")
				}
				if !bval {
					continue
				}
			}
			output = append(output, expression.Accept(i))
		}
		return output
	}
}

func (i *MainInterpreter) VisitBreakStmtStmt(BreakStmt) {
	i.breaking = true
}

func (i *MainInterpreter) VisitContinueStmtStmt(ContinueStmt) {
	i.continuing = true
}

func (i *MainInterpreter) VisitVarPathExpr(path VarPath) interface{} {
	i.error(path.Identifier, fmt.Sprintf("%T should not be visited directly", path))
	panic(UNREACHABLE)
}

func (i *MainInterpreter) VisitDeleteStmtStmt(del DeleteStmt) {
	for _, varPath := range del.Vars {
		identifier := varPath.Identifier
		identifierLexeme := identifier.GetLexeme()
		if len(varPath.Keys) == 0 {
			i.env.Delete(identifierLexeme)
		} else {
			modified := i.executeDelete(identifier, i.env.GetByToken(identifier), varPath.Keys)
			i.env.SetAndImplyType(identifier, modified)
		}
	}
}

func (i *MainInterpreter) executeDelete(identifier Token, value interface{}, keys []Expr) interface{} {
	if len(keys) == 0 {
		return value
	}

	key := keys[0].Accept(i)
	switch coerced := value.(type) {
	case []interface{}:
		idx, ok := key.(int64)
		if !ok {
			i.error(identifier, fmt.Sprintf("Array index must be an int, was %T (%v)", key, key))
		}
		if idx < 0 || idx >= int64(len(coerced)) {
			i.error(identifier, fmt.Sprintf("Array index out of bounds: %d > max idx %d", idx, len(coerced)-1))
		}

		if len(keys) == 1 {
			// end of the line, delete whatever we're pointing at
			coerced = append(coerced[:idx], coerced[idx+1:]...)
			return coerced
		} else {
			// we want to delete something deeper in the array, recurse
			coerced[idx] = i.executeDelete(identifier, coerced[idx], keys[1:])
			return coerced
		}

	case RslMap:
		keyStr, ok := key.(string)
		if !ok {
			// todo still unsure about this string constraint
			i.error(identifier, fmt.Sprintf("Map key must be a string, was %T (%v)", key, key))
		}
		val, exists := coerced.Get(keyStr)
		if !exists {
			i.error(identifier, fmt.Sprintf("Map key %q not found", keyStr))
		}
		if len(keys) == 1 {
			// end of the line, delete whatever we're pointing at
			coerced.Delete(keyStr)
			return coerced
		} else {
			// we want to delete something deeper in the map, recurse
			coerced.Set(keyStr, i.executeDelete(identifier, val, keys[1:]))
			return coerced
		}
	default:
		i.error(identifier, fmt.Sprintf("Expected collection for key %q, got %T", key, value))
		panic(UNREACHABLE)
	}
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

func (i *MainInterpreter) computeWithChildEnv(computable func() interface{}) interface{} {
	originalEnv := i.env
	env := originalEnv.NewChildEnv()
	i.env = &env
	result := computable()
	i.env = originalEnv
	return result
}
