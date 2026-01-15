package testing

import "testing"

func Test_RadBlock_MapContext_PrintFullContext(t *testing.T) {
	// This test shows the full context object structure for each row
	script := `
nums = [10, 20, 30]
display:
	fields nums
	nums:
		map fn(x, ctx) ctx
`
	setupAndRunCode(t, script, "--color=never")
	expected := "nums                                                 \n" +
		`{ "idx": 0, "src": [ 10, 20, 30 ], "field": "nums" }  ` + "\n" +
		`{ "idx": 1, "src": [ 10, 20, 30 ], "field": "nums" }  ` + "\n" +
		`{ "idx": 2, "src": [ 10, 20, 30 ], "field": "nums" }  ` + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_MapContext_Src(t *testing.T) {
	script := `
nums = [10, 20, 30]
display:
	fields nums
	nums:
		map fn(x, ctx) "{x} of {ctx.src.len()}"
`
	setupAndRunCode(t, script, "--color=never")
	expected := "nums    \n10 of 3  \n20 of 3  \n30 of 3  \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_MapContext_SingleParamStillWorks(t *testing.T) {
	// Ensure backward compatibility - single param lambdas unchanged
	script := `
nums = [10, 20, 30]
display:
	fields nums
	nums:
		map fn(x) x * 2
`
	setupAndRunCode(t, script, "--color=never")
	expected := "nums \n20    \n40    \n60    \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_FilterContext_Idx(t *testing.T) {
	script := `
nums = [10, 20, 30, 40, 50]
display:
	fields nums
	nums:
		filter fn(x, ctx) ctx.idx % 2 == 0
`
	setupAndRunCode(t, script, "--color=never")
	expected := "nums \n10    \n30    \n50    \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_FilterContext_Src(t *testing.T) {
	// Filter to keep only values less than the average
	script := `
nums = [10, 20, 30, 40, 50]
display:
	fields nums
	nums:
		filter fn(x, ctx) x < sum(ctx.src) / ctx.src.len()
`
	setupAndRunCode(t, script, "--color=never")
	expected := "nums \n10    \n20    \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_MapContext_MultipleFields(t *testing.T) {
	// Same lambda applied to multiple fields should see correct field name
	script := `
col1 = [1, 2]
col2 = [10, 20]
display:
	fields col1, col2
	col1, col2:
		map fn(x, ctx) "{ctx.field}={x}"
`
	setupAndRunCode(t, script, "--color=never")
	expected := "col1    col2    \ncol1=1  col2=10  \ncol1=2  col2=20  \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_MapContext_WithFilter(t *testing.T) {
	// Context idx should be based on filtered data indices
	script := `
nums = [10, 20, 30, 40]
display:
	fields nums
	nums:
		filter fn(x) x > 15
		map fn(x, ctx) "{ctx.idx}: {x}"
`
	setupAndRunCode(t, script, "--color=never")
	// After filter: [20, 30, 40], map sees idx 0, 1, 2 (post-filter indices)
	expected := "nums  \n0: 20  \n1: 30  \n2: 40  \n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_MapContext_SrcIsImmutableSnapshot(t *testing.T) {
	// ctx.src in rad block map should be an immutable snapshot
	// We use a helper function to attempt mutation and verify original is unchanged
	script := `
fn mutate_and_return(x, ctx):
	ctx.src[0] = 999
	return x

nums = [10, 20, 30]
display:
	fields nums
	nums:
		map mutate_and_return
print(nums)
`
	setupAndRunCode(t, script, "--color=never")
	// Original nums should be unchanged
	expected := "nums \n10    \n20    \n30    \n[ 10, 20, 30 ]\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_RadBlock_FilterContext_SrcIsImmutableSnapshot(t *testing.T) {
	// ctx.src in rad block filter should be an immutable snapshot
	script := `
fn mutate_and_keep(x, ctx):
	ctx.src[0] = 999
	return true

nums = [10, 20, 30]
display:
	fields nums
	nums:
		filter mutate_and_keep
print(nums)
`
	setupAndRunCode(t, script, "--color=never")
	// Original nums should be unchanged
	expected := "nums \n10    \n20    \n30    \n[ 10, 20, 30 ]\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
