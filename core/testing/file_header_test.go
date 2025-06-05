package testing

import "testing"

func Test_FileHeader_PrintsOneLinerIfOnlyThat(t *testing.T) {
	script := `
---
This is a one liner!
---
args:
	name string
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `This is a one liner!

Usage:
  <name> [OPTIONS]

Script args:
      --name string   
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
	name string
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `This is a one liner!

Here is
the rest!

Usage:
  <name> [OPTIONS]

Script args:
      --name string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
