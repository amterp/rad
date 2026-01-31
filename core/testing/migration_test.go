package testing

import "testing"

// ===== v0.7 Migration: For-loop index syntax =====

func Test_Migration_V07_ForLoopIndex_ShowsHelpfulError(t *testing.T) {
	// When user uses old syntax with 'idx' as first variable, show helpful hint
	script := `
for idx, item in [1, 2, 3]:
	print(idx, item)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20033", "Cannot unpack \"int\" into 2 values")
}

func Test_Migration_V07_ForLoopIndex_ThreeVars_ShowsHelpfulError(t *testing.T) {
	// Migration hint also works for 3+ variables
	script := `
for idx, item, extra in [1, 2, 3]:
	print(idx, item, extra)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20033", "Cannot unpack \"int\" into 3 values")
}

func Test_Migration_V07_ForLoopIndex_Underscore_ShowsHelpfulError(t *testing.T) {
	// Migration hint triggers for underscore (common pattern for discarding old auto-index)
	script := `
for _, item in [1, 2, 3]:
	print(item)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20033", "Cannot unpack \"int\" into 2 values")
}

// ===== v0.8 Migration: get_default removed =====

func Test_Migration_V08_GetDefault_ShowsHelpfulError(t *testing.T) {
	script := `
m = {"a": 1}
get_default(m, "b", 0)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "get_default was removed", "??")
}

func Test_Migration_V08_GetDefault_Ufcs_ShowsHelpfulError(t *testing.T) {
	script := `
m = {"a": 1}
m.get_default("b", 0)
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "get_default was removed", "??")
}
