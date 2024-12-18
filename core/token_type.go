package core

type TokenType string

//goland:noinspection GoCommentStart
const (
	// Single-character tokens

	LEFT_PAREN    TokenType = "LEFT_PAREN" // (
	RIGHT_PAREN   TokenType = "RIGHT_PAREN"
	LEFT_BRACKET  TokenType = "LEFT_BRACKET" // [
	RIGHT_BRACKET TokenType = "RIGHT_BRACKET"
	LEFT_BRACE    TokenType = "LEFT_BRACE" // {
	RIGHT_BRACE   TokenType = "RIGHT_BRACE"
	COMMA         TokenType = "COMMA"
	COLON         TokenType = "COLON"
	NEWLINE       TokenType = "NEWLINE"
	EQUAL         TokenType = "EQUAL"
	DOT           TokenType = "DOT"
	PIPE          TokenType = "PIPE"     // |
	QUESTION      TokenType = "QUESTION" // ?
	MINUS         TokenType = "MINUS"
	PLUS          TokenType = "PLUS"
	EXCLAMATION   TokenType = "EXCLAMATION"
	AT            TokenType = "AT"      // @
	LESS          TokenType = "LESS"    // <
	GREATER       TokenType = "GREATER" // >
	SLASH         TokenType = "SLASH"   // /
	STAR          TokenType = "STAR"    // *
	DOLLAR        TokenType = "DOLLAR"  // $

	// Two-character tokens

	BRACKETS      TokenType = "BRACKETS"
	EQUAL_EQUAL   TokenType = "EQUAL_EQUAL"
	NOT_EQUAL     TokenType = "NOT_EQUAL"
	LESS_EQUAL    TokenType = "LESS_EQUAL"
	GREATER_EQUAL TokenType = "GREATER_EQUAL"
	PLUS_EQUAL    TokenType = "PLUS_EQUAL"
	MINUS_EQUAL   TokenType = "MINUS_EQUAL"
	STAR_EQUAL    TokenType = "STAR_EQUAL"
	SLASH_EQUAL   TokenType = "SLASH_EQUAL"
	ARROW         TokenType = "ARROW" // ->

	// N-character tokens
	INDENT TokenType = "INDENT"
	DEDENT TokenType = "DEDENT"

	// Literals

	IDENTIFIER     TokenType = "IDENTIFIER"
	STRING_LITERAL TokenType = "STRING_LITERAL"
	INT_LITERAL    TokenType = "INT_LITERAL"
	FLOAT_LITERAL  TokenType = "FLOAT_LITERAL"
	BOOL_LITERAL   TokenType = "BOOL_LITERAL"
	FILE_HEADER    TokenType = "FILE_HEADER"
	ARG_COMMENT    TokenType = "ARG_COMMENT"

	// Keywords

	JSON     TokenType = "JSON"
	FOR      TokenType = "FOR"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	IN       TokenType = "IN"
	ARGS     TokenType = "ARGS"
	SWITCH   TokenType = "SWITCH"
	CASE     TokenType = "CASE"
	RAD      TokenType = "RAD"
	REQUEST  TokenType = "REQUEST"
	DISPLAY  TokenType = "DISPLAY"
	OR       TokenType = "OR"
	AND      TokenType = "AND"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	DELETE   TokenType = "DELETE"
	NOT      TokenType = "NOT"
	NOT_IN   TokenType = "NOT_IN"
	UNSAFE   TokenType = "UNSAFE"
	QUIET    TokenType = "QUIET"
	FAIL     TokenType = "FAIL"
	RECOVER  TokenType = "RECOVER"
	DEFER    TokenType = "DEFER"
	ERRDEFER TokenType = "ERRDEFER"
	MAP      TokenType = "MAP"

	// only in Args block
	STRING   TokenType = "STRING"
	INT      TokenType = "INT"
	FLOAT    TokenType = "FLOAT"
	BOOL     TokenType = "BOOL"
	ARRAY    TokenType = "ARRAY"
	REQUIRES TokenType = "REQUIRES"
	ONE_OF   TokenType = "ONE_OF"
	REGEX    TokenType = "REGEX"
	ENUM     TokenType = "ENUM"

	// only in rad block
	FIELDS   TokenType = "FIELDS"
	SORT     TokenType = "SORT"
	ASC      TokenType = "ASC"
	DESC     TokenType = "DESC"
	COLOR    TokenType = "COLOR"
	UNIQ     TokenType = "UNIQ"
	LIMIT    TokenType = "LIMIT"
	TABLE    TokenType = "TABLE"
	DEFAULT  TokenType = "DEFAULT"
	MARKDOWN TokenType = "MARKDOWN"

	EOF TokenType = "EOF"
)
