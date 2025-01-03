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

func Test_Parser_CanDefineListAcrossLines1(t *testing.T) {
	rsl := `
names = [
	"alice",
	"bob",
]
print(names)
`
	setupAndRunCode(t, rsl)
	expected := `[alice, bob]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Parser_CanDefineListAcrossLines2(t *testing.T) {
	rsl := `
names = ["alice",
	"bob",
]
print(names)
`
	setupAndRunCode(t, rsl)
	expected := `[alice, bob]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Parser_CanDefineListAcrossLines3(t *testing.T) {
	rsl := `
names = [

"alice","bob"
]
print(names)
`
	setupAndRunCode(t, rsl)
	expected := `[alice, bob]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Parser_CanDefineListAcrossLines4(t *testing.T) {
	rsl := `
names = ["alice","bob"
]
print(names)
`
	setupAndRunCode(t, rsl)
	expected := `[alice, bob]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
