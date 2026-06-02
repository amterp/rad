package check

import (
	"testing"

	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Phase 4b smoke tests for the null-predicate path through
// interpretCondition. We build a minimal typeChecker by hand so the
// tests stay focused on the predicate logic, not on the full
// parse/bind pipeline.

func TestStripNullFrom_OptionalReturnsInner(t *testing.T) {
	got := stripNullFrom(rl.NewOptionalType(rl.NewIntType()))
	require.NotNil(t, got)
	assert.Equal(t, rl.T_INT, got.Name())
}

func TestStripNullFrom_NonOptionalReturnsNil(t *testing.T) {
	assert.Nil(t, stripNullFrom(rl.NewIntType()))
	assert.Nil(t, stripNullFrom(rl.NewStrType()))
}

func TestStripNullFrom_UnionPeelsOptionalArm(t *testing.T) {
	// (int? | str) -> (int | str). The optional in arm 1 collapses
	// to its inner type; the other arm passes through.
	in := rl.NewUnionType(
		rl.NewOptionalType(rl.NewIntType()),
		rl.NewStrType(),
	)
	got := stripNullFrom(in)
	require.NotNil(t, got)
	u, ok := got.(*rl.TypingUnionT)
	require.True(t, ok, "union should remain a union")
	names := []string{u.Types()[0].Name(), u.Types()[1].Name()}
	assert.ElementsMatch(t, []string{rl.T_INT, rl.T_STR}, names)
}

func TestStripNullFrom_UnionAllNonOptionalReturnsNil(t *testing.T) {
	in := rl.NewUnionType(rl.NewIntType(), rl.NewStrType())
	assert.Nil(t, stripNullFrom(in))
}

// makeChecker wires a minimal typeChecker + Resolved + symbol such
// that interpretCondition has somewhere to look up the symbol's base
// type. Returns the checker, the identifier we'll compare against
// null, and the symbol it resolves to.
func makeChecker(baseType rl.TypingT) (*typeChecker, *rl.Identifier, *Symbol) {
	ident := rl.NewIdentifier(rl.Span{}, "x")
	sym := &Symbol{Name: "x", Kind: SymLocal}
	resolved := &Resolved{
		Uses: map[rl.Node]*Symbol{ident: sym},
	}
	info := &TypeInfo{
		SymbolTypes: map[*Symbol]rl.TypingT{sym: baseType},
		ExprTypes:   map[rl.Node]rl.TypingT{},
	}
	tc := &typeChecker{resolved: resolved, info: info, frame: NewFrame()}
	return tc, ident, sym
}

func TestInterpretCondition_NeqNullNarrowsTruthyToNonNull(t *testing.T) {
	// `x != null` where x: int?
	// Truthy branch narrows x to int; falsy branch narrows x to null.
	tc, ident, sym := makeChecker(rl.NewOptionalType(rl.NewIntType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNeq, ident, rl.NewLitNull(rl.Span{}))
	r := tc.interpretCondition(cond, nil)

	got, ok := r.WhenTrue[sym]
	require.True(t, ok, "truthy branch should narrow x to non-null")
	assert.Equal(t, rl.T_INT, got.Name())
	nullGot, ok := r.WhenFalse[sym]
	require.True(t, ok, "falsy branch should narrow x to null")
	assert.Equal(t, "null", nullGot.Name())
}

func TestInterpretCondition_EqNullNarrowsFalsyToNonNull(t *testing.T) {
	// `x == null` where x: str?
	// Truthy narrows x to null; falsy narrows x to str.
	tc, ident, sym := makeChecker(rl.NewOptionalType(rl.NewStrType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq, ident, rl.NewLitNull(rl.Span{}))
	r := tc.interpretCondition(cond, nil)

	got, ok := r.WhenFalse[sym]
	require.True(t, ok, "falsy branch should narrow x to non-null")
	assert.Equal(t, rl.T_STR, got.Name())
	nullGot, ok := r.WhenTrue[sym]
	require.True(t, ok, "truthy branch should narrow x to null")
	assert.Equal(t, "null", nullGot.Name())
}

func TestInterpretCondition_SwappedOperandsStillNarrows(t *testing.T) {
	// `null != x` is equivalent to `x != null`.
	tc, ident, sym := makeChecker(rl.NewOptionalType(rl.NewIntType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNeq, rl.NewLitNull(rl.Span{}), ident)
	r := tc.interpretCondition(cond, nil)

	got, ok := r.WhenTrue[sym]
	require.True(t, ok)
	assert.Equal(t, rl.T_INT, got.Name())
}

func TestInterpretCondition_NonOptionalBaseTypeNoOp(t *testing.T) {
	// `x != null` where x: int. The base type has no nullable
	// component, so neither side is refined. Falsy branch would
	// be Never if we had one, but we don't add diagnostics here.
	tc, ident, _ := makeChecker(rl.NewIntType())
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNeq, ident, rl.NewLitNull(rl.Span{}))
	r := tc.interpretCondition(cond, nil)

	assert.Empty(t, r.WhenTrue)
	assert.Empty(t, r.WhenFalse)
}

func TestInterpretCondition_UnknownPredicateReturnsEmpty(t *testing.T) {
	// `x < 5` isn't in the catalog. Should leave everything empty.
	tc, ident, _ := makeChecker(rl.NewIntType())
	cond := rl.NewOpBinary(rl.Span{}, rl.OpLt, ident, rl.NewLitInt(rl.Span{}, 5))
	r := tc.interpretCondition(cond, nil)

	assert.Empty(t, r.WhenTrue)
	assert.Empty(t, r.WhenFalse)
}

func TestInterpretCondition_NilConditionReturnsEmpty(t *testing.T) {
	tc, _, _ := makeChecker(rl.NewIntType())
	r := tc.interpretCondition(nil, nil)
	assert.Empty(t, r.WhenTrue)
	assert.Empty(t, r.WhenFalse)
}

// --- Phase 4c: type_of predicate ------------------------------------

// typeOfCall builds the AST for `type_of(<ident>)`.
func typeOfCall(ident *rl.Identifier) *rl.Call {
	fn := rl.NewIdentifier(rl.Span{}, "type_of")
	return rl.NewCall(rl.Span{}, fn, []rl.Node{ident}, nil)
}

func TestNarrowByTypeOf_UnionPartitions(t *testing.T) {
	// int | str, target="int" => truthy: int, falsy: str.
	base := rl.NewUnionType(rl.NewIntType(), rl.NewStrType())
	truthy, falsy := narrowByTypeOf(base, "int")
	require.NotNil(t, truthy)
	require.NotNil(t, falsy)
	assert.Equal(t, rl.T_INT, truthy.Name())
	assert.Equal(t, rl.T_STR, falsy.Name())
}

func TestNarrowByTypeOf_LeafMatch(t *testing.T) {
	truthy, falsy := narrowByTypeOf(rl.NewIntType(), "int")
	assert.Equal(t, rl.T_INT, truthy.Name())
	assert.Equal(t, rl.T_NEVER, falsy.Name())
}

func TestNarrowByTypeOf_LeafNoMatch(t *testing.T) {
	truthy, falsy := narrowByTypeOf(rl.NewIntType(), "str")
	assert.Equal(t, rl.T_NEVER, truthy.Name())
	assert.Equal(t, rl.T_INT, falsy.Name())
}

func TestNarrowByTypeOf_OptionalInnerMatches(t *testing.T) {
	// Optional<int>, target="int" => truthy: int, falsy: null
	// (TypingNullT - definite, the only remaining possibility).
	base := rl.NewOptionalType(rl.NewIntType())
	truthy, falsy := narrowByTypeOf(base, "int")
	require.NotNil(t, truthy)
	assert.Equal(t, rl.T_INT, truthy.Name())
	require.NotNil(t, falsy)
	assert.Equal(t, "null", falsy.Name(),
		"falsy is definite null after the inner-int branch is taken")
}

func TestNarrowByTypeOf_OptionalNullTarget(t *testing.T) {
	// Optional<int>, target="null" => truthy: null (definite),
	// falsy: int.
	base := rl.NewOptionalType(rl.NewIntType())
	truthy, falsy := narrowByTypeOf(base, "null")
	require.NotNil(t, truthy, "truthy is the definite-null arm")
	assert.Equal(t, "null", truthy.Name())
	require.NotNil(t, falsy)
	assert.Equal(t, rl.T_INT, falsy.Name())
}

func TestNarrowByTypeOf_UnionWithOptionalPreservesNullArm(t *testing.T) {
	// int?|str narrowed by type_of=="int".
	// Truthy: int (the optional's inner matched).
	// Falsy: int? (the optional's null half stays) and str (untouched).
	// This is the Phase 4h regression test for "union+optional drops null."
	base := rl.NewUnionType(
		rl.NewOptionalType(rl.NewIntType()),
		rl.NewStrType(),
	)
	truthy, falsy := narrowByTypeOf(base, "int")
	require.NotNil(t, truthy)
	assert.Equal(t, rl.T_INT, truthy.Name())
	require.NotNil(t, falsy)
	// Falsy should contain int? (null arm preserved) and str. Exact
	// representation depends on the join; what matters is the null
	// component isn't silently dropped.
	name := falsy.Name()
	assert.Contains(t, name, "?",
		"falsy must retain the nullable arm; got %q", name)
}

func TestNarrowByTypeOf_AnyShortCircuits(t *testing.T) {
	// any can't be partitioned without losing information.
	truthy, falsy := narrowByTypeOf(rl.NewAnyType(), "int")
	assert.Nil(t, truthy)
	assert.Nil(t, falsy)
}

func TestNarrowByTypeOf_StringEnumIsAStr(t *testing.T) {
	// A string-enum value still matches type_of(x) == "str".
	base := rl.NewStrEnumType("a", "b")
	truthy, falsy := narrowByTypeOf(base, "str")
	require.NotNil(t, truthy)
	assert.Equal(t, rl.T_NEVER, falsy.Name())
	// Truthy is the enum itself - more specific than plain str.
	assert.Equal(t, base.Name(), truthy.Name())
}

func TestInterpretCondition_TypeOfEqStrNarrows(t *testing.T) {
	// `type_of(x) == "int"` where x: int|str
	// Truthy: x is int. Falsy: x is str.
	tc, ident, sym := makeChecker(rl.NewUnionType(rl.NewIntType(), rl.NewStrType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq,
		typeOfCall(ident),
		rl.NewLitStringSimple(rl.Span{}, "int"))
	r := tc.interpretCondition(cond, nil)

	gotT, okT := r.WhenTrue[sym]
	gotF, okF := r.WhenFalse[sym]
	require.True(t, okT)
	require.True(t, okF)
	assert.Equal(t, rl.T_INT, gotT.Name())
	assert.Equal(t, rl.T_STR, gotF.Name())
}

func TestInterpretCondition_TypeOfNeqStrInverts(t *testing.T) {
	tc, ident, sym := makeChecker(rl.NewUnionType(rl.NewIntType(), rl.NewStrType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNeq,
		typeOfCall(ident),
		rl.NewLitStringSimple(rl.Span{}, "int"))
	r := tc.interpretCondition(cond, nil)

	// != flips: truthy is the non-matching arms (str), falsy is the
	// matching arm (int).
	gotT, _ := r.WhenTrue[sym]
	gotF, _ := r.WhenFalse[sym]
	assert.Equal(t, rl.T_STR, gotT.Name())
	assert.Equal(t, rl.T_INT, gotF.Name())
}

func TestInterpretCondition_TypeOfSwappedOperands(t *testing.T) {
	// "int" == type_of(x) is equivalent to type_of(x) == "int".
	tc, ident, sym := makeChecker(rl.NewUnionType(rl.NewIntType(), rl.NewStrType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq,
		rl.NewLitStringSimple(rl.Span{}, "int"),
		typeOfCall(ident))
	r := tc.interpretCondition(cond, nil)

	gotT, _ := r.WhenTrue[sym]
	assert.Equal(t, rl.T_INT, gotT.Name())
}

func TestInterpretCondition_TypeOfInvalidTargetMakesTruthyUnreachable(t *testing.T) {
	// type_of(x) == "frobnicate" - "frobnicate" isn't a valid type_of
	// return, so the equality is statically false. Truthy is Never.
	tc, ident, sym := makeChecker(rl.NewIntType())
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq,
		typeOfCall(ident),
		rl.NewLitStringSimple(rl.Span{}, "frobnicate"))
	r := tc.interpretCondition(cond, nil)

	gotT, _ := r.WhenTrue[sym]
	gotF, _ := r.WhenFalse[sym]
	require.NotNil(t, gotT)
	assert.Equal(t, rl.T_NEVER, gotT.Name())
	assert.Equal(t, rl.T_INT, gotF.Name())
}

// --- Phase 4d: string-enum and `in [lits]` narrowing -----------------

func TestPartitionStrEnum_BasicSplit(t *testing.T) {
	enum := rl.NewStrEnumType("a", "b", "c")
	truthy, falsy := partitionStrEnum(enum, map[string]bool{"b": true})
	require.NotNil(t, truthy)
	require.NotNil(t, falsy)

	tE, ok := truthy.(*rl.TypingStrEnumT)
	require.True(t, ok)
	assert.Equal(t, []string{"b"}, tE.Values())

	fE, ok := falsy.(*rl.TypingStrEnumT)
	require.True(t, ok)
	assert.Equal(t, []string{"a", "c"}, fE.Values())
}

func TestPartitionStrEnum_NoMatchTruthyIsNever(t *testing.T) {
	enum := rl.NewStrEnumType("a", "b")
	truthy, falsy := partitionStrEnum(enum, map[string]bool{"z": true})
	assert.Equal(t, rl.T_NEVER, truthy.Name())
	require.NotNil(t, falsy)
	fE, _ := falsy.(*rl.TypingStrEnumT)
	assert.Equal(t, []string{"a", "b"}, fE.Values())
}

func TestPartitionStrEnum_FullMatchFalsyIsNever(t *testing.T) {
	enum := rl.NewStrEnumType("a", "b")
	truthy, falsy := partitionStrEnum(enum, map[string]bool{"a": true, "b": true})
	require.NotNil(t, truthy)
	tE, _ := truthy.(*rl.TypingStrEnumT)
	assert.Equal(t, []string{"a", "b"}, tE.Values())
	assert.Equal(t, rl.T_NEVER, falsy.Name())
}

func TestInterpretCondition_StrEnumEqLiteral(t *testing.T) {
	// x: ["a", "b", "c"]; `x == "b"` narrows truthy to ["b"], falsy to ["a", "c"].
	tc, ident, sym := makeChecker(rl.NewStrEnumType("a", "b", "c"))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq, ident, rl.NewLitStringSimple(rl.Span{}, "b"))
	r := tc.interpretCondition(cond, nil)

	gotT, okT := r.WhenTrue[sym]
	gotF, okF := r.WhenFalse[sym]
	require.True(t, okT)
	require.True(t, okF)
	assert.Equal(t, `["b"]`, gotT.Name())
	assert.Equal(t, `["a", "c"]`, gotF.Name())
}

func TestInterpretCondition_StrEnumNeqLiteralInverts(t *testing.T) {
	tc, ident, sym := makeChecker(rl.NewStrEnumType("a", "b", "c"))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNeq, ident, rl.NewLitStringSimple(rl.Span{}, "b"))
	r := tc.interpretCondition(cond, nil)

	gotT, _ := r.WhenTrue[sym]
	gotF, _ := r.WhenFalse[sym]
	assert.Equal(t, `["a", "c"]`, gotT.Name())
	assert.Equal(t, `["b"]`, gotF.Name())
}

func TestInterpretCondition_StrEnumPlainStrNoNarrowing(t *testing.T) {
	// Plain str shouldn't narrow to a singleton enum - that surprises
	// users who declared the var as str.
	tc, ident, sym := makeChecker(rl.NewStrType())
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq, ident, rl.NewLitStringSimple(rl.Span{}, "x"))
	r := tc.interpretCondition(cond, nil)

	_, okT := r.WhenTrue[sym]
	_, okF := r.WhenFalse[sym]
	assert.False(t, okT)
	assert.False(t, okF)
}

func TestInterpretCondition_InStringList(t *testing.T) {
	// x: ["a", "b", "c", "d"]; `x in ["a", "c"]` narrows truthy to ["a","c"], falsy to ["b","d"].
	tc, ident, sym := makeChecker(rl.NewStrEnumType("a", "b", "c", "d"))
	listLit := rl.NewLitList(rl.Span{}, []rl.Node{
		rl.NewLitStringSimple(rl.Span{}, "a"),
		rl.NewLitStringSimple(rl.Span{}, "c"),
	})
	cond := rl.NewOpBinary(rl.Span{}, rl.OpIn, ident, listLit)
	r := tc.interpretCondition(cond, nil)

	gotT, _ := r.WhenTrue[sym]
	gotF, _ := r.WhenFalse[sym]
	assert.Equal(t, `["a", "c"]`, gotT.Name())
	assert.Equal(t, `["b", "d"]`, gotF.Name())
}

func TestInterpretCondition_NotInStringListInverts(t *testing.T) {
	tc, ident, sym := makeChecker(rl.NewStrEnumType("a", "b", "c"))
	listLit := rl.NewLitList(rl.Span{}, []rl.Node{
		rl.NewLitStringSimple(rl.Span{}, "a"),
	})
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNotIn, ident, listLit)
	r := tc.interpretCondition(cond, nil)

	gotT, _ := r.WhenTrue[sym]
	gotF, _ := r.WhenFalse[sym]
	assert.Equal(t, `["b", "c"]`, gotT.Name())
	assert.Equal(t, `["a"]`, gotF.Name())
}

func TestInterpretCondition_InListWithNonStringElementDisqualifies(t *testing.T) {
	// Mixed list (a string + an int) bails out of the pattern.
	tc, ident, sym := makeChecker(rl.NewStrEnumType("a", "b"))
	listLit := rl.NewLitList(rl.Span{}, []rl.Node{
		rl.NewLitStringSimple(rl.Span{}, "a"),
		rl.NewLitInt(rl.Span{}, 7),
	})
	cond := rl.NewOpBinary(rl.Span{}, rl.OpIn, ident, listLit)
	r := tc.interpretCondition(cond, nil)

	_, okT := r.WhenTrue[sym]
	_, okF := r.WhenFalse[sym]
	assert.False(t, okT)
	assert.False(t, okF)
}

func TestInterpretCondition_TypeOfOnDynamicNoNarrowing(t *testing.T) {
	tc, ident, sym := makeChecker(rl.NewDynamicType())
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq,
		typeOfCall(ident),
		rl.NewLitStringSimple(rl.Span{}, "int"))
	r := tc.interpretCondition(cond, nil)

	_, okT := r.WhenTrue[sym]
	_, okF := r.WhenFalse[sym]
	assert.False(t, okT)
	assert.False(t, okF)
}
