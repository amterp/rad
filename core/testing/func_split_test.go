package testing

import "testing"

func Test_Split_OneChar(t *testing.T) {
	script := `
print(split("a,b,c", ","))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "a", "b", "c" ]`+"\n")
	assertNoErrors(t)
}

func Test_Split_LongSplit(t *testing.T) {
	script := `
print(split("Alice      Smith", " "))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "Alice", "", "", "", "", "", "Smith" ]`+"\n")
	assertNoErrors(t)
}

func Test_Split_LongSplitRegex(t *testing.T) {
	script := `
print(split("Alice      Smith", " +"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"Alice\", \"Smith\" ]\n")
	assertNoErrors(t)
}

func Test_Split_CanSplitOnNoSeparater(t *testing.T) {
	script := `
print(split("Alice", ""))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "A", "l", "i", "c", "e" ]`+"\n")
	assertNoErrors(t)
}

func Test_Split_CanSplitTags(t *testing.T) {
	script := `
tags = ["0.0.1", "0.2.1", "0.0.3"]
tags = [split(t, "\.") for t in tags]
print(tags)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ "0", "0", "1" ], [ "0", "2", "1" ], [ "0", "0", "3" ] ]`+"\n")
	assertNoErrors(t)
}
