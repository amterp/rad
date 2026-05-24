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
	tc := &typeChecker{resolved: resolved, info: info}
	return tc, ident, sym
}

func TestInterpretCondition_NeqNullNarrowsTruthyToNonNull(t *testing.T) {
	// `x != null` where x: int?
	// Truthy branch should narrow x to int.
	tc, ident, sym := makeChecker(rl.NewOptionalType(rl.NewIntType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpNeq, ident, rl.NewLitNull(rl.Span{}))
	r := tc.interpretCondition(cond, nil)

	got, ok := r.WhenTrue[sym]
	require.True(t, ok, "truthy branch should narrow x")
	assert.Equal(t, rl.T_INT, got.Name())
	assert.Empty(t, r.WhenFalse, "falsy branch should record no narrowing (null side)")
}

func TestInterpretCondition_EqNullNarrowsFalsyToNonNull(t *testing.T) {
	// `x == null` where x: str? - inverse of the != case.
	tc, ident, sym := makeChecker(rl.NewOptionalType(rl.NewStrType()))
	cond := rl.NewOpBinary(rl.Span{}, rl.OpEq, ident, rl.NewLitNull(rl.Span{}))
	r := tc.interpretCondition(cond, nil)

	got, ok := r.WhenFalse[sym]
	require.True(t, ok, "falsy branch should narrow x to non-null")
	assert.Equal(t, rl.T_STR, got.Name())
	assert.Empty(t, r.WhenTrue, "truthy branch (null side) should record no narrowing")
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
