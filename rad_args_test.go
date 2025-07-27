package ra

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Basic(t *testing.T) {
	fs := NewCmd("test")

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

	parseErr := fs.ParseOrError([]string{})
	assert.Nil(t, parseErr)

	assert.Equal(t, true, *boolFlag)
	assert.Equal(t, "alice", *strFlag)
}

func Test_OptionalString(t *testing.T) {
	fs := NewCmd("test")

	strFlag, err := NewString("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetOptional(true).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{})
	assert.Nil(t, parseErr)

	assert.NotNil(t, strFlag)
	assert.Equal(t, "", *strFlag)
}

func Test_StringSliceMultiple(t *testing.T) {
	fs := NewCmd("test")

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{"--bar", "alice", "--bar", "bob"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceSeparator(t *testing.T) {
	fs := NewCmd("test")

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetSeparator("|").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{"--bar", "alice|bob"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceVariadic(t *testing.T) {
	fs := NewCmd("test")

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetVariadic(true).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{"--bar", "alice", "bob"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob"}, *strSliceFlag)
}

func Test_StringSliceVariadicAndSeparator(t *testing.T) {
	fs := NewCmd("test")

	strSliceFlag, err := NewStringSlice("bar").
		SetShort("b").
		SetUsage("bar usage here").
		SetVariadic(true).
		SetSeparator(",").
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{"--bar", "alice", "bob,charlie"})
	assert.Nil(t, parseErr)
	assert.Equal(t, []string{"alice", "bob", "charlie"}, *strSliceFlag)
}

func Test_IntRangeConstraint(t *testing.T) {
	fs := NewCmd("test")

	intFlag, err := NewInt("foo").
		SetMin(5, true).
		SetMax(10, true).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{"--foo", "7"})
	assert.Nil(t, parseErr)
	assert.Equal(t, 7, *intFlag)
}

func Test_IntRangeConstraintErrors(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt("foo").
		SetMin(5, true).
		SetMax(10, true).
		Register(fs)
	assert.NoError(t, err)

	parseErr := fs.ParseOrError([]string{"--foo", "70"})
	assert.NotNil(t, parseErr)
}

func Test_Cmds(t *testing.T) {
	addCmd := NewCmd("add")
	addFile, err := NewString("file").
		Register(addCmd)
	assert.NoError(t, err)

	rmCmd := NewCmd("rm")
	rmName, err := NewInt("name").
		Register(rmCmd)
	assert.NoError(t, err)

	rootCmd := NewCmd("root")

	addInvoked, err := rootCmd.RegisterCmd(addCmd)
	assert.NoError(t, err)

	rmInvoked, err := rootCmd.RegisterCmd(rmCmd)
	assert.NoError(t, err)

	parseErr := rootCmd.ParseOrError([]string{"add", "--file", "test.txt"})
	assert.Nil(t, parseErr)
	assert.True(t, *addInvoked)
	assert.False(t, *rmInvoked)
	assert.Equal(t, 0, *rmName) // rmName should have default value since rm command not used
	assert.Equal(t, "test.txt", *addFile)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Positional arguments: basic assignment rules
   ────────────────────────────────────────────────────────────────────────────
*/

// --arg1 already set, so "bbb" falls into arg2
func Test_PositionalAssignmentLeftToRight(t *testing.T) {
	fs := NewCmd("test")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewString("arg2").Register(fs)

	err := fs.ParseOrError([]string{"--arg1=aaa", "bbb"})
	assert.Nil(t, err)

	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, "bbb", *arg2)
}

// positional assignment then named flag override - named flag wins
func Test_PositionalThenNamedFlagOverride(t *testing.T) {
	fs := NewCmd("test")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewString("arg2").SetOptional(true).Register(fs)

	err := fs.ParseOrError([]string{"aaa", "--arg1=bbb"})
	assert.Nil(t, err)
	assert.Equal(t, "bbb", *arg1) // named flag overrides positional
	assert.Equal(t, "", *arg2)    // no value assigned
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Short‑flag clusters
   ────────────────────────────────────────────────────────────────────────────
*/

// -bcd where b,c,d are bools; “aaa” becomes first positional arg
func Test_ShortBoolCluster(t *testing.T) {
	fs := NewCmd("test")

	pos, _ := NewString("arg1").Register(fs)
	b, _ := NewBool("b").SetShort("b").Register(fs)
	c, _ := NewBool("c").SetShort("c").Register(fs)
	d, _ := NewBool("d").SetShort("d").Register(fs)
	e, _ := NewBool("e").SetShort("e").Register(fs) // never set

	err := fs.ParseOrError([]string{"-bcd", "aaa"})
	assert.Nil(t, err)

	assert.Equal(t, "aaa", *pos)
	assert.True(t, *b)
	assert.True(t, *c)
	assert.True(t, *d)
	assert.False(t, *e)
}

// cluster terminates at non‑bool flag
func Test_ShortClusterEndsWithNonBool(t *testing.T) {
	fs := NewCmd("test")

	a, _ := NewBool("a").SetShort("a").Register(fs)
	b, _ := NewBool("b").SetShort("b").Register(fs)
	c, _ := NewString("c").SetShort("c").Register(fs)

	err := fs.ParseOrError([]string{"-abc", "ddd"})
	assert.Nil(t, err)

	assert.True(t, *a)
	assert.True(t, *b)
	assert.Equal(t, "ddd", *c)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Negative numbers & “number‑shorts” mode
   ────────────────────────────────────────────────────────────────────────────
*/

// No int‑shorts defined → “-1” and “-2” are values, not flags.
func Test_NegativeIntsWithoutNumberShortMode(t *testing.T) {
	fs := NewCmd("test")

	val1, _ := NewInt("arg1").Register(fs)
	val2, _ := NewInt("arg2").Register(fs)

	err := fs.ParseOrError([]string{"-1", "--arg2", "-2"})
	assert.Nil(t, err)

	assert.Equal(t, -1, *val1)
	assert.Equal(t, -2, *val2)
}

// Defining an int‑short activates number‑shorts mode.
func Test_NumberShortsMode(t *testing.T) {
	fs := NewCmd("test")

	// arg1 is string to capture --arg1 value
	arg1, _ := NewString("arg1").Register(fs)
	// int short “2” – activates the mode
	arg2, _ := NewInt("arg2").SetShort("2").Register(fs)

	err := fs.ParseOrError([]string{"--arg1=-2", "-2", "42"})
	assert.Nil(t, err)

	assert.Equal(t, "-2", *arg1) // parsed via =, so literal -2
	assert.Equal(t, 42, *arg2)   // “-2” consumed flag, next token 42
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Variadic positional & multiple variadics
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_PositionalAndFlagVariadics(t *testing.T) {
	fs := NewCmd("test")

	posVar, _ := NewStringSlice("arg1").SetVariadic(true).Register(fs) // positional variadic
	flagVar, _ := NewStringSlice("arg2").
		SetVariadic(true).
		SetShort("e").
		Register(fs)

	err := fs.ParseOrError([]string{"aaa", "bbb", "--arg2", "ccc", "ddd", "-e", "eee"})
	assert.Nil(t, err)

	assert.Equal(t, []string{"aaa", "bbb"}, *posVar)
	assert.Equal(t, []string{"ccc", "ddd", "eee"}, *flagVar)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Configured() helper & unknown‑flag error
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_ConfiguredAndDefaults(t *testing.T) {
	fs := NewCmd("test")

	str, _ := NewString("foo").SetDefault("bar").Register(fs)
	assert.False(t, fs.Configured("foo")) // default only

	err := fs.ParseOrError([]string{"--foo", "baz"})
	assert.Nil(t, err)

	assert.True(t, fs.Configured("foo"))
	assert.Equal(t, "baz", *str)
}

func Test_UnknownFlagProducesError(t *testing.T) {
	fs := NewCmd("test")

	err := fs.ParseOrError([]string{"--does-not-exist"})
	assert.NotNil(t, err)
}

/*
   ───────────────────────────────────────────────────────
   Additional flag types: Int64, Float64, Bool slices
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_Int64Flag(t *testing.T) {
	fs := NewCmd("test")

	intFlag, err := NewInt64("value").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--value", "9223372036854775807"})
	assert.Nil(t, err)
	assert.Equal(t, int64(9223372036854775807), *intFlag)
}

func Test_Float64Flag(t *testing.T) {
	fs := NewCmd("test")

	floatFlag, err := NewFloat64("value").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--value", "3.14159"})
	assert.Nil(t, err)
	assert.Equal(t, 3.14159, *floatFlag)
}

func Test_BoolSliceFlag(t *testing.T) {
	fs := NewCmd("test")

	boolSliceFlag, err := NewBoolSlice("flags").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--flags", "true", "--flags", "false", "--flags", "1", "--flags", "0"})
	assert.Nil(t, err)
	assert.Equal(t, []bool{true, false, true, false}, *boolSliceFlag)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   String constraints: Enum and Regex
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_StringEnumConstraint(t *testing.T) {
	fs := NewCmd("test")

	enumFlag, err := NewString("level").
		SetEnumConstraint([]string{"debug", "info", "warn", "error"}).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--level", "info"})
	assert.Nil(t, err)
	assert.Equal(t, "info", *enumFlag)
}

func Test_StringEnumConstraintError(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewString("level").
		SetEnumConstraint([]string{"debug", "info", "warn", "error"}).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--level", "invalid"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid 'level' value: invalid (valid values: debug, info, warn, error)")
}

func Test_StringRegexConstraint(t *testing.T) {
	fs := NewCmd("test")

	regexFlag, err := NewString("email").
		SetRegexConstraint(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--email", "test@example.com"})
	assert.Nil(t, err)
	assert.Equal(t, "test@example.com", *regexFlag)
}

func Test_StringRegexConstraintError(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewString("name").
		SetRegexConstraint(regexp.MustCompile(`^[A-Z][a-z]*$`)).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--name", "alice"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid 'name' value: alice")
	assert.Contains(t, err.Error(), "must match regex: ^[A-Z][a-z]*$")
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Float64 constraints
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_Float64RangeConstraint(t *testing.T) {
	fs := NewCmd("test")

	floatFlag, err := NewFloat64("value").
		SetMin(0.0, true).
		SetMax(100.0, true).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--value", "42.5"})
	assert.Nil(t, err)
	assert.Equal(t, 42.5, *floatFlag)
}

func Test_Float64RangeConstraintError(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewFloat64("value").
		SetMin(0.0, true).
		SetMax(100.0, true).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--value", "150.0"})
	assert.NotNil(t, err)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Global flags
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_GlobalFlags(t *testing.T) {
	subCmd := NewCmd("sub")
	subArg, _ := NewString("subarg").Register(subCmd)

	rootCmd := NewCmd("root")
	globalFlag, err := NewBool("verbose").
		SetShort("v").
		SetDefault(false).
		Register(rootCmd, WithGlobal(true))
	assert.NoError(t, err)

	subInvoked, err := rootCmd.RegisterCmd(subCmd)
	assert.NoError(t, err)

	err = rootCmd.ParseOrError([]string{"--verbose", "sub", "--subarg", "test"})
	assert.Nil(t, err)
	assert.True(t, *globalFlag)
	assert.True(t, *subInvoked)
	assert.Equal(t, "test", *subArg)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Mutual exclusivity and requirements
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_MutuallyExclusiveFlags(t *testing.T) {
	fs := NewCmd("test")

	flag1, err := NewString("flag1").
		SetExcludes([]string{"flag2"}).
		SetOptional(true).
		Register(fs)
	assert.NoError(t, err)

	_, err = NewString("flag2").
		SetExcludes([]string{"flag1"}).
		SetOptional(true).
		Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--flag1", "value1", "--flag2", "value2"})
	assert.NotNil(t, err)

	// Test that using just one works
	err = fs.ParseOrError([]string{"--flag1", "value1"})
	assert.Nil(t, err)
	assert.Equal(t, "value1", *flag1)
}

func Test_ExcludesFlags_OneWay(t *testing.T) {
	fs := NewCmd("test")

	// Only flag1 declares flag2 as excluded, not the other way around
	flag1, err := NewString("flag1").
		SetExcludes([]string{"flag2"}).
		SetOptional(true).
		Register(fs)
	assert.NoError(t, err)

	flag2, err := NewString("flag2").SetOptional(true).Register(fs) // flag2 does NOT declare flag1 as excluded
	assert.NoError(t, err)

	// Should error when flag1 is used with flag2
	err = fs.ParseOrError([]string{"--flag1", "value1", "--flag2", "value2"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "excludes")

	// Should also error when flag2 is used with flag1 (one-way constraint should work both ways)
	err = fs.ParseOrError([]string{"--flag2", "value2", "--flag1", "value1"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "excludes")

	// Test that using just flag1 works
	err = fs.ParseOrError([]string{"--flag1", "value1"})
	assert.Nil(t, err)
	assert.Equal(t, "value1", *flag1)

	// Test that using just flag2 works
	err = fs.ParseOrError([]string{"--flag2", "value2"})
	assert.Nil(t, err)
	assert.Equal(t, "value2", *flag2)
}

func Test_ErrorMessages_Format(t *testing.T) {
	fs := NewCmd("test")

	// Test requires error message format
	_, err := NewString("a").SetRequires([]string{"b"}).Register(fs)
	assert.NoError(t, err)
	_, err = NewString("b").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--a", "value1"})
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid args: 'a' requires 'b', but 'b' was not set", err.Error())

	// Test excludes error message format
	fs2 := NewCmd("test2")
	_, err = NewString("file").SetExcludes([]string{"url"}).Register(fs2)
	assert.NoError(t, err)
	_, err = NewString("url").Register(fs2)
	assert.NoError(t, err)

	err = fs2.ParseOrError([]string{"--file", "test.txt", "--url", "http://example.com"})
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid args: 'file' excludes 'url', but 'url' was set", err.Error())
}

func Test_RequiresWithDefaults_Scenario(t *testing.T) {
	fs := NewCmd("test")

	// Setup: a (required), b (has default "bob" and requires c), c (required)
	_, err := NewString("a").Register(fs) // required, no default
	assert.NoError(t, err)

	_, err = NewString("b").SetDefault("bob").SetRequires([]string{"c"}).Register(fs) // has default, requires c
	assert.NoError(t, err)

	_, err = NewString("c").Register(fs) // required, no default
	assert.NoError(t, err)

	// Test: invoke with --a alice
	// What happens:
	// - 'a' is explicitly configured ✓
	// - 'b' gets default value "bob" (so b has a value)
	// - 'c' is required but not provided ✗
	//
	// Key insight:
	// - 'b' has a value (from default), and 'b' requires 'c'
	// - Since 'b' has a value, its requires constraint DOES apply
	// - 'c' is not set, so this should fail with requires constraint violation

	err = fs.ParseOrError([]string{"--a", "alice"})
	assert.NotNil(t, err, "Should fail because b (with default) requires c, but c was not provided")
	assert.Equal(t, "Invalid args: 'b' requires 'c', but 'c' was not set", err.Error())
}

func Test_RequiredFlags(t *testing.T) {
	fs := NewCmd("test")

	flag1, err := NewString("flag1").
		SetRequires([]string{"flag2"}).
		Register(fs)
	assert.NoError(t, err)

	flag2, err := NewString("flag2").Register(fs)
	assert.NoError(t, err)

	// Should error if flag1 is used without flag2
	err = fs.ParseOrError([]string{"--flag1", "value1"})
	assert.NotNil(t, err)

	// Should work if both are provided
	err = fs.ParseOrError([]string{"--flag1", "value1", "--flag2", "value2"})
	assert.Nil(t, err)
	assert.Equal(t, "value1", *flag1)
	assert.Equal(t, "value2", *flag2)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Flag-only and positional-only flags
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_FlagOnlyFlag(t *testing.T) {
	fs := NewCmd("test")

	flagOnly, err := NewString("flagonly").
		SetFlagOnly(true).
		Register(fs)
	assert.NoError(t, err)

	normalFlag, err := NewString("normal").Register(fs)
	assert.NoError(t, err)

	// Should work when used as flag
	err = fs.ParseOrError([]string{"--flagonly", "value1", "positional"})
	assert.Nil(t, err)
	assert.Equal(t, "value1", *flagOnly)
	assert.Equal(t, "positional", *normalFlag)
}

func Test_PositionalOnlyFlag(t *testing.T) {
	fs := NewCmd("test")

	posOnly, err := NewString("posonly").
		SetPositionalOnly(true).
		Register(fs)
	assert.NoError(t, err)

	normalFlag, err := NewString("normal").Register(fs)
	assert.NoError(t, err)

	// Should work when used positionally
	err = fs.ParseOrError([]string{"positional1", "--normal", "value2"})
	assert.Nil(t, err)
	assert.Equal(t, "positional1", *posOnly)
	assert.Equal(t, "value2", *normalFlag)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Examples from EXAMPLES.md
   ────────────────────────────────────────────────────────────────────────────
*/

// Example: mycmd aaa bbb
func Test_ExampleBasicPositional(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewString("arg2").Register(fs)

	err := fs.ParseOrError([]string{"aaa", "bbb"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, "bbb", *arg2)
}

// Example: mycmd aaa --arg2 bbb -c ddd -f
func Test_ExampleMixedPositionalAndFlags(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewString("arg2").Register(fs)
	arg3, _ := NewString("arg3").SetShort("c").Register(fs)
	arg4, _ := NewBool("arg4").SetShort("f").Register(fs)

	err := fs.ParseOrError([]string{"aaa", "--arg2", "bbb", "-c", "ddd", "-f"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, "bbb", *arg2)
	assert.Equal(t, "ddd", *arg3)
	assert.True(t, *arg4)
}

// Example: mycmd --arg1=aaa bbb  # assigns 'bbb' to 'arg2' because 'arg1' already assigned
func Test_ExamplePositionalAssignmentAfterFlag(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewString("arg2").Register(fs)

	err := fs.ParseOrError([]string{"--arg1=aaa", "bbb"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, "bbb", *arg2)
}

// Example: mycmd -bcd aaa  # since flags are bools, 'aaa' gets interpreted as the first positional arg
func Test_ExampleBoolClusterWithPositional(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewBool("arg2").SetShort("b").Register(fs)
	arg3, _ := NewBool("arg3").SetShort("c").Register(fs)
	arg4, _ := NewBool("arg4").SetShort("d").Register(fs)
	arg5, _ := NewBool("arg5").SetShort("e").Register(fs)

	err := fs.ParseOrError([]string{"-bcd", "aaa"})
	assert.Nil(t, err)

	assert.Equal(t, "aaa", *arg1)
	assert.True(t, *arg2)
	assert.True(t, *arg3)
	assert.True(t, *arg4)
	assert.False(t, *arg5)
}

// Example: mycmd -abc ddd  # last flag 'c' is a non-bool and so will read 'ddd'
func Test_ExampleBoolClusterEndingWithNonBool(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewBool("arg1").SetShort("a").Register(fs)
	arg2, _ := NewBool("arg2").SetShort("b").Register(fs)
	arg3, _ := NewString("arg3").SetShort("c").Register(fs)

	err := fs.ParseOrError([]string{"-abc", "ddd"})
	assert.Nil(t, err)

	assert.True(t, *arg1)
	assert.True(t, *arg2)
	assert.Equal(t, "ddd", *arg3)
}

// Example: mycmd -aaa (incrementing int shorts)
func Test_ExampleIncrementingIntShorts(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewInt("arg1").SetShort("a").SetDefault(0).Register(fs)

	err := fs.ParseOrError([]string{"-aaa"})
	assert.Nil(t, err)
	assert.Equal(t, 3, *arg1)
}

// Example: mycmd -1 --arg2 -2 -3.4 (negative numbers without number shorts mode)
func Test_ExampleNegativeNumbers(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewInt("arg1").Register(fs)
	arg2, _ := NewInt("arg2").Register(fs)
	arg3, _ := NewFloat64("arg3").Register(fs)

	err := fs.ParseOrError([]string{"-1", "--arg2", "-2", "-3.4"})
	assert.Nil(t, err)
	assert.Equal(t, -1, *arg1)
	assert.Equal(t, -2, *arg2)
	assert.Equal(t, -3.4, *arg3)
}

// Example: mycmd --arg1=-2 -2 aaa -a bbb ccc (number shorts mode)
func Test_ExampleNumberShortsMode(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewInt("arg1").Register(fs)
	arg2, _ := NewString("arg2").SetShort("2").Register(fs)
	arg3, _ := NewInt("arg3").Register(fs)
	arg4, _ := NewString("arg4").SetShort("a").Register(fs)

	err := fs.ParseOrError([]string{"--arg1=-2", "-2", "aaa", "-a", "bbb", "123"})
	assert.Nil(t, err)
	assert.Equal(t, -2, *arg1)
	assert.Equal(t, "aaa", *arg2)
	assert.Equal(t, 123, *arg3) // positional assignment
	assert.Equal(t, "bbb", *arg4)
}

// Example: mycmd aaa (positional variadic - empty)
func Test_ExamplePositionalVariadicEmpty(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)

	err := fs.ParseOrError([]string{"aaa"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, []string{}, *arg2)
}

// Example: mycmd aaa bbb (positional variadic - single item)
func Test_ExamplePositionalVariadicSingle(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)

	err := fs.ParseOrError([]string{"aaa", "bbb"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, []string{"bbb"}, *arg2)
}

// Example: mycmd aaa bbb ccc (positional variadic - multiple items)
func Test_ExamplePositionalVariadicMultiple(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)

	err := fs.ParseOrError([]string{"aaa", "bbb", "ccc"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, []string{"bbb", "ccc"}, *arg2)
}

// Example: mycmd aaa --arg2 (variadic flags - empty)
func Test_ExampleVariadicFlagEmpty(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)

	err := fs.ParseOrError([]string{"aaa", "--arg2"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, []string{}, *arg2)
}

// Example: mycmd aaa --arg2 bbb ccc (variadic flags - multiple items)
func Test_ExampleVariadicFlagMultiple(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)

	err := fs.ParseOrError([]string{"aaa", "--arg2", "bbb", "ccc"})
	assert.Nil(t, err)
	assert.Equal(t, "aaa", *arg1)
	assert.Equal(t, []string{"bbb", "ccc"}, *arg2)
}

// Example: mycmd --arg2 aaa bbb --arg1 ccc (variadic reads until next flag)
func Test_ExampleVariadicUntilNextFlag(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewString("arg1").Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)

	err := fs.ParseOrError([]string{"--arg2", "aaa", "bbb", "--arg1", "ccc"})
	assert.Nil(t, err)
	assert.Equal(t, "ccc", *arg1)
	assert.Equal(t, []string{"aaa", "bbb"}, *arg2)
}

// Example: mycmd aaa bbb --arg2 ccc ddd -e fff (multiple variadics)
func Test_ExampleMultipleVariadics(t *testing.T) {
	fs := NewCmd("mycmd")

	arg1, _ := NewStringSlice("arg1").SetVariadic(true).Register(fs)
	arg2, _ := NewStringSlice("arg2").SetVariadic(true).Register(fs)
	arg3, _ := NewBool("arg3").SetShort("e").Register(fs)
	arg4, _ := NewString("arg4").Register(fs)

	err := fs.ParseOrError([]string{"aaa", "bbb", "--arg2", "ccc", "ddd", "-e", "fff"})
	assert.Nil(t, err)
	assert.Equal(t, []string{"aaa", "bbb"}, *arg1)
	assert.Equal(t, []string{"ccc", "ddd"}, *arg2)
	assert.True(t, *arg3)
	assert.Equal(t, "fff", *arg4)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Hidden flags and help functionality
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_HiddenFlags(t *testing.T) {
	fs := NewCmd("test")

	visible, _ := NewString("visible").Register(fs)
	hidden, _ := NewString("hidden").SetHidden(true).Register(fs)

	// Both should work when parsed
	err := fs.ParseOrError([]string{"--visible", "value1", "--hidden", "value2"})
	assert.Nil(t, err)
	assert.Equal(t, "value1", *visible)
	assert.Equal(t, "value2", *hidden)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Unknown args handling
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_IgnoreUnknownArgs(t *testing.T) {
	fs := NewCmd("test")

	knownFlag, _ := NewString("known").Register(fs)

	err := fs.ParseOrError([]string{"--known", "value", "--unknown", "ignored", "positional"}, WithIgnoreUnknown(true))
	assert.Nil(t, err)
	assert.Equal(t, "value", *knownFlag)

	unknownArgs := fs.GetUnknownArgs()
	assert.Contains(t, unknownArgs, "--unknown")
	assert.Contains(t, unknownArgs, "ignored")
	assert.Contains(t, unknownArgs, "positional")
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Various slice flag options
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_IntSliceFlag(t *testing.T) {
	fs := NewCmd("test")

	intSlice, _ := NewIntSlice("values").Register(fs)

	err := fs.ParseOrError([]string{"--values", "1", "--values", "2", "--values", "3"})
	assert.Nil(t, err)
	assert.Equal(t, []int{1, 2, 3}, *intSlice)
}

func Test_Int64SliceFlag(t *testing.T) {
	fs := NewCmd("test")

	int64Slice, _ := NewInt64Slice("values").Register(fs)

	err := fs.ParseOrError([]string{"--values", "1", "--values", "2", "--values", "3"})
	assert.Nil(t, err)
	assert.Equal(t, []int64{1, 2, 3}, *int64Slice)
}

func Test_Float64SliceFlag(t *testing.T) {
	fs := NewCmd("test")

	floatSlice, _ := NewFloat64Slice("values").Register(fs)

	err := fs.ParseOrError([]string{"--values", "1.1", "--values", "2.2", "--values", "3.3"})
	assert.Nil(t, err)
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, *floatSlice)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Edge cases and error conditions
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_DuplicateFlagRegistration(t *testing.T) {
	fs := NewCmd("test")
	_, err := NewString("flag").Register(fs)
	assert.NoError(t, err)

	_, err = NewString("flag").Register(fs)
	assert.Error(t, err)
}

func Test_DuplicateShortFlagRegistration(t *testing.T) {
	fs := NewCmd("test")
	_, err := NewString("flag1").SetShort("f").Register(fs)
	assert.NoError(t, err)

	_, err = NewString("flag2").SetShort("f").Register(fs)
	assert.Error(t, err)
}

func Test_MissingRequiredFlagValue(t *testing.T) {
	fs := NewCmd("test")
	_, err := NewString("flag").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"--flag"})
	assert.NotNil(t, err)
}

func Test_InvalidFlagCluster(t *testing.T) {
	fs := NewCmd("test")
	_, err := NewString("flag1").SetShort("a").Register(fs)
	assert.NoError(t, err)
	_, err = NewBool("flag2").SetShort("b").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{"-ab"})
	assert.NotNil(t, err)
}

func Test_EmptyArgs(t *testing.T) {
	fs := NewCmd("test")
	_, err := NewString("flag").SetOptional(false).Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{})
	assert.NotNil(t, err)
}

func Test_EmptyArgsWithDefault(t *testing.T) {
	fs := NewCmd("test")
	flag, err := NewString("flag").SetDefault("default").Register(fs)
	assert.NoError(t, err)

	err = fs.ParseOrError([]string{})
	assert.Nil(t, err)
	assert.Equal(t, "default", *flag)
}

/*
   ────────────────────────────────────────────────────────────────────────────
   New Parse, Help, and Usage Tests
   ────────────────────────────────────────────────────────────────────────────
*/

// mockExitWriter is a test implementation of StderrWriter
type mockExitWriter struct {
	buffer bytes.Buffer
}

func (m *mockExitWriter) Write(p []byte) (int, error) {
	return m.buffer.Write(p)
}

// mockExit replaces the osExit function to test for calls to os.Exit.
func mockExit(t *testing.T) (func(), *int, *bytes.Buffer) {
	t.Helper()
	var exitCode int
	var mu sync.Mutex

	originalExit := osExit
	originalStderr := stderrWriter

	mockWriter := &mockExitWriter{}
	stderrWriter = mockWriter

	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
		// Use a panic to stop execution flow like os.Exit would.
		// The deferred function will recover from this.
		panic("os.Exit called")
	}

	cleanup := func() {
		osExit = originalExit
		stderrWriter = originalStderr
	}

	return cleanup, &exitCode, &mockWriter.buffer
}

func Test_ParseOrExit_ExitsOnError(t *testing.T) {
	cleanup, exitCode, stderr := mockExit(t)
	defer cleanup()

	// This will panic because we mocked os.Exit
	assert.PanicsWithValue(t, "os.Exit called", func() {
		fs := NewCmd("test")
		fs.ParseOrExit([]string{"--unknown-flag"})
	})

	assert.Equal(t, 1, *exitCode)
	assert.Contains(t, stderr.String(), "unknown flag")
	assert.Contains(t, stderr.String(), "Usage:")
}

func Test_ParseOrError_ReturnsError(t *testing.T) {
	cleanup, exitCode, _ := mockExit(t)
	defer cleanup()

	var err error
	assert.NotPanics(t, func() {
		fs := NewCmd("test")
		err = fs.ParseOrError([]string{"--unknown-flag"})
	})

	assert.NotNil(t, err)
	assert.Equal(t, 0, *exitCode) // os.Exit was not called
}

func Test_HelpFlags_Exit(t *testing.T) {
	// Test --help (long)
	cleanup, exitCode, _ := mockExit(t)
	assert.PanicsWithValue(t, "os.Exit called", func() {
		fs := NewCmd("test")
		NewString("my-flag").SetUsage("This is a test flag.").Register(fs)
		fs.ParseOrExit([]string{"--help"})
	})
	assert.Equal(t, 0, *exitCode)
	cleanup()

	// Test -h (short)
	cleanup, exitCode, _ = mockExit(t)
	assert.PanicsWithValue(t, "os.Exit called", func() {
		fs := NewCmd("test")
		NewString("my-flag").SetUsage("This is a test flag.").Register(fs)
		fs.ParseOrExit([]string{"-h"})
	})
	assert.Equal(t, 0, *exitCode)
	cleanup()
}

func Test_HiddenInLongHelp(t *testing.T) {
	cleanup, _, stderr := mockExit(t)
	defer cleanup()

	assert.Panics(t, func() {
		fs := NewCmd("test")
		NewString("visible-flag").Register(fs)
		NewString("hidden-flag").SetHiddenInLongHelp(true).Register(fs)
		fs.ParseOrExit([]string{"--help"})
	})

	output := stderr.String()
	assert.Contains(t, output, "visible-flag")
	assert.NotContains(t, output, "hidden-flag")
}

func Test_ShortHelpIsTheSameAsLongHelp(t *testing.T) {
	cleanup, _, longStderr := mockExit(t)
	assert.Panics(t, func() {
		fs := NewCmd("test")
		NewString("visible-flag").Register(fs)
		NewString("hidden-flag").SetHiddenInLongHelp(true).Register(fs)
		fs.ParseOrExit([]string{"--help"})
	})
	cleanup()

	cleanup, _, shortStderr := mockExit(t)
	assert.Panics(t, func() {
		fs := NewCmd("test")
		NewString("visible-flag").Register(fs)
		NewString("hidden-flag").SetHiddenInLongHelp(true).Register(fs)
		fs.ParseOrExit([]string{"-h"})
	})
	cleanup()

	// Per spec, the only difference is that HiddenInLongHelp flags are not shown in long help.
	// The structure should otherwise be identical.
	assert.Contains(t, longStderr.String(), "visible-flag")
	assert.NotContains(t, longStderr.String(), "hidden-flag")
	assert.Contains(t, shortStderr.String(), "visible-flag")
	assert.Contains(t, shortStderr.String(), "hidden-flag")
}

func Test_CustomUsage(t *testing.T) {
	var shortHelpCalled, longHelpCalled bool

	customUsageFunc := func(isLongHelp bool) {
		if isLongHelp {
			longHelpCalled = true
			stderrWriter.Write([]byte("Custom long help!"))
		} else {
			shortHelpCalled = true
			stderrWriter.Write([]byte("Custom short help!"))
		}
	}

	// Test long custom help
	cleanup, _, stderr := mockExit(t)
	assert.Panics(t, func() {
		fs := NewCmd("test")
		fs.SetCustomUsage(customUsageFunc)
		fs.ParseOrExit([]string{"--help"})
	})
	assert.True(t, longHelpCalled)
	assert.False(t, shortHelpCalled)
	assert.Contains(t, stderr.String(), "Custom long help!")
	cleanup()

	// Reset and test short custom help
	longHelpCalled, shortHelpCalled = false, false
	cleanup, _, stderr = mockExit(t)
	assert.Panics(t, func() {
		fs := NewCmd("test")
		fs.SetCustomUsage(customUsageFunc)
		fs.ParseOrExit([]string{"-h"})
	})
	assert.False(t, longHelpCalled)
	assert.True(t, shortHelpCalled)
	assert.Contains(t, stderr.String(), "Custom short help!")
	cleanup()
}

func Test_CustomUsageWithDefaultGenerator(t *testing.T) {
	cleanup, _, stderr := mockExit(t)
	defer cleanup()

	fs := NewCmd("test")
	NewString("my-flag").SetUsage("My flag usage.").Register(fs)
	fs.SetCustomUsage(func(isLongHelp bool) {
		// User captures fs in a closure
		if isLongHelp {
			stderrWriter.Write([]byte("--- Custom Header ---\n"))
			stderrWriter.Write([]byte(fs.GenerateLongUsage()))
			stderrWriter.Write([]byte("\n--- Custom Footer ---"))
		} else {
			stderrWriter.Write([]byte(fs.GenerateShortUsage()))
		}
	})

	assert.Panics(t, func() {
		fs.ParseOrExit([]string{"--help"})
	})

	output := stderr.String()
	assert.True(t, strings.HasPrefix(output, "--- Custom Header ---"))
	assert.True(t, strings.HasSuffix(output, "--- Custom Footer ---"))
	assert.Contains(t, output, "My flag usage.")
}

func Test_UsageStringFormat(t *testing.T) {
	cleanup, _, stderr := mockExit(t)
	defer cleanup()

	fs := NewCmd("hm")
	fs.SetDescription(
		"A rad-powered recreation of 'um', with the help of 'tldr'.\nAllows you to check the tldr for commands, but then also\nadd your own notes and customize the notes in their own\nentries.",
	)

	NewString("task").Register(fs)
	NewBool("edit").SetShort("e").Register(fs)
	NewBool("list").SetShort("l").SetUsage("Lists stored entries. Exits after.").Register(fs)
	NewBool("reconfigure").SetUsage("Enable to reconfigure hm.").Register(fs)

	// Global flags (help is added automatically)
	NewBool("debug").
		SetShort("d").
		SetUsage("Enables debug output. Intended for Rad script developers.").
		Register(fs, WithGlobal(true))
	NewString("color").
		SetUsage("Control output colorization.").
		SetEnumConstraint([]string{"auto", "always", "never"}).
		SetDefault("auto").
		Register(fs, WithGlobal(true))
	NewBool("quiet").
		SetShort("q").
		SetUsage("Suppresses some output.").
		Register(fs, WithGlobal(true))
	NewBool("confirm-shell").
		SetUsage("Confirm all shell commands before running them.").
		Register(fs, WithGlobal(true))
	NewString("src").
		SetUsage("Instead of running the target script, just print it out").
		SetHiddenInLongHelp(true).
		Register(fs, WithGlobal(true))

	assert.Panics(t, func() {
		fs.ParseOrExit([]string{"--help"})
	})

	expected := `A rad-powered recreation of 'um', with the help of 'tldr'.
Allows you to check the tldr for commands, but then also
add your own notes and customize the notes in their own
entries.

Usage:
  hm <task> [OPTIONS]

Arguments:
      --task str
  -e, --edit
  -l, --list          Lists stored entries. Exits after.
      --reconfigure   Enable to reconfigure hm.

Global options:
  -d, --debug           Enables debug output. Intended for Rad script developers.
      --color str       Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
  -h, --help            Print usage string.
`
	// Compare the full output as a string
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(stderr.String()))

	assert.NotContains(t, stderr.String(), "--src")
}

func Test_UsageStringFormatWithSubcommands(t *testing.T) {
	cleanup, _, stderr := mockExit(t)
	defer cleanup()

	rootCmd := NewCmd("git")
	rootCmd.SetDescription("A dummy git command.")

	NewString("author").Register(rootCmd, WithGlobal(true))

	addCmd := NewCmd("add")
	NewString("patch").SetShort("p").Register(addCmd)
	rootCmd.RegisterCmd(addCmd)

	commitCmd := NewCmd("commit")
	NewString("message").SetShort("m").Register(commitCmd)
	rootCmd.RegisterCmd(commitCmd)

	assert.Panics(t, func() {
		rootCmd.ParseOrExit([]string{"--help"})
	})

	expected := `
A dummy git command.

Usage:
  git [subcommand] [OPTIONS]

Commands:
  add
  commit

Global options:
      --author str
  -h, --help         Print usage string.
`
	// Compare the full output as a string
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(stderr.String()))
}

func Test_GlobalFlagAfterSubcmdRegistration(t *testing.T) {
	rootCmd := NewCmd("root")
	subCmd := NewCmd("sub")

	// Register subcommand first
	subInvoked, err := rootCmd.RegisterCmd(subCmd)
	assert.NoError(t, err)

	// Register global flag after subcommand registration
	globalFlag, err := NewBool("global").
		SetShort("g").
		Register(rootCmd, WithGlobal(true))
	assert.NoError(t, err)

	// Parse with global flag before subcommand
	err = rootCmd.ParseOrError([]string{"--global", "sub"})
	assert.NoError(t, err)
	assert.True(t, *subInvoked)
	assert.True(t, *globalFlag, "global flag should be accessible in subcommand even when registered after subcommand")
}

func Test_ConfiguredFunctionWithSubcommands(t *testing.T) {
	rootCmd := NewCmd("root")
	subCmd := NewCmd("sub")

	// Register global flag on root
	globalFlag, err := NewBool("global").
		SetShort("g").
		Register(rootCmd, WithGlobal(true))
	assert.NoError(t, err)

	// Register subcommand-specific flag
	subFlag, err := NewString("subflag").
		SetShort("s").
		Register(subCmd)
	assert.NoError(t, err)

	// Register subcommand
	subInvoked, err := rootCmd.RegisterCmd(subCmd)
	assert.NoError(t, err)

	// Parse with global flag AFTER subcommand (i.e., during subcommand parsing)
	err = rootCmd.ParseOrError([]string{"sub", "--global", "--subflag", "test"})
	assert.NoError(t, err)
	assert.True(t, *subInvoked)
	assert.True(t, *globalFlag)
	assert.Equal(t, "test", *subFlag)

	// Test Configured function - both should report the flag as configured
	// Global flag set during subcommand parsing should be visible from root
	assert.True(t, subCmd.Configured("global"), "global flag should be configured in subcommand")
	assert.True(
		t,
		rootCmd.Configured("global"),
		"global flag should be configured in root too when set during subcommand parsing",
	)

	// Subcommand-specific flag should be visible from root via recursive check
	assert.True(t, subCmd.Configured("subflag"), "subflag should be configured in subcommand")
	assert.True(t, rootCmd.Configured("subflag"), "subflag should be configured in root via recursive check")
}

func Test_ConfiguredFunctionDoesNotCheckUnusedSubcommands(t *testing.T) {
	rootCmd := NewCmd("root")
	subCmd1 := NewCmd("used")
	subCmd2 := NewCmd("unused")

	// Register flag on unused subcommand
	unusedFlag, err := NewBool("unused-flag").Register(subCmd2)
	assert.NoError(t, err)

	// Register both subcommands
	_, err = rootCmd.RegisterCmd(subCmd1)
	assert.NoError(t, err)
	_, err = rootCmd.RegisterCmd(subCmd2)
	assert.NoError(t, err)

	// Parse only the used subcommand
	err = rootCmd.ParseOrError([]string{"used"})
	assert.NoError(t, err)

	// unused-flag should not be reported as configured since subCmd2 wasn't used
	assert.False(t, rootCmd.Configured("unused-flag"), "should not report flags from unused subcommands as configured")
	assert.False(t, *unusedFlag, "unused flag should have default value")
}

func Test_GlobalFlagsAreFlagOnly(t *testing.T) {
	rootCmd := NewCmd("root")

	// Register global flag - should automatically become flag-only
	globalFlag, err := NewString("global").
		SetShort("g").
		SetOptional(true).
		Register(rootCmd, WithGlobal(true))
	assert.NoError(t, err)

	// Register regular flag - should be positional by default
	regularFlag, err := NewString("regular").
		SetShort("r").
		Register(rootCmd)
	assert.NoError(t, err)

	// Test 1: Global flag should NOT be assignable positionally
	// This should assign to regular flag, not global flag
	err = rootCmd.ParseOrError([]string{"somevalue"})
	assert.NoError(t, err)
	assert.Equal(t, "somevalue", *regularFlag, "value should go to regular flag")
	assert.Equal(t, "", *globalFlag, "global flag should remain empty when not specified")

	// Reset
	*globalFlag = ""
	*regularFlag = ""

	// Test 2: Global flag should work when used as flag
	err = rootCmd.ParseOrError([]string{"--global", "globalvalue", "regularvalue"})
	assert.NoError(t, err)
	assert.Equal(t, "globalvalue", *globalFlag, "global flag should be set via flag syntax")
	assert.Equal(t, "regularvalue", *regularFlag, "regular flag should be set via positional")

	// Test 3: Verify global flag is actually in positional list or not
	// This is the real test - check if global flag is added to positional list
	foundGlobalInPositional := false
	foundRegularInPositional := false
	for _, name := range rootCmd.positional {
		if name == "global" {
			foundGlobalInPositional = true
		}
		if name == "regular" {
			foundRegularInPositional = true
		}
	}
	assert.False(t, foundGlobalInPositional, "global flag should NOT be in positional list")
	assert.True(t, foundRegularInPositional, "regular flag should be in positional list")
}

/*
   ────────────────────────────────────────────────────────────────────────────
   Boolean flag relational constraints - only considered configured when true
   ────────────────────────────────────────────────────────────────────────────
*/

func Test_BoolFlagRequires_OnlyWhenTrue(t *testing.T) {
	fs := NewCmd("test")

	// authenticate is a bool flag that requires token
	authenticate, err := NewBool("authenticate").
		SetRequires([]string{"token"}).
		Register(fs)
	assert.NoError(t, err)

	token, err := NewString("token").Register(fs)
	assert.NoError(t, err)

	// Test 1: Only --token provided (authenticate defaults to false)
	// This should fail because authenticate is false and not considered configured for relational constraints
	err = fs.ParseOrError([]string{"--token", "mytoken"})
	assert.Nil(
		t,
		err,
		"should succeed when only token is provided since authenticate=false is not considered configured",
	)
	assert.False(t, *authenticate)
	assert.Equal(t, "mytoken", *token)
}

func Test_BoolFlagRequires_MutualRequirement(t *testing.T) {
	fs := NewCmd("test")

	// authenticate is a bool flag that requires token
	authenticate, err := NewBool("authenticate").
		SetRequires([]string{"token"}).
		Register(fs)
	assert.NoError(t, err)

	// token is a string flag that requires authenticate
	token, err := NewString("token").
		SetRequires([]string{"authenticate"}).
		Register(fs)
	assert.NoError(t, err)

	// Test 1: Only --token provided (authenticate defaults to false)
	// This should fail because token requires authenticate, but authenticate=false is not considered configured
	err = fs.ParseOrError([]string{"--token", "mytoken"})
	assert.NotNil(
		t,
		err,
		"should fail because token requires authenticate, but authenticate=false is not considered configured",
	)
	assert.Contains(t, err.Error(), "requires")
	assert.Contains(t, err.Error(), "token")
	assert.Contains(t, err.Error(), "authenticate")

	// Test 2: Both flags provided correctly should succeed
	err = fs.ParseOrError([]string{"--authenticate", "--token", "mytoken"})
	assert.Nil(t, err, "should succeed when both authenticate=true and token are provided")
	assert.True(t, *authenticate)
	assert.Equal(t, "mytoken", *token)
}

func Test_BoolFlagRequires_ExplicitlySetToFalse(t *testing.T) {
	fs := NewCmd("test")

	// authenticate is a bool flag that requires token
	authenticate, err := NewBool("authenticate").
		SetRequires([]string{"token"}).
		Register(fs)
	assert.NoError(t, err)

	token, err := NewString("token").Register(fs)
	assert.NoError(t, err)

	// Test: explicitly set authenticate to false with --authenticate=false
	// Even when explicitly set to false, bool flags should not be considered configured for relational constraints
	err = fs.ParseOrError([]string{"--authenticate=false", "--token", "mytoken"})
	assert.Nil(
		t,
		err,
		"should succeed when authenticate is explicitly set to false since false bools are not considered configured",
	)
	assert.False(t, *authenticate)
	assert.Equal(t, "mytoken", *token)
}

func Test_BoolFlagRequires_WhenTrue(t *testing.T) {
	fs := NewCmd("test")

	// authenticate is a bool flag that requires token
	authenticate, err := NewBool("authenticate").
		SetRequires([]string{"token"}).
		Register(fs)
	assert.NoError(t, err)

	token, err := NewString("token").SetOptional(true).Register(fs)
	assert.NoError(t, err)

	// Test 1: authenticate=true without token should fail
	err = fs.ParseOrError([]string{"--authenticate"})
	assert.NotNil(t, err, "should fail when authenticate=true but token is not provided")
	assert.Contains(t, err.Error(), "requires")

	// Test 2: authenticate=true with token should succeed
	err = fs.ParseOrError([]string{"--authenticate", "--token", "mytoken"})
	assert.Nil(t, err, "should succeed when both authenticate=true and token are provided")
	assert.True(t, *authenticate)
	assert.Equal(t, "mytoken", *token)
}

func Test_BoolFlagExcludes_OnlyWhenTrue(t *testing.T) {
	fs := NewCmd("test")

	// quiet is a bool flag that excludes verbose
	quiet, err := NewBool("quiet").
		SetExcludes([]string{"verbose"}).
		Register(fs)
	assert.NoError(t, err)

	verbose, err := NewBool("verbose").Register(fs)
	assert.NoError(t, err)

	// Test 1: Only --verbose provided (quiet defaults to false)
	// This should succeed because quiet=false is not considered configured for relational constraints
	err = fs.ParseOrError([]string{"--verbose"})
	assert.Nil(t, err, "should succeed when only verbose is provided since quiet=false is not considered configured")
	assert.False(t, *quiet)
	assert.True(t, *verbose)
}

func Test_BoolFlagExcludes_ExplicitlySetToFalse(t *testing.T) {
	fs := NewCmd("test")

	// quiet is a bool flag that excludes verbose
	quiet, err := NewBool("quiet").
		SetExcludes([]string{"verbose"}).
		Register(fs)
	assert.NoError(t, err)

	verbose, err := NewBool("verbose").Register(fs)
	assert.NoError(t, err)

	// Test: explicitly set quiet to false with --quiet=false
	// Even when explicitly set to false, bool flags should not be considered configured for relational constraints
	err = fs.ParseOrError([]string{"--quiet=false", "--verbose"})
	assert.Nil(
		t,
		err,
		"should succeed when quiet is explicitly set to false since false bools are not considered configured",
	)
	assert.False(t, *quiet)
	assert.True(t, *verbose)
}

func Test_BoolFlagExcludes_WhenTrue(t *testing.T) {
	// Test 1: both quiet=true and verbose=true should fail
	fs1 := NewCmd("test")
	_, err := NewBool("quiet").
		SetExcludes([]string{"verbose"}).
		Register(fs1)
	assert.NoError(t, err)
	_, err = NewBool("verbose").Register(fs1)
	assert.NoError(t, err)

	err = fs1.ParseOrError([]string{"--quiet", "--verbose"})
	assert.NotNil(t, err, "should fail when both quiet=true and verbose=true are provided")
	assert.Contains(t, err.Error(), "excludes")

	// Test 2: quiet=true alone should succeed
	fs2 := NewCmd("test")
	quiet2, err := NewBool("quiet").
		SetExcludes([]string{"verbose"}).
		Register(fs2)
	assert.NoError(t, err)
	verbose2, err := NewBool("verbose").Register(fs2)
	assert.NoError(t, err)

	err = fs2.ParseOrError([]string{"--quiet"})
	assert.Nil(t, err, "should succeed when only quiet=true is provided")
	assert.True(t, *quiet2)
	assert.False(t, *verbose2)
}

func Test_IntFlag_SetMin_Inclusive(t *testing.T) {
	fs := NewCmd("test")

	intFlag, err := NewInt("value").
		SetMin(5, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == min (should pass with inclusive=true)
	err = fs.ParseOrError([]string{"--value", "5"})
	assert.Nil(t, err)
	assert.Equal(t, 5, *intFlag)
}

func Test_IntFlag_SetMin_Exclusive(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt("value").
		SetMin(5, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == min (should fail with exclusive=false)
	err = fs.ParseOrError([]string{"--value", "5"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'value' value 5 is <= minimum (exclusive) 5")
}

func Test_IntFlag_SetMax_Inclusive(t *testing.T) {
	fs := NewCmd("test")

	intFlag, err := NewInt("value").
		SetMax(10, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == max (should pass with inclusive=true)
	err = fs.ParseOrError([]string{"--value", "10"})
	assert.Nil(t, err)
	assert.Equal(t, 10, *intFlag)
}

func Test_IntFlag_SetMax_Exclusive(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt("value").
		SetMax(10, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == max (should fail with exclusive=false)
	err = fs.ParseOrError([]string{"--value", "10"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'value' value 10 is >= maximum (exclusive) 10")
}

func Test_Int64Flag_SetMin_Inclusive(t *testing.T) {
	fs := NewCmd("test")

	intFlag, err := NewInt64("value").
		SetMin(5, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == min (should pass with inclusive=true)
	err = fs.ParseOrError([]string{"--value", "5"})
	assert.Nil(t, err)
	assert.Equal(t, int64(5), *intFlag)
}

func Test_Int64Flag_SetMin_Exclusive(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt64("value").
		SetMin(5, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == min (should fail with exclusive=false)
	err = fs.ParseOrError([]string{"--value", "5"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'value' value 5 is <= minimum (exclusive) 5")
}

func Test_Int64Flag_SetMax_Inclusive(t *testing.T) {
	fs := NewCmd("test")

	intFlag, err := NewInt64("value").
		SetMax(10, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == max (should pass with inclusive=true)
	err = fs.ParseOrError([]string{"--value", "10"})
	assert.Nil(t, err)
	assert.Equal(t, int64(10), *intFlag)
}

func Test_Int64Flag_SetMax_Exclusive(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt64("value").
		SetMax(10, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == max (should fail with exclusive=false)
	err = fs.ParseOrError([]string{"--value", "10"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'value' value 10 is >= maximum (exclusive) 10")
}

func Test_Float64Flag_SetMin_Inclusive(t *testing.T) {
	fs := NewCmd("test")

	floatFlag, err := NewFloat64("value").
		SetMin(5.0, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == min (should pass with inclusive=true)
	err = fs.ParseOrError([]string{"--value", "5.0"})
	assert.Nil(t, err)
	assert.Equal(t, 5.0, *floatFlag)
}

func Test_Float64Flag_SetMin_Exclusive(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewFloat64("value").
		SetMin(5.0, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == min (should fail with exclusive=false)
	err = fs.ParseOrError([]string{"--value", "5.0"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'value' value 5 is <= minimum (exclusive) 5")
}

func Test_Float64Flag_SetMax_Inclusive(t *testing.T) {
	fs := NewCmd("test")

	floatFlag, err := NewFloat64("value").
		SetMax(10.0, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == max (should pass with inclusive=true)
	err = fs.ParseOrError([]string{"--value", "10.0"})
	assert.Nil(t, err)
	assert.Equal(t, 10.0, *floatFlag)
}

func Test_Float64Flag_SetMax_Exclusive(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewFloat64("value").
		SetMax(10.0, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test value == max (should fail with exclusive=false)
	err = fs.ParseOrError([]string{"--value", "10.0"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'value' value 10 is >= maximum (exclusive) 10")
}

func Test_RangeError_ActualFormat(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt("age").
		SetMin(1, false). // exclusive
		Register(fs)
	assert.NoError(t, err)

	// Test the exact format requested: 'age' value 0 is <= minimum (exclusive) 1
	err = fs.ParseOrError([]string{"--age", "0"})
	assert.NotNil(t, err)
	assert.Equal(t, "'age' value 0 is <= minimum (exclusive) 1", err.Error())
}

func Test_RangeError_InclusiveFormat(t *testing.T) {
	fs := NewCmd("test")

	_, err := NewInt("score").
		SetMax(100, true). // inclusive
		Register(fs)
	assert.NoError(t, err)

	// Test inclusive format: 'score' value 101 is > maximum 100
	err = fs.ParseOrError([]string{"--score", "101"})
	assert.NotNil(t, err)
	assert.Equal(t, "'score' value 101 is > maximum 100", err.Error())
}

func Test_HelpEnabled_False_CustomHelpFlag(t *testing.T) {
	cmd := NewCmd("test")
	cmd.SetHelpEnabled(false) // Disable automatic help handling
	cmd.SetDescription("Test help disabled behavior")

	// User registers their own help flag
	help, err := NewBool("help").
		SetShort("h").
		SetUsage("Custom help handler").
		Register(cmd)
	assert.NoError(t, err)

	// Should parse without triggering usage print and exit
	err = cmd.ParseOrError([]string{"--help"})
	assert.NoError(t, err)
	assert.True(t, *help, "Custom help flag should be set to true")

	// Also test short form
	cmd2 := NewCmd("test2")
	cmd2.SetHelpEnabled(false)

	help2, err := NewBool("help").
		SetShort("h").
		SetUsage("Custom help handler").
		Register(cmd2)
	assert.NoError(t, err)

	err = cmd2.ParseOrError([]string{"-h"})
	assert.NoError(t, err)
	assert.True(t, *help2, "Custom help flag via short form should be set to true")
}

func Test_HelpEnabled_True_AutomaticHelp(t *testing.T) {
	// This test verifies that when helpEnabled is true (default),
	// help flags are automatically registered during parsing.

	cmd := NewCmd("test")
	// helpEnabled is true by default
	cmd.SetDescription("Test automatic help behavior")

	someFlag, err := NewString("flag").
		SetUsage("Some flag").
		Register(cmd)
	assert.NoError(t, err)

	// Trigger parsing to register automatic help flags
	// We'll parse empty args to avoid triggering help behavior
	err = cmd.ParseOrError([]string{})
	assert.Error(t, err) // Should error due to missing required flag

	_ = someFlag

	// The automatic help flag should now be available in the usage
	usage := cmd.GenerateUsage(false)
	assert.Contains(t, usage, "--help")
	assert.Contains(t, usage, "-h")
	assert.Contains(t, usage, "Print usage string")
}
