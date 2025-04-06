package testing

import (
	"testing"

	"github.com/amterp/color"
)

func Test_Color_CanPrint(t *testing.T) {
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
}

func Test_Color_RespectsNoColorFlag(t *testing.T) {
	rsl := `
print(red("Alice"))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := "Alice\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanConcat(t *testing.T) {
	rsl := `
print(red("Alice ") + blue("Bob ") + yellow("Charlie"))
`
	setupAndRunCode(t, rsl)
	expected := red("Alice ") + blue("Bob ") + yellow("Charlie") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanUpperLower(t *testing.T) {
	rsl := `
print(upper(red("Alice")))
print(lower(red("Alice")))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := red("ALICE") + "\n" + red("alice") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_ChangesDoNotAffectOriginalString(t *testing.T) {
	rsl := `
a = "Alice"
print(lower(red(a)))
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := red("alice") + "\n" + "Alice" + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanIndex(t *testing.T) {
	rsl := `
a = upper(red("Alice"))
print(a[2])
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := red("I") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanSlice(t *testing.T) {
	rsl := `
a = upper(red("Alice"))
print(a[2:4])
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := "IC" + "\n" // TODO this *should* be red
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanPrintEmojis(t *testing.T) {
	rsl := `
print(red("hi 👋"))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := red("hi 👋") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanPrintInArray(t *testing.T) {
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
}

// todo this should not be the case!!
//   - given the below equality test, this should just be a single [Alice] (drop all attrs?)
func Test_Color_UniqueConsidersColors(t *testing.T) {
	rsl := `
a = "Alice"
b = red(a)
c = red(a)
print(unique([a, b, c]))
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "[ \"Alice\", \"" + red("Alice") + "\" ]\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_Equality(t *testing.T) {
	rsl := `
a = "Alice"
b = red(a)
c = red(a)
print(a == b)
print(b == c)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := "true\ntrue\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_HyperlinkEquality(t *testing.T) {
	rsl := `
a = "Alice"
b = red(a).hyperlink("https://example.com")
c = red(a).hyperlink("https://example.com")
print(a == b)
print(b == c)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := "true\ntrue\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanHyperlink(t *testing.T) {
	rsl := `
a = "Alice"
print(a.hyperlink("https://example.com"))
print(a.red().hyperlink("https://example.com"))
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := color.New().Hyperlink("https://example.com").Sprintf("Alice") + "\n" + color.New(color.FgRed).Hyperlink("https://example.com").Sprintf("Alice") + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanBold(t *testing.T) {
	rsl := `
a = "Alice"
print(a.bold())
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[1mAlice\x1b[22m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanItalic(t *testing.T) {
	rsl := `
a = "Alice"
print(a.italic())
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[3mAlice\x1b[23m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanUnderline(t *testing.T) {
	rsl := `
a = "Alice"
print(a.underline())
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[4mAlice\x1b[24m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_CanColorInt(t *testing.T) {
	rsl := `
print(2.red())
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[31m2\x1b[0m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
