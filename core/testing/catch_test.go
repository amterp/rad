package testing

import (
	"testing"
)

func Test_Catch_ErrorsIfNotCatch(t *testing.T) {
	script := `
a = catch foo()
print("Got: {a}")

fn foo():
	return catch error("this is an error")
`
	setupAndRunCode(t, script)
	expected := `Got: this is an error
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
