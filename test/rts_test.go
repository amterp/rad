package rts_test

import (
	"testing"

	"github.com/amterp/rts"
)

func Test_CreateRts(t *testing.T) {
	rslTs, err := rts.NewRslParser()
	if err != nil {
		t.Fatalf("NewRslParser() failed: %v", err)
	}
	defer rslTs.Close()
}

func Test_CanParse(t *testing.T) {
	rslTs, _ := rts.NewRslParser()
	defer rslTs.Close()
	_ = rslTs.Parse("a = 2\nprint(a)")
}

func Test_Tree_Sexp(t *testing.T) {
	rslTs, _ := rts.NewRslParser()
	defer rslTs.Close()

	tree := rslTs.Parse("a = 2\nprint(a)")

	expected := "(source_file (assign left: (var_path) right: (expr base: (primary_expr (literal (int))))) (expr base: (call arg: (expr base: (var_path)))))"
	if tree.Sexp() != expected {
		t.Fatalf("Sexp failed: %v", tree.Sexp())
	}
}

func Test_Tree_CanGetShebang(t *testing.T) {
	rslTs, _ := rts.NewRslParser()
	defer rslTs.Close()

	rsl := `#!/usr/bin/env rsl
args:
	name string
print(name)
`
	tree := rslTs.Parse(rsl)
	shebang, _ := tree.FindShebang()
	if shebang == nil {
		t.Fatalf("Didn't find shebang")
	}
	if shebang.Src() != "#!/usr/bin/env rsl" {
		t.Fatalf("Shebang contents didn't match: <%v>", shebang.Src())
	}
}

func Test_Tree_CanGetFileHeader(t *testing.T) {
	rslTs, _ := rts.NewRslParser()
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
	tree := rslTs.Parse(rsl)
	fileHeader, ok := tree.FindFileHeader()
	if !ok {
		t.Fatalf("failed to find file header")
	}
	if fileHeader.Contents != "These are\nsome file headers.\n" {
		t.Fatalf("File header contents didn't match: <%v>", fileHeader.Contents)
	}
}

func Test_Tree_CanGetArgBlock(t *testing.T) {
	rslTs, _ := rts.NewRslParser()
	defer rslTs.Close()

	rsl := `
args:
    name string
    age int = 30 # An age.

    name enum ["alice", "bob"]
    name regex "^[A-Z][a-z]$"
`
	tree := rslTs.Parse(rsl)
	argBlock, ok := tree.FindArgBlock()
	if !ok {
		t.Fatalf("failed to find arg block")
	}
	_ = argBlock
}
