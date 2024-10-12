package testing

import (
	"testing"
)

func TestPrint(t *testing.T) {
	setupAndRunArgs(t, "./rads/print.rad")
	expected := `hi alice
hi bob
hi charlie
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestDebugNoDebugFlag(t *testing.T) {
	setupAndRunArgs(t, "./rads/debug.rad")
	expected := "one\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestDebugWithDebugFlag(t *testing.T) {
	setupAndRunArgs(t, "./rads/debug.rad", "--DEBUG")
	expected := "one\nDEBUG: two\nDEBUG: three\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestPrettyPrintDict(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json.results.Alice
request url:
    fields node
pprint(node)
`
	expected := `{
  "age":30,
  "hometown":"New York"
}
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/unique_keys.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPrettyPrintInt(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json.results.Alice.age
request url:
    fields node
pprint(node)
`
	expected := `30
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/unique_keys.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPrettyPrintArray(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json[].ids
request url:
    fields node
pprint(node)
`
	expected := `[
  [
    1,
    2,
    3
  ],
  [
    4,
    5,
    6,
    7,
    8
  ],
  [
    9,
    10
  ]
]
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/arrays.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPrettyPrintComplex(t *testing.T) {
	rsl := `
url = "https://google.com"
node = json
request url:
    fields node
pprint(node)
`
	expected := `[
  {
    "friends": [
      {
        "id":2,
        "name":"Bob"
      }
    ],
    "height":1.7,
    "id":1,
    "name":"Alice",
    "old":true
  },
  {
    "friends": [
      {
        "id":1,
        "name":"Alice"
      },
      {
        "height":null,
        "id":3,
        "name":"Charlie"
      },
      null
    ],
    "height":1.8,
    "id":2,
    "name":"Bob",
    "old":false
  },
  null
]
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/lots_of_types.json", "--NO-COLOR")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPrettyPrintBasics(t *testing.T) {
	rsl := `
pprint("alice")
pprint(21)
pprint(1.2)
pprint(false)
`
	expected := `alice
21
1.2
false
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestPrettyPrintMap(t *testing.T) {
	rsl := `
a = { "alice": 35, "bob": "bar", "charlie": [1, "hi"] }
pprint(a)
`
	expected := `{
  "alice":35,
  "bob":"bar",
  "charlie": [
    1,
    "hi"
  ]
}
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
