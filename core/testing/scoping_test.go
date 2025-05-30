package testing

import "testing"

func Test_Scoping_ChangingVarInIfBlockIsRemembered(t *testing.T) {
	script := `
a = 1
if true:
	a = 2
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Scoping_DefiningVarInIfBlockIsRemembered(t *testing.T) {
	script := `
if true:
	a = 1
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Scoping_ChangingVarInForBlockIsRemembered(t *testing.T) {
	script := `
a = 1
for i in range(3):
	a = i
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Scoping_DefiningVarInForBlockIsRemembered(t *testing.T) {
	script := `
for i in range(3):
	// do nothing
print(i)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Scoping_DefiningItemVarInForBlockIsRemembered(t *testing.T) {
	script := `
for i, item in ["a", "b", "c"]:
	// do nothing
print("i", i)
print("item", item)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "i 2\nitem c\n")
	assertNoErrors(t)
}

func Test_Scoping_DefiningKeyValueVarsInForBlockIsRemembered(t *testing.T) {
	script := `
for k, v in {"a": 1, "b": 2, "c": 3}:
	// do nothing
print("k", k)
print("v", v)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "k c\nv 3\n")
	assertNoErrors(t)
}

func Test_Scoping_LastValueOfRadLambdaIsNotRemembered(t *testing.T) {
	script := `
nums = [10]
i = 0
display:
	fields nums
	nums:
		map fn(i) i * 2
print("i", i)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `nums 
20    
i 0
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Scoping_LastValueOfRadLambdaIsNotDefined(t *testing.T) {
	script := `
nums = [10]
display:
	fields nums
	nums:
		map fn(i) i * 2
print("i", i)
`
	expected := `nums 
20    
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	expected = `Error at L7:12

  print("i", i)
             ^ Undefined variable: i
`
	assertOutput(t, stdErrBuffer, expected)
	assertExitCode(t, 1)
}
