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

func TestFrame_MembershipRecordedAndNotMutating(t *testing.T) {
	xs := &check.Symbol{Name: "xs"}
	f := &check.Symbol{Name: "f"}
	fact := check.MembershipFact{Container: xs, Element: f}

	base := check.NewFrame()
	withFact := base.WithMembership(fact)

	assert.True(t, withFact.HasMembership(fact))
	assert.False(t, base.HasMembership(fact), "WithMembership must not mutate the receiver")
}

func TestFrame_MembershipIsKeyedOnBothSymbols(t *testing.T) {
	xs := &check.Symbol{Name: "xs"}
	ys := &check.Symbol{Name: "ys"}
	f := &check.Symbol{Name: "f"}

	frame := check.NewFrame().WithMembership(check.MembershipFact{Container: xs, Element: f})

	assert.True(t, frame.HasMembership(check.MembershipFact{Container: xs, Element: f}))
	assert.False(t, frame.HasMembership(check.MembershipFact{Container: ys, Element: f}),
		"a guard on one container must not refine a lookup in another")
	assert.False(t, frame.HasMembership(check.MembershipFact{Container: f, Element: xs}),
		"container and element are not interchangeable")
}

func TestFrame_MembershipSurvivesNarrowing(t *testing.T) {
	xs := &check.Symbol{Name: "xs"}
	f := &check.Symbol{Name: "f"}
	other := &check.Symbol{Name: "other"}
	fact := check.MembershipFact{Container: xs, Element: f}

	frame := check.NewFrame().WithMembership(fact)

	assert.True(t, frame.With(other, rl.NewIntType()).HasMembership(fact),
		"With must carry the membership set forward")
	assert.True(t, frame.WithMany(map[*check.Symbol]rl.TypingT{other: rl.NewIntType()}).HasMembership(fact),
		"WithMany must carry the membership set forward")
}

func TestFrame_MembershipReaddIsANoOp(t *testing.T) {
	xs := &check.Symbol{Name: "xs"}
	f := &check.Symbol{Name: "f"}
	fact := check.MembershipFact{Container: xs, Element: f}

	frame := check.NewFrame().WithMembership(fact)
	assert.Same(t, frame, frame.WithMembership(fact), "re-proving a known fact should not allocate")
}

// A nil *Frame is passed around by interpretCondition's callers, so the whole
// Frame API has to tolerate it rather than dereference the receiver.
func TestFrame_NilReceiverIsSafe(t *testing.T) {
	var nilFrame *check.Frame
	xs := &check.Symbol{Name: "xs"}
	f := &check.Symbol{Name: "f"}
	fact := check.MembershipFact{Container: xs, Element: f}

	assert.False(t, nilFrame.HasMembership(fact))
	assert.True(t, nilFrame.WithMembership(fact).HasMembership(fact))
	assert.False(t, nilFrame.With(xs, rl.NewIntType()).HasMembership(fact))
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
