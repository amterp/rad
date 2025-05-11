package testing

import (
	"testing"
)

func Test_Args_Optional(t *testing.T) {
	rsl := `
args:
    name string
    age int
    role string?
    year int?

print(name, age, role, year, sep="|")
`
	setupAndRunCode(t, rsl, "hey", "30", "--color=never")
	expected := `hey|30|null|null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
