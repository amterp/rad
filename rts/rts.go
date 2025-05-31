package rts

import (
	rsl "github.com/amterp/tree-sitter-rad/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadParser struct {
	parser *ts.Parser
}

func NewRadParser() (rts *RadParser, err error) {
	parser := ts.NewParser()

	err = parser.SetLanguage(ts.NewLanguage(rsl.Language()))
	if err != nil {
		return nil, err
	}

	return &RadParser{
		parser: parser,
	}, nil
}

func (rts *RadParser) Close() {
	rts.parser.Close()
}

func (rts *RadParser) Parse(src string) *RadTree {
	tree := rts.parser.Parse([]byte(src), nil)
	return newRadTree(rts.parser, tree, src)
}
