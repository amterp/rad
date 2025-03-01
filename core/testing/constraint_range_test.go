package testing

import "testing"

func Test_Constraint_Range_Help(t *testing.T) {
	rsl := `
args:
    age1 int # The age1.
    age2 int
    age3 float # The age3.
    age4 float

    age1 range [0, 100]
    age2 range [-20,)
    age3 range (, 200.5]
    age4 range (10, 20)
`
	setupAndRunCode(t, rsl, "--help", "--COLOR=never")
	expected := `Usage:
  test <age1> <age2> <age3> <age4>

Script args:
      --age1 int     The age1. Range: [0, 100]
      --age2 int     Range: [-20, )
      --age3 float   The age3. Range: (, 200.5]
      --age4 float   Range: (10, 20)

` + globalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Range_Basic(t *testing.T) {
	rsl := `
args:
    age int
    age range [0, 100]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "40")
	expected := `Age: 40
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Range_BasicMin(t *testing.T) {
	rsl := `
args:
    age int
    age range [0, 100]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "0")
	expected := `Age: 0
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Range_BasicMax(t *testing.T) {
	rsl := `
args:
    age int
    age range [0, 100]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "100")
	expected := `Age: 100
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Range_ExclusiveMin(t *testing.T) {
	rsl := `
args:
    age int
    age range (0, 100)
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "0")
	expected := `Error at L3:5

      age int
      ^^^^^^^ 'age' value 0 is <= minimum (exclusive) 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Constraint_Range_ExclusiveMax(t *testing.T) {
	rsl := `
args:
    age int
    age range (0, 100)
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "100")
	expected := `Error at L3:5

      age int
      ^^^^^^^ 'age' value 100 is >= maximum (exclusive) 100
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Constraint_Range_FloatBasic(t *testing.T) {
	rsl := `
args:
    age float
    age range [0.5, 100]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "0.5")
	expected := `Age: 0.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Range_FloatMinExclusive(t *testing.T) {
	rsl := `
args:
    age float
    age range (0.5, 100]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "0.5")
	expected := `Error at L3:5

      age float
      ^^^^^^^^^ 'age' value 0.5 is <= minimum (exclusive) 0.5
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Constraint_Range_NoMax(t *testing.T) {
	rsl := `
args:
    age float
    age range (0.5,]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "9999")
	expected := `Age: 9999
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Range_BelowMinWhenNoMax(t *testing.T) {
	rsl := `
args:
    age float
    age range (0.5,]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "0.1")
	expected := `Error at L3:5

      age float
      ^^^^^^^^^ 'age' value 0.1 is <= minimum (exclusive) 0.5
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Constraint_Range_AboveMaxNoMin(t *testing.T) {
	rsl := `
args:
    age int
    age range (, 200]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "250")
	expected := `Error at L3:5

      age int
      ^^^^^^^ 'age' value 250 is > maximum 200
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Constraint_Range_NoMin(t *testing.T) {
	rsl := `
args:
    age int
    age range (, 200]
print("Age:", age)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "--age", "-300")
	expected := `Age: -300
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
