package testing

import "testing"

const (
	setupGenTblRsl = `
url = "https://google.com"

shortint = json[].shortint
longint = json[].longint
shortfloat = json[].shortfloat
longfloat = json[].longfloat
`
)

func TestVariousTypeLengths(t *testing.T) {
	rsl := setupGenTblRsl + `
rad url:
    fields shortint, longint, shortfloat, longfloat
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/numbers.json", "--NO-COLOR")
	expected := `shortint  longint              shortfloat  longfloat          
1         1234567899987654400  1.12        1234.5678999876543  
`
	assertExpected(t, stdOutBuffer, expected)
	assertExpected(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
