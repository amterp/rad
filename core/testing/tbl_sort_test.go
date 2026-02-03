package testing

import "testing"

func Test_RequestBlock_SortDoesMutate(t *testing.T) {
	script := `
Ages = json[].age
print("Before request:", Ages)
request "http://example.com":
	fields Ages
	sort Ages
print("After request:", Ages)
`
	setupAndRunCode(t, script, "--color=never", "--mock-response", "example.com:./resources/mock_ages.json")
	expected := `Before request: [ ]
After request: [ 10, 20, 30 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \"example.com\"): http://example.com\n")
	assertHttpInvocationUrls(t, "http://example.com")
	assertNoErrors(t)
}
