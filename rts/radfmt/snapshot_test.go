package radfmt_test

import (
	"testing"

	snap "github.com/amterp/go-snap"

	"github.com/amterp/rad/rts/radfmt"
)

// FmtSuite declares the formatter surface: intentionally-messy Rad source in,
// its canonical formatting out.
//
// A case marks INPUT and STDOUT `[raw]` when the rule under test is about bytes
// rather than lines - line endings, trailing whitespace, the exact newline at
// end of file. Those sections are stored as Go-quoted strings and compared
// byte-for-byte; everything else forgives a trailing newline, which a text file
// cannot pin down reliably.
//
// Exported so rules_test.go can read these files back without redeclaring the
// schema.
var FmtSuite = snap.Suite{
	Run:      runFmtCase,
	Inputs:   []snap.Input{{Name: "INPUT"}},
	Outputs:  []snap.Output{{Name: "STDOUT"}},
	Parallel: true,
}

func TestFmtSnapshots(t *testing.T) {
	snap.Run(t, "snapshots", &FmtSuite)
}

func runFmtCase(t *testing.T, c *snap.Case) {
	out, _, ok := radfmt.Format(c.Text("INPUT"))
	if !ok {
		t.Fatalf("Format returned ok=false (parse error?) for input:\n%s", c.Text("INPUT"))
	}
	c.Out("STDOUT", out)

	// Idempotence: formatting already-formatted output is a no-op.
	reformatted, changed, ok2 := radfmt.Format(out)
	if !ok2 {
		t.Fatalf("re-format returned ok=false")
	}
	if reformatted != out {
		t.Errorf("format is not idempotent:\n first: %q\nsecond: %q", out, reformatted)
	}
	if changed {
		t.Errorf("re-format reported a change (not idempotent)")
	}
}
