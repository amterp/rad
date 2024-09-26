package testing

import "testing"

const (
	setupSortingRsl = `
url = "https://google.com"

name = json[].name
age = json[].age
city = json[].city
`
)

func TestNoSorting(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Charlie  30   Paris        
Bob      40   London       
Alice    30   New York     
Bob      25   Los Angeles  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestGeneralAscNoToken(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestGeneralAscWithToken(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestGeneralDesc(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort desc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Charlie  30   Paris        
Bob      40   London       
Bob      25   Los Angeles  
Alice    30   New York     
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestExplicitAsc(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort name asc, age asc, city asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      25   Los Angeles  
Bob      40   London       
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestDescTiebreak(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort name asc, age desc, city
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Alice    30   New York     
Bob      40   London       
Bob      25   Los Angeles  
Charlie  30   Paris        
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMix(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort age, city desc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestLeavesInExtractionOrderIfNoTiebreaker(t *testing.T) {
	rsl := setupSortingRsl + `
rad url:
    fields name, age, city
    sort age asc
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./json/people.json", "--NO-COLOR")
	expected := `name     age  city        
Bob      25   Los Angeles  
Charlie  30   Paris        
Alice    30   New York     
Bob      40   London       
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
