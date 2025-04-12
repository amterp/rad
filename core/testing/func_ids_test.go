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

func Test_Func_nanoid(t *testing.T) {
	rsl := `
ids = [nanoid() for i in range(1000)]
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

func Test_Func_nanoid_short(t *testing.T) {
	rsl := `
ids = [nanoid(size=10) for i in range(1000)]
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

func Test_Func_nanoid_long(t *testing.T) {
	rsl := `
ids = [nanoid(size=255) for i in range(1000)]
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

func Test_Func_nanoid_ErrorsIfSize0(t *testing.T) {
	rsl := `
nanoid(size=0)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  nanoid(size=0)
  ^^^^^^^^^^^^^^ Size must be [1, 255]. Got 0.
`
	assertError(t, 1, expected)
}

func Test_Func_nanoid_ErrorsIfNegSize(t *testing.T) {
	rsl := `
nanoid(size=-10)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  nanoid(size=-10)
  ^^^^^^^^^^^^^^^^ Size must be [1, 255]. Got -10.
`
	assertError(t, 1, expected)
}

func Test_Func_nanoid_ErrorsIfSizeTooLong(t *testing.T) {
	rsl := `
nanoid(size=256)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  nanoid(size=256)
  ^^^^^^^^^^^^^^^^ Size must be [1, 255]. Got 256.
`
	assertError(t, 1, expected)
}
