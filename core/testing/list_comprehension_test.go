package testing

import "testing"

func TestListComprehensionAdding(t *testing.T) {
	rsl := `
a = [1, 2, 3]
print([x + 1 for x in a])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[2, 3, 4]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestListComprehensionUpping(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
print([upper(x) for x in a])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[A, B, C]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestListComprehensionLens(t *testing.T) {
	rsl := `
a = [[1, 2, 3], [4], [5, 6]]
print([len(x) for x in a])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[3, 1, 2]\n")
	assertNoErrors(t)
	resetTestState()
}

// todo non void list comprehensions not yet implemented
//func TestListComprehensionPrints(t *testing.T) {
//	rsl := `
//a = [1, 2, 3]
//[print(x) for x in a]
//`
//	setupAndRunCode(t, rsl)
//	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
//	assertNoErrors(t)
//	resetTestState()
//}
//
//func TestListComprehensionPrintsReturnsEmptyArray(t *testing.T) {
//	rsl := `
//a = [1, 2, 3]
//b = [print(x) for x in a]
//print(b)
//`
//	setupAndRunCode(t, rsl)
//	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n[]\n")
//	assertNoErrors(t)
//	resetTestState()
//}

func TestListComprehensionCanGetIndex(t *testing.T) {
	rsl := `
a = [10, 20, 30]
print([i * x for i, x in a])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[0, 20, 60]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestListComprehensionCanFilterNumbers(t *testing.T) {
	rsl := `
a = [5, 15, 20, 8]
print([x for x in a if x < 10])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[5, 8]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestListComprehensionCanFilterStringLengths(t *testing.T) {
	rsl := `
a = ["a", "aa", "aaa", "aaaa"]
print([x for x in a if len(x) < 3])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[a, aa]\n")
	assertNoErrors(t)
	resetTestState()
}