package testing

import "testing"

func Test_Split_OneChar(t *testing.T) {
	rsl := `
print(split("a,b,c", ","))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "a", "b", "c" ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Split_LongSplit(t *testing.T) {
	rsl := `
print(split("Alice      Smith", " "))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "Alice", "", "", "", "", "", "Smith" ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Split_LongSplitRegex(t *testing.T) {
	rsl := `
print(split("Alice      Smith", " +"))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"Alice\", \"Smith\" ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Split_CanSplitOnNoSeparater(t *testing.T) {
	rsl := `
print(split("Alice", ""))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "A", "l", "i", "c", "e" ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Split_CanSplitTags(t *testing.T) {
	rsl := `
tags = ["0.0.1", "0.2.1", "0.0.3"]
tags = [split(t, "\.") for t in tags]
print(tags)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ "0", "0", "1" ], [ "0", "2", "1" ], [ "0", "0", "3" ] ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}
