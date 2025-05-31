package testing

import "testing"

func TestSlice_Array_Basic(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[0:2])
print(a[1:3])
print(a[3:4])
print(a[0:len(a)])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 10, 20 ]\n[ 20, 30 ]\n[ 40 ]\n[ 10, 20, 30, 40, 50 ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_Negative(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[-5:-3])
print(a[-4:-2])
print(a[-2:-1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 10, 20 ]\n[ 20, 30 ]\n[ 40 ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_PositiveAndNegative(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[-3:3])
print(a[-2:2])
print(a[1:-1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 30 ]\n[ ]\n[ 20, 30, 40 ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_NoStartEndReturnsWholeArray(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[:])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 10, 20, 30, 40, 50 ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_OkayOutsideBounds(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[0:99])
print(a[-99:-1])
print(a[-99:99])
print(a[:99])
print(a[-99:])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 10, 20, 30, 40, 50 ]\n[ 10, 20, 30, 40 ]\n[ 10, 20, 30, 40, 50 ]\n[ 10, 20, 30, 40, 50 ]\n[ 10, 20, 30, 40, 50 ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_EmptySlices(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[0:0])
print(a[3:3])
print(a[-3:-3])
print(a[99:])
print(a[:-99])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ]\n[ ]\n[ ]\n[ ]\n[ ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_PartialSlices(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[2:])
print(a[:-2])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 30, 40, 50 ]\n[ 10, 20, 30 ]\n")
	assertNoErrors(t)
}

func TestSlice_Array_DoesNotErrorOnNonsenseSlices(t *testing.T) {
	script := `
a = [10, 20, 30, 40, 50]
print(a[3:2])
print(a[-2:-3])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ]\n[ ]\n")
	assertNoErrors(t)
}
