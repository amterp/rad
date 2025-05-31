package testing

import "testing"

func Test_Func_GetArgs(t *testing.T) {
	script := `
print(get_args())
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "--color=never" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
