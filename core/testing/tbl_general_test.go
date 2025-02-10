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

func TestRad_VariousTypeLengths(t *testing.T) {
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

func TestRad_RequestBlock(t *testing.T) {
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
	expected := `Names: [ "Charlie", "Bob", "Alice", "Bob" ]
Ages: [ 30, 40, 30, 25 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRad_RequestBlockComplainsIfNoUrl(t *testing.T) {
	rsl := `
url = "https://google.com"
Name = json[].name
Age = json[].age
request:
    fields Name, Age
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L5:8

  request:
          Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestRad_DisplayBlock(t *testing.T) {
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

func TestRad_DisplayBlockErrorsIfGivenUrl(t *testing.T) {
	rsl := `
url = "https://google.com"
Name = ["Alice", "Bob", "Charlie"]
Age = [30, 40, 25]
display url:
    fields Name, Age
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L5:9

  display url:
          ^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestRad_RequestThenDisplayBlocks(t *testing.T) {
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

func TestRad_RequiresBlockElseError(t *testing.T) {
	rsl := `
url = "https://google.com"
rad url
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/text.txt", "--NO-COLOR")
	expected := `Error at L3:1

  rad url
  ^^^^^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestRad_CanConditionallyApplySort(t *testing.T) {
	rsl := `
Name = ["Alice", "Bob", "Charlie"]
Age = [30, 40, 25]
should_sort = false
display:
	fields Name, Age
	if should_sort:
		sort desc
should_sort = true
display:
	fields Name, Age
	if should_sort:
		sort desc
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Name     Age 
Alice    30   
Bob      40   
Charlie  25   
Name     Age 
Charlie  25   
Bob      40   
Alice    30   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRad_CanFallBackToElse(t *testing.T) {
	rsl := `
Name = ["Alice", "Bob", "Charlie"]
Age = [30, 40, 25]
should_sort = false
display:
	fields Name, Age
	if should_sort:
		sort desc
	else:
		sort Age asc
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Name     Age 
Charlie  25   
Alice    30   
Bob      40   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRad_CanFallBackToElseIf(t *testing.T) {
	rsl := `
Name = ["Alice", "Bob", "Charlie"]
Age = [30, 40, 25]
should_sort = false
display:
	fields Name, Age
	if should_sort:
		sort desc
	else if true:
		sort Age asc
	else:
		sort Age desc
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Name     Age 
Charlie  25   
Alice    30   
Bob      40   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestRad_IfStmtWorksOnRadWithUrl(t *testing.T) {
	rsl := `
url = "https://google.com"
name = json[].name
city = json[].city
should_sort = true
rad url:
	fields name, city
	if should_sort:
		sort name asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json", "--NO-COLOR")
	expected := `name     city        
Alice    New York     
Bob      London       
Bob      Los Angeles  
Charlie  Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
