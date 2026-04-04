package testing

import (
	"strings"
	"testing"
)

func Test_MapKeys_IntAndStringAreDistinct(t *testing.T) {
	script := `
m = {}
m[1] = "int"
m["1"] = "str"
print(m.len())
print(m[1])
print(m["1"])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\nint\nstr\n")
	assertNoErrors(t)
}

func Test_MapKeys_FloatAndIntShare(t *testing.T) {
	script := `
m = {}
m[1] = "int"
m[1.0] = "float"
print(m.len())
print(m[1])
print(m[1.0])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\nfloat\nfloat\n")
	assertNoErrors(t)
}

func Test_MapKeys_BoolAndStringAreDistinct(t *testing.T) {
	script := `
m = {}
m[true] = "bool"
m["true"] = "str"
print(m.len())
print(m[true])
print(m["true"])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\nbool\nstr\n")
	assertNoErrors(t)
}

func Test_MapKeys_FloatAndStringAreDistinct(t *testing.T) {
	script := `
m = {}
m[1.5] = "float"
m["1.5"] = "str"
print(m.len())
print(m[1.5])
print(m["1.5"])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\nfloat\nstr\n")
	assertNoErrors(t)
}

func Test_MapKeys_BoolAndIntAreDistinct(t *testing.T) {
	script := `
m = {}
m[true] = "bool"
m[1] = "int"
print(m.len())
print(m[true])
print(m[1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\nbool\nint\n")
	assertNoErrors(t)
}

func Test_MapKeys_LiteralMixedTypes(t *testing.T) {
	script := `
m = {1: "int", "1": "str"}
print(m.len())
print(m[1])
print(m["1"])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\nint\nstr\n")
	assertNoErrors(t)
}

func Test_MapKeys_InOperatorRespectsType(t *testing.T) {
	script := `
m = {42: "int"}
print(42 in m)
print("42" in m)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\n")
	assertNoErrors(t)
}

func Test_MapKeys_DeleteRespectsType(t *testing.T) {
	script := `
m = {1: "int", "1": "str"}
del m[1]
print(m.len())
print(m["1"])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\nstr\n")
	assertNoErrors(t)
}

func Test_MapKeys_NullKeyAssignError(t *testing.T) {
	script := `
m = {}
m[null] = "value"
`
	setupAndRunCode(t, script, "--color=never")
	assertExitCode(t, 1)
	output := stdErrBuffer.String()
	if !strings.Contains(output, "null") {
		t.Errorf("Expected error mentioning null, got: %s", output)
	}
}

func Test_MapKeys_NullKeyLiteralError(t *testing.T) {
	script := `m = {null: "value"}`
	setupAndRunCode(t, script, "--color=never")
	assertExitCode(t, 1)
	output := stdErrBuffer.String()
	if !strings.Contains(output, "null") {
		t.Errorf("Expected error mentioning null, got: %s", output)
	}
}

func Test_MapKeys_NullKeyReadError(t *testing.T) {
	script := `
m = {"a": 1}
x = m[null]
`
	setupAndRunCode(t, script, "--color=never")
	assertExitCode(t, 1)
	output := stdErrBuffer.String()
	if !strings.Contains(output, "null") {
		t.Errorf("Expected error mentioning null, got: %s", output)
	}
}

func Test_MapKeys_NullInMapReturnsFalse(t *testing.T) {
	script := `
m = {"a": 1}
print(null in m)
print(null not in m)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "false\ntrue\n")
	assertNoErrors(t)
}
