package testing

import (
	"strconv"
	"strings"
	"testing"

	snap "github.com/amterp/go-snap"
	"github.com/amterp/go-snap/prompt"
)

// snapshotSuite declares rad's end-to-end script surface: a case is one
// invocation of the CLI, described by the script it runs, the arguments it gets,
// and the terminal it runs in.
//
// Cases run sequentially. The runner reports through the package-level
// stdin/stdout/stderr buffers in test_helpers.go, which the whole package shares.
var snapshotSuite = snap.Suite{
	Run: runSnapshotCase,
	Inputs: []snap.Input{
		{Name: "INPUT"},
		{Name: "ARGS", List: true},
		// RAW_ARGS is ARGS with the --color=never injection suppressed. Its
		// presence is the switch, so it is meaningful even with an empty body.
		{Name: "RAW_ARGS", List: true},
		{Name: "TERM_WIDTH"},
		prompt.KeysSection,
	},
	// Declaration order is write order, and matches the order these files were
	// already in: the run's output, then what it rendered, then how it ended.
	Outputs: []snap.Output{
		{Name: "STDOUT"},
		{Name: "STDERR"},
		prompt.FramesSection,
		{Name: "EXIT", Int: true},
	},
}

func TestSnapshots(t *testing.T) {
	snap.Run(t, "snapshots", &snapshotSuite)
}

func runSnapshotCase(t *testing.T, c *snap.Case) {
	args, rawArgs := c.List("ARGS"), c.Has("RAW_ARGS")
	if rawArgs {
		if len(args) > 0 {
			t.Fatalf("a case sets both ARGS and RAW_ARGS; they are the same channel, so use one")
		}
		args = c.List("RAW_ARGS")
	} else {
		// Snapshots assert on plain text, so color is off unless the case says
		// otherwise. RAW_ARGS opts out for cases that assert on arg validation,
		// where an injected flag would change what is being tested.
		if !hasColorFlag(args) {
			args = append(args, "--color=never")
		}
	}

	tp := NewTestParams(c.Text("INPUT"), args...)

	if w := c.Text("TERM_WIDTH"); w != "" {
		width, err := strconv.Atoi(w)
		if err != nil {
			t.Fatalf("TERM_WIDTH: %v", err)
		}
		tp.TermWidth(width)
		// Force UTF-8 mode for deterministic ellipsis output in truncation tests.
		setTerminalUtf8(t, true)
	}
	if keys := c.List(prompt.KeysSection.Name); len(keys) > 0 {
		tp.Keys(keys...)
	}

	setupAndRun(t, tp)

	c.Out("STDOUT", normalizeOutput(stdOutBuffer.String()))
	c.Out("STDERR", normalizeOutput(stdErrBuffer.String()))
	c.OutInt("EXIT", exitCode())
	c.Out(prompt.FramesSection.Name, normalizeOutput(captureFrames()))
}

func hasColorFlag(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--color") {
			return true
		}
	}
	return false
}

// exitCode reports the exit code from the last run, or 0 if nothing exited.
func exitCode() int {
	if errorOrExit.exitCode != nil {
		return *errorOrExit.exitCode
	}
	return 0
}

// normalizeOutput replaces the harness's stand-in script name with a stable
// placeholder. This is the only scrubbing the suite does; everything else that
// could vary is pinned by the fakes in test_helpers.go instead.
func normalizeOutput(output string) string {
	return strings.ReplaceAll(output, "--> TestCase:", "--> <script>:")
}
