package testing

import "testing"

func Test_Parser_CanHaveCommentsAtTheStartAndEndOfBlocks(t *testing.T) {
	rsl := `
if true:
	// comment
	print("alice")
	// at the end
for i in range(2):
	// comment
	print("bob")
	// at the end
`
	setupAndRunCode(t, rsl)
	expected := `alice
bob
bob
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
