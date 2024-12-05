package testing

import "testing"

const (
	truthy          = "truthy\n"
	falsy           = "falsy\n"
	truthyFalsyTest = `
if a:
	print("truthy")
else:
	print("falsy")
`
)

func TestTruthyFalsy_FalsyString(t *testing.T) {
	rsl := `a = ""` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_TruthyString(t *testing.T) {
	rsl := `a = "hi"` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_StringWithSpacesIsTruthy(t *testing.T) {
	rsl := `a = " "` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_FalsyInt(t *testing.T) {
	rsl := `a = 0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_TruthyInt(t *testing.T) {
	rsl := `a = 10` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_MinusZeroIntIsFalsy(t *testing.T) {
	rsl := `a = -0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_FalsyFloat(t *testing.T) {
	rsl := `a = 0.0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_TruthyFloat(t *testing.T) {
	rsl := `a = 10.2` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_MinusZeroFloatIsFalsy(t *testing.T) {
	rsl := `a = -0.0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_FalsyList(t *testing.T) {
	rsl := `a = []` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_TruthyList(t *testing.T) {
	rsl := `a = [1]` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_ListWith0IsTruthy(t *testing.T) {
	rsl := `a = [0]` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_FalsyMap(t *testing.T) {
	rsl := `a = {}` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_TruthyMap(t *testing.T) {
	rsl := `a = { "alice": 1 }` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_MapWithFalsyElementsIsStillTruthy(t *testing.T) {
	rsl := `a = { "": 0 }` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

// todo should writeable like below, but we don't properly allow expr stmts here
// 0 ? print("falsy") : print("truthy")
func TestTruthyFalsy_Ternary(t *testing.T) {
	rsl := `
a = 0 ? "truthy" : "falsy"
print(a)
a = 1 ? "truthy" : "falsy"
print(a)
a = "" ? "truthy" : "falsy"
print(a)
a = "hi" ? "truthy" : "falsy"
print(a)
a = 0.0 ? "truthy" : "falsy"
print(a)
a = 10.2 ? "truthy" : "falsy"
print(a)
a = [] ? "truthy" : "falsy"
print(a)
a = [1] ? "truthy" : "falsy"
print(a)
a = [0] ? "truthy" : "falsy"
print(a)
a = {} ? "truthy" : "falsy"
print(a)
a = { "alice": 1 } ? "truthy" : "falsy"
print(a)
a = { "": 0 } ? "truthy" : "falsy"
print(a)
`
	expected := `falsy
truthy
falsy
truthy
falsy
truthy
falsy
truthy
truthy
falsy
truthy
truthy
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestTruthyFalsy_ListComprehensionFilter(t *testing.T) {
	rsl := `
a = [0, 1, "", "hi", 0.0, 10.2, [], [1], [0], {}, { "alice": 1 }, { "": 0 }]
b = [x for x in a if x]
print(b)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, hi, 10.2, [1], [0], { alice: 1 }, { : 0 }]\n")
	assertNoErrors(t)
	resetTestState()
}
