package core

func (i *Interpreter) InitArgs(args []RslArg) {
	env := i.env

	for _, arg := range args {
		switch coerced := arg.(type) {
		case *BoolRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *BoolArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *StringRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *StringArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *IntRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *IntArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		case *FloatRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), coerced.Value))
		case *FloatArrRslArg:
			env.SetVar(coerced.Identifier, newRslValue(i, arg.GetNode(), NewRslListFromGeneric(i, arg.GetNode(), coerced.Value)))
		default:
			i.errorf(arg.GetNode(), "Unsupported arg type, cannot init: %T", arg)
		}
	}
}
