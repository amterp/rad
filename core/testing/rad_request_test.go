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

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./test_json/id_name.json")
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
