package testing

import "testing"

func TestTruncate(t *testing.T) {
	rsl := `
url = "https://google.com"
id = json[].id
words = json[].words
rad url:
	fields id, words
	truncate words 10
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/long_values.json", "--NO-COLOR")
	expected := `id  words      
1   Lorem ips…  
2   Ut placer…  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestTruncateMatchesColWidth(t *testing.T) {
	rsl := `
url = "https://google.com"
id = json[].id
name = json[].name
rad url:
	fields id, name
	truncate name 5
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/id_name.json", "--NO-COLOR")
	expected := `id  name  
1   Alice  
2   Bob    
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestTruncateErrorsIfInvalidField(t *testing.T) {
	rsl := `
url = "https://google.com"
name = json[].name
rad url:
	fields name
	truncate does_not_exist 5
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/id_name.json", "--NO-COLOR")
	assertError(t, 1, "Mocking response for url (matched \".*\"): https://google.com\nColumn to truncate 'does_not_exist' is not a valid header\n")
	resetTestState()
}
