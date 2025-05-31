package testing

import "testing"

func TestIf_True(t *testing.T) {
	script := `
a = ["a", "b", "c"]
if len(a) > 0:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
}

func TestIf_False(t *testing.T) {
	script := `
a = ["a", "b", "c"]
if len(a) > 99:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "empty\n")
	assertNoErrors(t)
}

func TestIf_CanRefVarDefinedOutside(t *testing.T) {
	script := `
name = "alice"
if true:
	print(upper(name))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n")
	assertNoErrors(t)
}

func TestIf_CanRefJsonVarDefinedOutside(t *testing.T) {
	script := `
url = "url"
name = json[].name
if true:
	id = json[].id
	rad url:
		fields id, name
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertOutput(t, stdOutBuffer, "id  name  \n1   Alice  \n2   Bob    \n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): url\n")
	assertNoErrors(t)
}

func TestIf_Or(t *testing.T) {
	script := `
t = true
f = false
if t or f:
	print("TRUE")
else:
	print("FALSE")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "TRUE\n")
	assertNoErrors(t)
}

func TestIf_And(t *testing.T) {
	script := `
t = true
f = false
if t and f:
	print("TRUE")
else:
	print("FALSE")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "FALSE\n")
	assertNoErrors(t)
}
