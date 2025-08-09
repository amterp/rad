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
  TestCase [OPTIONS]
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
  TestCase [OPTIONS]
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
  TestCase [OPTIONS]
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
	expected := "unknown flag: --help\n\nDocs here.\n\n\x1b[32;1mUsage:\x1b[0;22m\n  \x1b[1mTestCase\x1b[22m \x1b[36m[OPTIONS]\x1b[0m\n"
	assertError(t, 1, expected)
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
	expected := "unknown flag: --debug\n\nDocs here.\n\n\x1b[32;1mUsage:\x1b[0;22m\n  \x1b[1mTestCase\x1b[22m \x1b[36m[OPTIONS]\x1b[0m\n"
	assertError(t, 1, expected)
}

func Test_Macros_ErrorsIfArgsBlockDisabledButSpecified(t *testing.T) {
	script := `
---
Docs here.
@enable_args_block = 0
---
args:
	name str
print("hi")
`
	setupAndRunCode(t, script)
	expected := `Macro 'enable_args_block' disabled, but args block found.
`
	assertError(t, 1, expected)
}

// todo this test is actually bad because it invokes *another instance* of Rad to generate the usage string
func Test_Macros_DoesPassthroughOfHelp(t *testing.T) {
	t.Skip("TODO come back to this...")
	script := `
---
@enable_args_block = 0
@enable_global_options = 0
---

my_args = get_args().join(" ")
quiet $!'./rad_scripts/hello.rad {my_args}'
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  hello.rad <name> [OPTIONS]

Script args:
      --name str   

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Macros_DoesNotComplainOfUnusedPositionalArgsIfArgsBlocKDisabled(t *testing.T) {
	script := `
---
Docs here.
@enable_args_block = 0
---
print("hi", get_args())
`
	setupAndRunCode(t, script, "bob")
	expected := `hi [ "bob" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
