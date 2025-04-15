package testing

import "testing"

func Test_Func_GetDefault_CanGet(t *testing.T) {
	rsl := `
m = { 1: "one", "two": 2 }
v = m.get_default(1, "noo!")
print(v)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "one\n")
	assertNoErrors(t)
}

func Test_Func_GetDefault_DefaultsIfNotPresent(t *testing.T) {
	rsl := `
m = { 1: "one", "two": 2 }
v = m.get_default(2, "noo!")
print(v)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "noo!\n")
	assertNoErrors(t)
}
