package rts

import (
	rsl "github.com/amterp/tree-sitter-rsl/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslParser struct {
	parser *ts.Parser
}

func NewRslParser() (rts *RslParser, err error) {
	parser := ts.NewParser()

	err = parser.SetLanguage(ts.NewLanguage(rsl.Language()))
	if err != nil {
		return nil, err
	}

	return &RslParser{
		parser: parser,
	}, nil
}

func (rts *RslParser) Close() {
	rts.parser.Close()
}

func (rts *RslParser) Parse(src string) *RslTree {
	tree := rts.parser.Parse([]byte(src), nil)
	return newRslTree(rts.parser, tree, src)
}
