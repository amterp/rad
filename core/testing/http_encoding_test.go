package testing

import "testing"

func Test_Requester_PathSpaces(t *testing.T) {
	script := `
url = "http://example.com/hello world"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/hello%20world")
}

func Test_Requester_QuerySpaces(t *testing.T) {
	script := `
url = "http://example.com?name=John Doe"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com?name=John%20Doe")
}

func Test_Requester_PathAndQuerySpaces(t *testing.T) {
	script := `
url = "http://example.com/my path?q=foo bar"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/my%20path?q=foo%20bar")
}

func Test_Requester_AlreadyEncodedPath(t *testing.T) {
	script := `
url = "http://example.com/hello%20world"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/hello%20world")
}

func Test_Requester_AlreadyEncodedQuery(t *testing.T) {
	script := `
url = "http://example.com?q=hello%20world"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com?q=hello%20world")
}

func Test_Requester_LiteralPlusInQuery(t *testing.T) {
	script := `
url = "http://example.com?formula=a+b"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	// Literal + in raw URL is preserved and encoded as %2B
	assertHttpInvocationUrls(t, "http://example.com?formula=a%2Bb")
}

func Test_Requester_EncodedPlusInQuery(t *testing.T) {
	script := `
url = "http://example.com?formula=a%2Bb"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com?formula=a%2Bb")
}

func Test_Requester_MixedEncoding(t *testing.T) {
	script := `
url = "http://example.com/foo bar?q=already%20encoded&x=raw space"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/foo%20bar?q=already%20encoded&x=raw%20space")
}

func Test_Requester_SpecialCharsInPath(t *testing.T) {
	script := `
url = "http://example.com/path/with spaces/and:colons/file.json"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/path/with%20spaces/and:colons/file.json")
}

func Test_Requester_MultipleQueryParams(t *testing.T) {
	script := `
url = "http://example.com?first=value one&second=value two&third=value+three"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	// Spaces encoded as %20, literal + encoded as %2B
	assertHttpInvocationUrls(t, "http://example.com?first=value%20one&second=value%20two&third=value%2Bthree")
}

func Test_Requester_URLWithFragment(t *testing.T) {
	script := `
url = "http://example.com/path?q=test#section name"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/path?q=test#section%20name")
}

func Test_Requester_EmptyQuery(t *testing.T) {
	script := `
url = "http://example.com/path?"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/path?")
}

func Test_Requester_UnicodeInPath(t *testing.T) {
	script := `
url = "http://example.com/路径/文件"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com/%E8%B7%AF%E5%BE%84/%E6%96%87%E4%BB%B6")
}

func Test_Requester_UnicodeInQuery(t *testing.T) {
	script := `
url = "http://example.com?名字=值"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com?%E5%90%8D%E5%AD%97=%E5%80%BC")
}

func Test_Requester_URLWithPort(t *testing.T) {
	script := `
url = "http://example.com:8080/path with spaces?q=test value"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com:8080/path%20with%20spaces?q=test%20value")
}

func Test_Requester_InvalidURL(t *testing.T) {
	script := `
url = "://invalid"
resp = http_get(url)
print(resp["success"])
print(resp["error"])
`
	setupAndRunCode(t, script, "--color=never")
	expected := `false
Failed to create HTTP request: invalid URL: parse "://invalid": missing protocol scheme
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Requester_NoDoubleEncoding(t *testing.T) {
	script := `
url = "http://example.com?q=hello%20world"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com?q=hello%20world")
}

func Test_Requester_InvalidPercentEncoding(t *testing.T) {
	script := `
url = "http://example.com?discount=50%"
resp = http_get(url)
print(resp["success"])
print(resp["error"])
`
	setupAndRunCode(t, script, "--color=never")
	expected := `false
Failed to create HTTP request: invalid percent-encoding in query value for key "discount": invalid URL escape "%"
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Requester_AmpersandInQueryValue(t *testing.T) {
	script := `
url = "http://example.com?text=hello&world"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	// "world" has no = so it's treated as a flag-style parameter
	assertHttpInvocationUrls(t, "http://example.com?text=hello&world")
}

func Test_Requester_EqualsInQueryValue(t *testing.T) {
	script := `
url = "http://example.com?formula=x=5"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	assertHttpInvocationUrls(t, "http://example.com?formula=x%3D5")
}

func Test_Requester_PreservesParameterOrder(t *testing.T) {
	script := `
url = "http://example.com?zebra=1&alpha=2&middle=3"
pprint(http_get(url))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertNoErrors(t)
	// Order should be preserved as written: zebra, alpha, middle (not alphabetized)
	assertHttpInvocationUrls(t, "http://example.com?zebra=1&alpha=2&middle=3")
}
