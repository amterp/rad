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

func Test_Func_gen_nanoid(t *testing.T) {
	rsl := `
ids = [gen_nanoid() for i in range(1000)]
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

func Test_Func_gen_nanoid_short(t *testing.T) {
	rsl := `
ids = [gen_nanoid(size=10) for i in range(1000)]
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

func Test_Func_gen_nanoid_long(t *testing.T) {
	rsl := `
ids = [gen_nanoid(size=255) for i in range(1000)]
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

func Test_Func_gen_nanoid_ErrorsIfSize0(t *testing.T) {
	rsl := `
gen_nanoid(size=0)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  gen_nanoid(size=0)
  ^^^^^^^^^^^^^^^^^^ Size must be [1, 255]. Got 0.
`
	assertError(t, 1, expected)
}

func Test_Func_gen_nanoid_ErrorsIfNegSize(t *testing.T) {
	rsl := `
gen_nanoid(size=-10)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  gen_nanoid(size=-10)
  ^^^^^^^^^^^^^^^^^^^^ Size must be [1, 255]. Got -10.
`
	assertError(t, 1, expected)
}

func Test_Func_gen_nanoid_ErrorsIfSizeTooLong(t *testing.T) {
	rsl := `
gen_nanoid(size=256)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  gen_nanoid(size=256)
  ^^^^^^^^^^^^^^^^^^^^ Size must be [1, 255]. Got 256.
`
	assertError(t, 1, expected)
}

func Test_Func_gen_stid(t *testing.T) {
	rsl := `
ids = [gen_stid() for i in range(1000)]
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

func Test_Func_gen_stid_named_args(t *testing.T) {
	rsl := `
ids = [gen_stid(alphabet="1234567890abcdef", num_random_chars=8, time_granularity=round(1e3*60*60)) for i in range(1000)]
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

func Test_Func_gen_stid_ErrorsOnEmptyAlphabet(t *testing.T) {
	rsl := `
gen_stid(alphabet="")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  gen_stid(alphabet="")
  ^^^^^^^^^^^^^^^^^^^^^
  Error creating STID generator: alphabet must contain at least 2 characters
`
	assertError(t, 1, expected)
}

func Test_Func_gen_stid_ErrorsOnNegNumRandomChars(t *testing.T) {
	rsl := `
gen_stid(num_random_chars=-1)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:27

  gen_stid(num_random_chars=-1)
                            ^^ Number of random chars must be non-negative, got -1
`
	assertError(t, 1, expected)
}
