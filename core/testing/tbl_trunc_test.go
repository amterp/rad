package testing

import (
	"os"
	"testing"

	com "github.com/amterp/rad/core/common"
)

func TestTruncate(t *testing.T) {
	// Ensure UTF-8 mode for consistent ellipsis character across platforms
	os.Setenv("LANG", "en_US.UTF-8")
	com.TerminalIsUtf8 = true
	defer func() {
		os.Unsetenv("LANG")
		com.TerminalIsUtf8 = false
	}()
	script := `
url = "https://google.com"
id = json[].id
words = json[].words
rad url:
	fields id, words
	words:
		map fn(x) truncate(x, 10)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/long_values.json", "--color=never")
	expected := "id  words      \n1   Lorem ips…  \n2   Ut placer…  \n"
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestTruncateTwoFieldsAtOnce(t *testing.T) {
	// Ensure UTF-8 mode for consistent ellipsis character across platforms
	os.Setenv("LANG", "en_US.UTF-8")
	com.TerminalIsUtf8 = true
	defer func() {
		os.Unsetenv("LANG")
		com.TerminalIsUtf8 = false
	}()

	script := `
url = "https://google.com"
age = json[].age
name = json[].name
city = json[].city
rad url:
	fields age, name, city
	name, city:
		map fn(x) truncate(x, 5)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/people.json", "--color=never")
	expected := "age  name   city  \n30   Char…  Paris  \n40   Bob    Lond…  \n30   Alice  New …  \n25   Bob    Los …  \n"
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}
