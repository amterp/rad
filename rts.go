package rts

import (
	rsl "github.com/amterp/tree-sitter-rsl/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslTreeSitter struct {
	parser *ts.Parser
}

func NewRts() (rts *RslTreeSitter, err error) {
	parser := ts.NewParser()

	err = parser.SetLanguage(ts.NewLanguage(rsl.Language()))
	if err != nil {
		return nil, err
	}

	return &RslTreeSitter{
		parser: parser,
	}, nil
}

func (rts *RslTreeSitter) Close() {
	rts.parser.Close()
}

func (rts *RslTreeSitter) Parse(input string) (root *RtsTree, err error) {
	tree := rts.parser.Parse([]byte(input), nil)
	return NewRtsTree(tree, rts.parser, input), nil
}
