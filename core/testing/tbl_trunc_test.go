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
	expected := `id  words      
1   Lorem ips…  
2   Ut placer…  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestTruncateMatchesColWidth(t *testing.T) {
	script := `
url = "https://google.com"
id = json[].id
name = json[].name
rad url:
	fields id, name
	name:
		map fn(x) truncate(x, 5)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	expected := `id  name  
1   Alice  
2   Bob    
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}

func TestTruncateErrorsIfInvalidField(t *testing.T) {
	script := `
url = "https://google.com"
name = json[].name
rad url:
    fields name
    does_not_exist:
        map fn(x) truncate(x, 5)
`
	setupAndRunCode(t, script, "--mock-response", ".*:./responses/id_name.json", "--color=never")
	assertErrorContains(t, 1, "RAD20028", "Cannot modify undefined field", "does_not_exist")
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
	expected := `age  name   city  
30   Char…  Paris  
40   Bob    Lond…  
30   Alice  New …  
25   Bob    Los …  
`
	assertOutput(t, stdOutBuffer, expected)
	assertOutput(t, stdErrBuffer, "Mocking response for url (matched \".*\"): https://google.com\n")
	assertNoErrors(t)
}
