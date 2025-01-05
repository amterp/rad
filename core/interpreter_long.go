package core

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

var (
	FLOAT64_TYPE = reflect.TypeOf(float64(0))
	INT64_TYPE   = reflect.TypeOf(int64(0))
)

func (i *MainInterpreter) VisitBinaryExpr(binary Binary) interface{} {
	left := binary.Left.Accept(i)
	right := binary.Right.Accept(i)
	return i.executeOp(left, right, binary.Tkn, binary.Op)
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
	leftType := reflect.TypeOf(left)
	rightType := reflect.TypeOf(right)

	if (op == OP_EQUAL || op == OP_NOT_EQUAL) &&
		leftType != rightType &&
		!(leftType == FLOAT64_TYPE && rightType == INT64_TYPE) && !(leftType == INT64_TYPE && rightType == FLOAT64_TYPE) {
		// equality check between different types is false. unless the types are int and float, in which case
		// we fall through and can compare
		return op == OP_NOT_EQUAL
	}

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
				if coercedRight == 0 {
					// todo idk if this is what we want to do? should we have a nan concept?
					i.error(tkn, "Cannot divide by 0")
				}
				return float64(coercedLeft) / float64(coercedRight)
			case OP_MODULO:
				if coercedRight == 0 {
					i.error(tkn, "Cannot modulo by 0")
				}
				return coercedLeft % coercedRight
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
				if coercedRight == 0 {
					i.error(tkn, "Cannot divide by 0")
				}
				return float64(coercedLeft) / coercedRight
			case OP_MODULO:
				if coercedRight == 0 {
					i.error(tkn, "Cannot modulo by 0")
				}
				return math.Mod(float64(coercedLeft), coercedRight)
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
			default:
				i.error(tkn, "Invalid binary operator for int, list")
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
				if coercedRight == 0 {
					i.error(tkn, "Cannot divide by 0")
				}
				return coercedLeft / float64(coercedRight)
			case OP_MODULO:
				if coercedRight == 0 {
					i.error(tkn, "Cannot modulo by 0")
				}
				return math.Mod(coercedLeft, float64(coercedRight))
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
				if coercedRight == 0 {
					i.error(tkn, "Cannot divide by 0")
				}
				return coercedLeft / coercedRight
			case OP_MODULO:
				if coercedRight == 0 {
					i.error(tkn, "Cannot modulo by 0")
				}
				return math.Mod(coercedLeft, coercedRight)
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
				i.error(tkn, "Invalid binary operator for int, float")
			}
		case []interface{}:
			switch op {
			case OP_IN:
				return contains(coercedRight, coercedLeft)
			case OP_NOT_IN:
				return !contains(coercedRight, coercedLeft)
			default:
				i.error(tkn, "Invalid binary operator for float, list")
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
				i.error(tkn, "Invalid binary operator for string, list")
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
				i.error(tkn, "Invalid binary operator for bool, list")
			}
		default:
			i.error(tkn, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case []interface{}:
		switch coercedRight := right.(type) {
		case RslString:
			i.error(tkn, "Invalid binary operator for list, string")
		case int64:
			i.error(tkn, "Invalid binary operator for list, int")
		case float64:
			i.error(tkn, "Invalid binary operator for list, float")
		case bool:
			i.error(tkn, "Invalid binary operator for list, bool")
		case []interface{}:
			switch op {
			case OP_PLUS:
				return append(coercedLeft, coercedRight...)
			case OP_IN:
				return contains(coercedLeft, coercedRight)
			case OP_NOT_IN:
				return !contains(coercedLeft, coercedRight)
			default:
				i.error(tkn, "Invalid binary operator for list, list")
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
func contains(list []interface{}, val interface{}) bool {
	if a, ok := val.(RslString); ok {
		for _, v := range list {
			if b, ok := v.(RslString); ok && a.Equals(b) {
				return true
			}
		}
	} else {
		for _, v := range list {
			if v == val {
				return true
			}
		}
	}
	return false
}
