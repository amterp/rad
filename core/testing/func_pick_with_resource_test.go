package testing

import "testing"

const (
	websiteRsl = `
args:
    filter string

url = pick_with_resource("./resources/website.json", filter)
print(url)`
	websitesRsl = `
args:
    filter string

url, title = pick_with_resource("./resources/websites.json", filter)
print(url)
print(title)`
	peopleRsl = `
args:
    filter string

name, age = pick_with_resource("./resources/people.json", filter)
print(name)
print(age * 10)`
)

func TestSingleValueSameTypesPickWithResourceGh(t *testing.T) {
	setupAndRunCode(t, websiteRsl, "gh")
	expected := `github.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestSingleValueSameTypesPickWithResourceGl(t *testing.T) {
	setupAndRunCode(t, websiteRsl, "gl")
	expected := `gitlab.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickWithResourceGh(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "gh")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickWithResourceHub(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "hub")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickWithResourceGl(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "gl")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickWithResourceLab(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "lab")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueDifferentTypesPickWithResourceAlice(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "alice")
	expected := `Alice
250
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueDifferentTypesPickWithResourceBob(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueDifferentTypesPickWithResourceRobert(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "robert")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
