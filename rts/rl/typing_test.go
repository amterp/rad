package rl_test

import (
	"testing"

	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
)

// Pure unit tests over TypingT.IsCompatibleWith. The runtime currently exercises
// these only at function-signature boundaries; these tests pin the behavior we
// rely on there (and protect the recursive collection-element checks that are
// otherwise easy to silently regress).

func TestPrimitiveCompatibility(t *testing.T) {
	intT := rl.NewIntType()
	strT := rl.NewStrType()
	floatT := rl.NewFloatType()
	boolT := rl.NewBoolType()

	// IntT
	assert.True(t, intT.IsCompatibleWith(rl.NewIntSubject(5)))
	assert.False(t, intT.IsCompatibleWith(rl.NewFloatSubject(5.0)))
	assert.False(t, intT.IsCompatibleWith(rl.NewStrSubject("5")))
	assert.False(t, intT.IsCompatibleWith(rl.NewBoolSubject(true)))

	// FloatT accepts ints (widening)
	assert.True(t, floatT.IsCompatibleWith(rl.NewFloatSubject(5.5)))
	assert.True(t, floatT.IsCompatibleWith(rl.NewIntSubject(5)))
	assert.False(t, floatT.IsCompatibleWith(rl.NewStrSubject("5")))

	// StrT
	assert.True(t, strT.IsCompatibleWith(rl.NewStrSubject("x")))
	assert.False(t, strT.IsCompatibleWith(rl.NewIntSubject(1)))

	// BoolT
	assert.True(t, boolT.IsCompatibleWith(rl.NewBoolSubject(false)))
	assert.False(t, boolT.IsCompatibleWith(rl.NewIntSubject(0)))
}

func TestAnyMatchesEverything(t *testing.T) {
	anyT := rl.NewAnyType()
	assert.True(t, anyT.IsCompatibleWith(rl.NewIntSubject(1)))
	assert.True(t, anyT.IsCompatibleWith(rl.NewStrSubject("x")))
	assert.True(t, anyT.IsCompatibleWith(rl.NewNullSubject()))
	assert.True(t, anyT.IsCompatibleWith(rl.NewListSubject()))
}

func TestVoidOnlyMatchesVoid(t *testing.T) {
	voidT := rl.NewVoidType()
	assert.True(t, voidT.IsCompatibleWith(rl.NewVoidSubject()))
	assert.False(t, voidT.IsCompatibleWith(rl.NewIntSubject(0)))
	assert.False(t, voidT.IsCompatibleWith(rl.NewNullSubject()))
}

func TestOptionalAllowsNull(t *testing.T) {
	optInt := rl.NewOptionalType(rl.NewIntType())
	assert.True(t, optInt.IsCompatibleWith(rl.NewIntSubject(5)))
	assert.True(t, optInt.IsCompatibleWith(rl.NewNullSubject()))
	assert.False(t, optInt.IsCompatibleWith(rl.NewStrSubject("x")))
}

func TestUnionMatchesAnyBranch(t *testing.T) {
	intOrStr := rl.NewUnionType(rl.NewIntType(), rl.NewStrType())
	assert.True(t, intOrStr.IsCompatibleWith(rl.NewIntSubject(1)))
	assert.True(t, intOrStr.IsCompatibleWith(rl.NewStrSubject("x")))
	assert.False(t, intOrStr.IsCompatibleWith(rl.NewBoolSubject(true)))
}

func TestListShallowElements(t *testing.T) {
	intList := rl.NewListType(rl.NewIntType())

	// Empty list passes
	assert.True(t, intList.IsCompatibleWith(makeListSubject()))
	// All ints pass
	assert.True(t, intList.IsCompatibleWith(makeListSubject(int64(1), int64(2), int64(3))))
	// Mixed: one bad element fails (this works even pre-fix because outer check is value-aware)
	assert.False(t, intList.IsCompatibleWith(makeListSubject(int64(1), "bad")))
}

// Regression test for the NewSubject value-propagation bug. Pre-fix, the outer
// list iterates its elements but each inner []interface{} element gets wrapped
// in a Val-less subject by NewSubject, so the inner type check passes vacuously.
// Post-fix, the inner check sees the inner list contents and rejects "bad".
func TestListNestedRegression(t *testing.T) {
	intMatrix := rl.NewListType(rl.NewListType(rl.NewIntType()))

	good := makeListSubject(
		[]interface{}{int64(1), int64(2)},
		[]interface{}{int64(3)},
	)
	bad := makeListSubject(
		[]interface{}{int64(1), int64(2)},
		[]interface{}{int64(3), "bad"},
	)

	assert.True(t, intMatrix.IsCompatibleWith(good))
	assert.False(t, intMatrix.IsCompatibleWith(bad), "nested int[][] must reject string element")
}

