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

func Test_TruthyFalsy_FalsyString(t *testing.T) {
	rsl := `a = ""` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_TruthyString(t *testing.T) {
	rsl := `a = "hi"` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_StringWithSpacesIsTruthy(t *testing.T) {
	rsl := `a = " "` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_FalsyInt(t *testing.T) {
	rsl := `a = 0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_TruthyInt(t *testing.T) {
	rsl := `a = 10` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_MinusZeroIntIsFalsy(t *testing.T) {
	rsl := `a = -0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_FalsyFloat(t *testing.T) {
	rsl := `a = 0.0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_TruthyFloat(t *testing.T) {
	rsl := `a = 10.2` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_MinusZeroFloatIsFalsy(t *testing.T) {
	rsl := `a = -0.0` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_FalsyList(t *testing.T) {
	rsl := `a = []` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_TruthyList(t *testing.T) {
	rsl := `a = [1]` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_ListWith0IsTruthy(t *testing.T) {
	rsl := `a = [0]` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_FalsyMap(t *testing.T) {
	rsl := `a = {}` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, falsy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_TruthyMap(t *testing.T) {
	rsl := `a = { "alice": 1 }` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_MapWithFalsyElementsIsStillTruthy(t *testing.T) {
	rsl := `a = { "": 0 }` + truthyFalsyTest
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, truthy)
	assertNoErrors(t)
	resetTestState()
}

// todo should writeable like below, but we don't properly allow expr stmts here
//   - 0 ? print("falsy") : print("truthy")
func Test_TruthyFalsy_Ternary(t *testing.T) {
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

func Test_TruthyFalsy_ListComprehensionFilter(t *testing.T) {
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

func Test_TruthyFalsy_NotTruthy(t *testing.T) {
	rsl := `
a = []
if not a:
	print("first")
else:
	print("second")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "first\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Or_Int(t *testing.T) {
	rsl := `
print(0 or 0)
print(1 or 0)
print(0 or 1)
print(1 or 1)
`
	setupAndRunCode(t, rsl)
	expected := `false
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Or_Float(t *testing.T) {
	rsl := `
print(0.0 or 0.0)
print(1.0 or 0.0)
print(0.0 or 1.0)
print(1.0 or 1.0)
`
	setupAndRunCode(t, rsl)
	expected := `false
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Or_String(t *testing.T) {
	rsl := `
print("" or "")
print("hi" or "")
print("" or "hi")
print("hi" or "hi")
`
	setupAndRunCode(t, rsl)
	expected := `false
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Or_List(t *testing.T) {
	rsl := `
print([] or [])
print([0] or [])
print([] or [0])
print([0] or [0])
`
	setupAndRunCode(t, rsl)
	expected := `false
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Or_Map(t *testing.T) {
	rsl := `
print({} or {})
print({ "alice" : 1 } or {})
print({} or { "alice" : 1 })
print({ "alice" : 1 } or { "alice" : 1 })
`
	setupAndRunCode(t, rsl)
	expected := `false
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_And_Int(t *testing.T) {
	rsl := `
print(0 and 0)
print(1 and 0)
print(0 and 1)
print(1 and 1)
`
	setupAndRunCode(t, rsl)
	expected := `false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_And_Float(t *testing.T) {
	rsl := `
print(0.0 and 0.0)
print(1.0 and 0.0)
print(0.0 and 1.0)
print(1.0 and 1.0)
`
	setupAndRunCode(t, rsl)
	expected := `false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_And_String(t *testing.T) {
	rsl := `
print("" and "")
print("hi" and "")
print("" and "hi")
print("hi" and "hi")
`
	setupAndRunCode(t, rsl)
	expected := `false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_And_List(t *testing.T) {
	rsl := `
print([] and [])
print([0] and [])
print([] and [0])
print([0] and [0])
`
	setupAndRunCode(t, rsl)
	expected := `false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_And_Map(t *testing.T) {
	rsl := `
print({} and {})
print({ "alice" : 1 } and {})
print({} and { "alice" : 1 })
print({ "alice" : 1 } and { "alice" : 1 })
`
	setupAndRunCode(t, rsl)
	expected := `false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Not_Int(t *testing.T) {
	rsl := `
print(not 0)
print(not 1)
`
	setupAndRunCode(t, rsl)
	expected := `true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Not_Float(t *testing.T) {
	rsl := `
print(not 0.0)
print(not 1.0)
`
	setupAndRunCode(t, rsl)
	expected := `true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Not_String(t *testing.T) {
	rsl := `
print(not "")
print(not "hi")
`
	setupAndRunCode(t, rsl)
	expected := `true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Not_List(t *testing.T) {
	rsl := `
print(not [])
print(not [0])
`
	setupAndRunCode(t, rsl)
	expected := `true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_TruthyFalsy_Not_Map(t *testing.T) {
	rsl := `
print(not {})
print(not { "alice" : 1 })
`
	setupAndRunCode(t, rsl)
	expected := `true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
