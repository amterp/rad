package core

import (
	"fmt"
	"strings"
)

func (i *MainInterpreter) VisitBinaryExpr(binary Binary) interface{} {
	left := binary.Left.Accept(i)
	right := binary.Right.Accept(i)
	return i.executeOp(left, right, binary.Operator, binary.Operator.GetType())
}

func (i *MainInterpreter) VisitCompoundAssignStmt(assign CompoundAssign) {
	left := i.env.GetByToken(assign.Name)
	right := assign.Value.Accept(i)
	result := i.calculateResult(left, right, assign.Operator)
	i.env.SetAndImplyType(assign.Name, result)
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
		keyStr, ok := key.(string)
		if !ok {
			i.error(assign.Identifier, "Map key must be a string") // todo still unsure about this constraint
		}
		existing, ok := coerced.Get(keyStr)
		if !ok && assign.Operator.GetType() != EQUAL {
			i.error(assign.Operator, fmt.Sprintf("Cannot use compound assignment on non-existing map key %q", keyStr))
		}
		coerced.Set(keyStr, i.calculateResult(existing, assign.Value.Accept(i), assign.Operator))
		i.env.SetAndImplyType(assign.Identifier, coerced)
	default:
		i.error(assign.Operator, fmt.Sprintf("Expected collection, got %T", collection))
	}
}

func (i *MainInterpreter) calculateResult(left interface{}, right interface{}, operator Token) interface{} {
	var operatorType TokenType
	switch operator.GetType() {
	case EQUAL:
		// only used by collection entry assignment atm
		return right
	case PLUS_EQUAL:
		operatorType = PLUS
	case MINUS_EQUAL:
		operatorType = MINUS
	case STAR_EQUAL:
		operatorType = STAR
	case SLASH_EQUAL:
		operatorType = SLASH
	default:
		i.error(operator, "Invalid assignment operator")
	}
	return i.executeOp(left, right, operator, operatorType)
}

