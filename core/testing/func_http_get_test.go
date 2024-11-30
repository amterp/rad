package testing

import "testing"

func TestHttpGet_Basic(t *testing.T) {
	rsl := `
url = "http//www.google.com"
pprint(http_get(url))
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/id_name.json", "--NO-COLOR")
	expected := `{
  "body": [
    {
      "id":1,
      "name":"Alice"
    },
    {
      "id":2,
      "name":"Bob"
    }
  ],
  "status_code":200
}
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): http//www.google.com\n")
	assertNoErrors(t)
	resetTestState()
}
