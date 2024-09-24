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

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/id_name.json", "--NO-COLOR")
	// todo notice strange trailing whitespace in table below, would be good to trim probably
	expected := `Id  Name  
1   Alice  
2   Bob    
`
	assertExpected(t, stdOutBuffer, expected)
	assertExpected(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
