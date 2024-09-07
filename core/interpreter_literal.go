package core

type LiteralInterpreter struct {
	i                 *MainInterpreter
	ShouldInterpolate bool
}

func NewLiteralInterpreter(i *MainInterpreter) *LiteralInterpreter {
	return &LiteralInterpreter{
		i:                 i,
		ShouldInterpolate: true,
	}
}

func (l LiteralInterpreter) VisitStringLiteralLiteral(literal StringLiteral) interface{} {
	stringLiteral := literal.Value.Literal
	if l.ShouldInterpolate && l.i != nil {
		return performStringInterpolation(stringLiteral, l.i.env)
	}
	return stringLiteral
}

func (l LiteralInterpreter) VisitIntLiteralLiteral(literal IntLiteral) interface{} {
	return literal.Value.Literal
}

func (l LiteralInterpreter) VisitFloatLiteralLiteral(literal FloatLiteral) interface{} {
	return literal.Value.Literal
}

func (l LiteralInterpreter) VisitBoolLiteralLiteral(literal BoolLiteral) interface{} {
	return literal.Value.Literal
}

func (l LiteralInterpreter) VisitStringArrayLiteralArrayLiteral(literal StringArrayLiteral) interface{} {
	var values []string
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(string))
	}
	return values
}

func (l LiteralInterpreter) VisitIntArrayLiteralArrayLiteral(literal IntArrayLiteral) interface{} {
	var values []int
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(int))
	}
	return values
}

func (l LiteralInterpreter) VisitFloatArrayLiteralArrayLiteral(literal FloatArrayLiteral) interface{} {
	var values []float64
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(float64))
	}
	return values
}

func (l LiteralInterpreter) VisitBoolArrayLiteralArrayLiteral(literal BoolArrayLiteral) interface{} {
	var values []bool
	for _, v := range literal.Values {
		values = append(values, v.Accept(l).(bool))
	}
	return values
}

func (l LiteralInterpreter) VisitEmptyArrayLiteralArrayLiteral(EmptyArrayLiteral) interface{} {
	return []interface{}{}
}

func (l LiteralInterpreter) VisitLoaLiteralLiteralOrArray(literal LoaLiteral) interface{} {
	return literal.Value.Accept(l)
}

func (l LiteralInterpreter) VisitLoaArrayLiteralOrArray(array LoaArray) interface{} {
	return array.Value.Accept(l)
}
