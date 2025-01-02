package testing

import "testing"

func Test_Scoping_ChangingVarInIfBlockIsRemembered(t *testing.T) {
	rsl := `
a = 1
if true:
	a = 2
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_DefiningVarInIfBlockIsRemembered(t *testing.T) {
	rsl := `
if true:
	a = 1
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_ChangingVarInForBlockIsRemembered(t *testing.T) {
	rsl := `
a = 1
for i in range(3):
	a = i
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_DefiningVarInForBlockIsRemembered(t *testing.T) {
	rsl := `
for i in range(3):
	// do nothing
print(i)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_DefiningItemVarInForBlockIsRemembered(t *testing.T) {
	rsl := `
for i, item in ["a", "b", "c"]:
	// do nothing
print("i", i)
print("item", item)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "i 2\nitem c\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_DefiningKeyValueVarsInForBlockIsRemembered(t *testing.T) {
	rsl := `
for k, v in {"a": 1, "b": 2, "c": 3}:
	// do nothing
print("k", k)
print("v", v)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "k c\nv 3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_LastValueOfRadLambdaIsNotRemembered(t *testing.T) {
	rsl := `
nums = [10]
i = 0
display:
	fields nums
	nums:
		map i -> i * 2
print("i", i)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `nums 
20    
i 0
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Scoping_LastValueOfRadLambdaIsNotDefined(t *testing.T) {
	rsl := `
nums = [10]
display:
	fields nums
	nums:
		map i -> i * 2
print("i", i)
`
	expected := `nums 
20    
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "RslError at L7/12 on 'i': Undefined variable referenced: i\n")
	assertExitCode(t, 1)
	resetTestState()
}
