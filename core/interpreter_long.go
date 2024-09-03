package core

import (
	"fmt"
	"strconv"
)

func (i *MainInterpreter) VisitBinaryExpr(binary Binary) interface{} {
	left := binary.Left.Accept(i)
	right := binary.Right.Accept(i)

	switch left.(type) {
	case int:
		switch right.(type) {
		case int:
			switch binary.Operator.GetType() {
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
				i.error(binary.Operator, "Invalid binary operator for int, int")
			}
		case float64:
			switch binary.Operator.GetType() {
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
				i.error(binary.Operator, "Invalid binary operator for int, float")
			}
		default:
			i.error(binary.Operator, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case float64:
		switch right.(type) {
		case int, float64:
			switch binary.Operator.GetType() {
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
				i.error(binary.Operator, "Invalid binary operator for int, int")
			}
		default:
			i.error(binary.Operator, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case string:
		switch right.(type) {
		case string:
			switch binary.Operator.GetType() {
			case PLUS:
				return left.(string) + right.(string)
			case EQUAL_EQUAL:
				return left.(string) == right.(string)
			case NOT_EQUAL:
				return left.(string) != right.(string)
			default:
				i.error(binary.Operator, "Invalid binary operator for string, string")
			}
		case int:
			switch binary.Operator.GetType() {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(int))
			default:
				i.error(binary.Operator, "Invalid binary operator for string, int")
			}
		case float64:
			switch binary.Operator.GetType() {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(float64)) // todo check formatting
			default:
				i.error(binary.Operator, "Invalid binary operator for string, float")
			}
		case bool:
			switch binary.Operator.GetType() {
			case PLUS:
				return left.(string) + fmt.Sprintf("%v", right.(bool))
			default:
				i.error(binary.Operator, "Invalid binary operator for string, bool")
			}
		}
	case bool:
		i.error(binary.Operator, fmt.Sprintf("Invalid binary operator for bool: %v", right))
	case []string:
		switch right.(type) {
		case []string:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]string), right.([]string)...)
			default:
				i.error(binary.Operator, "Invalid binary operator for string[], string[]")
			}
		case string:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]string), right.(string))
			default:
				i.error(binary.Operator, "Invalid binary operator for string[], string")
			}
		case int:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(int)))
			default:
				i.error(binary.Operator, "Invalid binary operator for string[], int")
			}
		case float64:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(float64)))
			default:
				i.error(binary.Operator, "Invalid binary operator for string[], float")
			}
		case bool:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]string), fmt.Sprintf("%v", right.(bool)))
			default:
				i.error(binary.Operator, "Invalid binary operator for string[], bool")
			}
		case []int:
			i.error(binary.Operator, fmt.Sprintf("Cannot join two arrays of different types: string[], int[]"))
		case []float64:
			i.error(binary.Operator, fmt.Sprintf("Cannot join two arrays of different types: string[], float[]"))
		default:
			i.error(binary.Operator, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case []int:
		switch right.(type) {
		case []string:
			i.error(binary.Operator, fmt.Sprintf("Cannot join two arrays of different types: int[], string[]"))
		case string:
			switch binary.Operator.GetType() {
			case PLUS:
				parsed, err := strconv.Atoi(right.(string))
				if err != nil {
					i.error(binary.Operator, fmt.Sprintf("Cannot convert string to int: %v", right))
				}
				return append(left.([]int), parsed)
			default:
				i.error(binary.Operator, "Invalid binary operator for int[], string")
			}
		case int:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]int), right.(int))
			default:
				i.error(binary.Operator, "Invalid binary operator for int[], int")
			}
		case float64:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]int), int(right.(float64)))
			default:
				i.error(binary.Operator, "Invalid binary operator for int[], float")
			}
		case []int:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]int), right.([]int)...)
			default:
				i.error(binary.Operator, "Invalid binary operator for int[], int[]")
			}
		case []float64:
			i.error(binary.Operator, fmt.Sprintf("Cannot join two arrays of different types: int[], float[]"))
		default:
			i.error(binary.Operator, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	case []float64:
		switch right.(type) {
		case []string:
			i.error(binary.Operator, fmt.Sprintf("Cannot join two arrays of different types: float[], string[]"))
		case string:
			switch binary.Operator.GetType() {
			case PLUS:
				parsed, err := strconv.ParseFloat(right.(string), 64)
				if err != nil {
					i.error(binary.Operator, fmt.Sprintf("Cannot convert string to float: %v", right))
				}
				return append(left.([]float64), parsed)
			default:
				i.error(binary.Operator, "Invalid binary operator for float[], string")
			}
		case int:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]float64), float64(right.(int)))
			default:
				i.error(binary.Operator, "Invalid binary operator for float[], int")
			}
		case float64:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]float64), right.(float64))
			default:
				i.error(binary.Operator, "Invalid binary operator for float[], float")
			}
		case []int:
			i.error(binary.Operator, fmt.Sprintf("Cannot join two arrays of different types: float[], int[]"))
		case []float64:
			switch binary.Operator.GetType() {
			case PLUS:
				return append(left.([]float64), right.([]float64)...)
			default:
				i.error(binary.Operator, "Invalid binary operator for float[], float[]")
			}
		default:
			i.error(binary.Operator, fmt.Sprintf("Invalid binary operands: %v %v", left, right))
		}
	}
	panic(UNREACHABLE)
}
