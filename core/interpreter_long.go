package core

import (
	"fmt"
	"strings"
)

func (i *MainInterpreter) VisitBinaryExpr(binary Binary) interface{} {
	left := binary.Left.Accept(i)
	right := binary.Right.Accept(i)
	return i.executeOp(left, right, binary.Tkn, binary.Op)
}

func (i *MainInterpreter) VisitCollectionEntryAssignStmt(assign CollectionEntryAssign) {
	key := assign.Key.Accept(i)
	collection := i.env.GetByToken(assign.Identifier)
	switch coerced := collection.(type) {
	case []interface{}:
		idx, ok := key.(int64)
		if !ok {
			i.error(assign.Identifier, "Array key must be an int")
		}
		adjustedIdx := idx
		if adjustedIdx < 0 {
			adjustedIdx += int64(len(coerced))
		}
		arrLen := len(coerced)
		if adjustedIdx < 0 || adjustedIdx >= int64(arrLen) {
			i.error(assign.Identifier, fmt.Sprintf("Array index out of bounds: %d (list length: %d)", idx, arrLen))
		}
		coerced[adjustedIdx] = i.calculateResult(coerced[adjustedIdx], assign.Value.Accept(i), assign.Operator)
	case RslMap:
		keyStr, ok := key.(RslString)
		if !ok {
			i.error(assign.Identifier, fmt.Sprintf("Map key must be a string, was %s", TypeAsString(key))) // todo still unsure about this constraint
		}
		existing, ok := coerced.Get(keyStr)
		if !ok && assign.Operator.GetType() != EQUAL {
			i.error(assign.Operator, fmt.Sprintf("Cannot use compound assignment on non-existing map key %q", ToPrintable(keyStr)))
		}
		coerced.Set(keyStr, i.calculateResult(existing, assign.Value.Accept(i), assign.Operator))
		i.env.SetAndImplyType(assign.Identifier, coerced)
	default:
		i.error(assign.Operator, fmt.Sprintf("Expected collection, got %T", collection))
	}
}

func (i *MainInterpreter) calculateResult(left interface{}, right interface{}, operator Token) interface{} {
	if operator.GetType() == EQUAL {
		// only used by collection entry assignment atm
		return right
	}

	op, _ := TKN_TYPE_TO_OP_MAP[operator.GetType()]
	return i.executeOp(left, right, operator, op)
}

