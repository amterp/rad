package rl

import (
	"fmt"
)

// Actual interpreter types
type RadType int

const (
	RadStrT RadType = iota
	RadIntT
	RadFloatT
	RadBoolT
	RadListT
	RadMapT
	RadFnT
	RadNullT
	RadErrorT
)

func (r RadType) AsString() string {
	switch r {
	case RadStrT:
		return T_STR
	case RadIntT:
		return T_INT
	case RadFloatT:
		return T_FLOAT
	case RadBoolT:
		return T_BOOL
	case RadListT:
		return T_LIST
	case RadMapT:
		return T_MAP
	case RadFnT:
		return "function"
	case RadNullT:
		return "null"
	case RadErrorT:
		return T_ERROR
	default:
		panic(fmt.Sprintf("Bug! Unhandled Rad type: %v", r))
	}
}
