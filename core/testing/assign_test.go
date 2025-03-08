package testing

import "testing"

func Test_Assign_InsideCollection(t *testing.T) {
	rsl := `a = [1, 2]
a[0] = 3
print(a)`

	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 3, 2 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_MultiAssign_InsideCollectionViaSwitch(t *testing.T) {
	rsl := `a = [1, [2, 3], 4]
b = "alice"
a[1][0], a[2] = switch b:
	case "alice": 20, 30
	case "bob": 40, 50
print(a)`

	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ 20, 3 ], 30 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_MultiAssign_InsideCollectionViaFunc(t *testing.T) {
	rsl := `a = [1, [2, 3], 4]
a[1][0], a[2] = pick_from_resource("./resources/people.json", "alice")
print(a)`

	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ \"Alice\", 3 ], 25 ]\n")
	assertNoErrors(t)
	resetTestState()
}
