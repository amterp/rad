package rts

import (
	"testing"

	"github.com/amterp/rts"
)

func Test_Parsing_Str_Escaping(t *testing.T) {
	t.Skip("More complex than looking at Contents, revisit")
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	rsl := `
//a = "\\"
//b = "\n"
a = "this is \\ a string \n blah \ blahh"
`
	tree, _ := rslTs.Parse(rsl)
	nodes, err := rts.QueryNodes[*rts.StringNode](tree)
	if err != nil {
		t.Fatalf("Escaping failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("Found %d nodes, expected 2", len(nodes))
	}
	if nodes[0].RawLexeme != `\` {
		t.Fatalf("Node 0 contents didn't match: <%v>", nodes[0].RawLexeme)
	}
	if nodes[1].RawLexeme != "\n" {
		t.Fatalf("Node 1 contents didn't match: <%v>", nodes[1].RawLexeme)
	}
}

func Test_Parsing_Str_RawEscaping(t *testing.T) {
	rslTs, _ := rts.NewRts()
	defer rslTs.Close()

	rsl := `
a = r"\\"
b = r"\n"
c = r"\"
`
	tree, _ := rslTs.Parse(rsl)
	nodes, err := rts.QueryNodes[*rts.StringNode](tree)
	if err != nil {
		t.Fatalf("Escaping failed: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("Found %d nodes, expected 3", len(nodes))
	}
	if nodes[0].RawLexeme != `\\` {
		t.Fatalf("Node 0 contents didn't match: <%v>", nodes[0].RawLexeme)
	}
	if nodes[1].RawLexeme != `\n` {
		t.Fatalf("Node 1 contents didn't match: <%v>", nodes[1].RawLexeme)
	}
	if nodes[2].RawLexeme != `\` {
		t.Fatalf("Node 2 contents didn't match: <%v>", nodes[2].RawLexeme)
	}
}
