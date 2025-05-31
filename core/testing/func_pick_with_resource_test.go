package testing

import "testing"

const (
	websiteScript = `
args:
    filter string

url = pick_from_resource("./resources/website.json", filter)
print(url)`
	websitesScript = `
args:
    filter string

url, title = pick_from_resource("./resources/websites.json", filter)
print(url)
print(title)`
	peopleScript = `
args:
    filter string

name, age = pick_from_resource("./resources/people.json", filter)
print(name)
print(age * 10)`
)

func TestSingleValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websiteScript, "gh")
	expected := `github.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestSingleValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websiteScript, "gl")
	expected := `gitlab.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websitesScript, "gh")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceHub(t *testing.T) {
	setupAndRunCode(t, websitesScript, "hub")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websitesScript, "gl")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceLab(t *testing.T) {
	setupAndRunCode(t, websitesScript, "lab")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueDifferentTypesPickFromResourceAlice(t *testing.T) {
	setupAndRunCode(t, peopleScript, "alice")
	expected := `Alice
250
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueDifferentTypesPickFromResourceBob(t *testing.T) {
	setupAndRunCode(t, peopleScript, "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueDifferentTypesPickFromResourceRobert(t *testing.T) {
	setupAndRunCode(t, peopleScript, "robert")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestResourcePathIsRelativeToScript(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/people_resource.rad", "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
