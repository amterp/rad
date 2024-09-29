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
		case int64:
			switch operatorType {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(int64)))
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
		case []interface{}:
			arr, ok := AsStringArray(right.([]interface{}))
			if !ok {
				i.error(operatorToken, "Cannot join two arrays of different types: string[], mixed array")
			}
			return append(left.([]string), arr...)
		case []int64:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: string[], int[]"))
		case []float64:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: string[], float[]"))
		case []bool:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: string[], bool[]"))
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case []int64:
		switch right.(type) {
		case []string:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: int[], string[]"))
		case string:
			switch operatorType {
			case PLUS:
				parsed, err := strconv.ParseInt(right.(string), 10, 64)
				if err != nil {
					i.error(operatorToken, fmt.Sprintf("Cannot convert string to int: %v", right))
				}
				return append(left.([]int64), parsed)
			default:
				i.error(operatorToken, "Invalid binary operator for int[], string")
			}
		case int64:
			switch operatorType {
			case PLUS:
				return append(left.([]int64), right.(int64))
			default:
				i.error(operatorToken, "Invalid binary operator for int[], int")
			}
		case float64:
			switch operatorType {
			case PLUS:
				return append(left.([]int64), int64(right.(float64)))
			default:
				i.error(operatorToken, "Invalid binary operator for int[], float")
			}
		case []interface{}:
			arr, ok := AsIntArray(right.([]interface{}))
			if !ok {
				i.error(operatorToken, "Cannot join two arrays of different types: int[], mixed array")
			}
			return append(left.([]int64), arr...)
		case []int64:
			switch operatorType {
			case PLUS:
				return append(left.([]int64), right.([]int64)...)
			default:
				i.error(operatorToken, "Invalid binary operator for int[], int[]")
			}
		case []float64:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: int[], float[]"))
		case []bool:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: int[], bool[]"))
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
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
		case int64:
			switch operatorType {
			case PLUS:
				return append(left.([]float64), float64(right.(int64)))
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
		case []interface{}:
			arr, ok := AsFloatArray(right.([]interface{}))
			if !ok {
				i.error(operatorToken, "Cannot join two arrays of different types: float[], mixed array")
			}
			return append(left.([]float64), arr...)
		case []int64:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: float[], int[]"))
		case []float64:
			switch operatorType {
			case PLUS:
				return append(left.([]float64), right.([]float64)...)
			default:
				i.error(operatorToken, "Invalid binary operator for float[], float[]")
			}
		case []bool:
			i.error(operatorToken, fmt.Sprintf("Cannot join two arrays of different types: float[], bool[]"))
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case []bool:
		i.error(operatorToken, "bool[] operations not yet supported") // todo support bool[]
	case []interface{}:
		switch right.(type) {
		case []string:
			switch operatorType {
			case PLUS:
				array, _ := AsMixedArray(right.([]string))
				return append(left.([]interface{}), array...)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, string[]")
			}
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
		case []int64:
			switch operatorType {
			case PLUS:
				array, _ := AsMixedArray(right.([]int64))
				return append(left.([]interface{}), array...)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, int[]")
			}
		case []float64:
			switch operatorType {
			case PLUS:
				array, _ := AsMixedArray(right.([]float64))
				return append(left.([]interface{}), array...)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, float[]")
			}
		case []bool:
			switch operatorType {
			case PLUS:
				array, _ := AsMixedArray(right.([]bool))
				return append(left.([]interface{}), array...)
			default:
				i.error(operatorToken, "Invalid binary operator for mixed array, bool[]")
			}
		default:
			i.error(operatorToken, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	}
	panic(UNREACHABLE)
}
