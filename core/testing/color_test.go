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

func Test_Color_CanRgbColor(t *testing.T) {
	rsl := `
print("Hi".color_rgb(50, 110, 220))
print(2.color_rgb(50, 110, 220))
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[38;2;50;110;220mHi\x1b[0;22;0;0;0m\n"
	expected += "\x1b[38;2;50;110;220m2\x1b[0;22;0;0;0m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Color_ErrorsOnTooLargeRgb(t *testing.T) {
	rsl := `
"Hi".color_rgb(300, 110, 220)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:16

  "Hi".color_rgb(300, 110, 220)
                 ^^^ RGB values must be [0, 255]; got 300
`
	assertError(t, 1, expected)
}

func Test_Color_ErrorsOnNegativeRgb(t *testing.T) {
	rsl := `
"Hi".color_rgb(50, 110, -10)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:25

  "Hi".color_rgb(50, 110, -10)
                          ^^^ RGB values must be [0, 255]; got -10
`
	assertError(t, 1, expected)
}

func Test_Colorize_CanColorNames(t *testing.T) {
	rsl := `
names = ["Alice", "Bob", "Charlie", "David"]
for n in names:
	n.colorize(names).print()
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[38;2;230;38;25mAlice\x1b[0;22;0;0;0m\n\x1b[38;2;99;130;233mBob\x1b[0;22;0;0;0m\n\x1b[38;2;106;189;15mCharlie\x1b[0;22;0;0;0m\n\x1b[38;2;209;71;184mDavid\x1b[0;22;0;0;0m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Colorize_CanColorInDisplayBlock(t *testing.T) {
	rsl := `
names = ["Alice", "Bob", "Charlie", "David"]
display:
	fields names
	names:
		map fn(n) n.colorize(names)
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "\x1b[33mnames                                 \x1b[0m \n\x1b[38;2;230;38;25mAlice\x1b[0;22;0;0;0m                                   \n\x1b[38;2;99;130;233mBob\x1b[0;22;0;0;0m                                     \n\x1b[38;2;106;189;15mCharlie\x1b[0;22;0;0;0m                                 \n\x1b[38;2;209;71;184mDavid\x1b[0;22;0;0;0m                                   \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
