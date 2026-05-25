package rl_test

import (
	"testing"

	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
)

// Pure unit tests over TypingT.IsAssignableFrom - the static-checker's
// type-vs-type compatibility rule. Distinct from IsCompatibleWith (the runtime
// value-vs-type check). The rules baked in here:
//   - `any` is universally consistent (both directions).
//   - Primitives are identity, with int->float as the single implicit widening.
//   - Lists are covariant in element type; tuples, maps, and structs are
//     invariant.
//   - Function params are contravariant, returns covariant.
//   - Optional<T> accepts T directly; T does not accept Optional<T>.
//   - StrEnum is subset-based.
//   - Union<T,U> accepts X if any branch does (or, if X is itself a union,
//     every branch of X must fit).

func TestAssign_PrimitiveIdentity(t *testing.T) {
	intT := rl.NewIntType()
	strT := rl.NewStrType()
	floatT := rl.NewFloatType()
	boolT := rl.NewBoolType()

	assert.True(t, intT.IsAssignableFrom(rl.NewIntType()))
	assert.False(t, intT.IsAssignableFrom(strT))
	assert.False(t, intT.IsAssignableFrom(floatT))
	assert.False(t, intT.IsAssignableFrom(boolT))

	assert.True(t, strT.IsAssignableFrom(rl.NewStrType()))
	assert.False(t, strT.IsAssignableFrom(intT))

	assert.True(t, boolT.IsAssignableFrom(rl.NewBoolType()))
	assert.False(t, boolT.IsAssignableFrom(intT))
}

func TestAssign_IntWidensToFloat(t *testing.T) {
	floatT := rl.NewFloatType()
	intT := rl.NewIntType()
	// The one and only implicit numeric widening.
	assert.True(t, floatT.IsAssignableFrom(intT))
	// The reverse must NOT hold - assigning a float into an int target loses data.
	assert.False(t, intT.IsAssignableFrom(floatT))
}

func TestAssign_AnyIsUniversallyConsistent(t *testing.T) {
	anyT := rl.NewAnyType()
	intT := rl.NewIntType()
	strList := rl.NewListType(rl.NewStrType())

	// any accepts anything
	assert.True(t, anyT.IsAssignableFrom(intT))
	assert.True(t, anyT.IsAssignableFrom(strList))
	assert.True(t, anyT.IsAssignableFrom(rl.NewVoidType()))

	// anything accepts any (gradual consistency - typed code can pull from any
	// without complaint; the runtime catches real mismatches)
	assert.True(t, intT.IsAssignableFrom(anyT))
	assert.True(t, strList.IsAssignableFrom(anyT))
}

func TestAssign_DynamicIsUniversallyConsistent(t *testing.T) {
	dynT := rl.NewDynamicType()
	intT := rl.NewIntType()
	strList := rl.NewListType(rl.NewStrType())

	// Same shape as any: dynamic accepts everything and is accepted by
	// everything. The distinction from any is provenance (implicit vs
	// user-written) and only matters when a future strict mode wants to
	// flag implicit-dynamic flow.
	assert.True(t, dynT.IsAssignableFrom(intT))
	assert.True(t, dynT.IsAssignableFrom(strList))
	assert.True(t, dynT.IsAssignableFrom(rl.NewVoidType()))

	assert.True(t, intT.IsAssignableFrom(dynT))
	assert.True(t, strList.IsAssignableFrom(dynT))
}

func TestAssign_NeverIsBottom(t *testing.T) {
	neverT := rl.NewNeverType()
	intT := rl.NewIntType()
	strList := rl.NewListType(rl.NewStrType())

	// Never accepts only Never.
	assert.True(t, neverT.IsAssignableFrom(rl.NewNeverType()))
	assert.False(t, neverT.IsAssignableFrom(intT))
	assert.False(t, neverT.IsAssignableFrom(strList))
	assert.False(t, neverT.IsAssignableFrom(rl.NewVoidType()))

	// Every other type accepts Never as a source - it's vacuously a subtype
	// of everything because no value inhabits it. This is what makes
	// "switch exhausted all cases, residual is Never, post-switch code is
	// reachable as anything" work cleanly.
	assert.True(t, intT.IsAssignableFrom(neverT))
	assert.True(t, strList.IsAssignableFrom(neverT))
	assert.True(t, rl.NewAnyType().IsAssignableFrom(neverT))

	// Even void accepts Never (vacuously).
	assert.True(t, rl.NewVoidType().IsAssignableFrom(neverT))
}

