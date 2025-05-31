package testing

import "testing"

func TestSlice_String_Basic(t *testing.T) {
	script := `
a = "alice"
print(a[0:2])
print(a[1:3])
print(a[3:4])
print(a[0:len(a)])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "al\nli\nc\nalice\n")
	assertNoErrors(t)
}

func TestSlice_String_Negative(t *testing.T) {
	script := `
a = "alice"
print(a[-5:-3])
print(a[-4:-2])
print(a[-2:-1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "al\nli\nc\n")
	assertNoErrors(t)
}

func TestSlice_String_PositiveAndNegative(t *testing.T) {
	script := `
a = "alice"
print(a[-3:3])
print(a[-2:2])
print(a[1:-1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "i\n\nlic\n")
	assertNoErrors(t)
}

func TestSlice_String_NoStartEndReturnsWholeString(t *testing.T) {
	script := `
a = "alice"
print(a[:])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "alice\n")
	assertNoErrors(t)
}

func TestSlice_String_OkayOutsideBounds(t *testing.T) {
	script := `
a = "alice"
print(a[0:99])
print(a[-99:-1])
print(a[-99:99])
print(a[:99])
print(a[-99:])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "alice\nalic\nalice\nalice\nalice\n")
	assertNoErrors(t)
}

func TestSlice_String_EmptySlices(t *testing.T) {
	script := `
a = "alice"
print(a[0:0])
print(a[3:3])
print(a[-3:-3])
print(a[99:])
print(a[:-99])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n\n\n\n\n")
	assertNoErrors(t)
}

func TestSlice_String_PartialSlices(t *testing.T) {
	script := `
a = "alice"
print(a[2:])
print(a[:-2])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ice\nali\n")
	assertNoErrors(t)
}

func TestSlice_String_DoesNotErrorOnNonsenseSlices(t *testing.T) {
	script := `
a = "alice"
print(a[3:2])
print(a[-2:-3])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n\n")
	assertNoErrors(t)
}
