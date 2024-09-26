package testing

import "testing"

const (
	websiteRsl = `
args:
    filter string

url = pick_from_resource("./resources/website.json", filter)
print(url)`
	websitesRsl = `
args:
    filter string

url, title = pick_from_resource("./resources/websites.json", filter)
print(url)
print(title)`
	peopleRsl = `
args:
    filter string

name, age = pick_from_resource("./resources/people.json", filter)
print(name)
print(age * 10)`
)

func TestSingleValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websiteRsl, "gh")
	expected := `github.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestSingleValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websiteRsl, "gl")
	expected := `gitlab.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "gh")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickFromResourceHub(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "hub")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "gl")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueSameTypesPickFromResourceLab(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "lab")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueDifferentTypesPickFromResourceAlice(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "alice")
	expected := `Alice
250
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueDifferentTypesPickFromResourceBob(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMultiValueDifferentTypesPickFromResourceRobert(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "robert")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestResourcePathIsRelativeToScript(t *testing.T) {
	setupAndRunArgs(t, "./rads/people_resource.rad", "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
