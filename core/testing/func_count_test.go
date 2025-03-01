package testing

import "testing"

func Test_Count_Basic(t *testing.T) {
	rsl := `
print(count("banana", "n"))
print(count("abracadabra", "a"))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n5\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Count_Overlap(t *testing.T) {
	rsl := `
print(count("aaa", "aa"))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Count_Empty(t *testing.T) {
	rsl := `
print(count("aaa", ""))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "4\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Count_EmptyStr(t *testing.T) {
	rsl := `
print(count("", "a"))
print(count("", ""))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "0\n1\n")
	assertNoErrors(t)
	resetTestState()
}
