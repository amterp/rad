package ra

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_Basic(t *testing.T) {
	fs := NewFlagSet()

	boolFlag := fs.AddBool("foo").
		SetShort("f").
		SetUsage("foo usage here").
		SetDefault(true).
		Value
	strFlag := fs.AddString("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetDefault("alice").
		Value

	err := fs.Parse(os.Args)
	assert.Nil(t, err)

	assert.Equal(t, true, *boolFlag)
	assert.Equal(t, "alice", *strFlag)
}

func Test_OptionalString(t *testing.T) {
	fs := NewFlagSet()

	strFlag := fs.AddString("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetOptional(true).
		Value

	err := fs.Parse(os.Args)

	assert.Nil(t, err)
	assert.Nil(t, strFlag)
}

func Test_StringSliceMultiple(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag := fs.AddStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		Value

	err := fs.Parse([]string{"--bar", "alice", "--bar", "bob"})

	assert.Nil(t, err)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceSeparator(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag := fs.AddStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetSeparator("|").
		Value

	err := fs.Parse([]string{"--bar", "alice|bob"})

	assert.Nil(t, err)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceVariadic(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag := fs.AddStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetVariadic(true).
		Value

	err := fs.Parse([]string{"--bar", "alice", "bob"})

	assert.Nil(t, err)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceVariadicAndSeparator(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag := fs.AddStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetVariadic(true).
		SetSeparator(",").
		Value

	err := fs.Parse([]string{"--bar", "alice", "bob,charlie"})

	assert.Nil(t, err)
	assert.Equal(t, []string{"alice", "bob", "charlie"}, *strSliceFlag)
}

func Test_IntRangeConstraint(t *testing.T) {
	fs := NewFlagSet()

	intFlag := fs.AddInt("foo").
		SetMin(5).
		SetMax(10).
		Value

	err := fs.Parse([]string{"--foo", "7"})

	assert.Nil(t, err)
	assert.Equal(t, 7, *intFlag)
}

func Test_IntRangeConstraintErrors(t *testing.T) {
	fs := NewFlagSet()

	_ = fs.AddInt("foo").
		SetMin(5).
		SetMax(10).
		Value

	err := fs.Parse([]string{"--foo", "70"})

	assert.NotNil(t, err)
}
