package rl

import (
	"fmt"
	"strings"
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
		panic(fmt.Sprintf("Bug! Unhandled Rad type in AsString: %v", r))
	}
}

// TypingT is the representation of a Rad type annotation. Each implementation
// provides three things:
//
//	Name()             - the user-visible spelling used in error messages
//	IsCompatibleWith() - whether a runtime value matches this type
//	IsAssignableFrom() - whether a value of another declared type can flow here
//
// IsCompatibleWith is the runtime check; the interpreter calls it at function
// param/return boundaries (see core/type_fn.go). It compares a *value* against
// this type.
//
// IsAssignableFrom is the static check; the type checker calls it when
// matching one declared type against another (assignment RHS vs LHS, argument
// vs parameter, return value vs declared return). It compares two *types*. The
// rules implement gradual typing: `any` and `dynamic` are universally
// compatible in both directions (you can store anything in them, and they can
// flow anywhere); collections are invariant (allowing `int[]` into `(int|str)[]`
// would let the callee write a string into a list the caller still believes is
// int-typed); function parameters are contravariant and returns are covariant;
// `int` widens implicitly to `float` to mirror the runtime rule.
type TypingT interface {
	Name() string
	IsCompatibleWith(val TypingCompatVal) bool
	IsAssignableFrom(other TypingT) bool
}

// Compile-time guards: every concrete typing must satisfy TypingT.
var (
	_ TypingT = (*TypingStrT)(nil)
	_ TypingT = (*TypingIntT)(nil)
	_ TypingT = (*TypingFloatT)(nil)
	_ TypingT = (*TypingBoolT)(nil)
	_ TypingT = (*TypingErrorT)(nil)
	_ TypingT = (*TypingAnyT)(nil)
	_ TypingT = (*TypingDynamicT)(nil)
	_ TypingT = (*TypingNeverT)(nil)
	_ TypingT = (*TypingVoidT)(nil)
	_ TypingT = (*TypingAnyListT)(nil)
	_ TypingT = (*TypingListT)(nil)
	_ TypingT = (*TypingTupleT)(nil)
	_ TypingT = (*TypingAnyMapT)(nil)
	_ TypingT = (*TypingStructT)(nil)
	_ TypingT = (*TypingMapT)(nil)
	_ TypingT = (*TypingVarArgT)(nil)
	_ TypingT = (*TypingOptionalT)(nil)
	_ TypingT = (*TypingFnT)(nil)
	_ TypingT = (*TypingStrEnumT)(nil)
	_ TypingT = (*TypingUnionT)(nil)
)

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

// TypingDynamicT is the implicit-any type assigned by the static checker when
// inference can't pin a type. It is behaviorally identical to TypingAnyT
// (universally compatible in both directions), but tracked separately so that
// a future strict-mode flag can warn on implicit-dynamic flow into typed code
// without nagging users who wrote `any` deliberately. Users never write
// `dynamic` themselves - it appears only in inferred types.
type TypingDynamicT struct{}

func NewDynamicType() *TypingDynamicT {
	return &TypingDynamicT{}
}

func (t *TypingDynamicT) Name() string {
	return T_DYNAMIC
}

func (t *TypingDynamicT) IsCompatibleWith(TypingCompatVal) bool {
	return true
}

// TypingNeverT is the bottom type. No value inhabits it at runtime; the
// static checker synthesizes it when narrowing has eliminated every variant
// of a type (e.g. a switch over a string-enum that handles every literal
// leaves `Never` as the residual). Users never write `never` themselves.
//
// As a subtype of everything, Never is assignable into any slot - that's
// what makes the "you exhausted the switch" property compose naturally with
// the rest of the type checker. But nothing except Never is assignable TO
// Never; assigning a real value into a Never-typed slot signals a soundness
// issue (the checker thought a branch was unreachable, but the user reached
// it anyway).
type TypingNeverT struct{}

func NewNeverType() *TypingNeverT {
	return &TypingNeverT{}
}

func (t *TypingNeverT) Name() string {
	return T_NEVER
}

