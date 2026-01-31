package testing

import "testing"

func TestDefer_CanDefer(t *testing.T) {
	script := `
defer print("bye")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi\nbye\n")
	assertNoErrors(t)
}

func TestDefer_ExecutesLifo(t *testing.T) {
	script := `
defer print("bye1")
defer print("bye2")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi\nbye2\nbye1\n")
	assertNoErrors(t)
}

func TestDefer_Block(t *testing.T) {
	script := `
defer:
	print("bye1")
	print("bye2")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi\nbye1\nbye2\n")
	assertNoErrors(t)
}

func TestDefer_RunsDespiteCleanExit(t *testing.T) {
	script := `
defer print("bye")
print("hi")
exit(0)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi\nbye\n")
	assertNoErrors(t)
}

func TestDefer_RunsDespiteErrorExit(t *testing.T) {
	script := `
defer print("bye")
print("hi")
exit(1)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi\nbye\n")
	assertError(t, 1, "")
}

func TestDefer_RunsDespiteError(t *testing.T) {
	script := `
defer print("bye")
print("hi")
print(asd)
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, "hi\nbye\n")
	assertErrorContains(t, 1, "RAD20028", "Undefined variable: asd")
}

func TestDefer_AllDefersRunEvenIfOneFails(t *testing.T) {
	script := `
defer print("bye1")
defer print(asd)
defer print("bye2")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, "hi\nbye2\nbye1\n")
	assertErrorContains(t, 1, "RAD20028", "Undefined variable: asd")
}

func TestDefer_UsesNonZeroCodeFromLifodDeferredExitDespiteDeferredError(t *testing.T) {
	script := `
defer print("bye1")
defer print(asd)
defer exit(3)  // this one executed before 'asd' error, so we should use its code
defer print("bye2")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, "hi\nbye2\nbye1\n")
	assertErrorContains(t, 3, "RAD20028", "Undefined variable: asd")
}

func TestDefer_UsesErrorCodeLifodDeferredErrorOverLaterNonZeroExit(t *testing.T) {
	script := `
defer print("bye1")
defer exit(3)
defer print(asd)  // this error occurs before the exit above, so we use error code 1
defer print("bye2")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, "hi\nbye2\nbye1\n")
	assertErrorContains(t, 1, "RAD20028", "Undefined variable: asd")
}

func TestDefer_IgnoresZeroCodeFromLifodDeferredExitInsteadUsesDeferredError(t *testing.T) {
	script := `
defer print("bye1")
defer print(asd)
defer exit(0)  // this is a clean exit, so we should use the error from 'asd'
defer print("bye2")
print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, "hi\nbye2\nbye1\n")
	assertErrorContains(t, 1, "RAD20028", "Undefined variable: asd")
}
