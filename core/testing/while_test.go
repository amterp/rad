package testing

import "testing"

func Test_While_BasicLoop(t *testing.T) {
	rsl := `
a = 0
while a < 3:
	print(a)
	a++
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0\n1\n2\n")
	assertNoErrors(t)
}

func Test_While_NoConditionCanBreak(t *testing.T) {
	rsl := `
a = 0
while:
	print(a)
	a++
	if a == 3:
		break
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0\n1\n2\n")
	assertNoErrors(t)
}

func Test_While_CanContinue(t *testing.T) {
	rsl := `
a = 0
while a < 3:
	a++
	if a == 2:
		continue
	print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n3\n")
	assertNoErrors(t)
}

func Test_While_CanNext(t *testing.T) {
	rsl := `
a = 0
while a < 2:
	b = 0
	while b < 2:
		print(a, b)
		b++
	a++
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 0\n0 1\n1 0\n1 1\n")
	assertNoErrors(t)
}

func Test_While_DefinedValuesNotScoped(t *testing.T) {
	rsl := `
a = 0
while a < 2:
	a++
	b = a * 20
print(a, b)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2 40\n")
	assertNoErrors(t)
}
