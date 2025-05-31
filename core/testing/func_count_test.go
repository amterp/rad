package testing

import "testing"

func Test_Count_Basic(t *testing.T) {
	script := `
print(count("banana", "n"))
print(count("abracadabra", "a"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n5\n")
	assertNoErrors(t)
}

func Test_Count_Overlap(t *testing.T) {
	script := `
print(count("aaa", "aa"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Count_Empty(t *testing.T) {
	script := `
print(count("aaa", ""))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "4\n")
	assertNoErrors(t)
}

func Test_Count_EmptyStr(t *testing.T) {
	script := `
print(count("", "a"))
print(count("", ""))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0\n1\n")
	assertNoErrors(t)
}
