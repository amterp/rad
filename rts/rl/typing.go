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
	return "struct" // todo improve this, need evaluator
}

func (t *TypingStructT) IsCompatibleWith(val TypingCompatVal) bool {
	// 1. We have a *value*
	if val.Val != nil {
		actualMap, ok := (val.Val).(map[string]interface{})
		if !ok {
			// Not a map.
			return false
		}

		if val.Evaluator == nil {
			// Can't validate keys, take loose approach
			return true
		}

		// Validate each declared field
		seen := make(map[string]bool)
		for mapKey, typ := range t.named {
			expectedKeyName := (*val.Evaluator)(mapKey.Name.Node, mapKey.Name.Src)
			keyName, ok := expectedKeyName.(string)
			if !ok {
				// Key was not a string, unexpected.
				return false
			}

			actualVal, exists := actualMap[keyName]
			if !exists {
				if mapKey.IsOptional {
					continue // missing optional key is fine
				}
				return false // required key absent
			}

			seen[keyName] = true
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
	Name    string // empty if anonymous
	Params  []TypingFnParam
	ReturnT *TypingT
	// nils mean no typing
}

type TypingStrEnumT struct { // var: ["foo", "bar"] = "bar"
	strNodes []*RadNode
}

func NewStrEnumType(stringNodes ...*RadNode) *TypingStrEnumT {
	return &TypingStrEnumT{
		strNodes: stringNodes,
	}
}

func (t *TypingStrEnumT) Name() string {
	return "str enum"

	// TODO should do something like the below, but we only have nodes.
	//  probably should make the IsCompatibleWith method return error, with the msg.
	// var sb strings.Builder
	// sb.WriteString("[")
	// for i, v := range t.strNodes {
	// 	if i > 0 {
	// 		sb.WriteString(", ")
	// 	}
	// 	sb.WriteString(fmt.Sprintf("%q", v))
	// }
	// sb.WriteString("]")
	// return sb.String()
}

func (t *TypingStrEnumT) IsCompatibleWith(val TypingCompatVal) bool {
	if val.Val != nil {
		strVal, ok := (val.Val).(string)
		if !ok {
			return false
		}

		if val.Evaluator != nil {
			for _, strNode := range t.strNodes {
				out := (*val.Evaluator)(strNode.Node, strNode.Src)
				if out == strVal {
					return true
				}
			}
			return false
		}
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
	Name       *RadNode
	IsOptional bool
}

func NewMapNamedKey(name *RadNode, isOptional bool) MapNamedKey {
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
	Default    *RadNode    // CST-based default (nil if no default)
	DefaultAST *ASTDefault // AST-based default (set by converter)
}

// ASTDefault holds an AST node and source for a function parameter default value.
// Allows gradual migration: converter sets DefaultAST, interpreter reads it.
type ASTDefault struct {
	Node Node
	Src  string
}

func (t TypingFnParam) AnonymousOnly() bool {
	return strings.HasPrefix(t.Name, "_")
}
