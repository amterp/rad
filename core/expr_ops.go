package core

import (
	"fmt"
	"math"
	com "rad/core/common"
	"strings"

	"github.com/amterp/rts/rsl"

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
	case "and":
		return OP_AND
	case "or":
		return OP_OR
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
		case rsl.K_PLUS_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_PLUS)
		case rsl.K_MINUS_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_MINUS)
		case rsl.K_STAR_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_MULTIPLY)
		case rsl.K_SLASH_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_DIVIDE)
		case rsl.K_PERCENT_EQUAL:
			return i.executeOp(parentNode, left, right, opNode, OP_MODULO)
		default:
			i.errorf(opNode, "Invalid compound operator")
			panic(UNREACHABLE)
		}
	}()
	return newRslValue(i, parentNode, result)
}

func (i *Interpreter) executeUnaryOp(parentNode, argNode, opNode *ts.Node) RslValue {
	switch opNode.Kind() {
	case rsl.K_PLUS, rsl.K_MINUS, rsl.K_PLUS_PLUS, rsl.K_MINUS_MINUS:
		opStr := i.sd.Src[opNode.StartByte():opNode.EndByte()]
		argVal := i.evaluate(argNode, 1)[0]
		argVal.RequireType(i, argNode, fmt.Sprintf("Invalid operand type '%s' for op '%s'", TypeAsString(argVal), opStr), RslIntT, RslFloatT)

		intOp, floatOp := i.getUnaryOp(opNode)

		switch coerced := argVal.Val.(type) {
		case int64:
			return newRslValue(i, parentNode, intOp(coerced))
		case float64:
			return newRslValue(i, parentNode, floatOp(coerced))
		default:
			i.errorf(parentNode, fmt.Sprintf("Bug! Unhandled type for unary minus: %T", argVal.Val))
			panic(UNREACHABLE)
		}
	case rsl.K_NOT:
		return newRslValue(i, parentNode, !i.evaluate(argNode, 1)[0].TruthyFalsy())
	default:
		i.errorf(opNode, "Invalid unary operator")
		panic(UNREACHABLE)
	}
}

func (i *Interpreter) getUnaryOp(opNode *ts.Node) (func(int64) int64, func(float64) float64) {
	switch opNode.Kind() {
	case rsl.K_PLUS:
		return func(a int64) int64 { return a }, func(a float64) float64 { return a }
	case rsl.K_MINUS:
		return func(a int64) int64 { return -a }, func(a float64) float64 { return -a }
	case rsl.K_PLUS_PLUS:
		return func(a int64) int64 { return a + 1 }, func(a float64) float64 { return a + 1 }
	case rsl.K_MINUS_MINUS:
		return func(a int64) int64 { return a - 1 }, func(a float64) float64 { return a - 1 }
	default:
		i.errorf(opNode, "Invalid unary operator")
		panic(UNREACHABLE)
	}
}

func (i *Interpreter) executeOp(
	parentNode *ts.Node,
	leftNode *ts.Node,
	rightNode *ts.Node,
	opNode *ts.Node,
	op OpType,
) interface{} {
	left := com.Memoize(func() RslValue {
		return i.evaluate(leftNode, 1)[0]
	})
	right := com.Memoize(func() RslValue {
		return i.evaluate(rightNode, 1)[0]
	})

	if op == OP_AND {
		leftB := left().TruthyFalsy()
		if !leftB {
			return false
		}
		return right().TruthyFalsy()
	} else if op == OP_OR {
		leftB := left().TruthyFalsy()
		if leftB {
			// return actual value for falsy coalescing
			return left()
		}
		// return actual value for falsy coalescing
		return right()
	}

	if op == OP_EQUAL || op == OP_NOT_EQUAL {
		leftType := left().Type()
		rightType := right().Type()
		if leftType != rightType && !(leftType == RslFloatT && rightType == RslIntT) && !(leftType == RslIntT && rightType == RslFloatT) {
			// different types are not equal
			// UNLESS they're int/float, in which case we fall through to below and compare there
			return op == OP_NOT_EQUAL
		}
	}

	additionalErrMsg := ""
	leftV := left().Val
	rightV := right().Val
	switch coercedLeft := leftV.(type) {
	case int64:
		switch coercedRight := rightV.(type) {
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
			}
		case RslString:
			switch op {
			// todo python does not allow this, should we?
			case OP_IN:
				return strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			case OP_NOT_IN:
				return !strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(left())
			case OP_NOT_IN:
				return !coercedRight.Contains(left())
			}
		case *RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(left())
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(left())
			}
		}
	case float64:
		switch coercedRight := rightV.(type) {
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
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(left())
			case OP_NOT_IN:
				return !coercedRight.Contains(left())
			}
		case *RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(left())
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(left())
			}
		}
	case RslString:
		switch coercedRight := rightV.(type) {
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
			}
		case int64:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight))
			}
		case float64:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight)) // todo check formatting
			}
		case bool:
			switch op {
			case OP_PLUS:
				return coercedLeft.ConcatStr(fmt.Sprintf("%v", coercedRight))
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(left())
			case OP_NOT_IN:
				return !coercedRight.Contains(left())
			}
		case *RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(left())
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(left())
			}
		}
	case bool:
		switch coercedRight := rightV.(type) {
		case bool:
			switch op {
			case OP_EQUAL:
				return coercedLeft == coercedRight
			case OP_NOT_EQUAL:
				return coercedLeft != coercedRight
			}
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(left())
			case OP_NOT_IN:
				return !coercedRight.Contains(left())
			}
		case *RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(left())
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(left())
			}
		}
	case *RslList:
		switch coercedRight := rightV.(type) {
		case *RslList:
			switch op {
			case OP_PLUS:
				return coercedLeft.JoinWith(coercedRight)
			case OP_IN:
				return coercedLeft.Contains(right())
			case OP_NOT_IN:
				return !coercedLeft.Contains(right())
			}
		}
		switch op {
		case OP_PLUS:
			additionalErrMsg = ". Did you mean to wrap the right side in a list in order to append?"
		}
	case RslNull:
		switch coercedRight := rightV.(type) {
		case *RslList:
			switch op {
			case OP_IN:
				return coercedRight.Contains(left())
			case OP_NOT_IN:
				return !coercedRight.Contains(left())
			}
		case *RslMap:
			switch op {
			case OP_IN:
				return coercedRight.ContainsKey(left())
			case OP_NOT_IN:
				return !coercedRight.ContainsKey(left())
			}
		case RslNull:
			switch op {
			case OP_EQUAL:
				return true
			case OP_NOT_EQUAL:
				return false
			}
		}
	}

	opSrc := GetSrc(i.sd.Src, opNode)

	i.errorf(parentNode, "Invalid operand types: cannot do '%s %s %s'%s",
		TypeAsString(leftV), opSrc, TypeAsString(rightV), additionalErrMsg)
	panic(UNREACHABLE)
}
