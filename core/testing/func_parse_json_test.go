package testing

import "testing"

func TestParseJson_Int(t *testing.T) {
	script := `
a = parse_json("2")
print(a + 1)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func TestParseJson_Float(t *testing.T) {
	script := `
a = parse_json("2.1")
print(a + 1)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3.1\n")
	assertNoErrors(t)
}

func TestParseJson_Bool(t *testing.T) {
	script := `
a = parse_json("true")
print(a or false)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\n")
	assertNoErrors(t)
}

func TestParseJson_String(t *testing.T) {
	script := `
a = parse_json('"alice"')
print(a + "e")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "alicee\n")
	assertNoErrors(t)
}

func TestParseJson_Map(t *testing.T) {
	script := `
a = parse_json(r'{"name": "alice", "age": 20, "height": 5.5, "is_student": true, "cars": ["audi", "bmw"], "friends": {"bob": 1, "charlie": 2}}')
print(a["name"] + "e")
print(a["age"] + 1)
print(a["height"] + 1.1)
print(a["is_student"] or false)
print(a["cars"][0] + "e")
print(a["friends"]["bob"] + 1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `alicee
21
6.6
true
audie
2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestParseJson_ErrorsOnInvalidJson(t *testing.T) {
	script := `
parse_json(r'{asd asd}')
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  parse_json(r'{asd asd}')
  ^^^^^^^^^^^^^^^^^^^^^^^^
  Error parsing JSON: invalid character 'a' looking for beginning of object key string (RAD20011)
`
	assertError(t, 1, expected)
}

func TestParseJson_ErrorsOnInvalidType(t *testing.T) {
	script := `
parse_json(10)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:12

  parse_json(10)
             ^^ Value '10' (int) is not compatible with expected type 'str'
`
	assertError(t, 1, expected)
}

func TestParseJson_ErrorsOnNoArgs(t *testing.T) {
	script := `
parse_json()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  parse_json()
  ^^^^^^^^^^^^ Missing required argument '_str'
`
	assertError(t, 1, expected)
}

func TestParseJson_ErrorsOnTooManyArgs(t *testing.T) {
	script := `
parse_json("1", "2")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  parse_json("1", "2")
  ^^^^^^^^^^^^^^^^^^^^ Expected at most 1 args, but was invoked with 2
`
	assertError(t, 1, expected)
}

func TestParseJson_EmptyList(t *testing.T) {
	script := `
blop = r'{ "foo": [] }'
blop.parse_json().print()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"foo\": [ ] }\n")
	assertNoErrors(t)
}
