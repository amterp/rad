package rts_test

import (
	"testing"

	"github.com/amterp/rts"
)

func Test_CreateRts(t *testing.T) {
	rslTs, err := rts.NewRts()
	if err != nil {
		t.Fatalf("NewRts() failed: %v", err)
	}
	defer rslTs.Close()
}

func Test_CanParse(t *testing.T) {
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	_, err := rslTs.Parse("a = 2\nprint(a)")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}
}

func Test_Tree_Sexp(t *testing.T) {
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	tree, _ := rslTs.Parse("a = 2\nprint(a)")

	expected := "(source_file (assign left: (var_path root: (identifier)) right: (primary_expr (literal (int)))) (expr_stmt (primary_expr (call func: (identifier) args: (call_arg_list (primary_expr (var_path root: (identifier))))))))"
	if tree.Sexp() != expected {
		t.Fatalf("Sexp failed: %v", tree.String())
	}
}

func Test_Tree_CanGetShebang(t *testing.T) {
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	rsl := `#!/usr/bin/env rsl
args:
	name string
print(name)
`
	tree, _ := rslTs.Parse(rsl)
	shebang, ok := tree.GetShebang()
	if !ok {
		t.Fatalf("Didn't find shebang: %v", ok)
	}
	if shebang.Src() != "#!/usr/bin/env rsl" {
		t.Fatalf("Shebang contents didn't match: <%v>", shebang.Src())
	}
}

func Test_Tree_CanGetFileHeader(t *testing.T) {
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	rsl := `#!/usr/bin/env rsl
---
These are
some file headers.
---
args:
	name string
print(name)
`
	tree, _ := rslTs.Parse(rsl)
	fileHeader, ok := tree.GetFileHeader()
	if !ok {
		t.Fatalf("Didn't find file header: %v", ok)
	}
	if fileHeader.Contents != "These are\nsome file headers.\n" {
		t.Fatalf("File header contents didn't match: <%v>", fileHeader.Contents)
	}
}

func Test_Tree_Query_CanFindStrings(t *testing.T) {
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	rsl := `a = "hello"
b = "there {1 + 1}"
if true:
	c = "world!"
`
	tree, _ := rslTs.Parse(rsl)
	nodes, err := rts.QueryNodes[*rts.StringNode](tree)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("Found %d nodes, expected 3", len(nodes))
	}
	if nodes[0].Src() != "\"hello\"" {
		t.Fatalf("Node 0 src didn't match: <%v>", nodes[0].Src())
	}
	if nodes[1].Src() != "\"there {1 + 1}\"" {
		t.Fatalf("Node 1 src didn't match: <%v>", nodes[1].Src())
	}
	if nodes[2].Src() != "\"world!\"" {
		t.Fatalf("Node 2 src didn't match: <%v>", nodes[2].Src())
	}
}
