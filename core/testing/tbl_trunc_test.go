package testing

import "testing"

func TestTruncate(t *testing.T) {
	rsl := `
url = "https://google.com"
id = json[].id
words = json[].words
rad url:
	fields id, words
	words:
		truncate 10
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
	name:
		truncate 5
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
	does_not_exist:
		truncate 5
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/id_name.json", "--NO-COLOR")
	assertError(t, 1, "Mocking response for url (matched \".*\"): https://google.com\nColumn to truncate 'does_not_exist' is not a valid header\n")
	resetTestState()
}

func TestTruncateTwoFieldsAtOnce(t *testing.T) {
	rsl := `
url = "https://google.com"
age = json[].age
name = json[].name
city = json[].city
rad url:
	fields age, name, city
	name, city:
		truncate 5
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `age  name   city  
30   Char…  Paris  
40   Bob    Lond…  
30   Alice  New …  
25   Bob    Los …  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