// IsCompatibleWith returns false: no value has type Never at runtime. This
// matters when the static type happens to flow into a function call boundary
// check - any actual value will fail the compatibility test, which is the
// correct outcome since reaching such a call would mean narrowing was wrong.
func (t *TypingNeverT) IsCompatibleWith(TypingCompatVal) bool {
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
type TypingAnyListT struct{} // var: list i.e. any[]

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

type TypingListT struct { // var: int[]
	elem TypingT
}

func NewListType(elem TypingT) *TypingListT {
	return &TypingListT{elem: elem}
}

func (t *TypingListT) Name() string {
	return t.elem.Name() + "[]"
}

func (t *TypingListT) IsCompatibleWith(val TypingCompatVal) bool {
	// 1. Value‐level check
	if val.Val != nil {
		actualList, ok := val.Val.([]interface{})
		if !ok {
			return false
		}

		// every element must match t.elem
		for _, actualElem := range actualList {
			if !t.elem.IsCompatibleWith(NewSubject(actualElem)) {
				return false
			}
		}

		return true
	}

	// 2. Type‐level only: accept any list
	if val.Type != nil {
		return *val.Type == RadListT
	}

	// 3. no info
	return false // todo or should we say true?
}

type TypingTupleT struct { // var: [int, float]
	types []TypingT
}

func NewTupleType(types ...TypingT) *TypingTupleT {
	return &TypingTupleT{types: types}
}

func (t *TypingTupleT) Name() string {
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

func (t *TypingTupleT) IsCompatibleWith(val TypingCompatVal) bool {
	// 1. Value‐level check
	if val.Val != nil {
		actualList, ok := val.Val.([]interface{})
		if !ok {
			return false
		}

		if len(actualList) != len(t.types) {
			return false
		}

		for i, actualElem := range actualList {
			if !t.types[i].IsCompatibleWith(NewSubject(actualElem)) {
				return false
			}
		}
		return true
	}

	// 2. Type‐level only: accept any list
	if val.Type != nil {
		return *val.Type == RadListT
	}

	// 3. no info
	return false
}

type TypingAnyMapT struct{} // var: map i.e. { any: any }

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

type TypingStructT struct { // var: { "mykey": int, "mykey2"?: float }
	named map[MapNamedKey]TypingT
}

func NewStructType(named map[MapNamedKey]TypingT) *TypingStructT {
	return &TypingStructT{
		named: named,
	}
}

func (t *TypingStructT) Name() string {
	var sb strings.Builder
	sb.WriteString("{ ")
	i := 0
	for mapKey, typ := range t.named {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", mapKey.Name))
		if mapKey.IsOptional {
			sb.WriteString("?")
		}
		sb.WriteString(": ")
		sb.WriteString(typ.Name())
		i++
	}
	sb.WriteString(" }")
	return sb.String()
}

func (t *TypingStructT) IsCompatibleWith(val TypingCompatVal) bool {
	// 1. We have a *value*
	if val.Val != nil {
		actualMap, ok := (val.Val).(map[string]interface{})
		if !ok {
			return false
		}

		// Validate each declared field
		for mapKey, typ := range t.named {
			actualVal, exists := actualMap[mapKey.Name]
			if !exists {
				if mapKey.IsOptional {
					continue
				}
				return false
			}

			if !typ.IsCompatibleWith(NewSubject(actualVal)) {
				return false
			}
		}

		return true
	}

	// 2. Only RadType information available
	if val.Type != nil {
		return *val.Type == RadMapT
	}

	// 3. No information to work with
	return false
}

type TypingMapT struct { // var: { string: int }
	keyT TypingT
	valT TypingT
}

func NewMapType(key, val TypingT) *TypingMapT {
	return &TypingMapT{
		keyT: key,
		valT: val,
	}
}

func (t *TypingMapT) Name() string {
	return fmt.Sprintf("{ %s: %s }", t.keyT.Name(), t.valT.Name())
}

func (t *TypingMapT) IsCompatibleWith(val TypingCompatVal) bool {
	// 1. We have a *value*
	if val.Val != nil {
		actualMap, ok := (val.Val).(map[string]interface{})
		if !ok {
			// Not a map.
			return false
		}

		for k, v := range actualMap {
			if !t.keyT.IsCompatibleWith(NewSubject(k)) ||
				!t.valT.IsCompatibleWith(NewSubject(v)) {
				return false
			}
		}

		return true
	}

	// 2. Only RadType information available
	if val.Type != nil {
		return *val.Type == RadMapT
	}

	// 3. No information to work with
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
	FnName  string // declared name of the function (empty if anonymous/lambda)
	Params  []TypingFnParam
	ReturnT *TypingT
	// nils mean no typing
}

func (t *TypingFnT) Name() string {
	var sb strings.Builder
	sb.WriteString("fn(")
	for i, p := range t.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		if p.IsVariadic {
			sb.WriteString("*")
		}
		if p.Type != nil {
			sb.WriteString((*p.Type).Name())
		} else {
			sb.WriteString(T_ANY)
		}
	}
	sb.WriteString(")")
	if t.ReturnT != nil {
		sb.WriteString(" -> ")
		sb.WriteString((*t.ReturnT).Name())
	}
	return sb.String()
}

// IsCompatibleWith checks if a value is a function. We do NOT structurally
// compare param/return shapes today because TypingCompatVal.Val cannot carry
// function arity information (RadFn lives in core, not rl). That deferral is
// documented in docs/type_system.md.
func (t *TypingFnT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		// Value-level structural check deferred (see docs/type_system.md). Fall through
		// to the type-level signal which is the same thing for fns today.
	}
	if val.Type != nil {
		return *val.Type == RadFnT
	}
	return false
}

type TypingStrEnumT struct { // var: ["foo", "bar"] = "bar"
	values []string
}

func NewStrEnumType(values ...string) *TypingStrEnumT {
	return &TypingStrEnumT{
		values: values,
	}
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
			if enumVal == strVal {
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

func NewMapNamedKey(name string, isOptional bool) MapNamedKey {
	return MapNamedKey{
		Name:       name,
		IsOptional: isOptional,
	}
}

type TypingFnParam struct {
	Name       string
	Type       *TypingT
	IsVariadic bool // vararg
	NamedOnly  bool // if true, can only be passed as a named arg
	IsOptional bool
	// Default is the CST-based default. Still needed because typing
	// resolution sets it from CST, and the converter reads it to produce
	// DefaultAST. Built-in defaults are pre-converted in rts/signatures.go init().
	Default    *RadNode
	DefaultAST *ASTDefault // AST-based default (used by interpreter at call time)
}

// ASTDefault holds an AST node and source for a function parameter default value.
type ASTDefault struct {
	Node Node
	Src  string
}

func (t TypingFnParam) AnonymousOnly() bool {
	return strings.HasPrefix(t.Name, "_")
}
