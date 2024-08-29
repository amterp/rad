package interpreters

import (
	"rad/core"
)

type LiteralInterpreter struct{}

func (l *LiteralInterpreter) VisitStringLiteralLiteral(literal *core.StringLiteral) interface{} {
	return &literal.Value.(*core.StringLiteralToken).Literal
}

func (l *LiteralInterpreter) VisitIntLiteralLiteral(literal *core.IntLiteral) interface{} {
	return &literal.Value.(*core.IntLiteralToken).Literal
}

func (l *LiteralInterpreter) VisitFloatLiteralLiteral(literal *core.FloatLiteral) interface{} {
	return &literal.Value.(*core.FloatLiteralToken).Literal
}

func (l *LiteralInterpreter) VisitBoolLiteralLiteral(literal *core.BoolLiteral) interface{} {
	return &literal.Value.(*core.BoolLiteralToken).Literal
}

func (l *LiteralInterpreter) VisitStringArrayLiteralArrayLiteral(literal *core.StringArrayLiteral) interface{} {
	var values []string
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(string))
	}
	return values
}

func (l *LiteralInterpreter) VisitIntArrayLiteralArrayLiteral(literal *core.IntArrayLiteral) interface{} {
	var values []int
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(int))
	}
	return values
}

func (l *LiteralInterpreter) VisitFloatArrayLiteralArrayLiteral(literal *core.FloatArrayLiteral) interface{} {
	var values []float64
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(float64))
	}
	return values
}

func (l *LiteralInterpreter) VisitBoolArrayLiteralArrayLiteral(literal *core.BoolArrayLiteral) interface{} {
	var values []bool
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(bool))
	}
	return values
}

func (l *LiteralInterpreter) VisitHolderLiteralOrArray(holder *core.LiteralOrArrayHolder) interface{} {
	if holder.ArrayVal != nil {
		return holder.ArrayVal.Accept(l)
	} else {
		return holder.LiteralVal.Accept(l)
	}
}
