package testing

import "testing"

func TestMockResponse(t *testing.T) {
	script := `
url = "https://google.com"

Id = json[].id
Name = json[].name

rad url:
    fields Id, Name
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	// todo notice strange trailing whitespace in table below, would be good to trim probably
	expected := `Id  Name  
1   Alice  
2   Bob    
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}
