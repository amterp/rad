package testing

import "testing"

// todo need to mock out huh so that we can write tests that actually filter down further when prompted

func TestPickKvReturnsOnlyOption(t *testing.T) {
	script := `
keys = ["Chicken"]
values = ["Chicken Burger"]
print(pick_kv(keys, values))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Chicken Burger
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestPickKvFilterToOneOption(t *testing.T) {
	script := `
keys = ["Beef", "Chicken", "Fish"]
values = ["Hamburger", "Chicken Burger", "Fishwich"]
print(pick_kv(keys, values, "Bee"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hamburger
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestPickKvErrorsIfEmptyKeysValues(t *testing.T) {
	script := `
keys = []
values = []
pick_kv(keys, values)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L4:1

  pick_kv(keys, values)
  ^^^^^^^^^^^^^^^^^^^^^ Filtered 0 options to 0 with filters: []
`
	assertError(t, 1, expected)
}

func TestPickKvErrorsIfKeyValueArraysAreNotEqualLength(t *testing.T) {
	script := `
keys = ["Beef"]
values = ["Hamburger", "Chicken Burger"]
pick_kv(keys, values)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L4:1

  pick_kv(keys, values)
  ^^^^^^^^^^^^^^^^^^^^^
  Number of keys and values must match, but got 1 key and 2 values
`
	assertError(t, 1, expected)
}

func TestPickKvWorksWithMultipleTokens(t *testing.T) {
	script := `
keys = ["Beef", "Chicken", "Fish"]
values = ["Hamburger", "Chicken Burger", "Fishwich"]
print(pick_kv(keys, values, ["Be", "ef"]))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hamburger
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
