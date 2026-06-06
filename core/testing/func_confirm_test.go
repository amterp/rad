package testing

import (
	"testing"
)

func Test_Func_Confirm_Yes(t *testing.T) {
	script := `print(confirm("Proceed?"))`
	// Default harness auto-confirms.
	setupAndRunCode(t, script, "--color=never")

	assertConfirmCount(t, 1)
	assertOnlyOutput(t, stdOutBuffer, "true\n")
	assertNoErrors(t)
}

func Test_Func_Confirm_No(t *testing.T) {
	script := `print(confirm("Proceed?"))`
	no := func(title, prompt string) (bool, error) { return false, nil }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(no))

	assertConfirmCount(t, 1)
	assertOnlyOutput(t, stdOutBuffer, "false\n")
	assertNoErrors(t)
}
