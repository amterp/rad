package testing

import "testing"

func Test_Parser_CanHaveCommentsAtTheStartAndEndOfBlocks(t *testing.T) {
	script := `
if true:
	// comment
	print("alice")
	// at the end
for i in range(2):
	// comment
	print("bob")
	// at the end
`
	setupAndRunCode(t, script, "--color=never")
	expected := `alice
bob
bob
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineListAcrossLines1(t *testing.T) {
	script := `
names = [
	"alice",
	"bob",
]
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineListAcrossLines2(t *testing.T) {
	script := `
names = ["alice",
	"bob",
]
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineListAcrossLines3(t *testing.T) {
	script := `
names = [

"alice","bob"
]
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineListAcrossLines4(t *testing.T) {
	script := `
names = ["alice","bob"
]
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineListAcrossLines5(t *testing.T) {
	script := `
names = ["alice","bob"
	]
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineListAcrossLines6(t *testing.T) {
	script := `
if true:
	names = ["alice","bob"
		]
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineMapAcrossLines1(t *testing.T) {
	script := `
names = {
	"alice": 1,
	"bob": 2,
}
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "alice": 1, "bob": 2 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineMapAcrossLines2(t *testing.T) {
	script := `
names = {"alice": 1,
	"bob"     :2,
}
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "alice": 1, "bob": 2 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Parser_CanDefineMapAcrossLines3(t *testing.T) {
	script := `
names = {"alice": 1,
	"bob"     :2, "charlie": 3}
print(names)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "alice": 1, "bob": 2, "charlie": 3 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
