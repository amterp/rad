package testing

import "testing"

func TestMultipickErrorsIfEmptyOptions(t *testing.T) {
	script := `
opts = []
multipick(opts)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "Cannot multipick from empty options list")
}

func TestMultipickErrorsIfMinIsNegative(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, min=-1)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "min must be non-negative")
}

func TestMultipickErrorsIfMaxIsNegative(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, max=-1)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "max must be positive")
}

func TestMultipickErrorsIfMaxIsZero(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, max=0)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "max must be positive")
}

func TestMultipickErrorsIfMinGreaterThanMax(t *testing.T) {
	script := `
opts = ["Apple", "Banana", "Cherry"]
multipick(opts, min=5, max=3)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "min (5) cannot be greater than max (3)")
}

func TestMultipickErrorsIfMinGreaterThanOptionsLength(t *testing.T) {
	script := `
opts = ["Apple", "Banana"]
multipick(opts, min=3)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "min is 3 but only 2 options available")
}
