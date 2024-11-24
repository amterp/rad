package core

func runColor(i *MainInterpreter, function Token, args []interface{}, clr RslColor) RslString {
	clrStr := clr.String()
	if len(args) != 1 {
		i.error(function, clrStr+"() takes exactly one argument")
	}

	arg := args[0]
	switch coerced := arg.(type) {
	case RslString:
		return coerced.Color(clr)
	default:
		s := NewRslString(ToPrintable(arg))
		s.SetSegmentsColor(clr)
		return s
	}
}
