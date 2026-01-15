package testing

import "testing"

func Test_For_WithContext_PrintFullContext(t *testing.T) {
	// Print the full context object to verify structure
	script := `
items = ["a", "b", "c"]
for item in items with loop:
	print(loop)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "idx": 0, "src": [ "a", "b", "c" ] }
{ "idx": 1, "src": [ "a", "b", "c" ] }
{ "idx": 2, "src": [ "a", "b", "c" ] }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_WithContext_IdxAndItem(t *testing.T) {
	script := `
items = ["a", "b", "c"]
for item in items with loop:
	print(loop.idx, item)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 a\n1 b\n2 c\n")
	assertNoErrors(t)
}

func Test_For_WithContext_AccessSrc(t *testing.T) {
	// Use loop.src to access original collection for lookahead
	script := `
items = ["a", "b", "c"]
for item in items with loop:
	if loop.idx < loop.src.len() - 1:
		print(loop.src[loop.idx + 1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "b\nc\n")
	assertNoErrors(t)
}

func Test_For_WithContext_Map(t *testing.T) {
	script := `
m = {"a": 1, "b": 2, "c": 3}
for key in m with loop:
	print(loop.idx, key)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 a\n1 b\n2 c\n")
	assertNoErrors(t)
}

func Test_For_WithContext_MapKeyValue(t *testing.T) {
	script := `
m = {"a": 1, "b": 2}
for key, value in m with loop:
	print(loop.idx, key, value)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 a 1\n1 b 2\n")
	assertNoErrors(t)
}

func Test_For_WithContext_ListComprehension(t *testing.T) {
	script := `
items = ["a", "b", "c"]
result = ["{loop.idx}:{item}" for item in items with loop]
print(result)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"0:a\", \"1:b\", \"2:c\" ]\n")
	assertNoErrors(t)
}

func Test_For_WithContext_String(t *testing.T) {
	script := `
s = "abc"
for char in s with loop:
	print(loop.idx, char)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 a\n1 b\n2 c\n")
	assertNoErrors(t)
}

func Test_For_WithContext_CustomName(t *testing.T) {
	// User can use any name for context variable
	script := `
items = [1, 2, 3]
for item in items with meta:
	print(meta.idx)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0\n1\n2\n")
	assertNoErrors(t)
}

func Test_For_Unpacking_NoAutoIndex(t *testing.T) {
	script := `
pairs = [["a", 1], ["b", 2]]
for name, val in pairs:
	print(name, val)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a 1\nb 2\n")
	assertNoErrors(t)
}

func Test_For_Unpacking_WithContext(t *testing.T) {
	script := `
pairs = [["a", 1], ["b", 2]]
for name, val in pairs with loop:
	print(loop.idx, name, val)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 a 1\n1 b 2\n")
	assertNoErrors(t)
}

func Test_For_Context_SrcIsImmutableSnapshot(t *testing.T) {
	// Modifying loop.src should not affect the original list
	script := `
items = [10, 20, 30]
for item in items with loop:
	loop.src[0] = 999
print(items)
`
	setupAndRunCode(t, script, "--color=never")
	// Original items should be unchanged
	assertOnlyOutput(t, stdOutBuffer, "[ 10, 20, 30 ]\n")
	assertNoErrors(t)
}

func Test_For_Context_MapSrcIsImmutableSnapshot(t *testing.T) {
	// Modifying loop.src should not affect the original map
	script := `
m = {"a": 1, "b": 2}
for key in m with loop:
	loop.src["a"] = 999
print(m)
`
	setupAndRunCode(t, script, "--color=never")
	// Original map should be unchanged
	assertOnlyOutput(t, stdOutBuffer, `{ "a": 1, "b": 2 }`+"\n")
	assertNoErrors(t)
}
