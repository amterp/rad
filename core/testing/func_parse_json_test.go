package testing

import "testing"

func TestParseJson_Int(t *testing.T) {
	rsl := `
a = parse_json("2")
print(a + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_Float(t *testing.T) {
	rsl := `
a = parse_json("2.1")
print(a + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3.1\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_Bool(t *testing.T) {
	rsl := `
a = parse_json("true")
print(a or false)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_String(t *testing.T) {
	rsl := `
a = parse_json('"alice"')
print(a + "e")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "alicee\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_Map(t *testing.T) {
	rsl := `
a = parse_json(r'{"name": "alice", "age": 20, "height": 5.5, "is_student": true, "cars": ["audi", "bmw"], "friends": {"bob": 1, "charlie": 2}}')
print(a["name"] + "e")
print(a["age"] + 1)
print(a["height"] + 1.1)
print(a["is_student"] or false)
print(a["cars"][0] + "e")
print(a["friends"]["bob"] + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `alicee
21
6.6
true
audie
2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_ErrorsOnInvalidJson(t *testing.T) {
	rsl := `
parse_json(r'{asd asd}')
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  parse_json(r'{asd asd}')
  ^^^^^^^^^^^^^^^^^^^^^^^^
  Error parsing JSON: invalid character 'a' looking for beginning of object key string
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestParseJson_ErrorsOnInvalidType(t *testing.T) {
	rsl := `
parse_json(10)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:12

  parse_json(10)
             ^^ Got "int" as the 1st argument of parse_json(), but must be: string
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestParseJson_ErrorsOnNoArgs(t *testing.T) {
	rsl := `
parse_json()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  parse_json()
  ^^^^^^^^^^^^ parse_json() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestParseJson_ErrorsOnTooManyArgs(t *testing.T) {
	rsl := `
parse_json("1", "2")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  parse_json("1", "2")
  ^^^^^^^^^^^^^^^^^^^^ parse_json() requires at most 1 argument, but got 2
`
	assertError(t, 1, expected)
	resetTestState()
}