func TestAssign_NeverHasNoValues(t *testing.T) {
	// No runtime value should ever be considered compatible with Never. The
	// runtime never sees Never directly today, but the contract matters if
	// it ever flows through a call-boundary check.
	neverT := rl.NewNeverType()
	assert.False(t, neverT.IsCompatibleWith(rl.NewIntSubject(0)))
	assert.False(t, neverT.IsCompatibleWith(rl.NewStrSubject("")))
	assert.False(t, neverT.IsCompatibleWith(rl.NewNullSubject()))
}

func TestAssign_ErrorTypeSuppressesCascade(t *testing.T) {
	errT := rl.NewErrorTypeType()
	intT := rl.NewIntType()
	listInt := rl.NewListType(rl.NewIntType())

	// ErrorType accepts anything as a source - a failed expression's
	// poisoned type doesn't need to be "compatible" with anything that
	// happens to be assigned to it.
	assert.True(t, errT.IsAssignableFrom(intT))
	assert.True(t, errT.IsAssignableFrom(listInt))
	assert.True(t, errT.IsAssignableFrom(rl.NewVoidType()))

	// And it's accepted by anything as a target - this is the
	// cascade-suppression direction. A subsequent expression that consumes
	// a poisoned result doesn't fire its own type-mismatch diagnostic.
	assert.True(t, intT.IsAssignableFrom(errT))
	assert.True(t, listInt.IsAssignableFrom(errT))

	// Including by void - poison flows through "no return value" positions
	// too, so a poisoned call result doesn't double-error when discarded.
	assert.True(t, rl.NewVoidType().IsAssignableFrom(errT))
}

func TestAssign_ErrorTypeDistinctFromRuntimeError(t *testing.T) {
	// The static-checker poison type and the runtime `error` type are
	// completely separate. error is what parse_json returns when JSON is
	// malformed - a real value users handle. ErrorType is a placeholder
	// for "this expression failed to type-check" and never appears in user
	// code.
	staticErr := rl.TypingT(rl.NewErrorTypeType())
	runtimeErr := rl.TypingT(rl.NewErrorType())

	_, isRuntime := staticErr.(*rl.TypingErrorT)
	assert.False(t, isRuntime, "static ErrorType must not satisfy *TypingErrorT")
	_, isStatic := runtimeErr.(*rl.TypingErrorTypeT)
	assert.False(t, isStatic, "runtime error must not satisfy *TypingErrorTypeT")
	assert.NotEqual(t, staticErr.Name(), runtimeErr.Name(),
		"names must differ so users never see the static poison form")
}

func TestAssign_DynamicAndAnyAreDistinct(t *testing.T) {
	// They behave identically for IsAssignableFrom today, but they're not the
	// same type. The static checker must be able to tell them apart - that's
	// the whole point of having two.
	dynT := rl.TypingT(rl.NewDynamicType())
	anyT := rl.TypingT(rl.NewAnyType())

	_, isAny := dynT.(*rl.TypingAnyT)
	assert.False(t, isAny, "dynamic must not satisfy *TypingAnyT")
	_, isDyn := anyT.(*rl.TypingDynamicT)
	assert.False(t, isDyn, "any must not satisfy *TypingDynamicT")
	assert.NotEqual(t, dynT.Name(), anyT.Name(), "names must differ so error messages distinguish them")
}

func TestAssign_VoidIsExclusive(t *testing.T) {
	voidT := rl.NewVoidType()
	// Only void itself flows into void. Catches `x = print(...)`.
	assert.True(t, voidT.IsAssignableFrom(rl.NewVoidType()))
	assert.False(t, voidT.IsAssignableFrom(rl.NewIntType()))
	// Note: void does NOT accept `any` - users shouldn't be able to assign
	// arbitrary results to void slots.
	assert.False(t, voidT.IsAssignableFrom(rl.NewAnyType()))
}

func TestAssign_StrEnumIsAString(t *testing.T) {
	strT := rl.NewStrType()
	enumAB := rl.NewStrEnumType("a", "b")
	// A value of a string-enum type is, by definition, a string.
	assert.True(t, strT.IsAssignableFrom(enumAB))
	// But a string is not necessarily a member of the enum.
	assert.False(t, enumAB.IsAssignableFrom(strT))
}

func TestAssign_StrEnumSubsetRelation(t *testing.T) {
	enumABC := rl.NewStrEnumType("a", "b", "c")
	enumAB := rl.NewStrEnumType("a", "b")
	enumAX := rl.NewStrEnumType("a", "x")

	// Subset flows up: ["a","b"] fits into ["a","b","c"].
	assert.True(t, enumABC.IsAssignableFrom(enumAB))
	// Not a subset: ["a","x"] has "x" which isn't in the target.
	assert.False(t, enumABC.IsAssignableFrom(enumAX))
	// Superset doesn't fit into subset.
	assert.False(t, enumAB.IsAssignableFrom(enumABC))
}

