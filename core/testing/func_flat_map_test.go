package testing

import "testing"

// === List without function - all elements must be lists ===

func Test_Func_FlatMap_ListNoFn_NestedLists(t *testing.T) {
	script := `
a = [[1, 2], [3, 4]]
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2, 3, 4 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListNoFn_OneLevelOnly(t *testing.T) {
	script := `
a = [[[1, 2]], [[3, 4]]]
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ [ 1, 2 ], [ 3, 4 ] ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListNoFn_EmptyList(t *testing.T) {
	script := `
a = []
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListNoFn_EmptyNestedLists(t *testing.T) {
	script := `
a = [[], [1], []]
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListNoFn_ListOfStringLists(t *testing.T) {
	script := `
a = [["a", "b"], ["c", "d"]]
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "a", "b", "c", "d" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// === List without function - error cases ===

func Test_Func_FlatMap_ListNoFn_ErrorOnMixedElements(t *testing.T) {
	script := `
a = [1, [2, 3], 4]
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "flat_map requires all elements to be lists, but element at index 0 is int")
}

func Test_Func_FlatMap_ListNoFn_ErrorOnNonLists(t *testing.T) {
	script := `
a = [1, 2, 3]
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "flat_map requires all elements to be lists, but element at index 0 is int")
}

// === List with function - must return lists ===

func Test_Func_FlatMap_ListWithFn_Split(t *testing.T) {
	script := `
a = ["a-b", "c-d"]
a.flat_map(fn(e) e.split("-")).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "a", "b", "c", "d" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListWithFn_Duplicate(t *testing.T) {
	script := `
a = [1, 2]
a.flat_map(fn(x) [x, x * 10]).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 10, 2, 20 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListWithFn_Range(t *testing.T) {
	script := `
a = [1, 2, 3]
a.flat_map(fn(x) range(x)).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 0, 0, 1, 0, 1, 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListWithFn_NamedFunction(t *testing.T) {
	script := `
a = ["a-b", "c-d"]
splitter = fn(e) e.split("-")
a.flat_map(splitter).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "a", "b", "c", "d" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ListWithFn_EmptyListResults(t *testing.T) {
	script := `
a = [1, 2, 3]
a.flat_map(fn(x) []).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// === List with function - error cases ===

func Test_Func_FlatMap_ListWithFn_ErrorOnNonListResult(t *testing.T) {
	script := `
a = [1, 2, 3]
a.flat_map(fn(x) x * 2).print()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "flat_map function must return a list, but returned int for element at index 0")
}

// === Map - requires function ===

func Test_Func_FlatMap_MapNoFn_RequiresFunction(t *testing.T) {
	script := `
a = { "a": [1, 2], "b": [3, 4] }
a.flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "flat_map on maps requires a function argument")
}

func Test_Func_FlatMap_MapWithFn_ExtractValues(t *testing.T) {
	script := `
a = { "a": [1, 2], "b": [3, 4] }
a.flat_map(fn(k, v) v).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2, 3, 4 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_MapWithFn_CreatePairs(t *testing.T) {
	script := `
a = { "a": 1, "b": 2 }
a.flat_map(fn(k, v) [k, v]).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "a", 1, "b", 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_MapWithFn_TransformThenFlatten(t *testing.T) {
	script := `
a = { "a": "x-y", "b": "z-w" }
a.flat_map(fn(k, v) v.split("-")).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "x", "y", "z", "w" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// === Map with function - error cases ===

func Test_Func_FlatMap_MapWithFn_ErrorOnNonListResult(t *testing.T) {
	script := `
a = { "a": 1, "b": 2 }
a.flat_map(fn(k, v) v * 2).print()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "flat_map function must return a list, but returned int for key a")
}

// === Chaining ===

func Test_Func_FlatMap_ChainWithFilter(t *testing.T) {
	script := `
a = [[1, 2], [3, 4]]
a.flat_map().filter(fn(x) x > 2).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 3, 4 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ChainWithMap(t *testing.T) {
	script := `
a = [[1, 2], [3, 4]]
a.flat_map().map(fn(x) x * 2).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 2, 4, 6, 8 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_ChainMultiple(t *testing.T) {
	script := `
a = ["a-b", "c-d"]
a.flat_map(fn(e) e.split("-")).map(fn(x) x.upper()).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "A", "B", "C", "D" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// === Method vs Function syntax ===

func Test_Func_FlatMap_MethodSyntax(t *testing.T) {
	script := `
[[1], [2]].flat_map().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_FunctionSyntax(t *testing.T) {
	script := `
flat_map([[1], [2]]).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_FlatMap_FunctionSyntaxWithFn(t *testing.T) {
	script := `
flat_map([1, 2], fn(x) [x, x * 10]).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 10, 2, 20 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
