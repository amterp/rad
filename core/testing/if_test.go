package testing

import "testing"

func TestIf_True(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
if len(a) > 0:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
}

func TestIf_False(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
if len(a) > 99:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "empty\n")
	assertNoErrors(t)
}

func TestIf_CanRefVarDefinedOutside(t *testing.T) {
	rsl := `
name = "alice"
if true:
	print(upper(name))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n")
	assertNoErrors(t)
}

func TestIf_CanRefJsonVarDefinedOutside(t *testing.T) {
	rsl := `
url = "url"
name = json[].name
if true:
	id = json[].id
	rad url:
		fields id, name
`
	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertOutput(t, stdOutBuffer, "id  name  \n1   Alice  \n2   Bob    \n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): url\n")
	assertNoErrors(t)
}

func TestIf_Or(t *testing.T) {
	rsl := `
t = true
f = false
if t or f:
	print("TRUE")
else:
	print("FALSE")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "TRUE\n")
	assertNoErrors(t)
}

func TestIf_And(t *testing.T) {
	rsl := `
t = true
f = false
if t and f:
	print("TRUE")
else:
	print("FALSE")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "FALSE\n")
	assertNoErrors(t)
}
