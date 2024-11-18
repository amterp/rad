package testing

import "testing"

func TestMisc_SyntaxError(t *testing.T) {
	setupAndRunArgs(t, "./rsl_scripts/invalid_syntax.rad")
	assertError(t, 1, "RslError at L1/1 on '1': Expected Identifier\n")
	resetTestState()
}

func TestMisc_CanHaveVarNameThatIsJustAnUnderscore(t *testing.T) {
	rsl := `
_ = 2
print(_)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_CanHaveVarNameThatIsJustAnUnderscoreInForLoop(t *testing.T) {
	rsl := `
a = [1, 2, 3]
for _, _ in a:
	print(_)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_CanHaveNegativeNumbers(t *testing.T) {
	rsl := `
a = -10
print(a)
b = -20.2
print(b)
print("{-12}")
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "-10\n-20.2\n-12\n")
	assertNoErrors(t)
	resetTestState()
}
