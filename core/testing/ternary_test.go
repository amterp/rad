package testing

import "testing"

func TestTernary_True(t *testing.T) {
	rsl := `
a = "alice"
print(true ? a : "bob")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "alice\n")
	assertNoErrors(t)
	resetTestState()
}

func TestTernary_False(t *testing.T) {
	rsl := `
a = "alice"
print(false ? a : "bob")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "bob\n")
	assertNoErrors(t)
	resetTestState()
}

func TestTernary_Nested(t *testing.T) {
	rsl := `
print(true ? (false ? "bob" : "charlie") : "alice")
print(true ? false ? "bob" : "charlie" : "alice")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "charlie\ncharlie\n")
	assertNoErrors(t)
	resetTestState()
}

func TestTernary_Complex(t *testing.T) {
	rsl := `
a = 5
b = "alice"
c = "charlie"
print((c[0] == 'c' ? c : b)[(len(b) > 3 ? 1 : 2):5])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "harl\n")
	assertNoErrors(t)
	resetTestState()
}

func TestTernary_ErrorIfConditionNotBool(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
if len(a) > 0:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}
