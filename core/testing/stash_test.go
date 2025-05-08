package testing

import "testing"

func Test_Rad_home(t *testing.T) {
	rsl := `
get_rad_home().split("/")[-3:].print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "core", "testing", "rad_test_home" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_GetStashDirErrorsIfNoStashId(t *testing.T) {
	rsl := `
get_stash_dir()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  get_stash_dir()
  ^^^^^^^^^^^^^^^ Script ID is not set. Use set_stash_id() first to set it.
`
	assertError(t, 1, expected)
}

func Test_Stash_GetStashDir(t *testing.T) {
	rsl := `
set_stash_id("test_id")
get_stash_dir().split("/")[-5:].print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "core", "testing", "rad_test_home", "stashes", "test_id" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_GetStashDir_SubPath(t *testing.T) {
	rsl := `
set_stash_id("test_id")
get_stash_dir("some/path.txt").split("/")[-7:].print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "core", "testing", "rad_test_home", "stashes", "test_id", "some", "path.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadState(t *testing.T) {
	rsl := `
set_stash_id("with_stash")
state, existed = load_state()
print(state, existed)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "somekey": "somevalue" } true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadStateNoExisting(t *testing.T) {
	rsl := `
set_stash_id("with_no_stash")
state, existed = load_state()
print(state, existed)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ } false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_SaveAndLoadState(t *testing.T) {
	rsl := `
set_stash_id("with_stash")
m = load_state()
m.print()
save_state({ "changed": true })
load_state().print()
save_state(m)
load_state().print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "somekey": "somevalue" }
{ "changed": true }
{ "somekey": "somevalue" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadStashFileExisting(t *testing.T) {
	rsl := `
set_stash_id("with_stash")
r, existed = load_stash_file("existing.txt", "didn't find")
print(existed, r.content)
r.path.split("/")[-6:].print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `true hello there!
[ "testing", "rad_test_home", "stashes", "with_stash", "files", "existing.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_LoadStashFileNotExisting(t *testing.T) {
	rsl := `
set_stash_id("with_stash")
r, existed = load_stash_file("non_existing.txt", "didn't find")
print(existed, r.content)
r.path.split("/")[-6:].print()

p = get_path(r.path)
p.base_name.print()

// clean up
p.full_path.delete_path()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `false didn't find
[ "testing", "rad_test_home", "stashes", "with_stash", "files", "non_existing.txt" ]
non_existing.txt
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Stash_WriteStashFile(t *testing.T) {
	rsl := `
set_stash_id("with_stash")

write_stash_file("bloop.txt", "hello HELLO")
r = load_stash_file("bloop.txt", "didn't find")

r.content.print()
r.path.split("/")[-6:].print()

// clean up
r.path.delete_path()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hello HELLO
[ "testing", "rad_test_home", "stashes", "with_stash", "files", "bloop.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
