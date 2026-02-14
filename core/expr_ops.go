package core

import (
	"fmt"
	"math"
	"strings"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts/rl"
)

// executeUnaryOp handles unary operators (-, +, not).
func (i *Interpreter) executeUnaryOp(n *rl.OpUnary) RadValue {
	switch n.Op {
	case rl.OpNeg:
		argVal := i.eval(n.Operand).Val
		argVal.RequireType(i, n.Operand,
			fmt.Sprintf("Invalid operand type '%s' for op '-'", TypeAsString(argVal)),
			rl.RadIntT, rl.RadFloatT)
		switch coerced := argVal.Val.(type) {
		case int64:
			return newRadValue(i, n, -coerced)
		case float64:
			return newRadValue(i, n, -coerced)
		default:
			i.emitErrorf(rl.ErrInternalBug, n, "Bug: Unhandled type for unary minus: %T", argVal.Val)
			panic(UNREACHABLE)
		}
	case rl.OpAdd:
		// unary + is identity
		argVal := i.eval(n.Operand).Val
		argVal.RequireType(i, n.Operand,
			fmt.Sprintf("Invalid operand type '%s' for op '+'", TypeAsString(argVal)),
			rl.RadIntT, rl.RadFloatT)
		return newRadValue(i, n, argVal.Val)
	case rl.OpNot:
		return newRadValue(i, n, !i.eval(n.Operand).Val.TruthyFalsy())
	default:
		i.emitErrorf(rl.ErrUnsupportedOperation, n, "Invalid unary operator: %s", n.Op)
		panic(UNREACHABLE)
	}
}

