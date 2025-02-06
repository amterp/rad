package core

import (
	"fmt"
	"math"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

func getOp(str string) OpType {
	switch str {
	case "+":
		return OP_PLUS
	case "-":
		return OP_MINUS
	case "*":
		return OP_MULTIPLY
	case "/":
		return OP_DIVIDE
	case "%":
		return OP_MODULO
	case ">":
		return OP_GREATER
	case ">=":
		return OP_GREATER_EQUAL
	case "<":
		return OP_LESS
	case "<=":
		return OP_LESS_EQUAL
	case "==":
		return OP_EQUAL
	case "!=":
		return OP_NOT_EQUAL
	case "in":
		return OP_IN
	case "not in": // todo needs to work for if extra spaces between 'not in'
		return OP_NOT_IN
	default:
		panic("Bug! Unexpected operator: " + str)
	}
}

func (i *Interpreter) executeBinary(parentNode, leftNode, rightNode, opNode *ts.Node) RslValue {
	opStr := i.sd.Src[opNode.StartByte():opNode.EndByte()]
	op := getOp(opStr)
	return newRslValue(i, parentNode, i.executeOp(parentNode, leftNode, rightNode, opNode, op))
}

func (i *Interpreter) executeCompoundOp(parentNode, left, right, opNode *ts.Node) RslValue {
	result := func() interface{} {
		switch opNode.Kind() {
		case K_PLUS_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_PLUS)
		case K_MINUS_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_MINUS)
		case K_STAR_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_MULTIPLY)
		case K_SLASH_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_DIVIDE)
		case K_PERCENT_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_MODULO)
		default:
			i.errorf(opNode, "Invalid compound operator")
			panic(UNREACHABLE)
		}
	}()
	return newRslValue(i, parentNode, result)
}

func (i *Interpreter) executeUnary(opNode, argNode *ts.Node) RslValue {
	// todo
	return RslValue{}
}

func (i *Interpreter) executeOp(
	parentNode *ts.Node,
	leftNode *ts.Node,
	rightNode *ts.Node,
	opNode *ts.Node,
	op OpType,
) interface{} {
	left := Memoize(func() RslValue {
		return i.evaluate(leftNode, 1)[0]
	})
	right := Memoize(func() RslValue {
		return i.evaluate(rightNode, 1)[0]
	})

	if op == OP_EQUAL || op == OP_NOT_EQUAL {
		leftType := left().Type()
		rightType := right().Type()
		if leftType != rightType && !(leftType == RslFloatT && rightType == RslIntT) && !(leftType == RslIntT && rightType == RslFloatT) {
			// different types are not equal
			// UNLESS they're int/float, in which case we fall through to below and compare there
			return op == OP_NOT_EQUAL
		}
	}

	switch coercedLeft := left().Val.(type) {
	case int64:
		switch coercedRight := right().Val.(type) {
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
					i.errorf(rightNode, "Divisor was 0, cannot divide by 0")
				}
				return float64(coercedLeft) / float64(coercedRight)
			case OP_MODULO:
				if coercedRight == 0 {
					i.errorf(rightNode, "Value is 0, cannot modulo by 0")
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
				i.errorf(opNode, "Invalid binary operator for int, int")
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
					i.errorf(rightNode, "Divisor was 0, cannot divide by 0")
				}
				return float64(coercedLeft) / coercedRight
			case OP_MODULO:
				if coercedRight == 0 {
					i.errorf(rightNode, "Value is 0, cannot modulo by 0")
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
				i.errorf(opNode, "Invalid binary operator for int, float")
			}
		case RslString:
			switch op {
			// todo python does not allow this, should we?
			case OP_IN:
				return strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			case OP_NOT_IN:
				return !strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			default:
				i.errorf(opNode, "Invalid binary operator for int, string")
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(coercedLeft)
			case OP_NOT_IN:
				return !coercedRight.Contains(coercedLeft)
			default:
				i.errorf(opNode, "Invalid binary operator for int, list")
			}
		default:
			i.errorf(parentNode, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case float64:
		switch coercedRight := right().Val.(type) {
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
					i.errorf(rightNode, "Divisor was 0, cannot divide by 0")
				}
				return coercedLeft / float64(coercedRight)
			case OP_MODULO:
				if coercedRight == 0 {
					i.errorf(rightNode, "Value is 0, cannot modulo by 0")
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
				i.errorf(opNode, "Invalid binary operator for int, int")
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
					i.errorf(rightNode, "Divisor was 0, cannot divide by 0")
				}
				return coercedLeft / coercedRight
			case OP_MODULO:
				if coercedRight == 0 {
					i.errorf(rightNode, "Value is 0, cannot modulo by 0")
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
				i.errorf(opNode, "Invalid binary operator for int, float")
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(coercedLeft)
			case OP_NOT_IN:
				return !coercedRight.Contains(coercedLeft)
			default:
				i.errorf(opNode, "Invalid binary operator for float, list")
			}
		default:
			i.errorf(parentNode, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case RslString:
		switch coercedRight := right().Val.(type) {
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
				i.errorf(opNode, "Invalid binary operator for string, string")
			}
		case int64:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight))
			default:
				i.errorf(opNode, "Invalid binary operator for string, int")
			}
		case float64:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight)) // todo check formatting
			default:
				i.errorf(opNode, "Invalid binary operator for string, float")
			}
		case bool:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight))
			default:
				i.errorf(opNode, "Invalid binary operator for string, bool")
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(coercedLeft)
			case OP_NOT_IN:
				return !coercedRight.Contains(coercedLeft)
			default:
				i.errorf(opNode, "Invalid binary operator for string, list")
			}
		case *RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(left())
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(left())
			default:
				i.errorf(opNode, "Invalid binary operator for string, map")
			}
		default:
			i.errorf(parentNode, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case bool:
		switch coercedRight := right().Val.(type) {
		case bool:
			switch op {
			case OP_EQUAL:
				return coercedLeft == coercedRight
			case OP_NOT_EQUAL:
				return coercedLeft != coercedRight
			default:
				i.errorf(opNode, "Invalid binary operator for bool, bool")
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(coercedLeft)
			case OP_NOT_IN:
				return !coercedRight.Contains(coercedLeft)
			default:
				i.errorf(opNode, "Invalid binary operator for bool, list")
			}
		default:
			i.errorf(parentNode, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	case *RslList:
		switch coercedRight := right().Val.(type) {
		case RslString:
			i.errorf(opNode, "Invalid binary operator for list, string")
		case int64:
			i.errorf(opNode, "Invalid binary operator for list, int")
		case float64:
			i.errorf(opNode, "Invalid binary operator for list, float")
		case bool:
			i.errorf(opNode, "Invalid binary operator for list, bool")
		case *RslList:
			switch op {
			case OP_PLUS:
				return coercedLeft.JoinWith(coercedRight)
			case OP_IN:
				return coercedLeft.Contains(coercedRight)
			case OP_NOT_IN:
				return !coercedLeft.Contains(coercedRight)
			default:
				i.errorf(opNode, "Invalid binary operator for list, list")
			}
		default:
			i.errorf(parentNode, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
		}
	default:
		i.errorf(parentNode, fmt.Sprintf("Invalid binary operand types: %T, %T", left, right))
	}
	panic(UNREACHABLE)
}
