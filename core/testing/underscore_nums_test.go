package testing

import "testing"

func Test_NumUnderscores_CanWrite(t *testing.T) {
	script := `
print(1_234)
print(0.123_456)
print(12_34.56_78)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1234
0.123456
1234.5678
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_NumUnderscores_CanUseInArgs(t *testing.T) {
	script := `
args:
  a int = 1_234
  b float = 0.123_456

  a range [1_000, 2_000]

print(a, b)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1234 0.123456
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_NumUnderscores_CanUseInArgs_Range(t *testing.T) {
	script := `
args:
  a int = 1_234
  b float = 0.123_456

  a range [1_000, 2_000]

print(a, b)
`
	setupAndRunCode(t, script, "3000", "--color=never")
	expected := `Error at L3:3

    a int = 1_234
    ^^^^^^^^^^^^^ 'a' value 3000 is > maximum 2000
`
	assertError(t, 1, expected)
}
