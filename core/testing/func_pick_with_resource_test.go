package testing

import "testing"

const (
	websiteScript = `
args:
    filter str

url = pick_from_resource("./resources/website.json", filter)
print(url)`
	websitesScript = `
args:
    filter str

url, title = pick_from_resource("./resources/websites.json", filter)
print(url)
print(title)`
	peopleScript = `
args:
    filter str

name, age = pick_from_resource("./resources/people.json", filter)
print(name)
print(age * 10)`
)

func Test_SingleValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websiteScript, "gh")
	expected := `github.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_SingleValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websiteScript, "gl")
	expected := `gitlab.com
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueSameTypesPickFromResourceGh(t *testing.T) {
	setupAndRunCode(t, websitesScript, "gh")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueSameTypesPickFromResourceHub(t *testing.T) {
	setupAndRunCode(t, websitesScript, "hub")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueSameTypesPickFromResourceGl(t *testing.T) {
	setupAndRunCode(t, websitesScript, "gl")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueSameTypesPickFromResourceLab(t *testing.T) {
	setupAndRunCode(t, websitesScript, "lab")
	expected := `gitlab.com
GitLab
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueDifferentTypesPickFromResourceAlice(t *testing.T) {
	setupAndRunCode(t, peopleScript, "alice")
	expected := `Alice
250
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueDifferentTypesPickFromResourceBob(t *testing.T) {
	setupAndRunCode(t, peopleScript, "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_MultiValueDifferentTypesPickFromResourceRobert(t *testing.T) {
	setupAndRunCode(t, peopleScript, "robert")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ResourcePathIsRelativeToScript(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/people_resource.rad", "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ExactMatch_Priority(t *testing.T) {
	// "g" fuzzy-matches BOTH entries: gitlab (gl contains g) and github (has key "g")
	// But "g" exact-matches only github's key "g"
	// So with exact match priority, it should pick github immediately
	setupAndRunCode(t, websitesScript, "g")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ExactMatch_CaseInsensitive(t *testing.T) {
	// "G" should exactly match key "g" (case-insensitive)
	setupAndRunCode(t, websitesScript, "G")
	expected := `github.com
GitHub
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_JoinedLabelNotExactMatch(t *testing.T) {
	// "gh hub g" (the joined label) should NOT count as exact match
	// since we only check individual keys, not the joined label
	// This filter won't fuzzy-match individual keys,
	// so it should error with 0 matches
	setupAndRunCode(t, websitesScript, "--color=never", "gh hub g")
	expected := `Error at L5:14

  url, title = pick_from_resource("./resources/websites.json", filter)
               ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
               Filtered 2 options to 0 with filters: [gh hub g]
`
	assertError(t, 1, expected)
}

func Test_Pick_PrioExact_OptIn(t *testing.T) {
	// pick() with prefer_exact=true should prioritize exact matches
	// "g" fuzzy-matches both "grape" and "g", but exact-matches only "g"
	script := `
result = pick(["grape", "g"], "g", prefer_exact=true)
print(result)`
	setupAndRunCode(t, script)
	expected := `g
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_PickKv_PrioExact_OptIn(t *testing.T) {
	// pick_kv() with prefer_exact=true should prioritize exact matches
	script := `
result = pick_kv(["grape", "g"], [1, 2], "g", prefer_exact=true)
print(result)`
	setupAndRunCode(t, script)
	expected := `2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
