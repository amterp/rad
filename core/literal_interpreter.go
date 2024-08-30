package core

type LiteralInterpreter struct{}

func NewLiteralInterpreter() *LiteralInterpreter {
	return &LiteralInterpreter{}
}

func (l LiteralInterpreter) VisitStringLiteralLiteral(literal StringLiteral) interface{} {
	return literal.Value.(*StringLiteralToken).Literal
}

func (l LiteralInterpreter) VisitIntLiteralLiteral(literal IntLiteral) interface{} {
	return literal.Value.(*IntLiteralToken).Literal
}

func (l LiteralInterpreter) VisitFloatLiteralLiteral(literal FloatLiteral) interface{} {
	return literal.Value.(*FloatLiteralToken).Literal
}

func (l LiteralInterpreter) VisitBoolLiteralLiteral(literal BoolLiteral) interface{} {
	return literal.Value.(*BoolLiteralToken).Literal
}

func (l LiteralInterpreter) VisitNullLiteralLiteral(NullLiteral) interface{} {
	return nil
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

func (l LiteralInterpreter) VisitEmptyArrayLiteralArrayLiteral(literal EmptyArrayLiteral) interface{} {
	return []interface{}{}
}

func (l LiteralInterpreter) VisitLoaLiteralLiteralOrArray(literal LoaLiteral) interface{} {
	return literal.Value.Accept(l)
}

func (l LiteralInterpreter) VisitLoaArrayLiteralOrArray(array LoaArray) interface{} {
	return array.Value.Accept(l)
}
