package testing

import "testing"

// todo need to mock out huh so that we can write tests that actually filter down further when prompted

func TestPickNoFilterOneOption(t *testing.T) {
	rsl := `
opts = ["Hamburger"]
print(pick(opts))
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPickFilterWithOneOption(t *testing.T) {
	rsl := `
opts = ["Hamburger"]
print(pick(opts, "burg"))
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPickFilterToOneOption(t *testing.T) {
	rsl := `
opts = ["Hamburger", "Chicken Burger", "Sandwich", "Fish", "Chickwich"]
print(pick(opts, "Hamb"))
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
	resetTestState()
}

func TestPickErrorsIfEmptyOptions(t *testing.T) {
	rsl := `
opts = []
pick(opts)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/4 on 'pick': Filtered 0 options to 0 with filters: []\n")
	resetTestState()
}

func TestPickErrorsIfFilteredToZeroOptions(t *testing.T) {
	rsl := `
opts = ["Hamburger", "Chicken Burger", "Sandwich", "Fish", "Chickwich"]
pick(opts, "asdasdasd")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/4 on 'pick': Filtered 5 options to 0 with filters: [asdasdasd]\n")
	resetTestState()
}

func TestPickWorksWithMultipleTokens(t *testing.T) {
	rsl := `
filter = ["Ham", "ger"]
opts = ["Hamburger", "Chicken Burger", "Sandwich", "Fish", "Chickwich"]
print(pick(opts, filter))
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hamburger\n")
	assertNoErrors(t)
	resetTestState()
}
