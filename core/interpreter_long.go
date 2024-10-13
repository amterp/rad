package core

import (
	"fmt"
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
		arrLen := len(coerced)
		if idx < 0 || idx >= int64(arrLen) {
			i.error(assign.Identifier, fmt.Sprintf("Array index out of bounds: %d > max idx %d", idx, arrLen-1))
		}
		coerced[idx] = i.calculateResult(coerced[idx], assign.Value.Accept(i), assign.Operator)
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
	switch left.(type) {
	case int64:
		switch right.(type) {
		case int64:
			switch operatorType {
			case PLUS:
				return left.(int64) + right.(int64)
			case MINUS:
				return left.(int64) - right.(int64)
			case STAR:
				return left.(int64) * right.(int64)
			case SLASH:
				return left.(int64) / right.(int64)
			case GREATER:
				return left.(int64) > right.(int64)
			case GREATER_EQUAL:
				return left.(int64) >= right.(int64)
			case LESS:
				return left.(int64) < right.(int64)
			case LESS_EQUAL:
				return left.(int64) <= right.(int64)
			case EQUAL_EQUAL:
				return left.(int64) == right.(int64)
			case NOT_EQUAL:
				return left.(int64) != right.(int64)
			default:
				i.error(operatorToken, "Invalid binary operator for int, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return float64(left.(int64)) + right.(float64)
			case MINUS:
				return float64(left.(int64)) - right.(float64)
			case STAR:
				return float64(left.(int64)) * right.(float64)
			case SLASH:
				return float64(left.(int64)) / right.(float64)
			case GREATER:
				return float64(left.(int64)) > right.(float64)
			case GREATER_EQUAL:
				return float64(left.(int64)) >= right.(float64)
			case LESS:
				return float64(left.(int64)) < right.(float64)
			case LESS_EQUAL:
				return float64(left.(int64)) <= right.(float64)
			case EQUAL_EQUAL:
				return float64(left.(int64)) == right.(float64)
			case NOT_EQUAL:
				return float64(left.(int64)) != right.(float64)
			default:
				i.error(operatorToken, "Invalid binary operator for int, float")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case float64:
		switch right.(type) {
		case int64:
			switch operatorType {
			case PLUS:
				return left.(float64) + float64(right.(int64))
			case MINUS:
				return left.(float64) - float64(right.(int64))
			case STAR:
				return left.(float64) * float64(right.(int64))
			case SLASH:
				return left.(float64) / float64(right.(int64))
			case GREATER:
				return left.(float64) > float64(right.(int64))
			case GREATER_EQUAL:
				return left.(float64) >= float64(right.(int64))
			case LESS:
				return left.(float64) < float64(right.(int64))
			case LESS_EQUAL:
				return left.(float64) <= float64(right.(int64))
			case EQUAL_EQUAL:
				return left.(float64) == float64(right.(int64))
			case NOT_EQUAL:
				return left.(float64) != float64(right.(int64))
			default:
				i.error(operatorToken, "Invalid binary operator for int, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return left.(float64) + right.(float64)
			case MINUS:
				return left.(float64) - right.(float64)
			case STAR:
				return left.(float64) * right.(float64)
			case SLASH:
				return left.(float64) / right.(float64)
			case GREATER:
				return left.(float64) > right.(float64)
			case GREATER_EQUAL:
				return left.(float64) >= right.(float64)
			case LESS:
				return left.(float64) < right.(float64)
			case LESS_EQUAL:
				return left.(float64) <= right.(float64)
			case EQUAL_EQUAL:
				return left.(float64) == right.(float64)
			case NOT_EQUAL:
				return left.(float64) != right.(float64)
			default:
				i.error(operatorToken, "Invalid binary operator for int, float64")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case string:
		switch right.(type) {
		case string:
			switch operatorType {
			case PLUS:
				return left.(string) + right.(string)
			case EQUAL_EQUAL:
				return left.(string) == right.(string)
			case NOT_EQUAL:
				return left.(string) != right.(string)
			default:
				i.error(operatorToken, "Invalid binary operator for string, string")
			}
		case int64:
			switch operatorType {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(int64))
			default:
				i.error(operatorToken, "Invalid binary operator for string, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(float64)) // todo check formatting
			default:
				i.error(operatorToken, "Invalid binary operator for string, float")
			}
		case bool:
			switch operatorType {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(bool))
			default:
				i.error(operatorToken, "Invalid binary operator for string, bool")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case bool:
		i.error(operatorToken, fmt.Sprintf("Invalid binary operator for bool: %v", right))
	case []interface{}:
		switch right.(type) {
		case string:
			switch operatorType {
			case PLUS:
				return append(left.([]interface{}), right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, string")
			}
		case int64:
			switch operatorType {
			case PLUS:
				return append(left.([]interface{}), right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return append(left.([]interface{}), right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, float")
			}
		case bool:
			switch operatorType {
			case PLUS:
				return append(left.([]interface{}), right)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, bool")
			}
		case []interface{}:
			return append(left.([]interface{}), right.([]interface{})...)
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	default:
		i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
	}
	panic(UNREACHABLE)
}
