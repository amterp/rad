package testing

import (
	"testing"
)

func Test_Macros_ReadingStashId(t *testing.T) {
	rsl := `
---
Docs here.
Many lines!
@stash_id = abracadabra
---
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expected := `Docs here.
Many lines!

Usage:
 
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_ReadingStashIdWithSpace(t *testing.T) {
	rsl := `
---
Docs here.
Many lines!
@stash_id = abracadabra bloop
---
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expected := `Docs here.
Many lines!

Usage:
 
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_BecomesPartOfContentsIfFollowedByMoreDoc(t *testing.T) {
	rsl := `
---
Docs here.
Many lines!
@stash_id = abracadabra
Another line!
---
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expected := `Docs here.
Many lines!
@stash_id = abracadabra
Another line!

Usage:
 
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_HelpDisabledIfGlobalFlagsDisabled(t *testing.T) {
	rsl := `
---
Docs here.
@enable_global_flags = 0
---
print("hi")
`
	setupAndRunCode(t, rsl, "--help")
	expected := `hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_DisablingGlobalFlagsLeadsToComplaintsAboutThemIfSpecified(t *testing.T) {
	rsl := `
---
Docs here.
@enable_global_flags = 0
---
debug("hi1")
print("hi2")
`
	setupAndRunCode(t, rsl, "--debug")
	expected := `unknown flag: --debug

Docs here.

Usage:
 
`
	assertError(t, 1, expected)
}

func Test_Macros_ErrorsIfArgsBlockDisabledButSpecified(t *testing.T) {
	rsl := `
---
Docs here.
@enable_args_block = 0
---
args:
	name string
print("hi")
`
	setupAndRunCode(t, rsl)
	expected := `Macro 'enable_args_block' disabled, but args block found.
`
	assertError(t, 1, expected)
}

func Test_Macros_DoesPassthroughOfHelp(t *testing.T) {
	rsl := `
---
@enable_args_block = 0
@enable_global_flags = 0
---

my_args = get_args()[1:].join(" ")
quiet $!'./rsl_scripts/hello.rsl {my_args}'
`
	setupAndRunCode(t, rsl, "--help", "--color=never")
	expected := `Usage:
  hello.rsl <name>

Script args:
      --name string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
