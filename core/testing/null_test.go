package testing

import "testing"

func Test_Null_Print(t *testing.T) {
	script := `
print(null)
null.print()
[1, 2, null, 3].print()
{"key": null}.print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
null
[ 1, 2, null, 3 ]
{ "key": null }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Null_ErrorsIfUsedInWrongFunction(t *testing.T) {
	script := `
split(null, ",")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  split(null, ",")
        ^^^^ Got "null" as the 1st argument of split(), but must be: string
`
	assertError(t, 1, expected)
}

func Test_Null_ParseJsonGivesNull(t *testing.T) {
	script := `
j = r'{ "key": null }'
o = parse_json(j)
print(o)
print(o.key)
print(type_of(o.key))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "key": null }
null
null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Null_IsFalsy(t *testing.T) {
	script := `
a = null
if a:
	print("a is truthy")
else:
	print("a is falsy")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `a is falsy
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Null_OmittedNonDefaultArgsAreNull(t *testing.T) {
	script := `
args:
    aaa string
    bbb string

    aaa mutually excludes bbb

print(type_of(aaa), type_of(bbb))

if not bbb:
	print("aaa!")
`
	setupAndRunCode(t, script, "--aaa=hi", "--color=never")
	expected := `string null
aaa!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Null_Expressions(t *testing.T) {
	script := `
a = null
b = null
c = "not null"

print(a == b) // true
print(a == c) // false

print(a != b) // false
print(a != c) // true

print(a ? "t" : "f") // f

print(a in [1, 2, 3]) // false
print(a not in [1, 2, 3]) // true
print(a in [1, null, 3]) // true
print(a not in [1, null, 3]) // false
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
false
false
true
f
false
true
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo null coalesce operator
