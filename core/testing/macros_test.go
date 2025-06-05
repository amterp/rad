package testing

import (
	"testing"
)

func Test_Macros_ReadingStashId(t *testing.T) {
	script := `
---
Docs here.
Many lines!
@stash_id = abracadabra
---
`
	setupAndRunCode(t, script, "--color=never", "-h")
	expected := `Docs here.
Many lines!

Usage:
  [OPTIONS]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_ReadingStashIdWithSpace(t *testing.T) {
	script := `
---
Docs here.
Many lines!
@stash_id = abracadabra bloop
---
`
	setupAndRunCode(t, script, "--color=never", "-h")
	expected := `Docs here.
Many lines!

Usage:
  [OPTIONS]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_BecomesPartOfContentsIfFollowedByMoreDoc(t *testing.T) {
	script := `
---
Docs here.
Many lines!
@stash_id = abracadabra
Another line!
---
`
	setupAndRunCode(t, script, "--color=never", "-h")
	expected := `Docs here.
Many lines!
@stash_id = abracadabra
Another line!

Usage:
  [OPTIONS]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_HelpDisabledIfGlobalOptionsDisabled(t *testing.T) {
	script := `
---
Docs here.
@enable_global_options = 0
---
print("hi")
`
	setupAndRunCode(t, script, "--help")
	expected := `hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_DisablingGlobalOptionsLeadsToComplaintsAboutThemIfSpecified(t *testing.T) {
	script := `
---
Docs here.
@enable_global_options = 0
---
debug("hi1")
print("hi2")
`
	setupAndRunCode(t, script, "--debug")
	expected := `unknown flag: --debug

Docs here.

Usage:
 
`
	assertError(t, 1, expected)
}

func Test_Macros_ErrorsIfArgsBlockDisabledButSpecified(t *testing.T) {
	script := `
---
Docs here.
@enable_args_block = 0
---
args:
	name string
print("hi")
`
	setupAndRunCode(t, script)
	expected := `Macro 'enable_args_block' disabled, but args block found.
`
	assertError(t, 1, expected)
}

func Test_Macros_DoesPassthroughOfHelp(t *testing.T) {
	script := `
---
@enable_args_block = 0
@enable_global_options = 0
---

my_args = get_args()[1:].join(" ")
quiet $!'./rad_scripts/hello.rad {my_args}'
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  hello.rad <name>

Script args:
      --name string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
