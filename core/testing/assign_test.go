package testing

import "testing"

func Test_Assign_InsideCollection(t *testing.T) {
	script := `a = [1, 2]
a[0] = 3
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 3, 2 ]\n")
	assertNoErrors(t)
}

func Test_MultiAssign_InsideCollectionViaSwitch(t *testing.T) {
	script := `a = [1, [2, 3], 4]
b = "alice"
a[1][0], a[2] = switch b:
	case "alice" -> 20, 30
	case "bob" -> 40, 50
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ 20, 3 ], 30 ]\n")
	assertNoErrors(t)
}

func Test_MultiAssign_InsideCollectionViaFunc(t *testing.T) {
	script := `a = [1, [2, 3], 4]
a[1][0], a[2] = pick_from_resource("./resources/people.json", "alice")
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ \"Alice\", 3 ], 25 ]\n")
	assertNoErrors(t)
}

func Test_Assign_CannotAssignJsonFieldToIndexVarPath(t *testing.T) {
	script := `a = [0]
a[0] = json.id
`

	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  a[0] = json.id
  ^^^^ Json paths must be defined to plain identifiers
`
	assertError(t, 1, expected)
}
