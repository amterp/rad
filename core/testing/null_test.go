package testing

import "testing"

func Test_Null_Print(t *testing.T) {
	rsl := `
print(null)
null.print()
[1, 2, null, 3].print()
{"key": null}.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `null
null
[ 1, 2, null, 3 ]
{ "key": null }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Null_ErrorsIfUsedInWrongFunction(t *testing.T) {
	rsl := `
split(null, ",")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:7

  split(null, ",")
        ^^^^ Got "null" as the 1st argument of split(), but must be: string
`
	assertError(t, 1, expected)
}

func Test_Null_ParseJsonGivesNull(t *testing.T) {
	rsl := `
j = r'{ "key": null }'
o = parse_json(j)
print(o)
print(o.key)
print(type_of(o.key))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "key": null }
null
null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo null coalesce operator
