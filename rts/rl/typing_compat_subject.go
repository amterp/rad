package rl

import "fmt"

// first check Val, then Type
type TypingCompatVal struct {
	Val  interface{} // Specific int64, float64, string, bool, []interface{}, map[string]interface{}
	Type *RadType
}

func NewSubject(val interface{}) TypingCompatVal {
	switch coerced := val.(type) {
	case nil:
		// RadNull values flatten to nil via core's ToGoValue.
		return NewNullSubject()
	case int64:
		return NewIntSubject(coerced)
	case float64:
		return NewFloatSubject(coerced)
	case string:
		return NewStrSubject(coerced)
	case bool:
		return NewBoolSubject(coerced)
	case []interface{}:
		s := NewListSubject()
		s.Val = coerced
		return s
	case map[string]interface{}:
		s := NewMapSubject()
		s.Val = coerced
		return s
	case FnGoValue:
		// Function values inside collections come through as this sentinel so
		// the rl package doesn't have to import core.RadFn.
		return NewFnSubject()
	default:
		panic(fmt.Sprintf("Unhandled type for TypingCompatVal: %T", coerced))
	}
}

// FnGoValue is a sentinel used by core's ToGoValue when flattening a function
// value into a generic []interface{} / map[string]interface{}. NewSubject maps
// it back to NewFnSubject(). It lives here (not in core) because the rl
// package cannot import core where RadFn is defined.
type FnGoValue struct{}

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
