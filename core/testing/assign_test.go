package testing

import "testing"

func Test_Assign_InsideCollection(t *testing.T) {
	rsl := `a = [1, 2]
a[0] = 3
print(a)`

	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[3, 2]\n")
	assertNoErrors(t)
	resetTestState()
}

// todo: complete this test, and ensure it passes (needs more prod code work) RAD-54
//func Test_MultiAssign_InsideCollectionViaSwitch(t *testing.T) {
//	rsl := `a = [1, [2, 3], 4]
//a[1][0], a[2] = switch 0
//print(a)`
//
//	setupAndRunCode(t, rsl)
//	assertOnlyOutput(t, stdOutBuffer, "[3, 2]\n")
//	assertNoErrors(t)
//	resetTestState()
//}

func Test_MultiAssign_InsideCollectionViaSwitch(t *testing.T) {
	rsl := `a = [1, [2, 3], 4]
a[1][0], a[2] = pick_from_resource("./resources/people.json", "alice")
print(a)`

	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, [Alice, 3], 25]\n")
	assertNoErrors(t)
	resetTestState()
}
