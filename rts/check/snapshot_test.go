package check_test

import (
	"slices"
	"testing"

	snap "github.com/amterp/go-snap"
	tsapi "github.com/tree-sitter/go-tree-sitter"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// checkSuite declares the checker surface: a Rad script in, the deterministic
// dump of what the binder and type checker made of it out.
//
// A case opts into strict mode (advisory codes like RAD30011 surfaced as
// warnings) by putting --strict in ARGS, mirroring the real `rad check --strict`.
var checkSuite = snap.Suite{
	Run: runCheckCase,
	Inputs: []snap.Input{
		{Name: "INPUT"},
		{Name: "ARGS", List: true},
	},
	Outputs:  []snap.Output{{Name: "STDOUT"}},
	Parallel: true,
}

func TestCheckSnapshots(t *testing.T) {
	snap.Run(t, "snapshots", &checkSuite)
}

func runCheckCase(t *testing.T, c *snap.Case) {
	input := c.Text("INPUT")

	parser, err := rts.NewRadParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	defer parser.Close()

	tree := parser.Parse(input)

	// Run the full Check() pipeline (binder + type checker + AST-level checks)
	// rather than just Resolve + TypeCheck. Without the AST checks, snapshots can
	// claim "no diagnostics" while `rad check` emits real errors (return-outside-
	// fn, invalid LHS, and so on).
	file := safeConvertCST(tree.Root(), input)
	checker := check.NewCheckerWithTree(tree, parser, input, file)
	if slices.Contains(c.List("ARGS"), "--strict") {
		checker.SetStrict(true)
	}

	result, err := checker.Check()
	if err != nil {
		t.Fatalf("Check should not error: %v", err)
	}

	// Resolved and Types may be nil when the converter bailed on malformed input;
	// DumpForSnapshot tolerates that, which is what the parser-error heuristic
	// cases exercise.
	c.Out("STDOUT", check.DumpForSnapshot(file, result.Types, result.Resolved, result.Diagnostics))
}

// safeConvertCST wraps rts.ConvertCST with a panic recover, matching the
// production check pipeline (rts/check/check.go tryConvertAST). Returns nil when
// the converter bails on malformed input; the checker tolerates a nil ast.
func safeConvertCST(root *tsapi.Node, src string) (file *rl.SourceFile) {
	defer func() {
		if r := recover(); r != nil {
			file = nil
		}
	}()
	return rts.ConvertCST(root, src, "snapshot_test.rad")
}
