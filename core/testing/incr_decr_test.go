package testing

import "testing"

func Test_Increment_Basic(t *testing.T) {
	rsl := `
a = 1
a++
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Increment_Loop(t *testing.T) {
	rsl := `
a = 1
for i in range(1000):
	a++
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "1001\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Increment_InList(t *testing.T) {
	rsl := `
a = [1, [2]]
a[1][0]++
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ 3 ] ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Increment_InMap(t *testing.T) {
	rsl := `
a = {"a": 1, "b": {"c": 2}}
a.b.c++
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, `{ "a": 1, "b": { "c": 3 } }`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Decrement_Basic(t *testing.T) {
	rsl := `
a = 10
a--
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Decrement_Loop(t *testing.T) {
	rsl := `
a = 10
for i in range(1000):
	a--
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "-990\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Decrement_InList(t *testing.T) {
	rsl := `
a = [1, [2]]
a[1][0]--
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ 1 ] ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Decrement_InMap(t *testing.T) {
	rsl := `
a = {"a": 1, "b": {"c": 2}}
a.b.c--
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, `{ "a": 1, "b": { "c": 1 } }`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Increment_CannotChain(t *testing.T) {
	rsl := `
a = 1
a++++
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L3:1

  a++++
  ^^^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Decrement_CannotChain(t *testing.T) {
	rsl := `
a = 1
a----
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L3:1

  a----
  ^^^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_IncrDecr_CannotChain(t *testing.T) {
	rsl := `
a = 1
a++--
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L3:1

  a++--
  ^^^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_DecrIncr_CannotChain(t *testing.T) {
	rsl := `
a = 1
a--++
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L3:1

  a--++
  ^^^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Increment_CanIncrementFloat(t *testing.T) {
	rsl := `
a = 1.5
a++
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "2.5\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Decrement_CanDecrementFloat(t *testing.T) {
	rsl := `
a = 1.5
a--
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "0.5\n")
	assertNoErrors(t)
	resetTestState()
}
