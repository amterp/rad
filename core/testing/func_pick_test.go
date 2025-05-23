package testing

import "testing"

// todo need to mock out huh so that we can write tests that actually filter down further when prompted

func TestPickNoFilterOneOption(t *testing.T) {
	rsl := `
opts = ["Hamburger"]
print(pick(opts))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
}

func TestPickFilterWithOneOption(t *testing.T) {
	rsl := `
opts = ["Hamburger"]
print(pick(opts, "burg"))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
}

func TestPickFilterToOneOption(t *testing.T) {
	rsl := `
opts = ["Hamburger", "Chicken Burger", "Sandwich", "Fish", "Chickwich"]
print(pick(opts, "Hamb"))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
}

func TestPickErrorsIfEmptyOptions(t *testing.T) {
	rsl := `
opts = []
pick(opts)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:1

  pick(opts)
  ^^^^^^^^^^ Filtered 0 options to 0 with filters: []
`
	assertError(t, 1, expected)
}

func TestPickErrorsIfFilteredToZeroOptions(t *testing.T) {
	rsl := `
opts = ["Hamburger", "Chicken Burger", "Sandwich", "Fish", "Chickwich"]
pick(opts, "asdasdasd")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:1

  pick(opts, "asdasdasd")
  ^^^^^^^^^^^^^^^^^^^^^^^ Filtered 5 options to 0 with filters: [asdasdasd]
`
	assertError(t, 1, expected)
}

func TestPickWorksWithMultipleTokens(t *testing.T) {
	rsl := `
filter = ["Ham", "ger"]
opts = ["Hamburger", "Chicken Burger", "Sandwich", "Fish", "Chickwich"]
print(pick(opts, filter))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
}
