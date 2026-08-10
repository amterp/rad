package check_test

import (
	"testing"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// diagnosticsWithCode runs the full Check() pipeline - RAD40018 is an
// AST-level check, so Resolve + TypeCheck alone wouldn't see it - and returns
// every diagnostic carrying the given code.
func diagnosticsWithCode(t *testing.T, src string, code rl.Error) []check.Diagnostic {
	t.Helper()

	parser, err := rts.NewRadParser()
	require.NoError(t, err)
	defer parser.Close()

	tree := parser.Parse(src)
	file := safeConvertCST(tree.Root(), src)
	result, err := check.NewCheckerWithTree(tree, parser, src, file).Check()
	require.NoError(t, err)

	var out []check.Diagnostic
	for _, d := range result.Diagnostics {
		if d.Code != nil && *d.Code == code {
			out = append(out, d)
		}
	}
	return out
}

// hasShellCaptureWarning reports whether src trips RAD40018.
func hasShellCaptureWarning(t *testing.T, src string) bool {
	t.Helper()
	return len(diagnosticsWithCode(t, src, rl.ErrMisleadingShellCaptureName)) > 0
}

func countDiagnosticCode(t *testing.T, src string, code rl.Error) int {
	t.Helper()
	return len(diagnosticsWithCode(t, src, code))
}

func firstDiagnosticMessage(t *testing.T, src string, code rl.Error) string {
	t.Helper()
	diags := diagnosticsWithCode(t, src, code)
	require.NotEmpty(t, diags, "expected at least one %s diagnostic", code)
	return diags[0].Message
}

func TestShellCaptureNames_Flagged(t *testing.T) {
	// Every case here binds a variable to a stream its name contradicts.
	cases := map[string]string{
		"reserved name in stdout slot":     "code, output = $`cmd`\n",
		"reserved name in stderr slot":     "out, stdout = $`cmd`\n",
		"synonym alone takes stdout":       "exit_code = $`cmd`\n",
		"synonym in the stderr slot":       "_, output = $`cmd`\n",
		"old-order discard-then-capture":   "_, cpu = $`cmd`\n",
		"old-order three-target":           "_, pwd_output, _ = $`cmd`\n",
		"old-order skips leading discards": "_, _, tail = $`cmd`\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			assert.True(t, hasShellCaptureWarning(t, src), "expected RAD40018 for: %s", src)
		})
	}
}

func TestShellCaptureNames_Quiet(t *testing.T) {
	// These are all correct under the (stdout, stderr, code) order. A warning
	// on any of them would be firing on the forms the change recommends.
	cases := map[string]string{
		"recommended two-target":     "out, err = $`cmd`\n",
		"recommended single":         "out = $`cmd`\n",
		"swallow stdout take stderr": "_, err = $`cmd`\n",
		"named assignment":           "stdout, code = $`cmd`\n",
		"named single code":          "code = $`cmd`\n",
		"named all three reversed":   "code, stderr, stdout = $`cmd`\n",
		"no signal in any name":      "a, b = $`cmd`\n",
		"bare invocation":            "$`cmd`\n",
		"discard everything":         "_, _ = $`cmd`\n",
		"discard all three":          "_, _, _ = $`cmd`\n",
		// `status` reads as "exit status" but is just as often the output of
		// `git status`. Ambiguous names stay out of the synonym table.
		"ambiguous status name": "status = $`git status --porcelain`\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			assert.False(t, hasShellCaptureWarning(t, src), "unexpected RAD40018 for: %s", src)
		})
	}
}

func TestShellCaptureNames_OneDiagnosticPerStatement(t *testing.T) {
	// Both names are wrong here, for the same single reason. Reporting twice
	// would just be noise.
	count := countDiagnosticCode(t, "code, output = $`cmd`\n", rl.ErrMisleadingShellCaptureName)
	assert.Equal(t, 1, count)
}

func TestShellCaptureNames_ReservedMessageOmitsMigrationLink(t *testing.T) {
	// The reserved-name half is permanent - it describes a confusion that
	// exists on its own terms, not a version change - so it must not point at
	// the migration guide.
	msg := firstDiagnosticMessage(t, "code, output = $`cmd`\n", rl.ErrMisleadingShellCaptureName)
	assert.NotContains(t, msg, "v0.12")
}

func TestShellCaptureNames_SynonymMessageExplainsTheFlip(t *testing.T) {
	msg := firstDiagnosticMessage(t, "_, version = $`cmd`\n", rl.ErrMisleadingShellCaptureName)
	assert.Contains(t, msg, "Before v0.12")
}
