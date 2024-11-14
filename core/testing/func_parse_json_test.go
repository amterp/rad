package testing

import "testing"

func TestParseJson_Int(t *testing.T) {
	rsl := `
a = parse_json("2")
print(a + 1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_Float(t *testing.T) {
	rsl := `
a = parse_json("2.1")
print(a + 1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3.1\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_Bool(t *testing.T) {
	rsl := `
a = parse_json("true")
print(a or false)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "true\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_String(t *testing.T) {
	rsl := `
a = parse_json('"alice"')
print(a + "e")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "alicee\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_Map(t *testing.T) {
	rsl := `
a = parse_json('\{"name": "alice", "age": 20, "height": 5.5, "is_student": true, "cars": ["audi", "bmw"], "friends": \{"bob": 1, "charlie": 2}}')
print(a["name"] + "e")
print(a["age"] + 1)
print(a["height"] + 1.1)
print(a["is_student"] or false)
print(a["cars"][0] + "e")
print(a["friends"]["bob"] + 1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "alicee\n21\n6.6\ntrue\naudie\n2\n")
	assertNoErrors(t)
	resetTestState()
}

func TestParseJson_ErrorsOnInvalidJson(t *testing.T) {
	rsl := `
parse_json('\{asd asd}')
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/10 on 'parse_json': Error parsing JSON: invalid character 'a' looking for beginning of object key string\n")
	resetTestState()
}

func TestParseJson_ErrorsOnInvalidType(t *testing.T) {
	rsl := `
parse_json(10)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/10 on 'parse_json': parse_json() expects string, got int\n")
	resetTestState()
}

func TestParseJson_ErrorsOnNoArgs(t *testing.T) {
	rsl := `
parse_json()
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/10 on 'parse_json': parse_json() takes exactly one argument, got 0\n")
	resetTestState()
}

func TestParseJson_ErrorsOnTooManyArgs(t *testing.T) {
	rsl := `
parse_json("1", "2")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/10 on 'parse_json': parse_json() takes exactly one argument, got 2\n")
	resetTestState()
}
