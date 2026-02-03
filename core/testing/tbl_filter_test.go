package testing

import "testing"

func Test_RequestBlock_FilterDoesMutate(t *testing.T) {
	script := `
Ages = json[].age
print("Before request:", Ages)
request "http://example.com":
	fields Ages
	Ages:
		filter fn(a) a >= 18
print("After request:", Ages)
`
	setupAndRunCode(t, script, "--color=never", "--mock-response", "example.com:./resources/mock_ages.json")
	expected := `Before request: [ ]
After request: [ 30, 20 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \"example.com\"): http://example.com\n")
	assertHttpInvocationUrls(t, "http://example.com")
	assertNoErrors(t)
}
