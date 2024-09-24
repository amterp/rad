package testing

import "testing"

func TestIfStmtTrue(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
if len(a) > 0:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIfStmtFalse(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
if len(a) > 99:
	print("not empty")
else:
	print("empty")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "empty\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIfStmtCanRefVarDefinedOutside(t *testing.T) {
	rsl := `
name = "alice"
if true:
	print(upper(name))
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n")
	assertNoErrors(t)
	resetTestState()
}

func TestIfStmtCanRefJsonVarDefinedOutside(t *testing.T) {
	rsl := `
url = "url"
name = json[].name
if true:
	id = json[].id
	rad url:
		fields id, name
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/id_name.json", "--NO-COLOR")
	assertExpected(t, stdOutBuffer, "id  name  \n1   Alice  \n2   Bob    \n")
	assertExpected(t, stdErrBuffer, "Mocking response for url (matched \".*\"): url\n")
	assertNoErrors(t)
	resetTestState()
}