func TestTupleNested(t *testing.T) {
	// [int, int[]]
	tup := rl.NewTupleType(rl.NewIntType(), rl.NewListType(rl.NewIntType()))

	good := makeListSubject(int64(1), []interface{}{int64(2), int64(3)})
	badInner := makeListSubject(int64(1), []interface{}{int64(2), "x"})
	wrongLen := makeListSubject(int64(1))

	assert.True(t, tup.IsCompatibleWith(good))
	assert.False(t, tup.IsCompatibleWith(badInner), "nested element type must be checked")
	assert.False(t, tup.IsCompatibleWith(wrongLen))
}

func TestMapNested(t *testing.T) {
	// { str: int[] }
	mapT := rl.NewMapType(rl.NewStrType(), rl.NewListType(rl.NewIntType()))

	good := makeMapSubject(map[string]interface{}{
		"a": []interface{}{int64(1), int64(2)},
	})
	bad := makeMapSubject(map[string]interface{}{
		"a": []interface{}{int64(1), "no"},
	})

	assert.True(t, mapT.IsCompatibleWith(good))
	assert.False(t, mapT.IsCompatibleWith(bad), "nested map value type must be checked")
}

func TestStructNested(t *testing.T) {
	// { "outer": { "inner": int } }
	inner := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("inner", false): rl.NewIntType(),
	})
	outer := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("outer", false): inner,
	})

	good := makeMapSubject(map[string]interface{}{
		"outer": map[string]interface{}{"inner": int64(1)},
	})
	badInner := makeMapSubject(map[string]interface{}{
		"outer": map[string]interface{}{"inner": "no"},
	})
	missingRequired := makeMapSubject(map[string]interface{}{
		"outer": map[string]interface{}{},
	})

	assert.True(t, outer.IsCompatibleWith(good))
	assert.False(t, outer.IsCompatibleWith(badInner), "nested struct field type must be checked")
	assert.False(t, outer.IsCompatibleWith(missingRequired))
}

func TestStructOptionalField(t *testing.T) {
	s := rl.NewStructType(map[rl.MapNamedKey]rl.TypingT{
		rl.NewMapNamedKey("req", false): rl.NewIntType(),
		rl.NewMapNamedKey("opt", true):  rl.NewStrType(),
	})

	withOpt := makeMapSubject(map[string]interface{}{
		"req": int64(1),
		"opt": "hello",
	})
	withoutOpt := makeMapSubject(map[string]interface{}{
		"req": int64(1),
	})
	missingReq := makeMapSubject(map[string]interface{}{
		"opt": "hello",
	})

	assert.True(t, s.IsCompatibleWith(withOpt))
	assert.True(t, s.IsCompatibleWith(withoutOpt))
	assert.False(t, s.IsCompatibleWith(missingReq))
}

func TestFnNameFormatting(t *testing.T) {
	intT := rl.TypingT(rl.NewIntType())
	strT := rl.TypingT(rl.NewStrType())
	boolT := rl.TypingT(rl.NewBoolType())

	cases := []struct {
		name string
		fn   *rl.TypingFnT
		want string
	}{
		{
			name: "no params no return",
			fn:   &rl.TypingFnT{},
			want: "fn()",
		},
		{
			name: "with return",
			fn:   &rl.TypingFnT{ReturnT: &boolT},
			want: "fn() -> bool",
		},
		{
			name: "with params and return",
			fn: &rl.TypingFnT{
				Params: []rl.TypingFnParam{
					{Type: &intT},
					{Type: &strT},
				},
				ReturnT: &boolT,
			},
			want: "fn(int, str) -> bool",
		},
		{
			name: "untyped param renders as any",
			fn: &rl.TypingFnT{
				Params:  []rl.TypingFnParam{{}},
				ReturnT: &boolT,
			},
			want: "fn(any) -> bool",
		},
		{
			name: "variadic param prefixed with *",
			fn: &rl.TypingFnT{
				Params: []rl.TypingFnParam{{Type: &intT, IsVariadic: true}},
			},
			want: "fn(*int)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.fn.Name())
		})
	}
}

func TestFnIsCompatibleAtTypeLevel(t *testing.T) {
	fn := &rl.TypingFnT{}

	// A fn-typed value matches any fn declaration today (type-level only).
	assert.True(t, fn.IsCompatibleWith(rl.NewFnSubject()))
	// Non-fn values are rejected.
	assert.False(t, fn.IsCompatibleWith(rl.NewIntSubject(1)))
	assert.False(t, fn.IsCompatibleWith(rl.NewNullSubject()))
}

// --- helpers ---

func makeListSubject(elems ...interface{}) rl.TypingCompatVal {
	s := rl.NewListSubject()
	s.Val = []interface{}(elems)
	return s
}

func makeMapSubject(m map[string]interface{}) rl.TypingCompatVal {
	s := rl.NewMapSubject()
	s.Val = m
	return s
}
