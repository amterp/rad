package testing

import "testing"

// todo maps
//  - entryset
//  - pick functions integration

func Test_Map_CanDeclare(t *testing.T) {
	script := `
a = { "alice": 35, "bob": "bar", "charlie": [1, "hi"] }
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 35, \"bob\": \"bar\", \"charlie\": [ 1, \"hi\" ] }\n")
	assertNoErrors(t)
}

func Test_Map_CanExtract(t *testing.T) {
	script := `
a = { "alice": 35, "bob": "bar","charlie": [ 1, "hi" ] }
print(a["charlie"][0] + 1)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Map_CanDeclareWithExpressions(t *testing.T) {
	script := `
foo = "bar"
t = true
f = false
a = { "alice": 30 + 5, "bob": foo, upper("charlie"): [1, t or f] }
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 35, \"bob\": \"bar\", \"CHARLIE\": [ 1, true ] }\n")
	assertNoErrors(t)
}

func Test_Map_CanAddByKey(t *testing.T) {
	script := `
a = { "alice": 35, "bob": "bar"}
a["charlie"] = 20
a[upper("dave")] = "hi"
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 35, \"bob\": \"bar\", \"charlie\": 20, \"DAVE\": \"hi\" }\n")
	assertNoErrors(t)
}

func Test_Map_CanCompoundAssign(t *testing.T) {
	script := `
a = { "alice": 100, "bob": 200, "charlie": 300, "dave": 400 }
a["alice"] += 20
a["bob"] -= 20
a["charlie"] *= 2
a["dave"] /= 2
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 120, \"bob\": 180, \"charlie\": 600, \"dave\": 200 }\n")
	assertNoErrors(t)
}

func Test_Map_CompoundOpOnNonExistentKeyErrors(t *testing.T) {
	script := `
a = { "alice": 100, "bob": 200, "charlie": 300, "dave": 400 }
a["eve"] += 20
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:3

  a["eve"] += 20
    ^^^^^ Key not found: "eve" (RAD20028)
`
	assertError(t, 1, expected)
}

func Test_Map_CanModifyArrayNestedInMap(t *testing.T) {
	script := `
a = { "alice": 100, "bob": [10, 20, 30] }
a["bob"][1] = 200
a["bob"][2] += 5
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 100, \"bob\": [ 10, 200, 35 ] }\n")
	assertNoErrors(t)
}

func Test_Map_CanInStringKeys(t *testing.T) {
	script := `
a = { }
print("one" in a)
print("one" not in a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\n")
	assertNoErrors(t)
}

func Test_Map_CanInIntKeys(t *testing.T) {
	script := `
a = { }
print(2 in a)
print(2 not in a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\n")
	assertNoErrors(t)
}

func Test_Map_CanInFloatKeys(t *testing.T) {
	script := `
a = { }
print(2.1 in a)
print(2.1 not in a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\n")
	assertNoErrors(t)
}

func Test_Map_CanInBoolKeys(t *testing.T) {
	script := `
a = { }
print(false in a)
print(false not in a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\n")
	assertNoErrors(t)
}

func Test_Map_MissingKeyBracketSyntaxErrors(t *testing.T) {
	script := `
a = { "alice": 100 }
print(a["bob"])
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:9

  print(a["bob"])
          ^^^^^ Key not found: "bob" (RAD20028)
`
	assertError(t, 1, expected)
}

func Test_Map_MissingKeyDotSyntaxErrors(t *testing.T) {
	script := `
a = { "alice": 100 }
print(a.bob)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:9

  print(a.bob)
          ^^^ Key not found: "bob" (RAD20028)
`
	assertError(t, 1, expected)
}
