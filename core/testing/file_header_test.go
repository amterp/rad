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
	setupAndRunCode(t, rsl, "-h", "--NO-COLOR")
	expected := `This is a one liner!

Usage:
  test <name>

Script args:
      --name string   

` + globalFlagHelp
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_FileHeader_PrintsExtra(t *testing.T) {
	rsl := `
---
This is a one liner!

Here is
the rest!
---
args:
	name string
`
	setupAndRunCode(t, rsl, "-h", "--NO-COLOR")
	expected := `Here is
the rest!

Usage:
  test <name>

Script args:
      --name string   

` + globalFlagHelp
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}