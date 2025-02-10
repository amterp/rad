package testing

import "testing"

func TestColor_CanPrint(t *testing.T) {
	rsl := `
print(red("Alice"))
print(blue("Bob"))
print(yellow("Charlie"))
print(yellow(2))
print(blue([true, "hi", 10]))
`
	setupAndRunCode(t, rsl)
	expected := red("Alice") + "\n"
	expected += blue("Bob") + "\n"
	expected += yellow("Charlie") + "\n"
	expected += yellow("2") + "\n"
	expected += blue("[ true, \"hi\", 10 ]") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_RespectsNoColorFlag(t *testing.T) {
	rsl := `
print(red("Alice"))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := "Alice\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_CanConcat(t *testing.T) {
	rsl := `
print(red("Alice ") + blue("Bob ") + yellow("Charlie"))
`
	setupAndRunCode(t, rsl)
	expected := red("Alice ") + blue("Bob ") + yellow("Charlie") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_CanUpperLower(t *testing.T) {
	rsl := `
print(upper(red("Alice")))
print(lower(red("Alice")))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := red("ALICE") + "\n" + red("alice") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_ChangesDoNotAffectOriginalString(t *testing.T) {
	rsl := `
a = "Alice"
print(lower(red(a)))
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := red("alice") + "\n" + "Alice" + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_CanIndex(t *testing.T) {
	rsl := `
a = upper(red("Alice"))
print(a[2])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := red("I") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_CanSlice(t *testing.T) {
	rsl := `
a = upper(red("Alice"))
print(a[2:4])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := "IC" + "\n" // TODO this *should* be red
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_CanPrintEmojis(t *testing.T) {
	rsl := `
print(red("hi 👋"))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := red("hi 👋") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_CanPrintInArray(t *testing.T) {
	rsl := `
a = "Alice"
b = red(a)
c = red(a)
print([a, b, c])
`
	setupAndRunCode(t, rsl)
	expected := "[ \"Alice\", \"" + red("Alice") + "\", \"" + red("Alice") + "\" ]\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

// todo this should not be the case!!
//   - given the below equality test, this should just be a single [Alice] (drop all attrs?)
func TestColor_UniqueConsidersColors(t *testing.T) {
	rsl := `
a = "Alice"
b = red(a)
c = red(a)
print(unique([a, b, c]))
`
	setupAndRunCode(t, rsl)
	expected := "[ \"Alice\", \"" + red("Alice") + "\" ]\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestColor_Equality(t *testing.T) {
	rsl := `
a = "Alice"
b = red(a)
c = red(a)
print(a == b)
print(b == c)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := "true\ntrue\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
