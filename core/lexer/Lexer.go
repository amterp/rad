package lexer

import "rad/core"

type Lexer struct {
	source   *string
	start    int // index of start of the current lexeme (0 indexed)
	next     int // index of next character to be read (0 indexed)
	line     int // current line number (1 indexed)
	lineChar int // character number in the current line (1 indexed)
	tokens   []core.Token
}

func NewLexer(source *string) *Lexer {
	return &Lexer{
		source:   source,
		start:    0,
		next:     0,
		line:     1,
		lineChar: 1,
		tokens:   []core.Token{},
	}
}
