package testing

import "testing"

func Test_Map_Parse_Empty(t *testing.T) {
	rsl := `
a = {}
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{}\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Map_Parse_EmptyTwoLine(t *testing.T) {
	rsl := `
a = {
}
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{}\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Map_Parse_SingleSpaced(t *testing.T) {
	rsl := `
a = { "alice" : 1 }
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 1 }\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Map_Parse_SingleMultiline(t *testing.T) {
	rsl := `
a = {
	"alice": 1
}
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 1 }\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Map_Parse_SingleTrailingComma(t *testing.T) {
	rsl := `
a = {"alice": 1,}
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 1 }\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Map_Parse_DoubleWeird(t *testing.T) {
	rsl := `
a = {
	"a":	1,

	"b":	2,
}
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{ a: 1, b: 2 }\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Map_Parse_ErrorsOnCommaNoElements(t *testing.T) {
	rsl := `
a = {,}
print(a)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/6 on ',': Expected expression\n")
	resetTestState()
}

func Test_List_Parse(t *testing.T) {
	rsl := `
a = [
	1, 2, 3,
	4,
	5, 6, 7,
		8,
9,
	10,
]
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Mixed_Parse(t *testing.T) {
	rsl := `
a = [
	"a", "b",
	"c",
	[],
	[1, 2, 3],
	["nested",
		["deeply", "nested"],
	],
	[
		"mixed",
		1,
		2.5,
		[
			"another",
			"level",
		],
		"types",
	],
	{},
	{"key": "value"},
	{
		"another": "map",
		"with": [
			"nested",
			"list",
		],
	},
	1,
	2,
	3,
]
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[a, b, c, [], [1, 2, 3], [nested, [deeply, nested]], [mixed, 1, 2.5, [another, level], types], {}, { key: value }, { another: map, with: [nested, list] }, 1, 2, 3]\n")
	assertNoErrors(t)
	resetTestState()
}
