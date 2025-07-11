package ra

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Basic(t *testing.T) {
	fs := NewFlagSet()

	boolFlag, err := NewBool("foo").
		SetShort("f").
		SetUsage("foo usage here").
		SetDefault(true).
		Register(fs)
	assert.NoError(t, err)

	strFlag, err := NewString("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetDefault("alice").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse(os.Args)
	assert.Nil(t, parseErr)

	assert.Equal(t, true, *boolFlag)
	assert.Equal(t, "alice", *strFlag)
}

func Test_OptionalString(t *testing.T) {
	fs := NewFlagSet()

	strFlag, err := NewString("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetOptional(true).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse(os.Args)
	assert.Nil(t, parseErr)

	assert.NotNil(t, strFlag)
	assert.Equal(t, "", *strFlag)
}

func Test_StringSliceMultiple(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--bar", "alice", "--bar", "bob"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceSeparator(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetSeparator("|").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--bar", "alice|bob"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceVariadic(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetVariadic(true).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--bar", "alice", "bob"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceVariadicAndSeparator(t *testing.T) {
	fs := NewFlagSet()

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetVariadic(true).
		SetSeparator(",").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--bar", "alice", "bob,charlie"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob", "charlie"}, *strSliceFlag)
}

func Test_IntRangeConstraint(t *testing.T) {
	fs := NewFlagSet()

	intFlag, err := NewInt("foo").
		SetMin(5).
		SetMax(10).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--foo", "7"})
	assert.Nil(t, parseErr)
	assert.Equal(t, 7, *intFlag)
}

func Test_IntRangeConstraintErrors(t *testing.T) {
	fs := NewFlagSet()

	_, err := NewInt("foo").
		SetMin(5).
		SetMax(10).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--foo", "70"})
	assert.NotNil(t, parseErr)
}
