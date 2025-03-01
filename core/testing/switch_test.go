package testing

import "testing"

func Test_Switch_BasicAssign(t *testing.T) {
	rsl := `
base = "https://example.com"
endpoint = "cars"
title, url = switch endpoint:
    case "cars", "automobiles": "Cars", "{base}/automobiles"
    case "books": "Books", "{base}/reading?type=books"
print(title)
print(url)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "Cars\nhttps://example.com/automobiles\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Switch_BasicAssign2(t *testing.T) {
	rsl := `
name = "alice"
result1 = switch name:
	case "alice": "ALICE"
	case "bob": "BOB"
	case "charlie": "CHARLIE"
print(result1)

name = "bob"
result2 = switch name:
	case "alice": "ALICE"
	case "bob": "BOB"
	case "charlie": "CHARLIE"
print(result2)

name = "charlie"
result3 = switch name:
	case "alice": "ALICE"
	case "bob": "BOB"
	case "charlie": "CHARLIE"
print(result3)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\nBOB\nCHARLIE\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Switch_NoAssign(t *testing.T) {
	rsl := `
name = "alice"
switch name:
	case "alice": print("ALICE"), print("ANOTHER")
	case "bob": print("BOB")
	case "charlie": print("CHARLIE")
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\nANOTHER\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Switch_NoMatchErrors(t *testing.T) {
	rsl := `
name = "david"
switch name:
	case "alice": print("ALICE")
	case "bob": print("BOB")
	case "charlie": print("CHARLIE")
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:8

  switch name:
         ^^^^ No matching case found for switch
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Switch_MultipleMatchesErrors(t *testing.T) {
	rsl := `
name = "alice"
switch name:
	case "alice": print("ALICE")
	case "bob": print("BOB")
	case "charlie", name: print("CHARLIE")
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:8

  switch name:
         ^^^^ Multiple matching cases found for switch
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Switch_AssignNumMismatchErrors(t *testing.T) {
	rsl := `
name = "charlie"
one, two = switch name:
    case "alice": 1, 2
    case "bob": 3, 4
    case "charlie": 5
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L6:5

      case "charlie": 5
      ^^^^^^^^^^^^^^^^^ Expecting 2 values, got 1
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Switch_CanSelectCaseBasedOnUsedVars(t *testing.T) {
	t.Skip("syntax later became unsupported. here in case I change my mind.")
	rsl := `
name = "alice"
age = 42
result = switch:
	case: "foo: {name}"
	case: "foo: {name}, bar: {age}"
	case: "foo: {name}, bar: {age}, baz: {notdefined}"
print(result)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "foo: alice, bar: 42\n")
	assertNoErrors(t)
	resetTestState()
}
