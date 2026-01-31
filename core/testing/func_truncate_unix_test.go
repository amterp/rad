//go:build !windows

package testing

import "testing"

func Test_Truncate_MultiByte(t *testing.T) {
	script := `
s = "hello😀world"
print(truncate(s, 7))
`
	setupAndRunCode(t, script, "--color=never")
	// 7 - 1 = 6 chars kept + "…"
	assertOnlyOutput(t, stdOutBuffer, "hello😀…\n")
	assertNoErrors(t)
}

func Test_Truncate_MultiByte_ExactBoundary(t *testing.T) {
	script := `
s = "a😀b"
print(truncate(s, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a…\n")
	assertNoErrors(t)
}

func Test_Truncate_AllEmoji(t *testing.T) {
	script := `
s = "😀😀😀😀😀"
print(truncate(s, 3))
`
	setupAndRunCode(t, script, "--color=never")
	// 3 - 1 = 2 emojis kept + "…"
	assertOnlyOutput(t, stdOutBuffer, "😀😀…\n")
	assertNoErrors(t)
}

func Test_Truncate_MinLength(t *testing.T) {
	// Minimum length is 1 (for UTF-8 ellipsis "…")
	script := `
print(truncate("hello", 1))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "…\n")
	assertNoErrors(t)
}

func Test_Truncate_ErrorsForZero(t *testing.T) {
	script := `
print(truncate("hello", 0))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `error[RAD20017]: Requires at least 1, got 0
  --> TestCase:2:7
  |
1 | 
2 | print(truncate("hello", 0))
  |       ^^^^^^^^^^^^^^^^^^^^
3 | 
  |
   = info: rad explain RAD20017

`
	assertError(t, 1, expected)
}

func Test_Truncate_ErrorsForNegative(t *testing.T) {
	script := `
print(truncate("hello", -5))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `error[RAD20017]: Requires at least 1, got -5
  --> TestCase:2:7
  |
1 | 
2 | print(truncate("hello", -5))
  |       ^^^^^^^^^^^^^^^^^^^^^
3 | 
  |
   = info: rad explain RAD20017

`
	assertError(t, 1, expected)
}
