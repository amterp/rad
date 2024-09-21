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

func (b BaseToken) GetType() TokenType {
	return b.Type
}

func (b BaseToken) GetLexeme() string {
	return b.Lexeme
}

func (b BaseToken) GetCharStart() int {
	return b.CharStart
}

func (b BaseToken) GetLine() int {
	return b.Line
}

func (b BaseToken) GetCharLineStart() int {
	return b.CharLineStart
}

type StringLiteralToken struct {
	BaseToken
	Literal string
}

type IntLiteralToken struct {
	BaseToken
	Literal int64
}

type FloatLiteralToken struct {
	BaseToken
	Literal float64
}

type BoolLiteralToken struct {
	BaseToken
	Literal bool
}

type FilerHeaderToken struct {
	BaseToken
	OneLiner string
	Rest     *string
}

type ArgCommentToken struct {
	BaseToken
	Literal *string
}

type JsonPathElementToken struct {
	BaseToken
	Literal string
	// Whether the path element token is tied to an array
	IsArray bool
}

func NewToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
) Token {
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
	literal string,
) Token {
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
	literal int64,
) Token {
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

func NewFloatLiteralToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	literal float64,
) Token {
	return &FloatLiteralToken{
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
	literal bool,
) Token {
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

func NewFileHeaderToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	oneLiner string,
	rest *string,
) Token {
	return &FilerHeaderToken{
		BaseToken: BaseToken{
			Type:          tokenType,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		OneLiner: oneLiner,
		Rest:     rest,
	}
}

func NewArgCommentToken(
	tokenType TokenType,
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	comment *string,
) Token {
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

func NewJsonPathElementToken(
	lexeme string,
	charStart int,
	line int,
	charLineStart int,
	jsonPathElement string,
	isArray bool,
) Token {
	return &JsonPathElementToken{
		BaseToken: BaseToken{
			Type:          JSON_PATH_ELEMENT,
			Lexeme:        lexeme,
			CharStart:     charStart,
			Line:          line,
			CharLineStart: charLineStart,
		},
		Literal: jsonPathElement,
		IsArray: isArray,
	}
}
