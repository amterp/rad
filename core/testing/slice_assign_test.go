package testing

import "testing"

func Test_Slice_Assign_Equivalent(t *testing.T) {
	rsl := `
a = [0, 10, 20, 30, 40, 50]
a[1:3] = [100, 200]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 100, 200, 30, 40, 50 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Slice_Assign_Bigger(t *testing.T) {
	rsl := `
a = [0, 10, 20, 30, 40, 50]
a[1:5] = [100, 200]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 100, 200, 50 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Slice_Assign_Smaller(t *testing.T) {
	rsl := `
a = [0, 10, 20, 30, 40, 50]
a[1:2] = [100, 200]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 100, 200, 20, 30, 40, 50 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Slice_Assign_NoEnd(t *testing.T) {
	rsl := `
a = [0, 10, 20, 30, 40, 50]
a[1:] = [100, 200]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 100, 200 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Slice_Assign_NoStart(t *testing.T) {
	rsl := `
a = [0, 10, 20, 30, 40, 50]
a[:4] = [100, 200]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[ 100, 200, 40, 50 ]\n")
	assertNoErrors(t)
	resetTestState()
}
