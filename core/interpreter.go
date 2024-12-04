package core

import (
	"fmt"
)

type MainInterpreter struct {
	env           *Env
	LiteralI      *LiteralInterpreter
	argBlockI     *ArgBlockInterpreter
	radBlockI     *RadBlockInterpreter
	switchI       *SwitchInterpreter
	statements    []Stmt
	deferredStmts []DeferStmt

	breaking   bool
	continuing bool
}

func NewInterpreter(statements []Stmt) *MainInterpreter {
	i := &MainInterpreter{
		statements:    statements,
		deferredStmts: make([]DeferStmt, 0),
	}
	i.LiteralI = NewLiteralInterpreter(i)
	i.argBlockI = NewArgBlockInterpreter(i)
	i.radBlockI = NewRadBlockInterpreter(i)
	i.switchI = NewSwitchInterpreter(i)
	i.env = NewEnv(i)
	return i
}

func (i *MainInterpreter) InitArgs(args []RslArg) {
	for _, arg := range args {
		i.env.InitArg(arg)
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
		keyValString, ok := keyVal.(RslString)
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
			adjustedKey := coercedKey
			if adjustedKey < 0 {
				adjustedKey += int64(len(coerced))
			}
			if adjustedKey < 0 || adjustedKey >= int64(len(coerced)) {
				i.error(access.AccessOpener, fmt.Sprintf("Array index out of bounds: %d (list length: %d)", coercedKey, len(coerced)))
			}
			return coerced[adjustedKey]
		default:
			i.error(access.AccessOpener, "Array access key must be an int")
			panic(UNREACHABLE)
		}
	case RslString:
		switch coercedKey := key.(type) {
		case int64:
			adjustedKey := coercedKey
			if adjustedKey < 0 {
				adjustedKey += coerced.Len()
			}
			if adjustedKey < 0 || adjustedKey >= coerced.Len() {
				i.error(access.AccessOpener, fmt.Sprintf("String index out of bounds: %d (string length: %d)", coercedKey, coerced.Len()))
			}
			return coerced.IndexAt(adjustedKey)
		default:
			i.error(access.AccessOpener, "String index must be an int")
			panic(UNREACHABLE)
		}
	case RslMap:
		switch coercedKey := key.(type) {
		case RslString:
			val, exists := coerced.Get(coercedKey)
			if !exists {
				i.error(access.AccessOpener, fmt.Sprintf("Key '%s' not found in map", coercedKey))
				panic(UNREACHABLE)
			}
			return val
		default:
			i.error(access.AccessOpener, fmt.Sprintf("Map access key must be a string, was %v (%T)", key, key))
			panic(UNREACHABLE)
		}
	default:
		i.error(access.AccessOpener, "Bug! Should've failed earlier")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) VisitSliceAccessExpr(sliceAccess SliceAccess) interface{} {
	original := sliceAccess.ListOrString.Accept(i)

	switch coerced := original.(type) {
	case RslString:
		start, end := i.resolveStartEnd(sliceAccess, coerced.Len())
		return coerced.Slice(start, end)
	case []interface{}:
		start, end := i.resolveStartEnd(sliceAccess, int64(len(coerced)))
		return coerced[start:end]
	default:
		i.error(sliceAccess.AccessOpener, "Slice access must be on a string or array")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) VisitFunctionCallExpr(call FunctionCall) interface{} {
	return RunRslNonVoidFunction(i, call.Function, call.NumExpectedReturnValues, evalArgs(i, call.Args), call.NamedArgs)
}

func (i *MainInterpreter) VisitFunctionStmtStmt(functionStmt FunctionStmt) {
	RunRslFunction(i, functionStmt.Call)
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
		case IDENTIFIER:
			if unary.Operator.GetLexeme() == "not" {
				return !valBool
			} else {
				i.error(unary.Operator, fmt.Sprintf("Bug! Expected 'not' identifier, got %q", unary.Operator.GetLexeme()))
			}
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
	// todo have i not implemented the parser emitting these? are blocks currently not supported accidentally?
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

func (i *MainInterpreter) VisitTernaryExpr(ternary Ternary) interface{} {
	conditionResult := ternary.Condition.Accept(i)
	bval, ok := conditionResult.(bool)
	if !ok {
		i.error(ternary.QuestionMark, "Ternary condition must resolve to a bool")
	}
	if bval {
		return ternary.True.Accept(i)
	}
	return ternary.False.Accept(i)
}

func (i *MainInterpreter) VisitForStmtStmt(stmt ForStmt) {
	rangeValue := stmt.Range.Accept(i)

	switch coerced := rangeValue.(type) {
	case []interface{}:
		var valIdentifier Token
		var idxIdentifier *Token
		if stmt.Identifier2 != nil {
			idxIdentifier = &stmt.Identifier1
			valIdentifier = *stmt.Identifier2
		} else {
			valIdentifier = stmt.Identifier1
		}
		i.runWithChildEnv(runArrayForLoop(i, stmt, coerced, idxIdentifier, valIdentifier))
	case RslMap:
		i.runWithChildEnv(runMapForLoop(i, stmt, coerced, stmt.Identifier1, stmt.Identifier2))
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

func runArrayForLoop(
	i *MainInterpreter,
	stmt ForStmt,
	rangeArr []interface{},
	idxIdentifier *Token,
	valIdentifier Token,
) func() {
	return func() {
		for idx, val := range rangeArr {
			if idxIdentifier != nil {
				i.env.SetAndImplyType(*idxIdentifier, int64(idx))
			}
			i.env.SetAndImplyType(valIdentifier, val)
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

func runMapForLoop(
	i *MainInterpreter,
	stmt ForStmt,
	rangeMap RslMap,
	keyIdentifier Token,
	valIdentifier *Token,
) func() {
	return func() {
		keys := rangeMap.keys
		for _, key := range keys {
			i.env.SetAndImplyType(keyIdentifier, key)
			if valIdentifier != nil {
				val, ok := rangeMap.GetStr(key)
				if !ok {
					i.error(stmt.ForToken, fmt.Sprintf("Bug! Map contains key %q but lookup failed.", key))
				}
				i.env.SetAndImplyType(*valIdentifier, val)
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

func (i *MainInterpreter) VisitDeferStmtStmt(deferStmt DeferStmt) {
	i.deferredStmts = append(i.deferredStmts, deferStmt)
}

func (i *MainInterpreter) resolveStartEnd(sliceAccess SliceAccess, len int64) (int64, int64) {
	start := int64(0)
	end := len

	if sliceAccess.Start != nil {
		start = i.resolveSliceIndex(sliceAccess.AccessOpener, *sliceAccess.Start, len, true)
	}
	if sliceAccess.End != nil {
		end = i.resolveSliceIndex(sliceAccess.AccessOpener, *sliceAccess.End, len, false)
	}

	if start > end {
		start = end
	}

	return start, end
}

// todo currently these execute after an error is printed. Should they execute before?
func (i *MainInterpreter) ExecuteDeferredStmts(errCode int) {
	// execute backwards (LIFO)
	for j := len(i.deferredStmts) - 1; j >= 0; j-- {
		deferredStmt := i.deferredStmts[j]

		if deferredStmt.IsErrDefer && errCode == 0 {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					// we only debug log. we expect the error that occurred to already have been logged.
					// we might also be here only because a deferred statement invoked a clean exit, for example, so
					// this is arguably also sometimes just standard flow.
					RP.RadDebug(fmt.Sprintf("Recovered from panic in deferred statement: %v", r))
				}
			}()

			if deferredStmt.DeferredStmt != nil {
				// todo why does this need to be dereferenced but not the block below?
				(*deferredStmt.DeferredStmt).Accept(i)
			} else if deferredStmt.DeferredBlock != nil {
				deferredStmt.DeferredBlock.Accept(i)
			} else {
				i.error(deferredStmt.DeferToken, "Bug! Deferred statement should have either a statement or a block")
			}
		}()
	}
}

func (i *MainInterpreter) resolveSliceIndex(token Token, expr Expr, len int64, isStart bool) int64 {
	index := expr.Accept(i)
	rawIdx, ok := index.(int64)
	if !ok {
		i.error(token, fmt.Sprintf("Slice index must be an int, was %T (%v)", index, index))
	}

	var idx = rawIdx
	if rawIdx < 0 {
		idx = rawIdx + len
	}

	if isStart {
		if idx < 0 {
			// the start index is still negative, so we'll slice from the beginning
			idx = 0
		}
		if idx > len {
			// the start index is greater than the length of the list, so we'll slice to the end
			idx = len
		}
	} else {
		if idx > len {
			// the end index is greater than the length of the list, so we'll slice to the end
			idx = len
		}
		if idx < 0 {
			// the end index is still negative, so we'll slice from the end
			idx = 0
		}
	}

	return idx
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
		keyStr, ok := key.(RslString)
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
	i.errorWithCode(token, message, 1)
}

func (i *MainInterpreter) errorWithCode(token Token, message string, errorCode int) {
	RP.TokenErrorCodeExit(token, message+"\n", errorCode)
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
