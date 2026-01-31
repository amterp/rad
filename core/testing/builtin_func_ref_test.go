package testing

import "testing"

func Test_CanReferenceBuiltInFuncs(t *testing.T) {
	script := `
foo = upper
"test".foo().print()

["Test", "Foo"].map(lower).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `TEST
[ "test", "foo" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_CanSaveBuiltInFuncRefAndThenShadow(t *testing.T) {
	script := `
foo = upper
upper = "hi"
print(upper)
print(foo(upper))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ErrorsIfInvokingUndefinedSymbol(t *testing.T) {
	script := `
notarealsymbol()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD40003", "Cannot invoke unknown function: notarealsymbol")
}

func Test_ErrorsIfInvokingNonFunction(t *testing.T) {
	script := `
foo = "hi"
foo(2)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30001", "Cannot invoke 'foo' as a function: it is a str")
}
