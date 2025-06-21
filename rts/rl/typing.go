package rl

import (
	"fmt"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
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

type TypingT interface {
	Name() string
	IsCompatibleWith(val TypingCompatVal) bool
}

// Primitives / Union Primitives
type TypingStrT struct{} // var: str

func NewStrType() *TypingStrT {
	return &TypingStrT{}
}

func (t *TypingStrT) Name() string {
	return T_STR
}

func (t *TypingStrT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		_, ok := (val.Val).(string)
		return ok
	}
	if val.Type != nil {
		return *val.Type == RadStrT
	}
	return false
}

type TypingIntT struct{} // var: int

func NewIntType() *TypingIntT {
	return &TypingIntT{}
}

func (t *TypingIntT) Name() string {
	return T_INT
}

func (t *TypingIntT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		_, ok := (val.Val).(int64)
		return ok
	}
	if val.Type != nil {
		return *val.Type == RadIntT
	}
	return false
}

type TypingFloatT struct{} // var: float

func NewFloatType() *TypingFloatT {
	return &TypingFloatT{}
}

func (t *TypingFloatT) Name() string {
	return T_FLOAT
}

func (t *TypingFloatT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		switch val.Val.(type) {
		case float64, int64:
			return true
		}
	}
	if val.Type != nil {
		return *val.Type == RadFloatT || *val.Type == RadIntT
	}
	return false
}

type TypingBoolT struct{} // var: bool

func NewBoolType() *TypingBoolT {
	return &TypingBoolT{}
}

func (t *TypingBoolT) Name() string {
	return T_BOOL
}

func (t *TypingBoolT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		_, ok := (val.Val).(bool)
		return ok
	}
	if val.Type != nil {
		return *val.Type == RadBoolT
	}
	return false
}

type TypingErrorT struct{} // var: error

func NewErrorType() *TypingErrorT {
	return &TypingErrorT{}
}

func (t *TypingErrorT) Name() string {
	return T_ERROR
}

func (t *TypingErrorT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Type != nil {
		return *val.Type == RadErrorT
	}
	return false
}

type TypingAnyT struct{} // var: any

func NewAnyType() *TypingAnyT {
	return &TypingAnyT{}
}

func (t *TypingAnyT) Name() string {
	return T_ANY
}

func (t *TypingAnyT) IsCompatibleWith(TypingCompatVal) bool {
	return true
}

// TODO REMOVE THIS, UNNECESSARY, FLOAT IS ENOUGH
type TypingNumT struct{} // var: num i.e. int | float

func NewNumType() *TypingNumT {
	return &TypingNumT{}
}

func (t *TypingNumT) Name() string {
	return T_NUM
}

func (t *TypingNumT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		switch (val.Val).(type) {
		case int64, float64:
			return true
		default:
			return false
		}
	}
	if val.Type != nil {
		return *val.Type == RadIntT || *val.Type == RadFloatT
	}
	return false
}

type TypingVoidT struct{} // -> void

func NewVoidType() *TypingVoidT {
	return &TypingVoidT{}
}

func (t *TypingVoidT) Name() string {
	return T_VOID
}

func (t *TypingVoidT) IsCompatibleWith(val TypingCompatVal) bool {
	// Void is a special case, it means no return value.
	if val.Val == nil && val.Type == nil {
		return true
	}
	return false
}

// Collections
type TypingAnyListT struct{} // var: list i.e. [*any]

func NewAnyListType() *TypingAnyListT {
	return &TypingAnyListT{}
}

func (t *TypingAnyListT) Name() string {
	return T_LIST
}

func (t *TypingAnyListT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Type != nil {
		return *val.Type == RadListT
	}
	return false
}

type TypingListT struct { // var: [*int] OR [str, int] i.e. tuple
	types []TypingT
}

func NewListType(types ...TypingT) *TypingListT {
	return &TypingListT{
		types: types,
	}
}

func (t *TypingListT) Name() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, typ := range t.types {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(typ.Name())
	}
	sb.WriteString("]")
	return sb.String()
}

