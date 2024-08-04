package core

type Token struct {
	Type          TokenType
	Lexeme        string  // todo pointer?
	Literal       *string // todo do I want to be using inheritance here to give specific types?
	CharStart     int     // from start of the source file, 0 indexed
	Line          int     // line of the token, 1 indexed
	CharLineStart int     // from the start of the line, 1 indexed
}
