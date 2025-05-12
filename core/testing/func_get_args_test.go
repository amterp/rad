package testing

import "testing"

func Test_Func_GetArgs(t *testing.T) {
	rsl := `
print(get_args())
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "--color=never" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
