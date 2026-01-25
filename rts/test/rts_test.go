package rts_test

import (
	"testing"

	"github.com/amterp/rad/rts"
)

func Test_CreateRts(t *testing.T) {
	radParser, err := rts.NewRadParser()
	if err != nil {
		t.Fatalf("NewRadParser() failed: %v", err)
	}
	defer radParser.Close()
}

func Test_CanParse(t *testing.T) {
	radParser, _ := rts.NewRadParser()
	defer radParser.Close()
	_ = radParser.Parse("a = 2\nprint(a)")
}

func Test_Tree_Sexp(t *testing.T) {
	radParser, _ := rts.NewRadParser()
	defer radParser.Close()

	tree := radParser.Parse("a = 2\nprint(a)")

	expected := "(source_file (assign left: (var_path) right: (expr delegate: (ternary_expr delegate: (or_expr delegate: (and_expr delegate: (compare_expr delegate: (add_expr delegate: (mult_expr delegate: (unary_expr delegate: (fallback_expr delegate: (indexed_expr root: (primary_expr (literal (int)))))))))))))) (expr_stmt expr: (expr delegate: (ternary_expr delegate: (or_expr delegate: (and_expr delegate: (compare_expr delegate: (add_expr delegate: (mult_expr delegate: (unary_expr delegate: (fallback_expr delegate: (indexed_expr root: (primary_expr (call arg: (expr delegate: (ternary_expr delegate: (or_expr delegate: (and_expr delegate: (compare_expr delegate: (add_expr delegate: (mult_expr delegate: (unary_expr delegate: (fallback_expr delegate: (var_path))))))))))))))))))))))))"
	if tree.Sexp() != expected {
		t.Fatalf("Sexp failed: %v", tree.Sexp())
	}
}

func Test_Tree_CanGetShebang(t *testing.T) {
	radParser, _ := rts.NewRadParser()
	defer radParser.Close()

	script := `#!/usr/bin/env rl
args:
	name string
print(name)
`
	tree := radParser.Parse(script)
	shebang, _ := tree.FindShebang()
	if shebang == nil {
		t.Fatalf("Didn't find shebang")
	}
	if shebang.Src() != "#!/usr/bin/env rl" {
		t.Fatalf("Shebang contents didn't match: <%v>", shebang.Src())
	}
}

func Test_Tree_CanGetFileHeader(t *testing.T) {
	radParser, _ := rts.NewRadParser()
	defer radParser.Close()

	script := `#!/usr/bin/env rl
---
These are
some file headers.
---
args:
	name string
print(name)
`
	tree := radParser.Parse(script)
	fileHeader, ok := tree.FindFileHeader()
	if !ok {
		t.Fatalf("failed to find file header")
	}
	if fileHeader.Contents != "These are\nsome file headers." {
		t.Fatalf("File header contents didn't match: <%v>", fileHeader.Contents)
	}
}

func Test_Tree_CanGetArgBlock(t *testing.T) {
	radParser, _ := rts.NewRadParser()
	defer radParser.Close()

	script := `
args:
    name string
    age int = 30 # An age.

    name enum ["alice", "bob"]
    name regex "^[A-Z][a-z]$"
`
	tree := radParser.Parse(script)
	argBlock, ok := tree.FindArgBlock()
	if !ok {
		t.Fatalf("failed to find arg block")
	}
	_ = argBlock
}
