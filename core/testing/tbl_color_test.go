package testing

import (
	"github.com/fatih/color"
	"testing"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()

func TestRadColorStmts(t *testing.T) {
	rsl := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
	fields name, city
	name:
		color "red" "o[a-z]"
	city:
		color "blue" "o[a-z]"
`
	setupAndRunCode(t, rsl, "--MOCK-RESPONSE", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += "B" + red("ob") + "      L" + blue("on") + "d" + blue("on") + "       \n"
	expected += "Alice    New Y" + blue("or") + "k     \n"
	expected += "B" + red("ob") + "      L" + blue("os") + " Angeles  \n"
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
	resetTestState()
}

func TestRadColorErrorsOnInvalidColor(t *testing.T) {
	rsl := `
url = "https://google.com"
name = json[].name
color = "licorice"
rad url:
	fields name
	name:
		color color "o[a-z]"
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L8/8 on 'color': Invalid color value \"licorice\". Allowed: [black red green yellow blue magenta cyan white]\n")
	resetTestState()
}
