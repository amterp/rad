package testing

import "testing"

func TestAlgo_JsonNonRootArrayExtraction(t *testing.T) {
	script := `
url = "https://google.com"

Id = json.id
Names = json.names[]

request url:
    fields Id, Names
print(Id[0])
print(Names)
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/not_root_array.json", "--color=never")
	expected := `1
[ "Alice", "Bob", "Charlie" ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_KeyExtraction(t *testing.T) {
	script := `
url = "https://google.com"

Name = json.results.*
Age = json.results.*.age
Hometown = json.results.*.hometown

rad url:
    fields Name, Age, Hometown
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/unique_keys.json", "--color=never")
	expected := `Name   Age  Hometown    
Alice  30   New York     
Bob    40   Los Angeles  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_KeyArrayExtraction(t *testing.T) {
	script := `
url = "https://google.com"

Hometown = json.*
Name = json.*[].name
Age = json.*[].age

rad url:
    fields Name, Age, Hometown
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/unique_keys_array.json", "--color=never")
	expected := `Name       Age  Hometown 
Alice      30   London    
Bob        40   London    
Charlotte  35   Paris     
David      25   Paris     
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_NestedWildcardExtraction(t *testing.T) {
	script := `
url = "https://google.com"

city = json.*
country = json.*.*
name = json.*.*[].name
age = json.*.*[].age

rad url:
    fields city, country, name, age
`

	setupAndRunCode(t, script, "--mock-response", ".*:./responses/nested_wildcard.json", "--color=never")
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
}

func TestAlgo_WildcardListCapture(t *testing.T) {
	script := `
url = "https://google.com"

names = json.*
ids = json.*.ids

request url:
    fields names, ids
print(names)
print(ids)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/array_wildcard.json", "--color=never")
	expected := `[ "Alice", "Bob", "Charlie" ]
[ [ 1, 2, 3 ], [ 4, 5, 6, 7, 8 ], [ 9, 10 ] ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_WildcardListObjectCapture(t *testing.T) {
	script := `
url = "https://google.com"

names = json.*
ids = json.*.ids[].id

request url:
    fields names, ids
print(names)
print(ids)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/array_objects.json", "--color=never")
	expected := `[ "Alice", "Alice", "Alice", "Bob", "Charlie", "Charlie" ]
[ 1, 2, 3, 4, 5, 6 ]
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_ListOfObjectCapture(t *testing.T) {
	script := `
url = "https://google.com"
Building = json.buildings.*
issues = json.buildings.*.issues
request url:
    fields Building, issues
print([len(x) for x in issues])
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/issues.json", "--color=never")
	assertOutput(t, stdOutBuffer, "[ 2, 3 ]\n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_CaptureRootArray(t *testing.T) {
	script := `
url = "https://google.com"

ids = json[]

request url:
    fields ids
print(ids)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/root_prim_array.json", "--color=never")
	assertOutput(t, stdOutBuffer, "[ 1, 2, 3 ]\n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_CaptureNonArrayAndArray(t *testing.T) {
	script := `
url = "https://google.com"

len = json.len
ages = json.results[].age

request url:
    fields len, ages
print(len[0])
print(ages)
`
	expected := `2
[ 30, 40 ]
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/array_and_non_array.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_CaptureNonArrayAndWildcard(t *testing.T) {
	script := `
url = "https://google.com"

len = json.len
ages = json.results.*.age

request url:
    fields len, ages
print(len[0])
print(ages)
`
	expected := `2
[ 30, 40 ]
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/unique_keys.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_CaptureJsonNode(t *testing.T) {
	script := `
url = "https://google.com"
node = json.results.Alice
request url:
    fields node
print(sort("{node[0]}"))
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/unique_keys.json", "--color=never")
	assertOutput(t, stdOutBuffer, "      \"\"\"\"\"\",03::NYaeeeghkmnooortww{}\n")
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_CanCaptureWholeJson(t *testing.T) {
	script := `
url = "https://google.com"
node = json
request url:
    fields node
print(sort("{node[0]}")) // hack to get the test consistent, as the order of keys in a map is not guaranteed
`
	expected := `             """""""""""",,,12::::AB[]aabcddeeeiiilmmnno{{}}` + "\n"
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_CanCaptureWholeComplexJson(t *testing.T) {
	script := `
url = "https://google.com"
node = json
request url:
    fields node
pprint(node[0])
`
	expected := `[
  {
    "friends": [
      {
        "id": 2,
        "name": "Bob"
      }
    ],
    "height": 1.7,
    "id": 1,
    "name": "Alice",
    "old": true
  },
  {
    "friends": [
      {
        "id": 1,
        "name": "Alice"
      },
      {
        "height": null,
        "id": 3,
        "name": "Charlie"
      },
      null
    ],
    "height": 1.8,
    "id": 2,
    "name": "Bob",
    "old": false
  },
  null
]
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/lots_of_types.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestAlgo_HelpfulErrorIfRadBlockMixesArrayAndNoneArrayFields(t *testing.T) {
	script := `
url = "https://google.com"

Names = json.results[].name
Len = json.len

rad url:
    fields Names, Len
`
	expected := `Names  Len 
Alice  2    
Bob    2    
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/array_and_non_array.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
}
