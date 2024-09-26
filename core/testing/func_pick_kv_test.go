package testing

import "testing"

// todo need to mock out huh so that we can write tests that actually filter down further when prompted

func TestPickKvReturnsOnlyOption(t *testing.T) {
	rsl := `
keys = ["Chicken"]
values = ["Chicken Burger"]
print(pick_kv(keys, values))
`
	setupAndRunCode(t, rsl)
	expected := `Chicken Burger
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestPickKvFilterToOneOption(t *testing.T) {
	rsl := `
keys = ["Beef", "Chicken", "Fish"]
values = ["Hamburger", "Chicken Burger", "Fishwich"]
print(pick_kv(keys, values, "Bee"))
`
	setupAndRunCode(t, rsl)
	expected := `Hamburger
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestPickKvErrorsIfEmptyKeysValues(t *testing.T) {
	rsl := `
keys string[] = []
values string[] = []
pick_kv(keys, values)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L4/7 on 'pick_kv': pick_kv() requires keys and values to have at least one element\n")
	resetTestState()
}

func TestPickKvErrorsIfKeyValueArraysAreNotEqualLength(t *testing.T) {
	rsl := `
keys = ["Beef"]
values = ["Hamburger", "Chicken Burger"]
pick_kv(keys, values)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L4/7 on 'pick_kv': pick_kv() requires keys and values to be the same length, got 1 keys and 2 values\n")
	resetTestState()
}

func TestPickKvWorksWithMultipleTokens(t *testing.T) {
	rsl := `
keys = ["Beef", "Chicken", "Fish"]
values = ["Hamburger", "Chicken Burger", "Fishwich"]
print(pick_kv(keys, values, ["Be", "ef"]))
`
	setupAndRunCode(t, rsl)
	expected := `Hamburger
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