func TestAssign_ListsCovariant(t *testing.T) {
	listInt := rl.NewListType(rl.NewIntType())
	listFloat := rl.NewListType(rl.NewFloatType())
	listAny := rl.NewListType(rl.NewAnyType())

	// Identity holds.
	assert.True(t, listInt.IsAssignableFrom(rl.NewListType(rl.NewIntType())))

	// Covariance: List<int> flows into List<float> because int widens to
	// float at the scalar level. Likewise List<int> flows into List<any>.
	// Unsound under mutation+aliasing but accepted for ergonomics - see the
	// commentary on TypingListT.IsAssignableFrom.
	assert.True(t, listFloat.IsAssignableFrom(listInt))
	assert.True(t, listAny.IsAssignableFrom(listInt))

	// Narrowing direction stays refused.
	assert.False(t, listInt.IsAssignableFrom(listFloat))
}

func TestAssign_ListsCovariantOverUnions(t *testing.T) {
	// Regression locks for the round-2 LSP verification bugs (cards
	// list-covariance-t and nested-paren-union).
	intT := rl.NewIntType()
	strT := rl.NewStrType()
	boolT := rl.NewBoolType()
	intOrStr := rl.NewUnionType(intT, strT)
	intStrBoolFlat := rl.NewUnionType(intT, strT, boolT)
	intStrBoolNestedLeft := rl.NewUnionType(intOrStr, boolT)                   // ((int|str)|bool)
	intStrBoolNestedRight := rl.NewUnionType(intT, rl.NewUnionType(strT, boolT)) // (int|(str|bool))

	// `xs: (int|str)[] = [1, 2, 3]` — the common int[] -> (int|str)[] widening.
	assert.True(t, rl.NewListType(intOrStr).IsAssignableFrom(rl.NewListType(intT)))
	// `xs: (int|str)[] = ["a", "b"]` — same shape, str[] source.
	assert.True(t, rl.NewListType(intOrStr).IsAssignableFrom(rl.NewListType(strT)))
	// Nested-paren unions in the list element accept the flat union and vice versa.
	assert.True(t, rl.NewListType(intStrBoolNestedLeft).IsAssignableFrom(rl.NewListType(intStrBoolFlat)))
	assert.True(t, rl.NewListType(intStrBoolNestedRight).IsAssignableFrom(rl.NewListType(intStrBoolFlat)))
	assert.True(t, rl.NewListType(intStrBoolFlat).IsAssignableFrom(rl.NewListType(intStrBoolNestedLeft)))
	// Genuine mismatch still fires (int[] can't accept str[]).
	assert.False(t, rl.NewListType(intT).IsAssignableFrom(rl.NewListType(strT)))
}

func TestAssign_AnyListAcceptsAnyConcrete(t *testing.T) {
	anyList := rl.NewAnyListType()
	assert.True(t, anyList.IsAssignableFrom(rl.NewAnyListType()))
	assert.True(t, anyList.IsAssignableFrom(rl.NewListType(rl.NewIntType())))
	assert.True(t, anyList.IsAssignableFrom(rl.NewTupleType(rl.NewIntType(), rl.NewStrType())))
	assert.False(t, anyList.IsAssignableFrom(rl.NewIntType()))
}

func TestAssign_TuplesMatchPositionwise(t *testing.T) {
	intStr := rl.NewTupleType(rl.NewIntType(), rl.NewStrType())
	strInt := rl.NewTupleType(rl.NewStrType(), rl.NewIntType())
	intStrBool := rl.NewTupleType(rl.NewIntType(), rl.NewStrType(), rl.NewBoolType())

	assert.True(t, intStr.IsAssignableFrom(rl.NewTupleType(rl.NewIntType(), rl.NewStrType())))
	// Position matters.
	assert.False(t, intStr.IsAssignableFrom(strInt))
	// Length matters.
	assert.False(t, intStr.IsAssignableFrom(intStrBool))
}

func TestAssign_MapsInvariant(t *testing.T) {
	mapStrInt := rl.NewMapType(rl.NewStrType(), rl.NewIntType())
	mapStrFloat := rl.NewMapType(rl.NewStrType(), rl.NewFloatType())

	assert.True(t, mapStrInt.IsAssignableFrom(rl.NewMapType(rl.NewStrType(), rl.NewIntType())))
	// Same invariance argument as lists.
	assert.False(t, mapStrFloat.IsAssignableFrom(mapStrInt))
}

