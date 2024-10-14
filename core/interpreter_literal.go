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

func (l LiteralInterpreter) VisitMixedArrayLiteralArrayLiteral(literal MixedArrayLiteral) interface{} {
	var values []interface{}
	for _, v := range literal.Values {
		values = append(values, v.Accept(l))
	}
	return values
}

func (l LiteralInterpreter) VisitLoaLiteralLiteralOrArray(literal LoaLiteral) interface{} {
	return literal.Value.Accept(l)
}

func (l LiteralInterpreter) VisitLoaArrayLiteralOrArray(array LoaArray) interface{} {
	return array.Value.Accept(l)
}
