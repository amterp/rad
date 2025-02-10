package testing

import "testing"

func TestSplit_OneChar(t *testing.T) {
	rsl := `
print(split("a,b,c", ","))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, `[ "a", "b", "c" ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestSplit_LongSplit(t *testing.T) {
	rsl := `
print(split("Alice      Smith", " "))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, `[ "Alice", "", "", "", "", "", "Smith" ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestSplit_LongSplitRegex(t *testing.T) {
	rsl := `
print(split("Alice      Smith", " +"))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ \"Alice\", \"Smith\" ]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestSplit_CanSplitOnNoSeparater(t *testing.T) {
	rsl := `
print(split("Alice", ""))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, `[ "A", "l", "i", "c", "e" ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}
