package testing

import "testing"

func TestJsonNonRootArrayExtraction(t *testing.T) {
	rsl := `
url = "https://google.com"

Id = json.id
Names = json.names

request url:
    fields Id, Names
print(Id + 1)
print(Names)
`

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/not_root_array.json", "--NO-COLOR")
	expected := `2
[Alice, Bob, Charlie]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
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

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/unique_keys.json", "--NO-COLOR")
	expected := `Name   Age  Hometown    
Alice  30   New York     
Bob    40   Los Angeles  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
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

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/unique_keys_array.json", "--NO-COLOR")
	expected := `Name       Age  Hometown 
Alice      30   London    
Bob        40   London    
Charlotte  35   Paris     
David      25   Paris     
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
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

	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/nested_wildcard.json", "--NO-COLOR")
	expected := `city  country    name       age 
York  Australia  Charlotte  35   
York  Australia  David      25   
York  Australia  Eve        20   
York  England    Alice      30   
York  England    Bob        40   
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestWildcardListCapture(t *testing.T) {
	rsl := `
url = "https://google.com"

names = json.*
ids = json.*.ids

request url:
    fields names, ids
print(names)
print(ids)
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/array_wildcard.json", "--NO-COLOR")
	expected := `[Alice, Bob, Charlie]
[[1, 2, 3], [4, 5, 6, 7, 8], [9, 10]]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestWildcardListObjectCapture(t *testing.T) {
	rsl := `
url = "https://google.com"

names = json.*
ids = json.*.ids[].id

request url:
    fields names, ids
print(names)
print(ids)
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/array_objects.json", "--NO-COLOR")
	expected := `[Alice, Alice, Alice, Bob, Charlie, Charlie]
[1, 2, 3, 4, 5, 6]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestListOfObjectCapture(t *testing.T) {
	rsl := `
url = "https://google.com"
Building = json.buildings.*
issues = json.buildings.*.issues
request url:
    fields Building, issues
print([len(x) for x in issues])
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/issues.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, "[2, 3]\n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCaptureRootArray(t *testing.T) {
	rsl := `
url = "https://google.com"

ids = json

request url:
    fields ids
print(ids)
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/root_prim_array.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, "[1, 2, 3]\n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCaptureNonArrayAndArray(t *testing.T) {
	rsl := `
url = "https://google.com"

len = json.len
ages = json.results[].age

request url:
    fields len, ages
print(len + 1)
print(ages)
`
	expected := `3
[30, 40]
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/array_and_non_array.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCaptureNonArrayAndWildcard(t *testing.T) {
	rsl := `
url = "https://google.com"

len = json.len
ages = json.results.*.age

request url:
    fields len, ages
print(len + 1)
print(ages)
`
	expected := `3
[30, 40]
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/unique_keys.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCaptureJsonNode(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json.results.Alice
request url:
    fields node
print(node)
`
	expected := "{\"age\":30,\"hometown\":\"New York\"}\n"
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/unique_keys.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCanCaptureWholeJson(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json
request url:
    fields node
print(node)
`
	expected := "[{\"id\":1,\"name\":\"Alice\"}, {\"id\":2,\"name\":\"Bob\"}]\n"
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/id_name.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCanCaptureWholeComplexJson(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json
request url:
    fields node
print(node)
`
	expected := "[{\"friends\":[{\"id\":2,\"name\":\"Bob\"}],\"height\":1.7,\"id\":1,\"name\":\"Alice\",\"old\":true}, {\"friends\":[{\"id\":1,\"name\":\"Alice\"},{\"height\":null,\"id\":3,\"name\":\"Charlie\"},null],\"height\":1.8,\"id\":2,\"name\":\"Bob\",\"old\":false}, null]\n"
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/lots_of_types.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}
