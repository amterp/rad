package testing

import "testing"

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
