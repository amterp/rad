package rl

import "fmt"

// first check Val, then Type
type CompatSubject struct {
	Val  interface{} // Specific int64, float64, string, bool
	Type *RadType
}

func NewSubject(val interface{}) CompatSubject {
	switch coerced := val.(type) {
	case int64:
		return NewIntSubject(coerced)
	case float64:
		return NewFloatSubject(coerced)
	case string:
		return NewStrSubject(coerced)
	case bool:
		return NewBoolSubject(coerced)
	default:
		panic(fmt.Sprintf("Unhandled type for CompatSubject: %v", coerced))
	}
}

func NewIntSubject(val int64) CompatSubject {
	t := RadIntT
	return CompatSubject{
		Val:  val,
		Type: &t,
	}
}

func NewFloatSubject(val float64) CompatSubject {
	t := RadFloatT
	return CompatSubject{
		Val:  val,
		Type: &t,
	}
}

func NewStrSubject(val string) CompatSubject {
	t := RadStrT
	return CompatSubject{
		Val:  val,
		Type: &t,
	}
}

func NewBoolSubject(val bool) CompatSubject {
	t := RadBoolT
	return CompatSubject{
		Val:  val,
		Type: &t,
	}
}

func NewListSubject() CompatSubject {
	t := RadListT
	return CompatSubject{
		Type: &t,
	}
}

func NewMapSubject() CompatSubject {
	t := RadMapT
	return CompatSubject{
		Type: &t,
	}
}

func NewFnSubject() CompatSubject {
	t := RadFnT
	return CompatSubject{
		Type: &t,
	}
}

func NewNullSubject() CompatSubject {
	t := RadNullT
	return CompatSubject{
		Type: &t,
	}
}

func NewErrorSubject() CompatSubject {
	t := RadErrorT
	return CompatSubject{
		Type: &t,
	}
}

func NewVoidSubject() CompatSubject {
	return CompatSubject{}
}