func (i *Interpreter) executeOp(
	parentNode rl.Node,
	leftNode rl.Node,
	rightNode rl.Node,
	op rl.Operator,
	isCompound bool,
) interface{} {
	left := com.Memoize(func() RadValue {
		return i.eval(leftNode).Val
	})
	right := com.Memoize(func() RadValue {
		return i.eval(rightNode).Val
	})

	if op == rl.OpAnd {
		leftB := left().TruthyFalsy()
		if !leftB {
			return false
		}
		return right().TruthyFalsy()
	} else if op == rl.OpOr {
		leftB := left().TruthyFalsy()
		if leftB {
			// return actual value for falsy coalescing
			return left()
		}
		// return actual value for falsy coalescing
		return right()
	}

	// Equality is centralized in RadValue.Equals() so that ==, in, switch,
	// index_of, and deep collection comparison all agree on semantics.
	if op == rl.OpEq || op == rl.OpNeq {
		eq := left().Equals(right())
		return (op == rl.OpEq) == eq
	}

	additionalErrMsg := ""
	leftV := left().Val
	rightV := right().Val
	switch coercedLeft := leftV.(type) {
	case int64:
		switch coercedRight := rightV.(type) {
		case int64:
			switch op {
			case rl.OpAdd:
				return coercedLeft + coercedRight
			case rl.OpSub:
				return coercedLeft - coercedRight
			case rl.OpMul:
				return coercedLeft * coercedRight
			case rl.OpDiv:
				if coercedRight == 0 {
					// todo idk if this is what we want to do? should we have a nan concept?
					i.emitError(rl.ErrDivisionByZero, rightNode, "Divisor was 0, cannot divide by 0")
				}
				return float64(coercedLeft) / float64(coercedRight)
			case rl.OpMod:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Value is 0, cannot modulo by 0")
				}
				return coercedLeft % coercedRight
			case rl.OpGt:
				return coercedLeft > coercedRight
			case rl.OpGte:
				return coercedLeft >= coercedRight
			case rl.OpLt:
				return coercedLeft < coercedRight
			case rl.OpLte:
				return coercedLeft <= coercedRight
			}
		case float64:
			switch op {
			case rl.OpAdd:
				return float64(coercedLeft) + coercedRight
			case rl.OpSub:
				return float64(coercedLeft) - coercedRight
			case rl.OpMul:
				return float64(coercedLeft) * coercedRight
			case rl.OpDiv:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Divisor was 0, cannot divide by 0")
				}
				return float64(coercedLeft) / coercedRight
			case rl.OpMod:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Value is 0, cannot modulo by 0")
				}
				return math.Mod(float64(coercedLeft), coercedRight)
			case rl.OpGt:
				return float64(coercedLeft) > coercedRight
			case rl.OpGte:
				return float64(coercedLeft) >= coercedRight
			case rl.OpLt:
				return float64(coercedLeft) < coercedRight
			case rl.OpLte:
				return float64(coercedLeft) <= coercedRight
			}
		case RadString:
			switch op {
			// todo python does not allow this, should we?
			case rl.OpIn:
				return strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			case rl.OpNotIn:
				return !strings.Contains(coercedRight.Plain(), fmt.Sprintf("%v", coercedLeft))
			case rl.OpMul:
				return coercedRight.Repeat(coercedLeft)
			}
		case *RadList:
			switch op {
			case rl.OpIn:
				return coercedRight.Contains(left())
			case rl.OpNotIn:
				return !coercedRight.Contains(left())
			}
		case *RadMap:
			switch op {
			case rl.OpIn:
				return coercedRight.ContainsKey(left())
			case rl.OpNotIn:
				return !coercedRight.ContainsKey(left())
			}
		}
	case float64:
		switch coercedRight := rightV.(type) {
		case int64:
			switch op {
			case rl.OpAdd:
				return coercedLeft + float64(coercedRight)
			case rl.OpSub:
				return coercedLeft - float64(coercedRight)
			case rl.OpMul:
				return coercedLeft * float64(coercedRight)
			case rl.OpDiv:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Divisor was 0, cannot divide by 0")
				}
				return coercedLeft / float64(coercedRight)
			case rl.OpMod:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Value is 0, cannot modulo by 0")
				}
				return math.Mod(coercedLeft, float64(coercedRight))
			case rl.OpGt:
				return coercedLeft > float64(coercedRight)
			case rl.OpGte:
				return coercedLeft >= float64(coercedRight)
			case rl.OpLt:
				return coercedLeft < float64(coercedRight)
			case rl.OpLte:
				return coercedLeft <= float64(coercedRight)
			}
		case float64:
			switch op {
			case rl.OpAdd:
				return coercedLeft + coercedRight
			case rl.OpSub:
				return coercedLeft - coercedRight
			case rl.OpMul:
				return coercedLeft * coercedRight
			case rl.OpDiv:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Divisor was 0, cannot divide by 0")
				}
				return coercedLeft / coercedRight
			case rl.OpMod:
				if coercedRight == 0 {
					i.emitError(rl.ErrDivisionByZero, rightNode, "Value is 0, cannot modulo by 0")
				}
				return math.Mod(coercedLeft, coercedRight)
			case rl.OpGt:
				return coercedLeft > coercedRight
			case rl.OpGte:
				return coercedLeft >= coercedRight
			case rl.OpLt:
				return coercedLeft < coercedRight
			case rl.OpLte:
				return coercedLeft <= coercedRight
			}
		case *RadList:
			switch op {
			case rl.OpIn:
				return coercedRight.Contains(left())
			case rl.OpNotIn:
				return !coercedRight.Contains(left())
			}
		case *RadMap:
			switch op {
			case rl.OpIn:
				return coercedRight.ContainsKey(left())
			case rl.OpNotIn:
				return !coercedRight.ContainsKey(left())
			}
		}
	case RadString:
		switch coercedRight := rightV.(type) {
		case RadString:
			switch op {
			case rl.OpAdd:
				return coercedLeft.Concat(coercedRight)
			case rl.OpIn:
				return strings.Contains(coercedRight.Plain(), coercedLeft.Plain())
			case rl.OpNotIn:
				return !strings.Contains(coercedRight.Plain(), coercedLeft.Plain())
			}
		case int64:
			switch op {
			case rl.OpMul:
				return coercedLeft.Repeat(coercedRight)
			}
		case *RadList:
			switch op {
			case rl.OpIn:
				return coercedRight.Contains(left())
			case rl.OpNotIn:
				return !coercedRight.Contains(left())
			}
		case *RadMap:
			switch op {
			case rl.OpIn:
				return coercedRight.ContainsKey(left())
			case rl.OpNotIn:
				return !coercedRight.ContainsKey(left())
			}
		case *RadError:
			switch op {
			case rl.OpAdd:
				return coercedLeft.Concat(coercedRight.Msg())
			case rl.OpIn:
				return strings.Contains(coercedRight.Msg().Plain(), coercedLeft.Plain())
			case rl.OpNotIn:
				return !strings.Contains(coercedRight.Msg().Plain(), coercedLeft.Plain())
			}
		}
	case bool:
		switch coercedRight := rightV.(type) {
		case *RadList:
			switch op {
			case rl.OpIn:
				return coercedRight.Contains(left())
			case rl.OpNotIn:
				return !coercedRight.Contains(left())
			}
		case *RadMap:
			switch op {
			case rl.OpIn:
				return coercedRight.ContainsKey(left())
			case rl.OpNotIn:
				return !coercedRight.ContainsKey(left())
			}
		}
	case *RadList:
		switch coercedRight := rightV.(type) {
		case *RadList:
			switch op {
			case rl.OpAdd:
				return coercedLeft.JoinWith(coercedRight)
			case rl.OpIn:
				return coercedLeft.Contains(right())
			case rl.OpNotIn:
				return !coercedLeft.Contains(right())
			}
		}
		switch op {
		case rl.OpAdd:
			additionalErrMsg = ". Did you mean to wrap the right side in a list in order to append?"
		}
	case RadNull:
		switch coercedRight := rightV.(type) {
		case *RadList:
			switch op {
			case rl.OpIn:
				return coercedRight.Contains(left())
			case rl.OpNotIn:
				return !coercedRight.Contains(left())
			}
		case *RadMap:
			switch op {
			case rl.OpIn:
				return coercedRight.ContainsKey(left())
			case rl.OpNotIn:
				return !coercedRight.ContainsKey(left())
			}
		}
	case *RadError:
		switch coercedRight := rightV.(type) {
		case RadString:
			switch op {
			case rl.OpAdd:
				return coercedLeft.Msg().Concat(coercedRight)
			case rl.OpIn:
				return strings.Contains(coercedRight.Plain(), coercedLeft.Msg().Plain())
			case rl.OpNotIn:
				return !strings.Contains(coercedRight.Plain(), coercedLeft.Msg().Plain())
			}
		case *RadError:
			switch op {
			case rl.OpAdd:
				return coercedLeft.Msg().Concat(coercedRight.Msg())
			}
		case *RadList:
			switch op {
			case rl.OpIn:
				return coercedRight.Contains(left())
			case rl.OpNotIn:
				return !coercedRight.Contains(left())
			}
		case *RadMap:
			switch op {
			case rl.OpIn:
				return coercedRight.ContainsKey(left())
			case rl.OpNotIn:
				return !coercedRight.ContainsKey(left())
			}
		}
	}

	opStr := op.String()
	if isCompound {
		opStr += "="
	}
	i.emitErrorf(rl.ErrInvalidTypeForOp, parentNode, "Invalid operand types: cannot do '%s %s %s'%s",
		TypeAsString(leftV), opStr, TypeAsString(rightV), additionalErrMsg)
	panic(UNREACHABLE)
}
