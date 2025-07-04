package testing

import "testing"

func TestDelete_CanDeleteVariable(t *testing.T) {
	script := `
a = "alice"
b = "bob"
del a
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "b=\"bob\"\n")
	assertNoErrors(t)
}

func TestDelete_CanDeleteArray(t *testing.T) {
	script := `
a = [1, 2]
b = [3, 4]
del a
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "b=(3 4)\n")
	assertNoErrors(t)
}

func TestDelete_CanDeleteArrayEntry(t *testing.T) {
	script := `
a = [0, 10, 20, 30, 40]
del a[1]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 20, 30, 40 ]\n")
	assertNoErrors(t)
}

func TestDelete_CanDeleteNestedArrayEntry(t *testing.T) {
	script := `
a = [0, [10, [20, 30]], 40]
del a[1][1][0]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, [ 10, [ 30 ] ], 40 ]\n")
	assertNoErrors(t)
}

func TestDelete_CanDeleteMultipleThingsAtOnce(t *testing.T) {
	script := `
a = [0, 10, 20, 30, 40]
b = [0, -10, -20, -30, -40]
del a[1], b[2]
print(a)
print(b)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 20, 30, 40 ]\n[ 0, -10, -30, -40 ]\n")
	assertNoErrors(t)
}

func TestDelete_ArrayMultiDeleteAreInOrder(t *testing.T) {
	script := `
a = [0, 10, 20, 30, 40]
del a[1], a[1]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 30, 40 ]\n")
	assertNoErrors(t)
}

func TestDelete_ArrayDeletingLastEntryLeavesEmptyArray(t *testing.T) {
	script := `
a = [0]
del a[0]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ]\n")
	assertNoErrors(t)
}

func TestDelete_CanDeleteMapEntry(t *testing.T) {
	script := `
a = { "alice": 35, "bob": "bar", "charlie": [1, "hi"] }
del a["bob"]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 35, \"charlie\": [ 1, \"hi\" ] }\n")
	assertNoErrors(t)
}

func TestDelete_CanDeleteNestedMapEntry(t *testing.T) {
	script := `
a = { "alice": 35, "bob": { "car": "toyota", "shoes": "brooks" }, "charlie": [1, "hi"] }
del a["bob"]["shoes"]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(
		t,
		stdOutBuffer,
		"{ \"alice\": 35, \"bob\": { \"car\": \"toyota\" }, \"charlie\": [ 1, \"hi\" ] }\n",
	)
	assertNoErrors(t)
}

func TestDelete_CanDeleteArrayEntryNestedInMapEntry(t *testing.T) {
	script := `
a = { "alice": 35, "bob": { "car": "toyota", "ids": [10, 20, 30] }, "charlie": [1, "hi"] }
del a["bob"]["ids"][1]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(
		t,
		stdOutBuffer,
		"{ \"alice\": 35, \"bob\": { \"car\": \"toyota\", \"ids\": [ 10, 30 ] }, \"charlie\": [ 1, \"hi\" ] }\n",
	)
	assertNoErrors(t)
}

func TestDelete_CanDeleteFromListWithSlice(t *testing.T) {
	script := `
a = [0, 10, 20, 30, 40]
del a[1:3]
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 30, 40 ]\n")
	assertNoErrors(t)
}
