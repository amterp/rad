package testing

import "testing"

func Test_Func_uuid_v4(t *testing.T) {
	rsl := `
ids = [uuid_v4() for i in range(1000)]
uniq = ids.unique().len()
if ids.len() != uniq:
	print("IDs are not unique, got {uniq}")
else:
	print("IDs are unique")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "IDs are unique\n")
	assertNoErrors(t)
}

func Test_Func_uuid_v7(t *testing.T) {
	rsl := `
ids = [uuid_v7() for i in range(1000)]
uniq = ids.unique().len()
if ids.len() != uniq:
	print("IDs are not unique, got {uniq}")
else:
	print("IDs are unique")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "IDs are unique\n")
	assertNoErrors(t)
}
