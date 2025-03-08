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
		map x -> truncate(x, 10)
`
	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/long_values.json", "--color=never")
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
		map x -> truncate(x, 5)
`
	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/id_name.json", "--color=never")
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
        map x -> truncate(x, 5)
`
	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	expected := `Error at L6:5

      does_not_exist:
      ^^^^^^^^^^^^^^ Cannot modify undefined field "does_not_exist"
`
	assertError(t, 1, expected)
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
		map x -> truncate(x, 5)
`
	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/people.json", "--color=never")
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