func (i *MainInterpreter) executeOp(left interface{}, right interface{}, operatorToken Token, operatorType TokenType) interface{} {
	switch coercedLeft := left.(type) {
	case int64:
		switch coercedRight := right.(type) {
		case int64:
			switch operatorType {
			case PLUS:
				return coercedLeft + coercedRight
			case MINUS:
				return coercedLeft - coercedRight
			case STAR:
				return coercedLeft * coercedRight
			case SLASH:
				return coercedLeft / coercedRight
			case GREATER:
				return coercedLeft > coercedRight
			case GREATER_EQUAL:
				return coercedLeft >= coercedRight
			case LESS:
				return coercedLeft < coercedRight
			case LESS_EQUAL:
				return coercedLeft <= coercedRight
			case EQUAL_EQUAL:
				return coercedLeft == coercedRight
			case NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.error(operatorToken, "Invalid binary operator for int, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return float64(coercedLeft) + coercedRight
			case MINUS:
				return float64(coercedLeft) - coercedRight
			case STAR:
				return float64(coercedLeft) * coercedRight
			case SLASH:
				return float64(coercedLeft) / coercedRight
			case GREATER:
				return float64(coercedLeft) > coercedRight
			case GREATER_EQUAL:
				return float64(coercedLeft) >= coercedRight
			case LESS:
				return float64(coercedLeft) < coercedRight
			case LESS_EQUAL:
				return float64(coercedLeft) <= coercedRight
			case EQUAL_EQUAL:
				return float64(coercedLeft) == coercedRight
			case NOT_EQUAL:
				return float64(coercedLeft) != coercedRight
			default:
				i.error(operatorToken, "Invalid binary operator for int, float")
			}
		case string:
			switch operatorType {
			// todo python does not allow this, should we?
			case IN:
				return strings.Contains(coercedRight, fmt.Sprintf("%v", coercedLeft))
			case NOT_IN:
				return !strings.Contains(coercedRight, fmt.Sprintf("%v", coercedLeft))
			default:
				i.error(operatorToken, "Invalid binary operator for int, string")
			}
		case []interface{}:
			switch operatorType {
			case IN:
				return contains(coercedRight, coercedLeft)
			case NOT_IN:
				return !contains(coercedRight, coercedLeft)
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case float64:
		switch coercedRight := right.(type) {
		case int64:
			switch operatorType {
			case PLUS:
				return coercedLeft + float64(coercedRight)
			case MINUS:
				return coercedLeft - float64(coercedRight)
			case STAR:
				return coercedLeft * float64(coercedRight)
			case SLASH:
				return coercedLeft / float64(coercedRight)
			case GREATER:
				return coercedLeft > float64(coercedRight)
			case GREATER_EQUAL:
				return coercedLeft >= float64(coercedRight)
			case LESS:
				return coercedLeft < float64(coercedRight)
			case LESS_EQUAL:
				return coercedLeft <= float64(coercedRight)
			case EQUAL_EQUAL:
				return coercedLeft == float64(coercedRight)
			case NOT_EQUAL:
				return coercedLeft != float64(coercedRight)
			default:
				i.error(operatorToken, "Invalid binary operator for int, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return coercedLeft + coercedRight
			case MINUS:
				return coercedLeft - coercedRight
			case STAR:
				return coercedLeft * coercedRight
			case SLASH:
				return coercedLeft / coercedRight
			case GREATER:
				return coercedLeft > coercedRight
			case GREATER_EQUAL:
				return coercedLeft >= coercedRight
			case LESS:
				return coercedLeft < coercedRight
			case LESS_EQUAL:
				return coercedLeft <= coercedRight
			case EQUAL_EQUAL:
				return coercedLeft == coercedRight
			case NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.error(operatorToken, "Invalid binary operator for int, float64")
			}
		case []interface{}:
			switch operatorType {
			case IN:
				return contains(coercedRight, coercedLeft)
			case NOT_IN:
				return !contains(coercedRight, coercedLeft)
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case string:
		switch coercedRight := right.(type) {
		case string:
			switch operatorType {
			case PLUS:
				return coercedLeft + coercedRight
			case EQUAL_EQUAL:
				return coercedLeft == coercedRight
			case NOT_EQUAL:
				return coercedLeft != coercedRight
			case IN:
				return strings.Contains(coercedRight, coercedLeft)
			case NOT_IN:
				return !strings.Contains(coercedRight, coercedLeft)
			default:
				i.error(operatorToken, "Invalid binary operator for string, string")
			}
		case int64:
			switch operatorType {
			case PLUS:
				return coercedLeft + fmt.Sprintf("%v", coercedRight)
			default:
				i.error(operatorToken, "Invalid binary operator for string, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return coercedLeft + fmt.Sprintf("%v", coercedRight) // todo check formatting
			default:
				i.error(operatorToken, "Invalid binary operator for string, float")
			}
		case bool:
			switch operatorType {
			case PLUS:
				return coercedLeft + fmt.Sprintf("%v", coercedRight)
			default:
				i.error(operatorToken, "Invalid binary operator for string, bool")
			}
		case []interface{}:
			switch operatorType {
			case IN:
				return contains(coercedRight, coercedLeft)
			case NOT_IN:
				return !contains(coercedRight, coercedLeft)
			}
		case RslMap:
			switch operatorType {
			case IN:
				return coercedRight.ContainsKey(coercedLeft)
			case NOT_IN:
				return !coercedRight.ContainsKey(coercedLeft)
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case bool:
		switch coercedRight := right.(type) {
		case bool:
			switch operatorType {
			case AND:
				return coercedLeft && coercedRight
			case OR:
				return coercedLeft || coercedRight
			case EQUAL_EQUAL:
				return coercedLeft == coercedRight
			case NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.error(operatorToken, "Invalid binary operator for bool, bool")
			}
		case []interface{}:
			switch operatorType {
			case IN:
				return contains(coercedRight, coercedLeft)
			case NOT_IN:
				return !contains(coercedRight, coercedLeft)
			default:
				i.error(operatorToken, "Invalid binary operator for bool, array")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case []interface{}:
		switch coercedRight := right.(type) {
		case string:
			switch operatorType {
			case PLUS:
				return append(coercedLeft, right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, string")
			}
		case int64:
			switch operatorType {
			case PLUS:
				return append(coercedLeft, right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return append(coercedLeft, right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, float")
			}
		case bool:
			switch operatorType {
			case PLUS:
				return append(coercedLeft, right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, bool")
			}
		case []interface{}:
			switch operatorType {
			case PLUS:
				return append(coercedLeft, coercedRight...)
			case IN:
				return contains(coercedLeft, coercedRight)
			case NOT_IN:
				return !contains(coercedLeft, coercedRight)
			default:
				i.error(operatorToken, "Invalid binary operator for array, array")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	default:
		i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
	}
	panic(UNREACHABLE)
}

func contains(array []interface{}, val interface{}) bool {
	for _, v := range array {
		if v == val {
			return true
		}
	}
	return false
}
