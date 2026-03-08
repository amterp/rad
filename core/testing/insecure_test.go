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

// --- rad block with insecure keyword ---

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

func Test_RadBlock_InsecureTrue(t *testing.T) {
	script := `
id = json[].id
rad "http://example.com":
    insecure true
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

func Test_RadBlock_InsecureFalse(t *testing.T) {
	script := `
id = json[].id
rad "http://example.com":
    insecure false
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, false)
}

func Test_RadBlock_InsecureExpression(t *testing.T) {
	script := `
id = json[].id
flag = true
rad "http://example.com":
    insecure flag
    fields id
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInsecure(t, true)
}

// --- insecure with non-URL source is silently accepted ---
// The source expression could legitimately resolve to a URL on one code path
// and a list/map on another, so we don't error at runtime.

func Test_RadBlock_InsecureIgnoredOnNonUrlSource(t *testing.T) {
	script := `
data = [{"id": 1}]
id = json[].id
rad data:
    insecure
    fields id
`
	setupAndRunCode(t, script, "--color=never")
	assertNoErrors(t)
}
