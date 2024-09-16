package core

import (
	"fmt"
	"strconv"
)

func (i *MainInterpreter) VisitBinaryExpr(binary Binary) interface{} {
	left := binary.Left.Accept(i)
	right := binary.Right.Accept(i)
	return i.execute(left, right, binary.Operator, binary.Operator.GetType())
}

func (i *MainInterpreter) VisitCompoundAssignStmt(assign CompoundAssign) {
	variable := i.env.GetByToken(assign.Name)
	operand := assign.Value.Accept(i)
	var operatorType TokenType
	switch assign.Operator.GetType() {
	case PLUS_EQUAL:
		operatorType = PLUS
	case MINUS_EQUAL:
		operatorType = MINUS
	case STAR_EQUAL:
		operatorType = STAR
	case SLASH_EQUAL:
		operatorType = SLASH
	default:
		i.error(assign.Operator, "Invalid compound assignment operator")
	}
	result := i.execute(variable.value, operand, assign.Operator, operatorType)
	i.env.SetAndImplyType(assign.Name, result)
}

func (i *MainInterpreter) execute(left interface{}, right interface{}, operatorToken Token, operatorType TokenType) interface{} {
	switch left.(type) {
	case int:
		switch right.(type) {
		case int:
			switch operatorType {
			case PLUS:
				return left.(int) + right.(int)
			case MINUS:
				return left.(int) - right.(int)
			case STAR:
				return left.(int) * right.(int)
			case SLASH:
				return left.(int) / right.(int)
			case GREATER:
				return left.(int) > right.(int)
			case GREATER_EQUAL:
				return left.(int) >= right.(int)
			case LESS:
				return left.(int) < right.(int)
			case LESS_EQUAL:
				return left.(int) <= right.(int)
			case EQUAL_EQUAL:
				return left.(int) == right.(int)
			case NOT_EQUAL:
				return left.(int) != right.(int)
			default:
				i.error(operatorToken, "Invalid binary operator for int, int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return float64(left.(int)) + right.(float64)
			case MINUS:
				return float64(left.(int)) - right.(float64)
			case STAR:
				return float64(left.(int)) * right.(float64)
			case SLASH:
				return float64(left.(int)) / right.(float64)
			case GREATER:
				return float64(left.(int)) > right.(float64)
			case GREATER_EQUAL:
				return float64(left.(int)) >= right.(float64)
			case LESS:
				return float64(left.(int)) < right.(float64)
			case LESS_EQUAL:
				return float64(left.(int)) <= right.(float64)
			case EQUAL_EQUAL:
				return float64(left.(int)) == right.(float64)
			case NOT_EQUAL:
				return float64(left.(int)) != right.(float64)
			default:
				i.error(operatorToken, "Invalid binary operator for int, float")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case float64:
		switch right.(type) {
		case int:
			switch operatorType {
			case PLUS:
				return left.(float64) + float64(right.(int))
			case MINUS:
				return left.(float64) - float64(right.(int))
			case STAR:
				return left.(float64) * float64(right.(int))
			case SLASH:
				return left.(float64) / float64(right.(int))
			case GREATER:
				return left.(float64) > float64(right.(int))
			case GREATER_EQUAL:
				return left.(float64) >= float64(right.(int))
			case LESS:
				return left.(float64) < float64(right.(int))
			case LESS_EQUAL:
				return left.(float64) <= float64(right.(int))
			case EQUAL_EQUAL:
				return left.(float64) == float64(right.(int))
			case NOT_EQUAL:
				return left.(float64) != float64(right.(int))
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
			i.error(operatorToken, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
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
		case int:
			switch operatorType {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(int))
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
		}
	case bool:
		i.error(operatorToken, fmt.Sprintf("Invalid binary operator for bool: %v", right))
	case []string:
		switch right.(type) {
		case []string:
			switch operatorType {
			case PLUS:
				return append(left.([]string), right.([]string)...)
			default:
				i.error(operatorToken, "Invalid binary operator for string[], string[]")
			}
		case string:
			switch operatorType {
			case PLUS:
				return append(left.([]string), right.(string))
			default:
				i.error(operatorToken, "Invalid binary operator for string[], string")
			}
		case int:
			switch operatorType {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(int)))
			default:
				i.error(operatorToken, "Invalid binary operator for string[], int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(float64)))
			default:
				i.error(operatorToken, "Invalid binary operator for string[], float")
			}
		case bool:
			switch operatorType {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(bool)))
			default:
				i.error(operatorToken, "Invalid binary operator for string[], bool")
			}
		case []int:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: string[], int[]"))
		case []float64:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: string[], float[]"))
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case []int:
		switch right.(type) {
		case []string:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: int[], string[]"))
		case string:
			switch operatorType {
			case PLUS:
				parsed, err := strconv.Atoi(right.(string))
				if err != nil {
					i.error(operatorToken, fmt.Sprintf("Cannot convert string to int: %v", right))
				}
				return append(left.([]int), parsed)
			default:
				i.error(operatorToken, "Invalid binary operator for int[], string")
			}
		case int:
			switch operatorType {
			case PLUS:
				return append(left.([]int), right.(int))
			default:
				i.error(operatorToken, "Invalid binary operator for int[], int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return append(left.([]int), int(right.(float64)))
			default:
				i.error(operatorToken, "Invalid binary operator for int[], float")
			}
		case []int:
			switch operatorType {
			case PLUS:
				return append(left.([]int), right.([]int)...)
			default:
				i.error(operatorToken, "Invalid binary operator for int[], int[]")
			}
		case []float64:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: int[], float[]"))
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case []float64:
		switch right.(type) {
		case []string:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: float[], string[]"))
		case string:
			switch operatorType {
			case PLUS:
				parsed, err := strconv.ParseFloat(right.(string), 64)
				if err != nil {
					i.error(operatorToken, fmt.Sprintf("Cannot convert string to float: %v", right))
				}
				return append(left.([]float64), parsed)
			default:
				i.error(operatorToken, "Invalid binary operator for float[], string")
			}
		case int:
			switch operatorType {
			case PLUS:
				return append(left.([]float64), float64(right.(int)))
			default:
				i.error(operatorToken, "Invalid binary operator for float[], int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return append(left.([]float64), right.(float64))
			default:
				i.error(operatorToken, "Invalid binary operator for float[], float")
			}
		case []int:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: float[], int[]"))
		case []float64:
			switch operatorType {
			case PLUS:
				return append(left.([]float64), right.([]float64)...)
			default:
				i.error(operatorToken, "Invalid binary operator for float[], float[]")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	}
	panic(UNREACHABLE)
}