func (t *TypingListT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		switch coerced := (val.Val).(type) {
		case []interface{}:

			// TODO INCORRECT, NEEDS TO SUPPORT E.G. [*int]
			if len(coerced) != len(t.types) {
				return false
			}
			for i, elem := range coerced {
				if !t.types[i].IsCompatibleWith(NewSubject(elem)) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
	if val.Type != nil {
		return *val.Type == RadListT
	}
	return false
}

type TypingAnyMapT struct{} // var: map i.e. { any: any }
type TypingMapT struct {    // var: { string: int } OR { "mykey": int, "mykey2"?: float }
	keyT  TypingT
	valT  TypingT
	named map[MapNamedKey]TypingT
}

func NewAnyMapType() *TypingAnyMapT {
	return &TypingAnyMapT{}
}

func (t *TypingAnyMapT) Name() string {
	return T_MAP
}

func (t *TypingAnyMapT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Type != nil {
		return *val.Type == RadMapT
	}
	return false
}

// Modifiers
type TypingVarArgT struct { // var: *int
	t TypingT
}

func NewVarArgType(t TypingT) *TypingVarArgT {
	return &TypingVarArgT{
		t: t,
	}
}

func (t *TypingVarArgT) Name() string {
	return fmt.Sprintf("*%s", t.t.Name())
}

func (t *TypingVarArgT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		switch coerced := (val.Val).(type) {
		case []interface{}:
			for _, elem := range coerced {
				if !t.t.IsCompatibleWith(NewSubject(elem)) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
	if val.Type != nil {
		return *val.Type == RadListT
	}
	return false
}

type TypingOptionalT struct { // var: int?
	t TypingT
}

func NewOptionalType(t TypingT) *TypingOptionalT {
	return &TypingOptionalT{
		t: t,
	}
}

func (t *TypingOptionalT) Name() string {
	return fmt.Sprintf("%s?", t.t.Name())
}

func (t *TypingOptionalT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		return t.t.IsCompatibleWith(val)
	}
	if val.Type != nil {
		return *val.Type == RadNullT || t.t.IsCompatibleWith(val)
	}
	return false
}

// Complex / Unions / Misc
type TypingFnT struct { // var: fn(int, int) -> int
	Params  []TypingFnParam
	ReturnT *TypingT
	// nils mean no typing
}

type TypingStrEnumT struct { // var: ["foo", "bar"] = "bar"
	values []string
}

func (t *TypingStrEnumT) Name() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range t.values {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", v))
	}
	sb.WriteString("]")
	return sb.String()
}

func (t *TypingStrEnumT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		strVal, ok := (val.Val).(string)
		if !ok {
			return false
		}
		for _, enumVal := range t.values {
			if strVal == enumVal {
				return true
			}
		}
		return false
	}
	if val.Type != nil {
		// best effort, we can't know for sure cause value is not provided
		return *val.Type == RadStrT
	}
	return false
}

func (t *TypingFnT) ByName() map[string]TypingFnParam {
	name := make(map[string]TypingFnParam)
	for _, param := range t.Params {
		name[param.Name] = param
	}
	return name
}

type TypingUnionT struct { // var: int | float
	types []TypingT
}

func NewUnionType(types ...TypingT) *TypingUnionT {
	return &TypingUnionT{
		types: types,
	}
}

func (t *TypingUnionT) Name() string {
	var sb strings.Builder
	for i, typ := range t.types {
		if i > 0 {
			sb.WriteString("|")
		}
		sb.WriteString(typ.Name())
	}
	return sb.String()
}

func (t *TypingUnionT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		for _, typ := range t.types {
			if typ.IsCompatibleWith(val) {
				return true
			}
		}
		return false
	}
	if val.Type != nil {
		for _, typ := range t.types {
			if typ.IsCompatibleWith(val) {
				return true
			}
		}
		return false
	}
	return false
}

type MapNamedKey struct {
	Name       string
	IsOptional bool
}

type TypingFnParam struct {
	Name       string
	Type       *TypingT
	NamedOnly  bool // if true, can only be passed as a named arg
	IsOptional bool
	Default    *ts.Node // if no default, this is nil
}

func (t TypingFnParam) AnonymousOnly() bool {
	return strings.HasPrefix(t.Name, "_")
}
