package testing

import "testing"

func Test_RadBlock_FilterDoesMutate(t *testing.T) {
	script := `
Ages = json[].age
print("Before rad:", Ages)
rad "http://example.com":
	noprint
	fields Ages
	Ages:
		filter fn(a) a >= 18
print("After rad:", Ages)
`
	setupAndRunCode(t, script, "--color=never", "--mock-response", "example.com:./resources/mock_ages.json")
	expected := `Before rad: [ ]
After rad: [ 30, 20 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \"example.com\"): http://example.com\n")
	assertHttpInvocationUrls(t, "http://example.com")
	assertNoErrors(t)
}
