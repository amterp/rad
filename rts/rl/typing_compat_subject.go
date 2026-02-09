package rl

import "fmt"

// first check Val, then Type
type TypingCompatVal struct {
	Val  interface{} // Specific int64, float64, string, bool, []interface{}, map[string]interface{}
	Type *RadType
}

func NewSubject(val interface{}) TypingCompatVal {
	switch coerced := val.(type) {
	case int64:
		return NewIntSubject(coerced)
	case float64:
		return NewFloatSubject(coerced)
	case string:
		return NewStrSubject(coerced)
	case bool:
		return NewBoolSubject(coerced)
	case []interface{}:
		return NewListSubject() // todo should we give value, recurse?
	case map[string]interface{}:
		return NewMapSubject() // todo should we give value, recurse?
	default:
		panic(fmt.Sprintf("Unhandled type for TypingCompatVal: %v", coerced))
	}
}

func NewIntSubject(val int64) TypingCompatVal {
	t := RadIntT
	return TypingCompatVal{
		Val:  val,
		Type: &t,
	}
}

func NewFloatSubject(val float64) TypingCompatVal {
	t := RadFloatT
	return TypingCompatVal{
		Val:  val,
		Type: &t,
	}
}

func NewStrSubject(val string) TypingCompatVal {
	t := RadStrT
	return TypingCompatVal{
		Val:  val,
		Type: &t,
	}
}

func NewBoolSubject(val bool) TypingCompatVal {
	t := RadBoolT
	return TypingCompatVal{
		Val:  val,
		Type: &t,
	}
}

func NewListSubject() TypingCompatVal {
	t := RadListT
	return TypingCompatVal{
		Type: &t,
	}
}

func NewMapSubject() TypingCompatVal {
	t := RadMapT
	return TypingCompatVal{
		Type: &t,
	}
}

func NewFnSubject() TypingCompatVal {
	t := RadFnT
	return TypingCompatVal{
		Type: &t,
	}
}

func NewNullSubject() TypingCompatVal {
	t := RadNullT
	return TypingCompatVal{
		Type: &t,
	}
}

func NewErrorSubject() TypingCompatVal {
	t := RadErrorT
	return TypingCompatVal{
		Type: &t,
	}
}

func NewVoidSubject() TypingCompatVal {
	return TypingCompatVal{}
}
