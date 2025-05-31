package rts_test

import (
	"testing"

	"github.com/amterp/rad/rts"
)

func Test_Tree_Query_CanFindStrings(t *testing.T) {
	rslTs, _ := rts.NewRslParser()
	defer rslTs.Close()

	rsl := `a = "hello"
b = "there {1 + 1}"
if true:
	c = "world!"
`
	tree := rslTs.Parse(rsl)
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
