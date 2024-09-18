package testing

import "testing"

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

func TestKeyExtraction(t *testing.T) {
	rsl := `
url = "https://google.com"

Name = json.results.*
Age = json.results.*.age
Hometown = json.results.*.hometown

rad url:
    fields Name, Age, Hometown
`

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./test_json/unique_keys.json")
	expected := `Mocking response for url (matched ".*"): https://google.com
NAME   AGE  HOMETOWN    
Alice  30   New York     
Bob    40   Los Angeles  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestKeyArrayExtraction(t *testing.T) {
	rsl := `
url = "https://google.com"

Hometown = json.*
Name = json.*[].name
Age = json.*[].age

rad url:
    fields Name, Age, Hometown
`

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./test_json/unique_keys_array.json")
	expected := `Mocking response for url (matched ".*"): https://google.com
NAME       AGE  HOMETOWN 
Alice      30   London    
Bob        40   London    
Charlotte  35   Paris     
David      25   Paris     
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestNestedWildcardExtraction(t *testing.T) {
	rsl := `
url = "https://google.com"

city = json.*
country = json.*.*[]
name = json.*.*[].name
age = json.*.*[].age

rad url:
    fields city, country, name, age
`

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./test_json/nested_wildcard.json")
	expected := `Mocking response for url (matched ".*"): https://google.com
CITY  COUNTRY    NAME       AGE 
York  England    Alice      30   
York  England    Bob        40   
York  Australia  Charlotte  35   
York  Australia  David      25   
York  Australia  Eve        20   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
