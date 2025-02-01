package rts

import (
	"testing"
)

func Test_CreateRts(t *testing.T) {
	rts, err := NewRts()
	if err != nil {
		t.Fatalf("NewRts() failed: %v", err)
	}
	defer rts.Close()
}

func Test_CanParse(t *testing.T) {
	rts, _ := NewRts()
	defer rts.Close()

	_, err := rts.Parse("a = 2\nprint(a)")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}
}

func Test_Tree_CanPrint(t *testing.T) {
	rts, _ := NewRts()
	defer rts.Close()

	tree, _ := rts.Parse("a = 2\nprint(a)")

	expected := "(source_file (assign left: (var_path root: (identifier)) right: (primary_expr (literal (int)))) (expr_stmt (primary_expr (call func: (identifier) args: (call_arg_list (primary_expr (var_path root: (identifier))))))))"
	if tree.String() != expected {
		t.Fatalf("String() failed: %v", tree.String())
	}
}

func Test_Tree_CanGetShebang(t *testing.T) {
	rts, _ := NewRts()
	defer rts.Close()

	rsl := `#!/usr/bin/env rsl
args:
	name string
print(name)
`
	tree, _ := rts.Parse(rsl)
	shebang, ok := tree.GetShebang()
	if !ok {
		t.Fatalf("Didn't find shebang: %v", ok)
	}
	if shebang.Src() != "#!/usr/bin/env rsl" {
		t.Fatalf("Shebang contents didn't match: %v", shebang.Src())
	}
}

func Test_Tree_CanGetFileHeader(t *testing.T) {
	rts, _ := NewRts()
	defer rts.Close()

	rsl := `#!/usr/bin/env rsl
---
These are
some file headers.
---
args:
	name string
print(name)
`
	tree, _ := rts.Parse(rsl)
	fileHeader, ok := tree.GetFileHeader()
	if !ok {
		t.Fatalf("Didn't find file header: %v", ok)
	}
	if fileHeader.Contents != "These are\nsome file headers.\n" {
		t.Fatalf("File header contents didn't match: %v", fileHeader.Contents)
	}
}

func Test_Tree_Query_CanFindStrings(t *testing.T) {
	rts, _ := NewRts()
	defer rts.Close()

	rsl := `a = "hello"
b = "there {1 + 1}"
if true:
	c = "world!"
`
	tree, _ := rts.Parse(rsl)
	nodes, err := QueryNodes[*StringNode](tree)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("Found %d nodes, expected 3", len(nodes))
	}
	if nodes[0].Src() != "\"hello\"" {
		t.Fatalf("Node 0 src didn't match: %v", nodes[0].Src())
	}
	if nodes[1].Src() != "\"there {1 + 1}\"" {
		t.Fatalf("Node 1 src didn't match: %v", nodes[1].Src())
	}
	if nodes[2].Src() != "\"world!\"" {
		t.Fatalf("Node 2 src didn't match: %v", nodes[2].Src())
	}
}

func Test_Tree_Query_CanFindMultiline(t *testing.T) {
	rts, _ := NewRts()
	defer rts.Close()

	rsl := `
a = """
This is a
multiline string
"""
a = """   
just whitespace
"""
a = """     // asd
whitespace
and comment
"""
`
	tree, _ := rts.Parse(rsl)
	nodes, err := QueryNodes[*StringNode](tree)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("Found %d nodes, expected 3", len(nodes))
	}
	if nodes[0].Contents != "This is a\nmultiline string\n" {
		t.Fatalf("Node 0 contents didn't match: %v", nodes[0].Contents)
	}
	if nodes[1].Contents != "just whitespace\n" {
		t.Fatalf("Node 1 contents didn't match: %v", nodes[1].Contents)
	}
	if nodes[2].Contents != "whitespace\nand comment\n" {
		t.Fatalf("Node 2 contents didn't match: %v", nodes[2].Contents)
	}
}
