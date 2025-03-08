package testing

import "testing"

func TestAlgoV2_RepeatsColumnValuesToMatchNumRows(t *testing.T) {
	rsl := `
url = "https://google.com"

country = json.people.country
name = json.people.names[]
age = json.people.ages[]

rad url:
    fields name, age, country
`

	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/parallel_arrays.json", "--color=never")
	expected := `name   age  country 
Alice  25   US       
Bob    30   US       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestAlgoV2_CanExtractFromArrayWithIndex(t *testing.T) {
	rsl := `
url = "https://google.com"

Name = json.*
FirstId = json.*.ids[0]
SecondId = json.*.ids[1]

rad url:
    fields Name, FirstId, SecondId
`

	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/array_wildcard.json", "--color=never")
	expected := `Name     FirstId  SecondId 
Alice    1        2         
Bob      4        5         
Charlie  9        10        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestAlgoV2_CanChainArrayLookups(t *testing.T) {
	rsl := `
url = "https://google.com"

Nums = json[][][][1]

request url:
    fields Nums
print(Nums)
`

	setupAndRunCode(t, rsl, "--mock-response", ".*:./responses/deeply_nested_arrays.json", "--color=never")
	expected := `[ 2, 4, 6, 8 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
