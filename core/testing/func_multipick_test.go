package testing

import "testing"

func TestMultipickErrorsIfEmptyOptions(t *testing.T) {
	script := `
opts = []
multipick(opts)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:1

  multipick(opts)
  ^^^^^^^^^^^^^^^ Cannot multipick from empty options list
`
	assertError(t, 1, expected)
}

func TestMultipickErrorsIfMinIsNegative(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, min=-1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:1

  multipick(opts, min=-1)
  ^^^^^^^^^^^^^^^^^^^^^^^ min must be non-negative, got -1
`
	assertError(t, 1, expected)
}

func TestMultipickErrorsIfMaxIsNegative(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, max=-1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:1

  multipick(opts, max=-1)
  ^^^^^^^^^^^^^^^^^^^^^^^ max must be positive, got -1
`
	assertError(t, 1, expected)
}

func TestMultipickErrorsIfMaxIsZero(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, max=0)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:1

  multipick(opts, max=0)
  ^^^^^^^^^^^^^^^^^^^^^^ max must be positive, got 0
`
	assertError(t, 1, expected)
}

func TestMultipickErrorsIfMinGreaterThanMax(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, min=5, max=3)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:1

  multipick(opts, min=5, max=3)
  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ min (5) cannot be greater than max (3)
`
	assertError(t, 1, expected)
}

func TestMultipickErrorsIfMinGreaterThanOptionsLength(t *testing.T) {
	script := `
opts = ["Apple", "Banana"]
multipick(opts, min=3)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:1

  multipick(opts, min=3)
  ^^^^^^^^^^^^^^^^^^^^^^ min is 3 but only 2 options available
`
	assertError(t, 1, expected)
}
