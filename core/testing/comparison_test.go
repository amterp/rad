package testing

import "testing"

func Test_Comparison_Equality_Int(t *testing.T) {
	rsl := `
print(1 == 1)
print(1 == 2)
print(2 == 1)
print(2 == 2)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_Equality_String(t *testing.T) {
	rsl := `
print("a" == "a")
print("a" == "b")
print("b" == "a")
print("b" == "b")
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_Equality_Float(t *testing.T) {
	rsl := `
print(1.0 == 1.0)
print(1.0 == 2.0)
print(2.0 == 1.0)
print(2.0 == 2.0)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

// todo RAD-92
func Test_Comparison_Equality_List(t *testing.T) {
	t.Skip("TODO RAD-92")
	rsl := `
print([1, 2, 3] == [1, 2, 3])
print([1, 2, 3] == [1, 2, 4])
print([1, 2, 3] == [1, 2])
print([1, 2, 3] == [1, 2, 3, [4])
print([1, 2, 3, [4]] == [1, 2, 3, [4])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

// todo RAD-92
func Test_Comparison_Equality_Map(t *testing.T) {
	t.Skip("TODO RAD-92")
	rsl := `
print({"a": 1, "b": 2} == {"a": 1, "b": 2})
print({"a": 1, "b": 2} == {"a": 1, "b": 3})
print({"a": 1, "b": 2} == {"a": 1})
print({"a": 1, "b": 2} == {"a": 1, "b": 2, "c": {"d": 3}})
print({"a": 1, "b": 2, {"d": 3}} == {"a": 1, "b": 2, "c": {"d": 3}})
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
false
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_Equality_Mixed(t *testing.T) {
	rsl := `
print(1 == 1.0)
print(1.0 == 1)
print(1 == "1")
print("1" == 1)
print(1.0 == "1")
print("1" == 1.0)
print(1 == [1])
print([1] == 1)
print(1 == {"a": 1})
print({"a": 1} == 1)
print(1 == true)
print(1 == false)
print(true == 1)
print(false == 1)
print(0 == true)
print(0 == false)
print(true == 0)
print(false == 0)
print(true == 2)
print(2 == true)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
true
false
false
false
false
false
false
false
false
false
false
false
false
false
false
false
false
false
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_NotEquality_Mixed(t *testing.T) {
	rsl := `
print(1 != 1.0)
print(1.0 != 1)
print(1 != "1")
print("1" != 1)
print(1.0 != "1")
print("1" != 1.0)
print(1 != [1])
print([1] != 1)
print(1 != {"a": 1})
print({"a": 1} != 1)
print(1 != true)
print(1 != false)
print(true != 1)
print(false != 1)
print(0 != true)
print(0 != false)
print(true != 0)
print(false != 0)
print(true != 2)
print(2 != true)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `false
false
true
true
true
true
true
true
true
true
true
true
true
true
true
true
true
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_GtLt_Int(t *testing.T) {
	rsl := `
print(1 > 1)
print(1 > 2)
print(2 > 1)
print(1 >= 1)
print(1 >= 2)
print(2 >= 1)

print(1 < 1)
print(1 < 2)
print(2 < 1)
print(1 <= 1)
print(1 <= 2)
print(2 <= 1)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `false
false
true
true
false
true
false
true
false
true
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_GtLt_Float(t *testing.T) {
	rsl := `
print(1.1 > 1.1)
print(1.1 > 2.1)
print(2.1 > 1.1)
print(1.1 >= 1.1)
print(1.1 >= 2.1)
print(2.1 >= 1.1)

print(1.1 < 1.1)
print(1.1 < 2.1)
print(2.1 < 1.1)
print(1.1 <= 1.1)
print(1.1 <= 2.1)
print(2.1 <= 1.1)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `false
false
true
true
false
true
false
true
false
true
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Comparison_GtLt_Mixed(t *testing.T) {
	rsl := `
print(1.1 > 1)
print(1.1 > 2)
print(2.1 > 1)
print(1.1 >= 1)
print(1.1 >= 2)
print(2.1 >= 1)

print(1.1 < 1)
print(1.1 < 2)
print(2.1 < 1)
print(1.1 <= 1)
print(1.1 <= 2)
print(2.1 <= 1)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `true
false
true
true
false
true
false
true
false
false
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
