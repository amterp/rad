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
}

func TestSingleValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websiteRsl, "gl")
	expected := `gitlab.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "gh")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceHub(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "hub")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "gl")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueSameTypesPickFromResourceLab(t *testing.T) {
	setupAndRunCode(t, websitesRsl, "lab")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueDifferentTypesPickFromResourceAlice(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "alice")
	expected := `Alice
250
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueDifferentTypesPickFromResourceBob(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestMultiValueDifferentTypesPickFromResourceRobert(t *testing.T) {
	setupAndRunCode(t, peopleRsl, "robert")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestResourcePathIsRelativeToScript(t *testing.T) {
	setupAndRunArgs(t, "./rsl_scripts/people_resource.rad", "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
