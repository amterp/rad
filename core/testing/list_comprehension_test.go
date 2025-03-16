package testing

import "testing"

func Test_ListComprehension_Adding(t *testing.T) {
	rsl := `
a = [1, 2, 3]
print([x + 1 for x in a])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 2, 3, 4 ]\n")
	assertNoErrors(t)
}

func Test_ListComprehension_Upping(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
print([upper(x) for x in a])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ "A", "B", "C" ]`+"\n")
	assertNoErrors(t)
}

func Test_ListComprehension_Lens(t *testing.T) {
	rsl := `
a = [[1, 2, 3], [4], [5, 6]]
print([len(x) for x in a])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 3, 1, 2 ]\n")
	assertNoErrors(t)
}

func Test_ListComprehension_Prints(t *testing.T) {
	rsl := `
a = [1, 2, 3]
[print(x) for x in a]
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
	assertNoErrors(t)
}

func Test_ListComprehension_PrintsReturnsEmptyArray(t *testing.T) {
	rsl := `
a = [1, 2, 3]
b = [print(x) for x in a]
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n[ ]\n")
	assertNoErrors(t)
}

func Test_ListComprehension_CanGetIndex(t *testing.T) {
	rsl := `
a = [10, 20, 30]
print([i * x for i, x in a])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 20, 60 ]\n")
	assertNoErrors(t)
}

func Test_ListComprehension_CanFilterNumbers(t *testing.T) {
	rsl := `
a = [5, 15, 20, 8]
print([x for x in a if x < 10])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 5, 8 ]\n")
	assertNoErrors(t)
}

func Test_ListComprehension_CanFilterStringLengths(t *testing.T) {
	rsl := `
a = ["a", "aa", "aaa", "aaaa"]
print([x for x in a if len(x) < 3])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"a\", \"aa\" ]\n")
	assertNoErrors(t)
}

func Test_ListComprehension_FunctionReturningMultipleThingsKeepsOnlyFirst(t *testing.T) {
	rsl := `
a = ["1", "2"]
print([parse_int(x) for x in a])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, 2 ]\n")
	assertNoErrors(t)
}
