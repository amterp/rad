//go:build windows

package testing

import "testing"

// Windows uses "..." (3 chars) instead of "…" (1 char) for ellipsis,
// so minimum truncate length is 3 and outputs differ accordingly.

func Test_Truncate_MultiByte(t *testing.T) {
	script := `
s = "hello😀world"
print(truncate(s, 7))
`
	setupAndRunCode(t, script, "--color=never")
	// 7 - 3 = 4 chars kept + "..."
	assertOnlyOutput(t, stdOutBuffer, "hell...\n")
	assertNoErrors(t)
}

func Test_Truncate_MultiByte_ExactBoundary(t *testing.T) {
	// On Windows min is 3, test at that boundary with a 4-char string
	script := `
s = "a😀bc"
print(truncate(s, 3))
`
	setupAndRunCode(t, script, "--color=never")
	// 3 - 3 = 0 chars kept + "..."
	assertOnlyOutput(t, stdOutBuffer, "...\n")
	assertNoErrors(t)
}

func Test_Truncate_AllEmoji(t *testing.T) {
	script := `
s = "😀😀😀😀😀"
print(truncate(s, 3))
`
	setupAndRunCode(t, script, "--color=never")
	// 3 - 3 = 0 emojis kept + "..."
	assertOnlyOutput(t, stdOutBuffer, "...\n")
	assertNoErrors(t)
}

func Test_Truncate_MinLength(t *testing.T) {
	// Minimum length is 3 (for ASCII ellipsis "...")
	script := `
print(truncate("hello", 3))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "...\n")
	assertNoErrors(t)
}

func Test_Truncate_ErrorsForTwo(t *testing.T) {
	// 2 is below minimum of 3 on Windows
	script := `
print(truncate("hello", 2))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  print(truncate("hello", 2))
        ^^^^^^^^^^^^^^^^^^^^ Requires at least 3, got 2 (RAD20017)
`
	assertError(t, 1, expected)
}

func Test_Truncate_ErrorsForZero(t *testing.T) {
	script := `
print(truncate("hello", 0))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  print(truncate("hello", 0))
        ^^^^^^^^^^^^^^^^^^^^ Requires at least 3, got 0 (RAD20017)
`
	assertError(t, 1, expected)
}

func Test_Truncate_ErrorsForNegative(t *testing.T) {
	script := `
print(truncate("hello", -5))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  print(truncate("hello", -5))
        ^^^^^^^^^^^^^^^^^^^^^ Requires at least 3, got -5 (RAD20017)
`
	assertError(t, 1, expected)
}
