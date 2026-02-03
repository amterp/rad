package testing

import (
	"os"
	"testing"
)

func Test_Stash_WriteStashFile(t *testing.T) {
	defer os.Remove("rad_test_home/stashes/with_stash/files/bloop.txt")
	script := `
---
@stash_id = with_stash
---

write_stash_file("bloop.txt", "hello HELLO")
r = load_stash_file("bloop.txt", "didn't find")

r.content.print()
r.full_path.split("/")[-6:].print()

// clean up
r.full_path.delete_path()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `hello HELLO
[ "testing", "rad_test_home", "stashes", "with_stash", "files", "bloop.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_WriteStashFileReturnsNullOnSuccess(t *testing.T) {
	defer os.Remove("rad_test_home/stashes/write_test/files/test_return.txt")
	script := `
---
@stash_id = write_test
---
result = write_stash_file("test_return.txt", "test content")
print(result)
print(type_of(result))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
