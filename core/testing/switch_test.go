package testing

import "testing"

func Test_Switch_BasicAssign(t *testing.T) {
	script := `
base = "https://example.com"
endpoint = "cars"
title, url = switch endpoint:
    case "cars", "automobiles" -> "Cars", "{base}/automobiles"
    case "books" -> "Books", "{base}/reading?type=books"
print(title)
print(url)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Cars\nhttps://example.com/automobiles\n")
	assertNoErrors(t)
}

func Test_Switch_BasicAssign2(t *testing.T) {
	script := `
name = "alice"
result1 = switch name:
	case "alice" -> "ALICE"
	case "bob" -> "BOB"
	case "charlie" -> "CHARLIE"
print(result1)

name = "bob"
result2 = switch name:
	case "alice" -> "ALICE"
	case "bob" -> "BOB"
	case "charlie" -> "CHARLIE"
print(result2)

name = "charlie"
result3 = switch name:
	case "alice" -> "ALICE"
	case "bob" -> "BOB"
	case "charlie" -> "CHARLIE"
print(result3)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\nBOB\nCHARLIE\n")
	assertNoErrors(t)
}

func Test_Switch_NoAssign(t *testing.T) {
	script := `
name = "alice"
switch name:
	case "alice" -> print("ALICE"), print("ANOTHER")
	case "bob" -> print("BOB")
	case "charlie" -> print("CHARLIE")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\nANOTHER\n")
	assertNoErrors(t)
}

func Test_Switch_NoMatchErrors(t *testing.T) {
	script := `
name = "david"
switch name:
	case "alice" -> print("ALICE")
	case "bob" -> print("BOB")
	case "charlie" -> print("CHARLIE")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:8

  switch name:
         ^^^^ No matching case found for switch
`
	assertError(t, 1, expected)
}

func Test_Switch_MultipleMatchesErrors(t *testing.T) {
	script := `
name = "alice"
switch name:
	case "alice" -> print("ALICE")
	case "bob" -> print("BOB")
	case "charlie", name -> print("CHARLIE")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:8

  switch name:
         ^^^^ Multiple matching cases found for switch
`
	assertError(t, 1, expected)
}

func Test_Switch_AssignNumMismatchErrors(t *testing.T) {
	script := `
name = "charlie"
one, two = switch name:
    case "alice" -> 1, 2
    case "bob" -> 3, 4
    case "charlie" -> 5
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L6:20

      case "charlie" -> 5
                     ^^^^ Cannot assign 1 values to 2 variables
`
	assertError(t, 1, expected)
}

func Test_Switch_BasicDefaultAssign(t *testing.T) {
	script := `
a, b = switch 4:
    case 1, 2 -> 10, 20
    case 3 -> 30, 40
    default -> -1, -2
print(a, b)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1 -2\n")
	assertNoErrors(t)
}

func Test_Switch_BasicBlocks(t *testing.T) {
	script := `
switch 2:
    case 1, 2:
        print(10, 20)
    case 3:
        print(30, 40)
    default:
        print(0)

switch 4:
    case 1, 2:
        print(10, 20)
    case 3:
        print(30, 40)
    default:
        print(0)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10 20\n0\n")
	assertNoErrors(t)
}

func Test_Switch_Mixed(t *testing.T) {
	script := `
a = switch 2:
    case 1, 2 -> 10
    case 3:
        print(30, 40)
        yield 30
    default -> 50
print(a)

a = switch 3:
    case 1, 2 -> 10
    case 3:
        print(30, 40)
        yield 30
    default -> 50
print(a)

a = switch 4:
    case 1, 2 -> 10
    case 3:
        print(30, 40)
        yield 30
    default -> 50
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n30 40\n30\n50\n")
	assertNoErrors(t)
}

func Test_Switch_CanYieldEvenIfNoAssign(t *testing.T) {
	script := `
switch 1:
    case 1:
        yield 10, print("hi")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi\n")
	assertNoErrors(t)
}

func Test_Switch_CanYieldJsonPaths(t *testing.T) {
	script := `
a, b = switch 1:
    case 1:
        yield json.id, json[].name
print(a, b)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ] [ ]\n")
	assertNoErrors(t)
}

func Test_Switch_DontNeedToYieldIfBreak(t *testing.T) {
	script := `
for i in range(5):
    a = switch i:
		case 0:
			yield 10
		case 1:
			break
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n")
	assertNoErrors(t)
}

func Test_Switch_DontNeedToYieldIfContinue(t *testing.T) {
	script := `
for i in range(5):
    a = switch i:
		case 0:
			yield 10
		case 1:
			continue
		default:
			yield 20
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "20\n")
	assertNoErrors(t)
}

func Test_Switch_CanSelectCaseBasedOnUsedVars(t *testing.T) {
	t.Skip("syntax later became unsupported. here in case I change my mind.")
	script := `
name = "alice"
age = 42
result = switch:
	case: "foo: {name}"
	case: "foo: {name}, bar: {age}"
	case: "foo: {name}, bar: {age}, baz: {notdefined}"
print(result)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "foo: alice, bar: 42\n")
	assertNoErrors(t)
}
