package testing

import "testing"

func Test_Func_Str_String(t *testing.T) {
	script := `
a = "hello"
print(str(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}

func Test_Func_Str_Int(t *testing.T) {
	script := `
a = 10
print(str(a)+"bob")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10bob\n")
	assertNoErrors(t)
}

func Test_Func_Str_Float(t *testing.T) {
	script := `
a = 10.2
print(str(a)+"bob")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10.2bob\n")
	assertNoErrors(t)
}

func Test_Func_Str_Bool(t *testing.T) {
	script := `
a = false
print(str(a)+"bob")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "falsebob\n")
	assertNoErrors(t)
}

func Test_Func_Str_List(t *testing.T) {
	script := `
a = [10, 20]
print(str(a)+"bob")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 10, 20 ]bob\n")
	assertNoErrors(t)
}

func Test_Func_Str_Map(t *testing.T) {
	script := `
a = { "a": 10, "b": 20 }
print(str(a)+"bob")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `{ "a": 10, "b": 20 }bob`+"\n")
	assertNoErrors(t)
}
