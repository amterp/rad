package rl

// This file implements IsAssignableFrom for every TypingT. The variance rules
// (invariant collections, contravariant params + covariant returns on
// functions, int->float widening on scalars, universal consistency for
// any/dynamic/error_type) are documented on the TypingT interface in
// typing.go. Implementations live here to keep all the static-checker
// compatibility logic together; the runtime IsCompatibleWith methods remain
// alongside their types.

// isAnyLike reports whether other is a "flows-into-anything" type: `any`
// (user-written escape hatch), `dynamic` (the implicit form when inference
// can't pin a type), `never` (the bottom type, vacuously a subtype of
// everything because no value inhabits it), or `<error>` (the static
// checker's poison marker for already-failed expressions). Every concrete
// IsAssignableFrom checks this first so values of these types can flow into
// any target without false negatives.
func isAnyLike(other TypingT) bool {
	switch other.(type) {
	case *TypingAnyT, *TypingDynamicT, *TypingNeverT, *TypingErrorTypeT:
		return true
	}
	return false
}

// typesEqual reports strict structural equality between two static types. Used
// by the invariant variance checks on collections - allowing `int[]` to satisfy
// `(int|str)[]` would let the callee push a string into a list the caller
// still believes is int-typed, so collection element types must match exactly
// rather than via the looser IsAssignableFrom. Treats nil ReturnT / param Type
// as `any` to mirror the rendering convention in Name().
func typesEqual(a, b TypingT) bool {
	if a == nil || b == nil {
		return a == b
	}
	switch ac := a.(type) {
	case *TypingStrT:
		_, ok := b.(*TypingStrT)
		return ok
	case *TypingIntT:
		_, ok := b.(*TypingIntT)
		return ok
	case *TypingFloatT:
		_, ok := b.(*TypingFloatT)
		return ok
	case *TypingBoolT:
		_, ok := b.(*TypingBoolT)
		return ok
	case *TypingErrorT:
		_, ok := b.(*TypingErrorT)
		return ok
	case *TypingAnyT:
		_, ok := b.(*TypingAnyT)
		return ok
	case *TypingDynamicT:
		_, ok := b.(*TypingDynamicT)
		return ok
	case *TypingErrorTypeT:
		_, ok := b.(*TypingErrorTypeT)
		return ok
	case *TypingNeverT:
		_, ok := b.(*TypingNeverT)
		return ok
	case *TypingVoidT:
		_, ok := b.(*TypingVoidT)
		return ok
	case *TypingAnyListT:
		_, ok := b.(*TypingAnyListT)
		return ok
	case *TypingAnyMapT:
		_, ok := b.(*TypingAnyMapT)
		return ok
	case *TypingListT:
		bc, ok := b.(*TypingListT)
		return ok && typesEqual(ac.elem, bc.elem)
	case *TypingTupleT:
		bc, ok := b.(*TypingTupleT)
		if !ok || len(ac.types) != len(bc.types) {
			return false
		}
		for i := range ac.types {
			if !typesEqual(ac.types[i], bc.types[i]) {
				return false
			}
		}
		return true
	case *TypingMapT:
		bc, ok := b.(*TypingMapT)
		return ok && typesEqual(ac.keyT, bc.keyT) && typesEqual(ac.valT, bc.valT)
	case *TypingStructT:
		bc, ok := b.(*TypingStructT)
		if !ok || len(ac.named) != len(bc.named) {
			return false
		}
		for k, t := range ac.named {
			bt, exists := bc.named[k]
			if !exists || !typesEqual(t, bt) {
				return false
			}
		}
		return true
	case *TypingVarArgT:
		bc, ok := b.(*TypingVarArgT)
		return ok && typesEqual(ac.t, bc.t)
	case *TypingOptionalT:
		bc, ok := b.(*TypingOptionalT)
		return ok && typesEqual(ac.t, bc.t)
	case *TypingFnT:
		bc, ok := b.(*TypingFnT)
		if !ok || len(ac.Params) != len(bc.Params) {
			return false
		}
		for i := range ac.Params {
			if !typesEqual(paramTypeOrAny(ac.Params[i]), paramTypeOrAny(bc.Params[i])) {
				return false
			}
		}
		return typesEqual(returnTypeOrAny(ac), returnTypeOrAny(bc))
	case *TypingStrEnumT:
		bc, ok := b.(*TypingStrEnumT)
		if !ok || len(ac.values) != len(bc.values) {
			return false
		}
		set := make(map[string]bool, len(ac.values))
		for _, v := range ac.values {
			set[v] = true
		}
		for _, v := range bc.values {
			if !set[v] {
				return false
			}
		}
		return true
	case *TypingUnionT:
		bc, ok := b.(*TypingUnionT)
		if !ok || len(ac.types) != len(bc.types) {
			return false
		}
		used := make([]bool, len(bc.types))
		for _, at := range ac.types {
			matched := false
			for i, bt := range bc.types {
				if !used[i] && typesEqual(at, bt) {
					used[i] = true
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
		return true
	}
	return false
}

// paramTypeOrAny extracts a parameter's declared type, defaulting to `any`
// when no annotation is present. Matches the convention used by Name().
func paramTypeOrAny(p TypingFnParam) TypingT {
	if p.Type != nil {
		return *p.Type
	}
	return NewAnyType()
}

// returnTypeOrAny extracts a function's declared return type, defaulting to
// `any` when no annotation is present.
func returnTypeOrAny(fn *TypingFnT) TypingT {
	if fn.ReturnT != nil {
		return *fn.ReturnT
	}
	return NewAnyType()
}

// --- Primitives & wildcards ---

func (t *TypingStrT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	switch other.(type) {
	case *TypingStrT:
		return true
	case *TypingStrEnumT:
		// A string-enum value is, by definition, a string.
		return true
	}
	return false
}

func (t *TypingIntT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	_, ok := other.(*TypingIntT)
	return ok
}

// Float accepts int via the one and only implicit numeric widening rule (the
// same rule the runtime enforces in IsCompatibleWith). No other widening
// happens implicitly.
func (t *TypingFloatT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	switch other.(type) {
	case *TypingFloatT, *TypingIntT:
		return true
	}
	return false
}

func (t *TypingBoolT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	_, ok := other.(*TypingBoolT)
	return ok
}

func (t *TypingErrorT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	_, ok := other.(*TypingErrorT)
	return ok
}

// Any is the user-written escape hatch. Universally compatible: it accepts
// every other type, and is accepted by every other type (Siek-Taha gradual
// consistency). Distinct from Dynamic (which carries the same semantics but
// signals "implicit, not user-opted-in" so a future strict mode can flag it).
func (t *TypingAnyT) IsAssignableFrom(TypingT) bool {
	return true
}

// Dynamic is the implicit-any type the static checker assigns when inference
// can't pin a value down. Behaviorally identical to Any in assignability so
// no spurious errors fire today; the distinction exists for the future
// strict-mode flag.
func (t *TypingDynamicT) IsAssignableFrom(TypingT) bool {
	return true
}

// Void is the type of expressions that produce no value (e.g. print()).
// Only Void itself, Never (vacuous), and ErrorType (cascade suppression)
// are assignable to it. Notably `any` and `dynamic` are NOT - that's how
// `x = print(...)` gets caught instead of being silently swallowed under
// gradual consistency.
func (t *TypingVoidT) IsAssignableFrom(other TypingT) bool {
	switch other.(type) {
	case *TypingNeverT, *TypingErrorTypeT:
		return true
	}
	_, ok := other.(*TypingVoidT)
	return ok
}

// Never is the bottom type: only Never itself can flow into a Never slot.
// (Other types DO accept Never as a source because Never is a vacuous
// subtype of everything - that's handled by the isAnyLike short-circuit at
// the top of every other IsAssignableFrom.)
func (t *TypingNeverT) IsAssignableFrom(other TypingT) bool {
	_, ok := other.(*TypingNeverT)
	return ok
}

// ErrorType is the static-checker's poison marker. It accepts anything as a
// source, and (via the isAnyLike short-circuit on every other type) is
// accepted by anything as a target. That bidirectional permissiveness is
// what suppresses cascading diagnostics: one expression's failure produces
// one error, not a flood of derivative errors as the failure flows through
// subsequent expressions.
func (t *TypingErrorTypeT) IsAssignableFrom(TypingT) bool {
	return true
}

// --- Collections (invariant) ---

// AnyList is the unparameterized list type. It accepts any concrete list or
// tuple shape, since the caller has expressed no requirement on the element
// type.
func (t *TypingAnyListT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	switch other.(type) {
	case *TypingAnyListT, *TypingListT, *TypingTupleT:
		return true
	}
	return false
}

// List<T> is invariant in T. List<int> does NOT accept List<float>: the
// callee could write a float into the list, and the caller still believes
// the list is int-typed.
func (t *TypingListT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingListT)
	if !ok {
		return false
	}
	return typesEqual(t.elem, o.elem)
}

// Tuples must match position-for-position with exact element types.
func (t *TypingTupleT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingTupleT)
	if !ok || len(t.types) != len(o.types) {
		return false
	}
	for i := range t.types {
		if !typesEqual(t.types[i], o.types[i]) {
			return false
		}
	}
	return true
}

// AnyMap accepts any concrete map or struct shape - no key/value constraint.
func (t *TypingAnyMapT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	switch other.(type) {
	case *TypingAnyMapT, *TypingMapT, *TypingStructT:
		return true
	}
	return false
}

// Struct types match strictly: same key set (including each key's optional
// flag) with equal value types. Width subtyping (accepting a struct with extra
// fields) is intentionally not supported in v1 - it would let mutation through
// one alias corrupt the view through another.
func (t *TypingStructT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingStructT)
	if !ok || len(t.named) != len(o.named) {
		return false
	}
	for k, ty := range t.named {
		oTy, exists := o.named[k]
		if !exists || !typesEqual(ty, oTy) {
			return false
		}
	}
	return true
}