func TestAssign_StructStrictMatch(t *testing.T) {
	abReq := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("a", false): rl.NewIntType(),
		rl.NewMapNamedKey("b", false): rl.NewStrType(),
	})
	abReqAgain := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("a", false): rl.NewIntType(),
		rl.NewMapNamedKey("b", false): rl.NewStrType(),
	})
	abReqDifferentType := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("a", false): rl.NewIntType(),
		rl.NewMapNamedKey("b", false): rl.NewBoolType(),
	})
	abReqWithExtra := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("a", false): rl.NewIntType(),
		rl.NewMapNamedKey("b", false): rl.NewStrType(),
		rl.NewMapNamedKey("c", false): rl.NewBoolType(),
	})
	abReqOptional := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("a", false): rl.NewIntType(),
		rl.NewMapNamedKey("b", true):  rl.NewStrType(),
	})

	assert.True(t, abReq.IsAssignableFrom(abReqAgain))
	assert.False(t, abReq.IsAssignableFrom(abReqDifferentType))
	// Width subtyping not supported in v1.
	assert.False(t, abReq.IsAssignableFrom(abReqWithExtra))
	// Optionality is part of the key identity.
	assert.False(t, abReq.IsAssignableFrom(abReqOptional))
}

func TestAssign_OptionalWrappingAndUnwrapping(t *testing.T) {
	optInt := rl.NewOptionalType(rl.NewIntType())
	optFloat := rl.NewOptionalType(rl.NewFloatType())
	intT := rl.NewIntType()

	// Optional<int> accepts int directly (the "definitely not null" case).
	assert.True(t, optInt.IsAssignableFrom(intT))
	// And accepts another optional with a compatible element.
	assert.True(t, optInt.IsAssignableFrom(rl.NewOptionalType(rl.NewIntType())))
	// Optional<float> accepts Optional<int> via int->float widening.
	assert.True(t, optFloat.IsAssignableFrom(optInt))

	// The reverse is unsafe: int cannot hold null, so Optional<int> can't flow
	// into a plain int slot.
	assert.False(t, intT.IsAssignableFrom(optInt))
}

func TestAssign_FunctionContravariantParamsCovariantReturn(t *testing.T) {
	intT := rl.TypingT(rl.NewIntType())
	floatT := rl.TypingT(rl.NewFloatType())
	boolT := rl.TypingT(rl.NewBoolType())
	anyT := rl.TypingT(rl.NewAnyType())

	// declared: fn(int) -> float
	// supplied: fn(any) -> int
	// Should be assignable: param is contravariant (any accepts int), return is
	// covariant (float accepts int via widening).
	declared := &rl.TypingFnT{
		Params:  []rl.TypingFnParam{{Type: &intT}},
		ReturnT: &floatT,
	}
	supplied := &rl.TypingFnT{
		Params:  []rl.TypingFnParam{{Type: &anyT}},
		ReturnT: &intT,
	}
	assert.True(t, declared.IsAssignableFrom(supplied))

	// Reverse direction: declared fn(any)->int, supplied fn(int)->float. The
	// supplied function only knows how to handle int, but the caller may pass
	// any. Reject.
	assert.False(t, supplied.IsAssignableFrom(declared))

	// Arity mismatch always rejects.
	twoArg := &rl.TypingFnT{
		Params: []rl.TypingFnParam{{Type: &intT}, {Type: &intT}},
	}
	oneArg := &rl.TypingFnT{
		Params: []rl.TypingFnParam{{Type: &intT}},
	}
	assert.False(t, twoArg.IsAssignableFrom(oneArg))
	assert.False(t, oneArg.IsAssignableFrom(twoArg))

	// Incompatible return: fn()->bool can't fit into a slot expecting fn()->int.
	wantInt := &rl.TypingFnT{ReturnT: &intT}
	hasBool := &rl.TypingFnT{ReturnT: &boolT}
	assert.False(t, wantInt.IsAssignableFrom(hasBool))
}

func TestAssign_UnionBranchMatch(t *testing.T) {
	intOrStr := rl.NewUnionType(rl.NewIntType(), rl.NewStrType())
	intOrFloat := rl.NewUnionType(rl.NewIntType(), rl.NewFloatType())

	// Non-union fits into union if any branch accepts it.
	assert.True(t, intOrStr.IsAssignableFrom(rl.NewIntType()))
	assert.True(t, intOrStr.IsAssignableFrom(rl.NewStrType()))
	assert.False(t, intOrStr.IsAssignableFrom(rl.NewBoolType()))

	// Union-to-union: every branch of source must fit somewhere in target.
	intOrFloatTarget := rl.NewUnionType(rl.NewFloatType(), rl.NewStrType()) // accepts float (and int via widening), str
	assert.True(t, intOrFloatTarget.IsAssignableFrom(intOrFloat))            // int->float, float->float, both fit
	assert.False(t, intOrStr.IsAssignableFrom(intOrFloat))                   // float doesn't fit in int|str
}
