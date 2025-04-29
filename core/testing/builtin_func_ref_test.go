package testing

import "testing"

func Test_CanReferenceBuiltInFuncs(t *testing.T) {
	rsl := `
foo = upper
"test".foo().print()

["Test", "Foo"].map(lower).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `TEST
[ "test", "foo" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_CanSaveBuiltInFuncRefAndThenShadow(t *testing.T) {
	rsl := `
foo = upper
upper = "hi"
print(upper)
print(foo(upper))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ErrorsIfInvokingUndefinedSymbol(t *testing.T) {
	rsl := `
notarealsymbol()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  notarealsymbol()
  ^^^^^^^^^^^^^^ Cannot invoke unknown function: notarealsymbol
`
	assertError(t, 1, expected)
}

func Test_ErrorsIfInvokingNonFunction(t *testing.T) {
	rsl := `
foo = "hi"
foo(2)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:1

  foo(2)
  ^^^ Cannot invoke 'foo' as a function: it is a string
`
	assertError(t, 1, expected)
}
