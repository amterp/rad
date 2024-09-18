package testing

import "testing"

func TestMockResponse(t *testing.T) {
	rsl := `
url = "https://google.com"

Id = json[].id
Name = json[].name

rad url:
    fields Id, Name
`

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./test_json/mock.json")
	// todo notice strange trailing whitespace in table below, would be good to trim probably
	expected := `Mocking response for url (matched ".*"): https://google.com
ID  NAME  
1   Alice  
2   Bob    
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestJsonNonRootArrayExtraction(t *testing.T) {
	rsl := `
url = "https://google.com"

Id = json.id
Names = json.names

rad url:
    fields Id, Names
`

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./test_json/not_root_array.json")
	expected := `Mocking response for url (matched ".*"): https://google.com
ID  NAMES               
1   [Alice Bob Charlie]  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
