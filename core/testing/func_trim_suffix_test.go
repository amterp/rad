package testing

import "testing"

// Test that trim_suffix removes literal suffix (not character set)
func Test_Func_Trim_Suffix_Literal(t *testing.T) {
	script := `
a = "aaabbb"
print(trim_suffix(a, "b"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "aaabb\n") // ONE "b" removed, not all
	assertNoErrors(t)
}

func Test_Func_Trim_Suffix_MultiChar(t *testing.T) {
	script := `
print(trim_suffix("hello world", " world"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Suffix_NoMatch(t *testing.T) {
	script := `
print(trim_suffix("hello", "x"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Suffix_MultipleOccurrences(t *testing.T) {
	// trim_suffix should only remove the suffix once
	script := `
print(trim_suffix("bbb", "b"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "bb\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Suffix_EmptySuffix(t *testing.T) {
	// Empty suffix should not change the string
	script := `
print(trim_suffix("hello", ""))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}
