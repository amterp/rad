package rts

import (
	rad "github.com/amterp/tree-sitter-rad/bindings/go"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadParser struct {
	parser   *ts.Parser
	language *ts.Language
}

func NewRadParser() (rts *RadParser, err error) {
	parser := ts.NewParser()
	language := ts.NewLanguage(rad.Language())

	err = parser.SetLanguage(language)
	if err != nil {
		return nil, err
	}

	return &RadParser{
		parser:   parser,
		language: language,
	}, nil
}

// Language returns the tree-sitter language for this parser. It is
// safe to use concurrently with Parse - Language objects are
// immutable and shared across all parsers built from the same
// grammar.
func (rts *RadParser) Language() *ts.Language {
	return rts.language
}

func (rts *RadParser) Close() {
	rts.parser.Close()
}

// Parse builds a fresh RadTree. The returned tree retains only the
// language pointer (immutable), not the parser - so calls into the
// tree don't race against further Parse calls on this parser.
func (rts *RadParser) Parse(src string) *RadTree {
	tree := rts.parser.Parse([]byte(src), nil)
	return newRadTree(rts.language, tree, src)
}
