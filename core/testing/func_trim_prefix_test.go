package testing

import "testing"

// Test that trim_prefix removes literal prefix (not character set)
func Test_Func_Trim_Prefix_Literal(t *testing.T) {
	script := `
a = "aaabbb"
print(trim_prefix(a, "a"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "aabbb\n") // ONE "a" removed, not all
	assertNoErrors(t)
}

func Test_Func_Trim_Prefix_MultiChar(t *testing.T) {
	script := `
print(trim_prefix("hello world", "hello "))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "world\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Prefix_NoMatch(t *testing.T) {
	script := `
print(trim_prefix("hello", "x"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Prefix_MultipleOccurrences(t *testing.T) {
	// trim_prefix should only remove the prefix once
	script := `
print(trim_prefix("aaa", "a"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "aa\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Prefix_EmptyPrefix(t *testing.T) {
	// Empty prefix should not change the string
	script := `
print(trim_prefix("hello", ""))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}