// Map<K,V> is invariant in both K and V for the same mutation-safety reason
// as List.
func (t *TypingMapT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingMapT)
	if !ok {
		return false
	}
	return typesEqual(t.keyT, o.keyT) && typesEqual(t.valT, o.valT)
}

// --- Modifiers ---

// VarArg only appears inside function signatures; equality on the underlying
// element type is what callers will check via the enclosing fn comparison.
func (t *TypingVarArgT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingVarArgT)
	if !ok {
		return false
	}
	return typesEqual(t.t, o.t)
}

// Optional<T> accepts Optional<U> when T accepts U, and also accepts T
// directly (the "definitely not null" case). The reverse - assigning an
// Optional<T> into a slot typed T - is rejected, because T can't hold null.
func (t *TypingOptionalT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	if o, ok := other.(*TypingOptionalT); ok {
		return t.t.IsAssignableFrom(o.t)
	}
	return t.t.IsAssignableFrom(other)
}

// --- Higher-order ---

// Function types: parameters contravariant, return covariant. fn(any) -> int
// is assignable to fn(int) -> int because the caller will supply an int, and
// the callee's wider any parameter type can accept it. Reverse direction is
// unsafe.
func (t *TypingFnT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingFnT)
	if !ok {
		return false
	}
	if len(t.Params) != len(o.Params) {
		return false
	}
	for i := range t.Params {
		thisP := paramTypeOrAny(t.Params[i])
		otherP := paramTypeOrAny(o.Params[i])
		// Contravariant: the supplied function's parameter must accept anything
		// the declared parameter would supply.
		if !otherP.IsAssignableFrom(thisP) {
			return false
		}
	}
	// Covariant return.
	return returnTypeOrAny(t).IsAssignableFrom(returnTypeOrAny(o))
}

// StrEnum["a","b","c"] accepts a string-enum value whose set of allowed values
// is a subset of this enum's set.
func (t *TypingStrEnumT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	o, ok := other.(*TypingStrEnumT)
	if !ok {
		return false
	}
	allowed := make(map[string]bool, len(t.values))
	for _, v := range t.values {
		allowed[v] = true
	}
	for _, v := range o.values {
		if !allowed[v] {
			return false
		}
	}
	return true
}

// Union: when assigning a union into a union, every branch of the source must
// fit into the target. When assigning a non-union into a union, the source
// only needs to fit into one branch.
func (t *TypingUnionT) IsAssignableFrom(other TypingT) bool {
	if isAnyLike(other) {
		return true
	}
	if o, ok := other.(*TypingUnionT); ok {
		for _, ot := range o.types {
			if !t.IsAssignableFrom(ot) {
				return false
			}
		}
		return true
	}
	for _, tt := range t.types {
		if tt.IsAssignableFrom(other) {
			return true
		}
	}
	return false
}
