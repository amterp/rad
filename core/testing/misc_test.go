package testing

import (
	"strings"
	"testing"

	"github.com/amterp/rad/core"
)

func Test_Misc_Version(t *testing.T) {
	setupAndRunCode(t, "", "--version")
	assertOnlyOutput(t, stdOutBuffer, "rad "+core.Version+"\n")
	assertNoErrors(t)
}

func Test_Misc_VersionShort(t *testing.T) {
	setupAndRunCode(t, "", "-v")
	assertOnlyOutput(t, stdOutBuffer, "rad "+core.Version+"\n")
	assertNoErrors(t)
}

func Test_Misc_GlobalVersionFlagBypassesValidation(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--version", "--color=never")
	expected := "rad " + core.Version + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_GlobalCstTreeFlagBypassesValidation(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--cst-tree", "--color=never")
	// Just check that it starts with the expected tree format and doesn't error
	output := stdOutBuffer.String()
	if !strings.Contains(output, "source_file") || !strings.Contains(output, "arg_block") {
		t.Errorf("Expected CST output, got: %s", output)
	}
	assertNoErrors(t)
}

func Test_Misc_GlobalAstTreeFlagBypassesValidation(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--ast-tree", "--color=never")
	output := stdOutBuffer.String()
	if !strings.Contains(output, "SourceFile") {
		t.Errorf("Expected AST output containing 'SourceFile', got: %s", output)
	}
	assertNoErrors(t)
}

func Test_Misc_GlobalRadArgsDumpFlag(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--rad-args-dump", "--color=never")
	output := stdOutBuffer.String()
	if !strings.Contains(output, "Ra Command Dump") {
		t.Errorf("Expected Ra dump in output, got: %s", output)
	}
	assertNoErrors(t)
}

func Test_Misc_InvalidSyntax_WithCstTreeFlag(t *testing.T) {
	script := `foo = [11, 12, 13
`
	setupAndRunCode(t, script, "--cst-tree", "--color=never")
	output := stdOutBuffer.String()
	if !strings.Contains(output, "source_file") || !strings.Contains(output, "ERROR") {
		t.Errorf("Expected CST with ERROR node, got: %s", output)
	}
	assertNoErrors(t)
}

func globalFlagHelpWithout(s string) string {
	original := scriptGlobalFlagHelp
	removeLineWith := "--" + s
	lines := strings.Split(original, "\n")
	var result []string
	for _, line := range lines {
		if !strings.Contains(line, removeLineWith) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func Test_Misc_StackTraceShownInNestedFunctionError(t *testing.T) {
	script := `
fn inner():
    x = undefined_var

fn outer():
    inner()

outer()
`
	setupAndRunCode(t, script, "--color=never")
	// Get the error output before it gets reset
	output := stdErrBuffer.String()
	t.Logf("Full error output:\n%s", output)
	// Verify basic error
	if !strings.Contains(output, "RAD20028") {
		t.Errorf("Expected RAD20028 in output")
	}
	if !strings.Contains(output, "undefined_var") {
		t.Errorf("Expected 'undefined_var' in output")
	}
	// Stack trace should show nested function calls
	if !strings.Contains(output, "= stack:") {
		t.Errorf("Expected '= stack:' in error output for nested function error")
	}
	assertExitCode(t, 1)
}
