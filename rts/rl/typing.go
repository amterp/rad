package rl

import (
	"fmt"
	"sort"
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
	_ TypingT = (*TypingErrorTypeT)(nil)
	_ TypingT = (*TypingNeverT)(nil)
	_ TypingT = (*TypingVoidT)(nil)
	_ TypingT = (*TypingNullT)(nil)
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

// TypingErrorTypeT is the static-checker's "poison" type, distinct from the
// runtime TypingErrorT (the user-facing `error` type returned by builtins
// like parse_json). When an expression fails to type-check, the checker
// assigns it ErrorType instead of returning Go-nil. ErrorType is universally
// compatible in both directions, so subsequent expressions that consume the
// failed expression don't fire their own (cascading) diagnostics. The user
// sees one error - the original - instead of ten.
//
// Users never write or see this type in normal flow; if it ever surfaces in
// an error message (rendered as "<error>"), that itself indicates a checker
// bug.
type TypingErrorTypeT struct{}

func NewErrorTypeType() *TypingErrorTypeT {
	return &TypingErrorTypeT{}
}

func (t *TypingErrorTypeT) Name() string {
	return T_ERROR_TYPE
}

// IsCompatibleWith returns true: a poisoned expression's value, if it
// somehow flows to runtime, shouldn't cause additional confusion. Real
// callers should never see this - the failed check that produced the
// ErrorType already aborted the path.
func (t *TypingErrorTypeT) IsCompatibleWith(TypingCompatVal) bool {
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

// TypingNullT is the static type of the `null` literal. It is NOT
// user-writable as a standalone type annotation - users spell
// nullable as `T?` (TypingOptionalT). TypingNullT exists so that:
//
//   - the LitNull synth has a sound type instead of falling back to
//     Dynamic and silently fitting any slot;
//   - narrowing the false branch of `if x != null:` has a definite
//     answer instead of a no-op;
//   - inferred returns that mix a value and `null` synthesize to
//     `T?` rather than `T|dynamic`.
//
// Assignability: TypingNullT only flows into slots that admit null
// (Optional<T>, unions containing null, any/dynamic). Non-nullable
// slots reject it.
type TypingNullT struct{}

func NewNullType() *TypingNullT {
	return &TypingNullT{}
}

func (t *TypingNullT) Name() string {
	return "null"
}

func (t *TypingNullT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Type != nil {
		return *val.Type == RadNullT
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
	return parenWrapIfUnion(t.elem) + "[]"
}

// parenWrapIfUnion returns t's name wrapped in parentheses when t is
// a union, otherwise the bare name. Used by surface-syntax printers
// (TypingListT.Name, TypingOptionalT.Name) where omitting parens
// changes how a human re-reads the type: `int|str[]` parses as
// `int | str[]`, not as a list of int-or-str.
func parenWrapIfUnion(t TypingT) string {
	if _, ok := t.(*TypingUnionT); ok {
		return "(" + t.Name() + ")"
	}
	return t.Name()
}

// Elem returns the list'\”s element type. Exposed for narrowing -
// the for-loop walker uses it to type the loop variable.
func (t *TypingListT) Elem() TypingT {
	return t.elem
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

// Types returns the tuple's positional element types. Exposed for
// structural inspection (e.g. collapsing tuples with Never positions
// in the type checker's union-join pipeline).
func (t *TypingTupleT) Types() []TypingT {
	return t.types
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

// Field looks up a declared key by name, returning its type, whether
// it's optional, and whether it exists at all. Used by the type
// checker to resolve `m.key` / `m["key"]` access against a known
// struct shape (e.g. a builtin's typed-map return value).
func (t *TypingStructT) Field(name string) (typ TypingT, optional bool, found bool) {
	for mapKey, fieldT := range t.named {
		if mapKey.Name == name {
			return fieldT, mapKey.IsOptional, true
		}
	}
	return nil, false, false
}

// Fields exposes the declared keys for enumeration (LSP completion,
// hover). The returned map is the live backing store - callers must
// not mutate it.
func (t *TypingStructT) Fields() map[MapNamedKey]TypingT {
	return t.named
}

func (t *TypingStructT) Name() string {
	// Sort keys so the rendered name is deterministic - it surfaces
	// in diagnostics, hover, and snapshot output, all of which need
	// stable ordering (Go map iteration is randomized). We don't
	// retain declaration order: `named` is a map, so it's already
	// lost by here; alphabetical is the stable choice available.
	keys := make([]MapNamedKey, 0, len(t.named))
	for mapKey := range t.named {
		keys = append(keys, mapKey)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name < keys[j].Name })

	var sb strings.Builder
	sb.WriteString("{ ")
	for i, mapKey := range keys {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", mapKey.Name))
		if mapKey.IsOptional {
			sb.WriteString("?")
		}
		sb.WriteString(": ")
		// Guard against a nil field type: a struct built with a nil
		// value (not reachable via the type resolver today, but the
		// NewStructType API permits it) would otherwise panic here.
		if typ := t.named[mapKey]; typ != nil {
			sb.WriteString(typ.Name())
		} else {
			sb.WriteString(T_ANY)
		}
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

// KeyT and ValT expose the map'\”s key and value types. Used by the
// for-loop walker: a single-var iteration over a map binds the var
// to the key type; a two-var iteration binds (key, value).
func (t *TypingMapT) KeyT() TypingT { return t.keyT }
func (t *TypingMapT) ValT() TypingT { return t.valT }

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
	return parenWrapIfUnion(t.t) + "?"
}

// Inner returns the non-null component of an optional. Exposed for
// flow-sensitive narrowing: after `if x != null:` we want the
// underlying T from Optional<T>.
func (t *TypingOptionalT) Inner() TypingT {
	return t.t
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

// Values returns the closed set of strings this enum permits. Exposed
// for narrowing so the static checker can partition the set against a
// predicate like `x == "foo"` or `x in ["a", "b"]`.
func (t *TypingStrEnumT) Values() []string {
	return t.values
}

// Contains reports whether value is one of the enum's members.
func (t *TypingStrEnumT) Contains(value string) bool {
	for _, v := range t.values {
		if v == value {
			return true
		}
	}
	return false
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

// Types returns the union's component types. Exposed for narrowing -
// the static checker walks arms to subtract null, peel off enum
// variants, etc.
func (t *TypingUnionT) Types() []TypingT {
	return t.types
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
	Name string
	// NameSpan covers just the parameter-name identifier in source.
	// Zero for synthesised params (e.g. fn_type entries that have no
	// names) and for built-in signatures (constructed in Go, not parsed).
	// The binder uses this as the symbol's DeclSpan so LSP rename /
	// find-refs / goto-def land on the name token rather than the
	// owning fn span.
	NameSpan   Span
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

// TypingFromArgTypeName maps an args-block type-name string to its
// TypingT. The args-block grammar admits only the eight scalar / list
// forms (str, int, float, bool, plus their `[]` variants) and is the
// authoritative caller list - everywhere else we go through the
// general type parser. Returns nil for unrecognised names so the
// binder can no-op rather than panic on malformed input.
func TypingFromArgTypeName(name string) TypingT {
	switch name {
	case T_STR:
		return NewStrType()
	case T_INT:
		return NewIntType()
	case T_FLOAT:
		return NewFloatType()
	case T_BOOL:
		return NewBoolType()
	case T_STR_LIST:
		return NewListType(NewStrType())
	case T_INT_LIST:
		return NewListType(NewIntType())
	case T_FLOAT_LIST:
		return NewListType(NewFloatType())
	case T_BOOL_LIST:
		return NewListType(NewBoolType())
	}
	return nil
}
