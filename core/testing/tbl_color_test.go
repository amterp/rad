package testing

import (
	"testing"

	"github.com/amterp/color"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()

func TestRadColor_NonOverlappingMatches(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
    fields name, city
    city:
       color "red" "Los"
       color "blue" "Angeles"
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += "Bob      London       \n"
	expected += "Alice    New York     \n"
	expected += "Bob      " + red("Los") + " " + blue("Angeles") + "  \n"

	assertOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadColor_OverlappingMatches(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
    fields name, city
    city:
       color "red" "New"
       color "blue" "York"
       color "yellow" "New York"
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += "Bob      London       \n"
	expected += "Alice    " + yellow("New York") + "     \n"
	expected += "Bob      Los Angeles  \n"

	assertOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadColor_PartialOverlapPriority(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
    fields name, city
    name:
       color "blue" "Bo"
       color "red" "ob"
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += blue("B") + red("ob") + "      London       \n"
	expected += "Alice    New York     \n"
	expected += blue("B") + red("ob") + "      Los Angeles  \n"

	assertOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadColor_NoMatches(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
    fields name, city
    city:
       color "green" "Berlin"
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += "Bob      London       \n"
	expected += "Alice    New York     \n"
	expected += "Bob      Los Angeles  \n"

	assertOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadColor_Complex(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
	fields name, city
	name:
		color "red" "o[a-z]"
	city:
		color "blue" "o[a-z]"
		color "yellow" "ndon"
		color "red" "ndo"
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += "B" + red("ob") + "      L" + blue("o") + red("ndo") + yellow("n") + "       \n"
	expected += "Alice    New Y" + blue("or") + "k     \n"
	expected += "B" + red("ob") + "      L" + blue("os") + " Angeles  \n"
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestRadColor_Plain(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
city = json[].city
rad url:
	fields name, city
	city:
		color "red" "London"
		color "plain" "ndo"
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json")
	expected := yellow("name   ") + "  " + yellow("city       ") + " \n"
	expected += "Charlie  Paris        \n"
	expected += "Bob      " + red("Lo") + "ndo" + red("n") + "       \n"
	expected += "Alice    New York     \n"
	expected += "Bob      Los Angeles  \n"

	assertOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRadColorErrorsOnInvalidColor(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
color = "licorice"
rad url:
    fields name
    name:
        color color "o[a-z]"
`
	setupAndRunCode(t, script, "--color=never")
	// todo 'bold', 'italic', 'underline' are not colors and not supported for rad block color stmts
	assertErrorContains(t, 1, "RAD20025", "Invalid color value", "licorice")
}
