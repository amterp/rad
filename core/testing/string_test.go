package testing

import "testing"

func TestString_SimpleIndexing(t *testing.T) {
	script := `
a = "alice"
print(a[0])
print(a[1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\nl\n")
	assertNoErrors(t)
}

func TestString_NegativeIndexing(t *testing.T) {
	script := `
a = "alice"
print(a[-1])
print(a[-2])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "e\nc\n")
	assertNoErrors(t)
}

func TestString_ComplexIndexing(t *testing.T) {
	script := `
a = "alice"
print(a[len(a) - 3])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "i\n")
	assertNoErrors(t)
}

func TestString_MultiByteIndexing(t *testing.T) {
	script := `
s = "a😀b"
print(s[0])
print(s[1])
print(s[2])
print(s[-1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\n😀\nb\nb\n")
	assertNoErrors(t)
}

// Tests indexing across segment boundaries when first segment contains multi-byte chars.
// The emoji takes 4 bytes but 1 rune.
func TestString_MultiSegmentMultiByteIndexing(t *testing.T) {
	script := `
s = blue("a😀") + "bc"
print(s[0])
print(s[1])
print(s[2])
print(s[3])
`
	setupAndRunCode(t, script, "--color=never")
	expected := blue("a") + "\n"
	expected += blue("😀") + "\n"
	expected += "b\n"
	expected += "c\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
