package testing

import (
	"os"
	"testing"
)

func Test_Rad_home(t *testing.T) {
	script := `
get_rad_home().split("/")[-3:].print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "core", "testing", "rad_test_home" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_GetStashDirErrorsIfNoStashId(t *testing.T) {
	script := `
get_stash_dir()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20022", "Script ID is not set. Set the 'stash_id' macro in the file header.")
}

func Test_Stash_GetStashDir(t *testing.T) {
	script := `
---
@stash_id = test_id
---
get_stash_dir().split("/")[-5:].print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "core", "testing", "rad_test_home", "stashes", "test_id" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_GetStashDir_SubPath(t *testing.T) {
	script := `
---
@stash_id = test_id
---
get_stash_dir("some/path.txt").split("/")[-7:].print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "core", "testing", "rad_test_home", "stashes", "test_id", "some", "path.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadState(t *testing.T) {
	script := `
---
@stash_id = with_stash
---
state = load_state()
print(state)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "somekey": "somevalue" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadStateNoExisting(t *testing.T) {
	script := `
---
@stash_id = with_no_stash
---
state = load_state()
print(state)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_SaveAndLoadState(t *testing.T) {
	script := `
---
@stash_id = with_stash
---
m = load_state()
m.print()
save_state({ "changed": true })
load_state().print()
save_state(m)
load_state().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "somekey": "somevalue" }
{ "changed": true }
{ "somekey": "somevalue" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadStashFileExisting(t *testing.T) {
	script := `
---
@stash_id = with_stash
---
r = load_stash_file("existing.txt", "didn't find")
print(r.created, r.content)
r.full_path.split("/")[-6:].print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `false hello there!
[ "testing", "rad_test_home", "stashes", "with_stash", "files", "existing.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadStashFileNotExisting(t *testing.T) {
	script := `
---
@stash_id = with_stash
---
r = load_stash_file("non_existing.txt", "didn't find")
print(r.created, r.content)
r.full_path.split("/")[-6:].print()

p = get_path(r.full_path)
p.base_name.print()

// clean up
p.full_path.delete_path()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true didn't find
[ "testing", "rad_test_home", "stashes", "with_stash", "files", "non_existing.txt" ]
non_existing.txt
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

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

func Test_Stash_SaveStateReturnsNullOnSuccess(t *testing.T) {
	script := `
---
@stash_id = save_state_test
---
result = save_state({ "test": "value" })
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

func Test_Stash_SaveStateReturnsErrorOnFailure(t *testing.T) {
	script := `
result = save_state({ "test": "value" }) catch:
	pass
print(type_of(result))
print("Error contains 'Script ID':", "Script ID" in result)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `error
Error contains 'Script ID': true
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

func Test_Stash_WriteStashFileReturnsErrorOnFailure(t *testing.T) {
	script := `
result = write_stash_file("test.txt", "content") catch:
	pass
print(type_of(result))
print("Error contains 'Script ID':", "Script ID" in result)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `error
Error contains 'Script ID': true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
