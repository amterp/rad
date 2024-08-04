package core

type TokenType string

//goland:noinspection GoCommentStart
const (
	// Single-character tokens

	LEFT_PAREN  TokenType = "LEFT_PAREN"
	RIGHT_PAREN TokenType = "RIGHT_PAREN"
	COMMA       TokenType = "COMMA"
	COLON       TokenType = "COLON"
	NEWLINE     TokenType = "NEWLINE"
	EQUAL       TokenType = "EQUAL"
	DOT         TokenType = "DOT"
	PIPE        TokenType = "PIPE"     // |
	QUESTION    TokenType = "QUESTION" // ?
	MINUS       TokenType = "MINUS"
	PLUS        TokenType = "PLUS"
	EXCLAMATION TokenType = "EXCLAMATION"
	NULL        TokenType = "NULL"
	AT          TokenType = "AT" // @
	LESS        TokenType = "LESS"
	GREATER     TokenType = "GREATER"

	// Two-character tokens

	BRACKETS      TokenType = "BRACKETS"
	EQUAL_EQUAL   TokenType = "EQUAL_EQUAL"
	NOT_EQUAL     TokenType = "NOT_EQUAL"
	LESS_EQUAL    TokenType = "LESS_EQUAL"
	GREATER_EQUAL TokenType = "GREATER_EQUAL"

	// Three-character tokens
	TRIPLE_QUOTE TokenType = "TRIPLE_QUOTE"

	// N-character tokens
	INDENT TokenType = "INDENT" // todo remains to be seen that this is correct...

	// Literals

	IDENTIFIER        TokenType = "IDENTIFIER"
	STRING_LITERAL    TokenType = "STRING_LITERAL"
	INT_LITERAL       TokenType = "INT_LITERAL"
	ARG_COMMENT       TokenType = "ARG_COMMENT"
	JSON_PATH_ELEMENT TokenType = "JSON_PATH_ELEMENT"

	// Keywords TODO might not need these, we might treat them as identifiers in lexer, then check in parser

	FOR      TokenType = "FOR"
	IN       TokenType = "IN"
	ARGS     TokenType = "ARGS"
	CHOICE   TokenType = "CHOICE"
	FROM     TokenType = "FROM"
	ON       TokenType = "ON"
	RAD      TokenType = "RAD" // todo continue to rethink this being the keyword
	OR       TokenType = "OR"
	AND      TokenType = "AND"
	NOT      TokenType = "NOT"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RESOURCE TokenType = "RESOURCE"

	// only in args block
	STRING   TokenType = "STRING"
	INT      TokenType = "INT"
	BOOL     TokenType = "BOOL"
	REQUIRES TokenType = "REQUIRES"
	ONE_OF   TokenType = "ONE_OF"
	REGEX    TokenType = "REGEX" // todo this is pretty sad to not make available for users as flag

	// only in rad block
	SORT      TokenType = "SORT"
	ASC       TokenType = "ASC"
	DESC      TokenType = "DESC"
	COLOR     TokenType = "COLOR"
	MAX_WIDTH TokenType = "MAX_WIDTH"
	UNIQ      TokenType = "UNIQ"
	QUIET     TokenType = "QUIET"
	LIMIT     TokenType = "LIMIT"
	TABLE     TokenType = "TABLE"
	DEFAULT   TokenType = "DEFAULT"
	MARKDOWN  TokenType = "MARKDOWN"

	EOF TokenType = "EOF"
)
