package check_test

import (
	"testing"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
)

func TestFrame_NewIsEmpty(t *testing.T) {
	f := check.NewFrame()
	sym := &check.Symbol{Name: "x"}
	_, ok := f.Lookup(sym)
	assert.False(t, ok, "empty frame should have no narrowings")
}

func TestFrame_WithRecordsNarrowing(t *testing.T) {
	f := check.NewFrame()
	sym := &check.Symbol{Name: "x"}
	intT := rl.NewIntType()

	f2 := f.With(sym, intT)
	got, ok := f2.Lookup(sym)
	assert.True(t, ok)
	assert.Equal(t, rl.T_INT, got.Name())

	// Parent frame must be untouched.
	_, ok = f.Lookup(sym)
	assert.False(t, ok, "With must not mutate the receiver")
}

func TestFrame_LookupWalksParentChain(t *testing.T) {
	x := &check.Symbol{Name: "x"}
	y := &check.Symbol{Name: "y"}

	f1 := check.NewFrame().With(x, rl.NewIntType())
	f2 := f1.With(y, rl.NewStrType())

	gotX, okX := f2.Lookup(x)
	gotY, okY := f2.Lookup(y)
	assert.True(t, okX)
	assert.True(t, okY)
	assert.Equal(t, rl.T_INT, gotX.Name())
	assert.Equal(t, rl.T_STR, gotY.Name())
}

func TestFrame_ChildShadowsParent(t *testing.T) {
	x := &check.Symbol{Name: "x"}
	f1 := check.NewFrame().With(x, rl.NewIntType())
	f2 := f1.With(x, rl.NewStrType())

	got1, _ := f1.Lookup(x)
	got2, _ := f2.Lookup(x)
	assert.Equal(t, rl.T_INT, got1.Name(), "ancestor lookup must see ancestor's binding")
	assert.Equal(t, rl.T_STR, got2.Name(), "descendant lookup must see closer binding first")
}

func TestFrame_WithManyEmpty(t *testing.T) {
	f := check.NewFrame()
	// Empty map should return the same frame, not a new one.
	assert.Same(t, f, f.WithMany(nil))
	assert.Same(t, f, f.WithMany(map[*check.Symbol]rl.TypingT{}))
}

func TestFrame_WithManyBindsMultiple(t *testing.T) {
	x := &check.Symbol{Name: "x"}
	y := &check.Symbol{Name: "y"}
	f := check.NewFrame().WithMany(map[*check.Symbol]rl.TypingT{
		x: rl.NewIntType(),
		y: rl.NewStrType(),
	})
	gotX, _ := f.Lookup(x)
	gotY, _ := f.Lookup(y)
	assert.Equal(t, rl.T_INT, gotX.Name())
	assert.Equal(t, rl.T_STR, gotY.Name())
}

func TestRefinement_NegateSwapsSides(t *testing.T) {
	x := &check.Symbol{Name: "x"}
	r := check.Refinement{
		WhenTrue:  map[*check.Symbol]rl.TypingT{x: rl.NewIntType()},
		WhenFalse: map[*check.Symbol]rl.TypingT{x: rl.NewStrType()},
	}
	neg := r.Negate()
	assert.Equal(t, rl.T_STR, neg.WhenTrue[x].Name())
	assert.Equal(t, rl.T_INT, neg.WhenFalse[x].Name())
}
