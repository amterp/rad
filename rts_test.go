package rts

import (
	"fmt"
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
	fmt.Println(ok)
	fmt.Println(shebang.Src)
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
	fmt.Println(ok)
	fmt.Println("||" + fileHeader.Src + "||")
	fmt.Println("||" + fileHeader.Contents + "||")
}
