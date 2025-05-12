package testing

import "testing"

func Test_Func_GetArgs(t *testing.T) {
	rsl := `
myargs = get_args()
myargs[0] = myargs[0].split("/")[-1]
print(myargs)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "___Test_Func_GetArgs_in_rad_core_testing.test", "--color=never" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
