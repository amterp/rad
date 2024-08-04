package core

type Token interface {
	GetType() TokenType
	GetLexeme() string
	GetCharStart() int
	GetLine() int
	GetCharLineStart() int
}

type BaseToken struct {
	Type          TokenType
	Lexeme        string
	CharStart     int // from start of the source file, 0 indexed
	Line          int // line of the token, 1 indexed
	CharLineStart int // from the start of the line, 1 indexed
}

func (b *BaseToken) GetType() TokenType {
	return b.Type
}

func (b *BaseToken) GetLexeme() string {
	return b.Lexeme
}

func (b *BaseToken) GetCharStart() int {
	return b.CharStart
}

func (b *BaseToken) GetLine() int {
	return b.Line
}

func (b *BaseToken) GetCharLineStart() int {
	return b.CharLineStart
}

type StringLiteralToken struct {
	BaseToken
	Literal *string
}

type IntLiteralToken struct {
	BaseToken
	Literal *int
}

type BoolLiteralToken struct {
	BaseToken
	Literal *bool
}

type ArgCommentToken struct {
	BaseToken
	Literal *string
}

type IndentToken struct {
	BaseToken
	NumSpaces int
	NumTabs   int
}

type JsonPathElementToken struct {
	BaseToken
	Literal string
	IsArray bool
}

func NewToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int) Token {
	return &BaseToken{
		Type:          tokenType,
		Lexeme:        lexeme,
		CharStart:     charStart,
		Line:          line,
		CharLineStart: charLineStart,
	}
}

func NewStringLiteralToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	literal *string) Token {
	return &StringLiteralToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		Literal: literal,
	}
}

func NewIntLiteralToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	literal *int) Token {

	return &IntLiteralToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		Literal: literal,
	}
}

func NewBoolLiteralToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	literal *bool) Token {

	return &BoolLiteralToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		Literal: literal,
	}
}

func NewArgCommentToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	comment *string) Token {

	return &ArgCommentToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		Literal: comment,
	}
}

func NewIndentToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	numSpaces int,
	numTabs int) Token {

	return &IndentToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		NumSpaces: numSpaces,
		NumTabs:   numTabs,
	}
}

func NewJsonPathElementToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	jsonPathElement string,
	isArray bool) Token {

	return &JsonPathElementToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		Literal: jsonPathElement,
		IsArray: isArray,
	}
}
