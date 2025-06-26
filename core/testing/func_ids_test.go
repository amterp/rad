package testing

import (
	"testing"
)

func Test_Func_uuid_v4(t *testing.T) {
	script := `
ids = [uuid_v4() for i in range(1000)]
uniq = ids.unique().len()
if ids.len() != uniq:
	print("IDs are not unique, got {uniq}")
else:
	print("IDs are unique")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "IDs are unique\n")
	assertNoErrors(t)
}

func Test_Func_uuid_v7(t *testing.T) {
	script := `
ids = [uuid_v7() for i in range(1000)]
uniq = ids.unique().len()
if ids.len() != uniq:
	print("IDs are not unique, got {uniq}")
else:
	print("IDs are unique")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "IDs are unique\n")
	assertNoErrors(t)
}

func Test_Func_gen_fid(t *testing.T) {
	script := `
ids = [gen_fid() for i in range(1000)]
uniq = ids.unique().len()
if ids.len() != uniq:
	print("IDs are not unique, got {uniq}")
else:
	print("IDs are unique")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "IDs are unique\n")
	assertNoErrors(t)
}

func Test_Func_gen_fid_named_args(t *testing.T) {
	script := `
ids = [gen_fid(alphabet="1234567890abcdef", num_random_chars=8, tick_size_ms=round(1e3*60*60)) for i in range(1000)]
uniq = ids.unique().len()
if ids.len() != uniq:
	print("IDs are not unique, got {uniq}")
else:
	print("IDs are unique")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "IDs are unique\n")
	assertNoErrors(t)
}

func Test_Func_gen_fid_ErrorsOnEmptyAlphabet(t *testing.T) {
	script := `
gen_fid(alphabet="")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  gen_fid(alphabet="")
  ^^^^^^^^^^^^^^^^^^^^
  Error creating FID generator: alphabet must contain at least 2 characters
`
	assertError(t, 1, expected)
}

func Test_Func_gen_fid_ErrorsOnNegNumRandomChars(t *testing.T) {
	script := `
gen_fid(num_random_chars=-1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  gen_fid(num_random_chars=-1)
  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^ Number of random chars must be non-negative, got -1
`
	assertError(t, 1, expected)
}
