package testing

import "testing"

func Test_Fn_CustomFirst(t *testing.T) {
	script := `
fn add(x, y):
	return x + y

print(add(4, 5))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
}

func Test_Fn_TopLevelBottom(t *testing.T) {
	script := `
print(add(4, 5))

fn add(x, y):
	return x + y
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
}

func Test_Fn_CanDefineNewFunctionInBlock(t *testing.T) {
	script := `
if true:
	fn add(x, y):
		return x + y
	print(add(4, 5))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
}

func Test_Fn_CanDefineNewFunctionInBlockButOrderingMatters(t *testing.T) {
	script := `
if true:
	print(add(4, 5))
	fn add(x, y):
		return x + y
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD40003", "Cannot invoke unknown function: add")
}

func Test_Fn_CanDefineNewFunctionInBlockAndDoesNotFallOut(t *testing.T) {
	script := `
if true:
	fn add(x, y):
		return x + y
print(add(4, 5))
`
	setupAndRunCode(t, script, "--color=never")
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
}

func Test_Fn_WithTypes(t *testing.T) {
	script := `
print(add(4, 5))

fn add(x: int, y: int) -> int:
	return x + y
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
}

func Test_Fn_OneLiner(t *testing.T) {
	script := `
print(add(4, 5))

fn add(x: int, y: int) x + y
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "9\n")
	assertNoErrors(t)
}
