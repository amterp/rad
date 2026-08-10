package rts_test

import (
	"testing"

	snap "github.com/amterp/go-snap"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// stSuite declares the syntax-tree surface: a Rad script in, the concrete tree
// tree-sitter produced and the AST converted from it out.
var stSuite = snap.Suite{
	Run:      runSTCase,
	Inputs:   []snap.Input{{Name: "INPUT"}},
	Outputs:  []snap.Output{{Name: "CST"}, {Name: "AST"}},
	Parallel: true,
}

func TestSTSnapshots(t *testing.T) {
	snap.Run(t, "test/st_snapshots", &stSuite)
}

func runSTCase(t *testing.T, c *snap.Case) {
	input := c.Text("INPUT")

	parser, err := rts.NewRadParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	defer parser.Close()

	tree := parser.Parse(input)
	c.Out("CST", tree.Dump())

	// A tree with parse errors has no AST to dump, and conversion can panic on
	// inputs the converter cannot represent (out-of-range numbers, say). Both
	// leave the dump empty rather than failing the case: the CST is what a
	// parser-error case exists to pin.
	astDump := ""
	if !tree.Root().HasError() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("AST conversion failed: %v", r)
				}
			}()
			astDump = rl.AstDump(rts.ConvertCST(tree.Root(), input, "test.rad"))
		}()
	}
	c.Out("AST", astDump)
}
