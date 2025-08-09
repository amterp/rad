package testing

import "testing"

func Test_NumExponential_CanWrite(t *testing.T) {
	script := `
print(1.0e9)
print(1e9)
print(1E1_8)
print(1.2e9)
print(12.3e9)
print(12.3e1_8)
print(1.23e-1_8)
print(12.3e-1_9)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1000000000
1000000000
1000000000000000000
1200000000
12300000000
12300000000000000000
0.00000000000000000123
0.00000000000000000123
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_NumExponential_CanUseInArgs(t *testing.T) {
	script := `
args:
    a float = 2e3
    a range [1e3, 3e3]
`
	setupAndRunCode(t, script, "4000", "--color=never")
	expected := `'a' value 4000 is > maximum 3000

Usage:
  TestCase [a] [OPTIONS]

Script args:
      --a float   Range: [1000, 3000] (default 2000)

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}
