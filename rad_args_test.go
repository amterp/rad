package ra

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Basic(t *testing.T) {
	fs := NewCmd()

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
	fs := NewCmd()

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
	fs := NewCmd()

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
	fs := NewCmd()

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
	fs := NewCmd()

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
	fs := NewCmd()

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
	fs := NewCmd()

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
	fs := NewCmd()

	_, err := NewInt("foo").
		SetMin(5).
		SetMax(10).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.Parse([]string{"--foo", "70"})
	assert.NotNil(t, parseErr)
}

func Test_Cmds(t *testing.T) {
	addCmd := NewCmd()
	addFile, err := NewString("file").
		Register(addCmd)
	assert.NoError(t, err)

	rmCmd := NewCmd()
	rmName, err := NewInt("name").
		Register(rmCmd)
	assert.NoError(t, err)

	rootCmd := NewCmd()

	addInvoked, err := rootCmd.RegisterCmd("add", addCmd)
	assert.NoError(t, err)

	rmInvoked, err := rootCmd.RegisterCmd("rm", rmCmd)
	assert.NoError(t, err)

	parseErr := rootCmd.Parse([]string{"add", "--file", "test.txt"})
	assert.NotNil(t, parseErr)
	assert.True(t, *addInvoked)
	assert.False(t, *rmInvoked)
	assert.Nil(t, rmName)
	assert.Equal(t, "test.txt", *addFile)
}
