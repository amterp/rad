package testing

import "testing"

func TestHttpPost_Basic(t *testing.T) {
	rsl := `
url = "http//www.google.com"
pprint(http_post(url))
`
	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/id_name.json", "--color=never")
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
  "duration_seconds":0,
  "status_code":200,
  "success":true
}
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): http//www.google.com\n")
	assertNoErrors(t)
}
