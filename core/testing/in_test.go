package testing

import "testing"

func TestIn_String(t *testing.T) {
	rsl := `
a = "alice"
print("li" in a)
print("hello" in a)
print(2 in "123")
print(2 in "456")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\ntrue\nfalse\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_Array(t *testing.T) {
	rsl := `
a = [40, 50, 60]
print(50 in a)
print(70 in a)
print(true in [true, false])
print(true in [false])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\ntrue\nfalse\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_StringArray(t *testing.T) {
	rsl := `
a = ["alice", "bob", "charlie"]
print("alice" in a)
print("ALICE" in a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_NotInArray(t *testing.T) {
	rsl := `
a = [40, 50, 60]
print(50 not in a) 
print(70 not in a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_Map(t *testing.T) {
	rsl := `
a = { "alice": 40, "bob": "bar", "charlie": [1, "hi"] }
print("bob" in a)
print("dave" in a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_CanBeUsedInIfStatement(t *testing.T) {
	rsl := `
a = [40, 50, 60]

if 30 + 20 in a:
	print("50 is in a")
else:
	print("50 is not in a")

if 70 in a:	
	print("70 is in a")
else:
	print("70 is not in a")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "50 is in a\n70 is not in a\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_CanUseExpressions(t *testing.T) {
	rsl := `
a = [40, 50, 60]
b = [4, 5, 6]
print(30 + 20 in a)
print(b[0] * 10 in a)
print(b[0] * 100 in a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "true\ntrue\nfalse\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIn_CanNestIn(t *testing.T) {
	rsl := `
a = [true]
b = [false]
print(true in b in a)
print(false in b in a)
print(true in a in b)
print(false in a in b)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\nfalse\ntrue\n")
	assertNoErrors(t)
	resetTestState()
}
