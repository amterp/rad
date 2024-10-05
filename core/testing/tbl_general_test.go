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
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/numbers.json", "--NO-COLOR")
	expected := `shortint  longint              shortfloat  longfloat          
1         1234567899987654400  1.12        1234.5678999876543  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRequestBlock(t *testing.T) {
	rsl := `
url = "https://google.com"
Name = json[].name
Age = json[].age
request url:
    fields Name, Age
print("Names:", Name)
print("Ages:", Age)
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `Names: [Charlie, Bob, Alice, Bob]
Ages: [30, 40, 30, 25]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRequestBlockComplainsIfNoUrl(t *testing.T) {
	rsl := `
url = "https://google.com"
Name = json[].name
Age = json[].age
request:
    fields Name, Age
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L5/8 on ':': Expecting url or other source for request statement\n")
	resetTestState()
}

func TestDisplayBlock(t *testing.T) {
	rsl := `
Name = ["Alice", "Bob", "Charlie"]
Age = [30, 40, 25]
display:
    fields Name, Age
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Name     Age 
Alice    30   
Bob      40   
Charlie  25   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestDisplayBlockErrorsIfGivenUrl(t *testing.T) {
	rsl := `
url = "https://google.com"
Name = ["Alice", "Bob", "Charlie"]
Age = [30, 40, 25]
display url:
    fields Name, Age
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertError(t, 1, "RslError at L5/11 on 'url': Expecting ':' to immediately follow \"display\", preceding indented block\n")
	resetTestState()
}

func TestRequestThenDisplayBlocks(t *testing.T) {
	rsl := `
url = "https://google.com"
Name = json[].name
ids = json[].ids
request url:
	fields Name, ids
NumIds = [len(x) for x in ids]
display:
	fields Name, NumIds
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/arrays.json", "--NO-COLOR")
	expected := `Name     NumIds 
Alice    3       
Bob      5       
Charlie  2       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPassthroughRadBlock(t *testing.T) {
	rsl := `
url = "https://google.com"
rad url
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/text.txt", "--NO-COLOR")
	expected := `This is just some text
to emulate a non-structured
response. woo!
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