func (i *MainInterpreter) executeOp(left interface{}, right interface{}, tkn Token, op OpType) interface{} {
	switch coercedLeft := left.(type) {
	case int64:
		switch coercedRight := right.(type) {
		case int64:
			switch op {
			case OP_PLUS:
				return coercedLeft + coercedRight
			case OP_MINUS:
				return coercedLeft - coercedRight
			case OP_MULTIPLY:
				return coercedLeft * coercedRight
			case OP_DIVIDE:
				return coercedLeft / coercedRight
			case OP_GREATER:
				return coercedLeft > coercedRight
			case OP_GREATER_EQUAL:
				return coercedLeft >= coercedRight
			case OP_LESS:
				return coercedLeft < coercedRight
			case OP_LESS_EQUAL:
				return coercedLeft <= coercedRight
			case OP_EQUAL:
				return coercedLeft == coercedRight
			case OP_NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.error(tkn, "Invalid binary operator for int, int")
			}
		case float64:
			switch op {
			case OP_PLUS:
				return float64(coercedLeft) + coercedRight
			case OP_MINUS:
				return float64(coercedLeft) - coercedRight
			case OP_MULTIPLY:
				return float64(coercedLeft) * coercedRight
			case OP_DIVIDE:
				return float64(coercedLeft) / coercedRight
			case OP_GREATER:
				return float64(coercedLeft) > coercedRight
			case OP_GREATER_EQUAL:
				return float64(coercedLeft) >= coercedRight
			case OP_LESS:
				return float64(coercedLeft) < coercedRight
			case OP_LESS_EQUAL:
				return float64(coercedLeft) <= coercedRight
			case OP_EQUAL:
				return float64(coercedLeft) == coercedRight
			case OP_NOT_EQUAL:
				return float64(coercedLeft) != coercedRight
			default:
				i.error(tkn, "Invalid binary operator for int, float")
			}
		case RslString:
			switch op {
			// todo python does not allow this, should we?
			case OP_IN:
				return strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			case OP_NOT_IN:
				return !strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			default:
				i.error(tkn, "Invalid binary operator for int, string")
			}
		case []interface{}:
			switch op {
			case OP_IN:
				return contains(coercedRight, coercedLeft)
			case OP_NOT_IN:
				return !contains(coercedRight, coercedLeft)
			}
		default:
			i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case float64:
		switch coercedRight := right.(type) {
		case int64:
			switch op {
			case OP_PLUS:
				return coercedLeft + float64(coercedRight)
			case OP_MINUS:
				return coercedLeft - float64(coercedRight)
			case OP_MULTIPLY:
				return coercedLeft * float64(coercedRight)
			case OP_DIVIDE:
				return coercedLeft / float64(coercedRight)
			case OP_GREATER:
				return coercedLeft > float64(coercedRight)
			case OP_GREATER_EQUAL:
				return coercedLeft >= float64(coercedRight)
			case OP_LESS:
				return coercedLeft < float64(coercedRight)
			case OP_LESS_EQUAL:
				return coercedLeft <= float64(coercedRight)
			case OP_EQUAL:
				return coercedLeft == float64(coercedRight)
			case OP_NOT_EQUAL:
				return coercedLeft != float64(coercedRight)
			default:
				i.error(tkn, "Invalid binary operator for int, int")
			}
		case float64:
			switch op {
			case OP_PLUS:
				return coercedLeft + coercedRight
			case OP_MINUS:
				return coercedLeft - coercedRight
			case OP_MULTIPLY:
				return coercedLeft * coercedRight
			case OP_DIVIDE:
				return coercedLeft / coercedRight
			case OP_GREATER:
				return coercedLeft > coercedRight
			case OP_GREATER_EQUAL:
				return coercedLeft >= coercedRight
			case OP_LESS:
				return coercedLeft < coercedRight
			case OP_LESS_EQUAL:
				return coercedLeft <= coercedRight
			case OP_EQUAL:
				return coercedLeft == coercedRight
			case OP_NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.error(tkn, "Invalid binary operator for int, float64")
			}
		case []interface{}:
			switch op {
			case OP_IN:
				return contains(coercedRight, coercedLeft)
			case OP_NOT_IN:
				return !contains(coercedRight, coercedLeft)
			default:
				i.error(tkn, "Invalid binary operator for float64, array")
			}
		default:
			i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case RslString:
		switch coercedRight := right.(type) {
		case RslString:
			switch op {
			case OP_PLUS:
				return coercedLeft.Concat(coercedRight)
			case OP_EQUAL:
				return coercedLeft.Equals(coercedRight)
			case OP_NOT_EQUAL:
				return !coercedLeft.Equals(coercedRight)
			case OP_IN:
				return strings.Contains(coercedRight.Plain(), coercedLeft.Plain())
			case OP_NOT_IN:
				return !strings.Contains(coercedRight.Plain(), coercedLeft.Plain())
			default:
				i.error(tkn, "Invalid binary operator for string, string")
			}
		case int64:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight))
			default:
				i.error(tkn, "Invalid binary operator for string, int")
			}
		case float64:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight)) // todo check formatting
			default:
				i.error(tkn, "Invalid binary operator for string, float")
			}
		case bool:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight))
			default:
				i.error(tkn, "Invalid binary operator for string, bool")
			}
		case []interface{}:
			switch op {
			case OP_IN:
				return contains(coercedRight, coercedLeft)
			case OP_NOT_IN:
				return !contains(coercedRight, coercedLeft)
			default:
				i.error(tkn, "Invalid binary operator for string, array")
			}
		case RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(coercedLeft)
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(coercedLeft)
			default:
				i.error(tkn, "Invalid binary operator for string, map")
			}
		default:
			i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case bool:
		switch coercedRight := right.(type) {
		case bool:
			switch op {
			case OP_EQUAL:
				return coercedLeft == coercedRight
			case OP_NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.error(tkn, "Invalid binary operator for bool, bool")
			}
		case []interface{}:
			switch op {
			case OP_IN:
				return contains(coercedRight, coercedLeft)
			case OP_NOT_IN:
				return !contains(coercedRight, coercedLeft)
			default:
				i.error(tkn, "Invalid binary operator for bool, array")
			}
		default:
			i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case []interface{}:
		switch coercedRight := right.(type) {
		case RslString:
			i.error(tkn, "Invalid binary operator for mixed array, string")
		case int64:
			i.error(tkn, "Invalid binary operator for mixed array, int")
		case float64:
			i.error(tkn, "Invalid binary operator for mixed array, float")
		case bool:
			i.error(tkn, "Invalid binary operator for mixed array, bool")
		case []interface{}:
			switch op {
			case OP_PLUS:
				return append(coercedLeft, coercedRight...)
			case OP_IN:
				return contains(coercedLeft, coercedRight)
			case OP_NOT_IN:
				return !contains(coercedLeft, coercedRight)
			default:
				i.error(tkn, "Invalid binary operator for array, array")
			}
		default:
			i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	default:
		i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
	}
	panic(UNREACHABLE)
}

// todo should probably handle RslMap directly
func contains(array []interface{}, val interface{}) bool {
	if a, ok := val.(RslString); ok {
		for _, v := range array {
			if b, ok := v.(RslString); ok && a.Equals(b) {
				return true
			}
		}
	} else {
		for _, v := range array {
			if v == val {
				return true
			}
		}
	}
	return false
}
