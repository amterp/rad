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

func Test_Macros_NoGlobalFlagsInUsageIfDisabled(t *testing.T) {
	rsl := `
---
Docs here.
@enable_global_flags = 0
---
`
	setupAndRunCode(t, rsl, "--color=never", "--help")
	expected := `Docs here.

Usage:
 
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
