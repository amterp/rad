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

func (l *LiteralInterpreter) VisitStringLiteralLiteral(literal StringLiteral) interface{} {
	if l.ShouldInterpolate && l.i != nil {
		return l.performStringInterpolation(literal)
	} else {
		fullString := literal.Value[len(literal.Value)-1].FullStringLiteral
		return NewRslString(fullString)
	}
}

func (l *LiteralInterpreter) VisitIntLiteralLiteral(literal IntLiteral) interface{} {
	if literal.IsNegative {
		return -literal.Value.Literal
	}
	return literal.Value.Literal
}

func (l *LiteralInterpreter) VisitFloatLiteralLiteral(literal FloatLiteral) interface{} {
	if literal.IsNegative {
		return -literal.Value.Literal
	}
	return literal.Value.Literal
}

func (l *LiteralInterpreter) VisitBoolLiteralLiteral(literal BoolLiteral) interface{} {
	return literal.Value.Literal
}

func (l *LiteralInterpreter) VisitMixedArrayLiteralArrayLiteral(literal MixedArrayLiteral) interface{} {
	var values []interface{}
	for _, v := range literal.Values {
		values = append(values, v.Accept(l))
	}
	return values
}

func (l *LiteralInterpreter) VisitLoaLiteralLiteralOrArray(literal LoaLiteral) interface{} {
	return literal.Value.Accept(l)
}

func (l *LiteralInterpreter) VisitLoaArrayLiteralOrArray(array LoaArray) interface{} {
	return array.Value.Accept(l)
}

func (l *LiteralInterpreter) VisitIdentifierLiteralLiteral(literal IdentifierLiteral) interface{} {
	return NewRslString(literal.Tkn.GetLexeme())
}

func (l *LiteralInterpreter) VisitSyntheticIntLiteral(val SyntheticInt) interface{} {
	return val.Val
}
