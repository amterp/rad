package testing

import "testing"

func Test_FileHeader_PrintsOneLinerIfOnlyThat(t *testing.T) {
	rsl := `
---
This is a one liner!
---
args:
	name string
`
	setupAndRunCode(t, rsl, "-h", "--color=never")
	expected := `This is a one liner!

Usage:
  <name>

Script args:
      --name string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FileHeader_PrintsAll(t *testing.T) {
	rsl := `
---
This is a one liner!

Here is
the rest!
---
args:
	name string
`
	setupAndRunCode(t, rsl, "-h", "--color=never")
	expected := `This is a one liner!

Here is
the rest!

Usage:
  <name>

Script args:
      --name string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
