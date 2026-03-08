package testing

import "testing"

// --- http_* functions with insecure named arg ---

func Test_HttpGet_Insecure(t *testing.T) {
	script := `
url = "http://example.com/api"
pprint(http_get(url, insecure=true))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

func Test_HttpGet_InsecureFalse(t *testing.T) {
	script := `
url = "http://example.com/api"
pprint(http_get(url, insecure=false))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, false)
}

func Test_HttpGet_InsecureDefault(t *testing.T) {
	script := `
url = "http://example.com/api"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, false)
}

func Test_HttpPost_Insecure(t *testing.T) {
	script := `
url = "http://example.com/api"
pprint(http_post(url, insecure=true))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

// --- --tls-insecure CLI flag ---

func Test_TlsInsecure_Flag(t *testing.T) {
	script := `
url = "http://example.com/api"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--tls-insecure", "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	// The CLI flag sets insecure globally on the Requester, but the RequestDef.Insecure
	// field reflects per-request intent. The actual TLS bypass happens at the client level
	// via Requester.insecure, not via RequestDef.Insecure.
	assertHttpInsecure(t, false)
	assertRequesterInsecure(t, true)
}

// --- request block with insecure keyword ---

func Test_RequestBlock_Insecure(t *testing.T) {
	script := `
id = json[].id
request "http://example.com":
    insecure
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

func Test_RequestBlock_InsecureTrue(t *testing.T) {
	script := `
id = json[].id
request "http://example.com":
    insecure true
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

func Test_RequestBlock_InsecureFalse(t *testing.T) {
	script := `
id = json[].id
request "http://example.com":
    insecure false
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, false)
}

func Test_RequestBlock_InsecureExpression(t *testing.T) {
	script := `
id = json[].id
flag = true
request "http://example.com":
    insecure flag
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

// --- rad block with insecure ---

func Test_RadBlock_Insecure(t *testing.T) {
	script := `
id = json[].id
rad "http://example.com":
    insecure
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

// --- display block with insecure should error ---

func Test_DisplayBlock_InsecureError(t *testing.T) {
	script := `
data = [{"id": 1}]
id = json[].id
display data:
    insecure
    fields id
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "'insecure' is not valid in display blocks")
}
