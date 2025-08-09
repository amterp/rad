package testing

import "testing"

func Test_FileHeader_PrintsOneLinerIfOnlyThat(t *testing.T) {
	script := `
---
This is a one liner!
---
args:
	name str
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `This is a one liner!

Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FileHeader_PrintsAll(t *testing.T) {
	script := `
---
This is a one liner!

Here is
the rest!
---
args:
	name str
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `This is a one liner!

Here is
the rest!

Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
