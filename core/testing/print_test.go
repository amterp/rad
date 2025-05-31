package testing

import (
	"testing"
)

func Test_Print(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/print.rad")
	expected := `hi alice
hi bob
hi charlie
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_DebugNoDebugFlag(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/debug.rad")
	expected := "one\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_DebugWithDebugFlag(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/debug.rad", "--debug")
	expected := "one\nDEBUG: two\nDEBUG: three\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_PrettyPrint_Dict(t *testing.T) {
	script := `
url = "https://google.com"
node = json.results.Alice
request url:
    fields node
pprint(node[0])
`
	expected := `{
  "age": 30,
  "hometown": "New York"
}
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/unique_keys.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func Test_PrettyPrint_Int(t *testing.T) {
	script := `
url = "https://google.com"
node = json.results.Alice.age
request url:
    fields node
pprint(node[0])
`
	expected := `30
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/unique_keys.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func Test_PrettyPrint_Array(t *testing.T) {
	script := `
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
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/arrays.json", "--color=never")
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func Test_PrettyPrint_Complex(t *testing.T) {
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

func Test_PrettyPrint_Basics(t *testing.T) {
	script := `
pprint("alice")
pprint(21)
pprint(1.2)
pprint(false)
`
	expected := `"alice"
21
1.2
false
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_PrettyPrint_Map(t *testing.T) {
	script := `
a = { "alice": 35, "bob": "bar", "charlie": [1, "hi"] }
pprint(a)
`
	expected := `{
  "alice": 35,
  "bob": "bar",
  "charlie": [
    1,
    "hi"
  ]
}
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Print_CanPrintEmojis(t *testing.T) {
	script := `
print("👋")`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "👋\n")
	assertNoErrors(t)
}

func Test_Print_CanCustomizeEnd(t *testing.T) {
	script := `
print("hello", "there", "claire", end="bloop")`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello there clairebloop")
	assertNoErrors(t)
}

func Test_Print_CanUseEndToRemoveNewlines(t *testing.T) {
	script := `
print("hello", end="")
print("there", end="")
print("claire", end="")`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hellothereclaire")
	assertNoErrors(t)
}

func Test_Print_CanCustomizeSep(t *testing.T) {
	script := `
print("hello", "there", "claire", sep="_")`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello_there_claire\n")
	assertNoErrors(t)
}

func Test_Print_PrettyPrintEmptyList(t *testing.T) {
	script := `
blop = { "foo": [] }
blop.pprint()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{
  "foo": []
}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_PrintErr(t *testing.T) {
	script := `
print_err("hi alice")
print_err("hi", "bob", sep="_")
print_err("hi", end="_charlie\n")
`
	setupAndRunCode(t, script)
	expected := `hi alice
hi_bob
hi_charlie
`
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertNoErrors(t)
}
