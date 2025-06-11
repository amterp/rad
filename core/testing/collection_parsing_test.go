package testing

import "testing"

func Test_Map_Parse_Empty(t *testing.T) {
	script := `
a = { }
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ }\n")
	assertNoErrors(t)
}

func Test_Map_Parse_EmptyTwoLine(t *testing.T) {
	script := `
a = {
}
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ }\n")
	assertNoErrors(t)
}

func Test_Map_Parse_SingleSpaced(t *testing.T) {
	script := `
a = { "alice" : 1 }
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 1 }\n")
	assertNoErrors(t)
}

func Test_Map_Parse_SingleMultiline(t *testing.T) {
	script := `
a = {
	"alice": 1
}
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 1 }\n")
	assertNoErrors(t)
}

func Test_Map_Parse_SingleTrailingComma(t *testing.T) {
	script := `
a = {"alice": 1,}
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 1 }\n")
	assertNoErrors(t)
}

func Test_Map_Parse_DoubleWeird(t *testing.T) {
	script := `
a = {
	"a":	1,

	"b":	2,
}
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"a\": 1, \"b\": 2 }\n")
	assertNoErrors(t)
}

func Test_Map_Parse_ErrorsOnCommaNoElements(t *testing.T) {
	script := `
a = {,}
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:6

  a = {,}
       ^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_List_Parse(t *testing.T) {
	script := `
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
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 ]\n")
	assertNoErrors(t)
}

func Test_Mixed_Parse(t *testing.T) {
	script := `
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
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(
		t,
		stdOutBuffer,
		"[ \"a\", \"b\", \"c\", [ ], [ 1, 2, 3 ], [ \"nested\", [ \"deeply\", \"nested\" ] ], [ \"mixed\", 1, 2.5, [ \"another\", \"level\" ], \"types\" ], { }, { \"key\": \"value\" }, { \"another\": \"map\", \"with\": [ \"nested\", \"list\" ] }, 1, 2, 3 ]\n",
	)
	assertNoErrors(t)
}
