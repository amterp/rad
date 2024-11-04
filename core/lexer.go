package core

import (
	"fmt"
	"strconv"
	"strings"
)

type Lexer struct {
	source                 []rune
	start                  int   // index of start of the current lexeme (0 indexed)
	next                   int   // index of next character to be read (0 indexed)
	lineIndex              int   // current line number (1 indexed)
	lineCharIndex          int   // index of latest parsed char in the current line (1 indexed)
	indentStack            []int // stack of indents to, to emit indent/dedent tokens
	userUsingSpacesForTabs *bool // nil until we see the first case of a space indent
	inStringStarter        *rune // the character that started the current string literal we're in, if we are in one
	inStringStartIndex     int   // index of the start of the current string literal we're in
	escaping               bool  // true if we are currently escaping a character in a string using \
	Tokens                 []Token
}

func NewLexer(source string) *Lexer {
	runes := []rune(source)
	return &Lexer{
		source:                 runes,
		start:                  0,
		next:                   0,
		lineIndex:              1,
		lineCharIndex:          0,
		indentStack:            []int{0},
		userUsingSpacesForTabs: nil,
		inStringStarter:        nil,
		inStringStartIndex:     -1,
		escaping:               false,
		Tokens:                 []Token{},
	}
}

func (l *Lexer) Lex() []Token {
	for !l.isAtEnd() {
		l.scanToken()
	}

	// emit dedent tokens for any remaining indents
	for i := len(l.indentStack) - 1; i > 0; i-- {
		lexeme := string(l.source[l.next:l.next])
		token := NewToken(DEDENT, lexeme, l.next, l.lineIndex, l.lineCharIndex)
		l.Tokens = append(l.Tokens, token)
	}

	l.Tokens = append(l.Tokens, NewToken(EOF, "", l.next, l.lineIndex, l.lineCharIndex))
	return l.Tokens
}

func (l *Lexer) isAtEnd() bool {
	return l.next >= len(l.source)
}

func (l *Lexer) scanToken() {
	l.start = l.next
	c := l.advance()

	if l.lineCharIndex == 1 {
		if l.userUsingSpacesForTabs != nil {
			if *l.userUsingSpacesForTabs {
				l.rewind(1)
				l.lexSpaceIndent()
			} else {
				l.lexTabIndent()
			}
		} else {
			if c == ' ' {
				l.rewind(1)
				l.lexSpaceIndent()
			} else if c == '\t' {
				l.lexTabIndent()
			}
		}
	}

	switch c {
	case '(':
		l.addToken(LEFT_PAREN)
	case ')':
		l.addToken(RIGHT_PAREN)
	case '[':
		if l.match(']') {
			l.addToken(BRACKETS)
		} else {
			l.addToken(LEFT_BRACKET)
		}
	case ']':
		l.addToken(RIGHT_BRACKET)
	case '{':
		l.addToken(LEFT_BRACE)
	case '}':
		if l.inStringStarter != nil {
			l.lexStringLiteral(*l.inStringStarter)
		} else {
			l.addToken(RIGHT_BRACE)
		}
	case ',':
		l.addToken(COMMA)
	case ':':
		l.addToken(COLON)
	case '\n':
		l.addToken(NEWLINE)
	case '=':
		if l.match('=') {
			l.addToken(EQUAL_EQUAL)
		} else {
			l.addToken(EQUAL)
		}
	case '!':
		if l.match('=') {
			l.addToken(NOT_EQUAL)
		} else {
			l.addToken(EXCLAMATION)
		}
	case '?':
		l.addToken(QUESTION)
	case '<':
		if l.match('=') {
			l.addToken(LESS_EQUAL)
		} else {
			l.addToken(LESS)
		}
	case '>':
		if l.match('=') {
			l.addToken(GREATER_EQUAL)
		} else {
			l.addToken(GREATER)
		}
	case '|':
		l.addToken(PIPE)
	case '+':
		if l.match('=') {
			l.addToken(PLUS_EQUAL)
		} else {
			l.addToken(PLUS)
		}
	case '-':
		if l.matchString("--") && l.match('\n') {
			l.lexFileHeader()
		} else if l.match('=') {
			l.addToken(MINUS_EQUAL)
		} else {
			l.addToken(MINUS)
		}
	case '@':
		l.addToken(AT)
	case '$':
		l.addToken(DOLLAR)
	case '#':
		if l.match('!') && l.lineCharIndex == 2 {
			l.lexShebang()
		} else {
			l.lexArgComment()
		}
	case '"':
		l.lexStringLiteral('"')
	case '\'':
		l.lexStringLiteral('\'')
	case '`':
		l.lexStringLiteral('`')
	case '.':
		l.addToken(DOT)
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		} else if l.match('=') {
			l.addToken(SLASH_EQUAL)
		} else {
			l.addToken(SLASH)
		}
	case '*':
		if l.match('=') {
			l.addToken(STAR_EQUAL)
		} else {
			l.addToken(STAR)
		}
	case ' ', '\t':
		// ignore whitespace if not at start of line
	default:
		if isDigit(c) {
			l.lexNumber()
		} else if isAlpha(c) || c == '_' {
			l.lexIdentifier()
		} else {
			l.error("Unexpected character")
		}
	}
}

func (l *Lexer) advance() rune {
	r := l.source[l.next]
	if r == '\n' {
		l.lineIndex++
		l.lineCharIndex = 0
	} else {
		l.lineCharIndex++
	}
	l.next++
	return r
}

func (l *Lexer) match(expected rune) bool {
	return l.matchAny(expected)
}

func (l *Lexer) matchAny(expected ...rune) bool {
	if l.isAtEnd() {
		return false
	}

	nextRune := l.source[l.next]
	for _, r := range expected {
		if nextRune == r {
			if nextRune == '\n' {
				// todo: this results in bad errors for multiline tokens
				//  should only do this *after* the token is emitted
				//  issue is that lineindex is incremented before the token is emitted
				l.lineIndex++
				l.lineCharIndex = 0
			} else {
				l.lineCharIndex++
			}
			l.next++
			return true
		}
	}
	return false
}

func (l *Lexer) matchString(expected string) bool {
	for i, c := range expected {
		if l.next+i >= len(l.source) || rune(l.source[l.next+i]) != c {
			return false
		}
	}
	l.next += len(expected)
	return true
}

func (l *Lexer) peekEquals(toCheck string) bool {
	for i, c := range toCheck {
		if l.next+i >= len(l.source) || rune(l.source[l.next+i]) != c {
			return false
		}
	}
	return true
}

func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return rune(l.source[l.next])
}

func (l *Lexer) expectAndEmit(expected rune, tokenType TokenType, errorMessage string) {
	if !l.match(expected) {
		l.error(errorMessage)
	}
	l.addToken(tokenType)
	l.start = l.next
}

func (l *Lexer) expectNoEmit(expected rune, errorMessage string) {
	if !l.match(expected) {
		l.error(errorMessage)
	}
}

func (l *Lexer) rewind(num int) {
	l.next -= num
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (l *Lexer) lexStringLiteral(endChar rune) {
	if l.inStringStarter == nil {
		// we're beginning a truly new string
		l.inStringStarter = &endChar
		l.inStringStartIndex = l.start + 1 // +1 to exclude the starting quote
	}
	value := ""
	for !l.match(endChar) {
		if l.isAtEnd() {
			l.error("Unterminated string")
		}
		if l.match('\\') {
			l.escaping = !l.escaping
		} else {
			l.escaping = false
		}
		if l.match('{') {
			if !l.escaping {
				l.addStringLiteralToken(value, true)
				l.start = l.next
				return
			} else {
				l.rewind(1)
			}
		}
		value = value + string(l.advance())
	}
	l.addStringLiteralToken(value, false)
	if l.inStringStarter != nil && *l.inStringStarter == endChar {
		// we're ending the final part of the string broken up by inline exprs
		l.inStringStarter = nil
		l.inStringStartIndex = -1
	}
}

func (l *Lexer) lexNumber() {
	for isDigit(l.peek()) {
		l.advance()
	}

	isFloat := l.match('.')
	if isFloat {
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	lexeme := l.currentLexeme()

	if isFloat {
		literal, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			l.error("Invalid float")
		}
		l.addFloatLiteralToken(literal)
	} else {
		// int
		literal, err := strconv.ParseInt(lexeme, 10, 64) // what happens to ints starting with 0? e.g. 012?
		if err != nil {
			l.error("Invalid integer")
		}
		l.addIntLiteralToken(literal)
	}
}

func (l *Lexer) lexIdentifier() {
	nextChar := l.peek()
	for isAlpha(nextChar) || isDigit(nextChar) || nextChar == '_' {
		l.advance()
		nextChar = l.peek()
	}

	text := l.currentLexeme()

	if text == "true" {
		l.addBoolLiteralToken(true)
	} else if text == "false" {
		l.addBoolLiteralToken(false)
	} else {
		l.addToken(IDENTIFIER)
	}
}

func (l *Lexer) lexShebang() {
	// "#!" already matched at start of line

	if l.lineIndex != 1 {
		l.error("Shebangs are only allowed on the first line")
	}

	for l.peek() != '\n' && !l.isAtEnd() {
		l.advance()
	}
}

func (l *Lexer) lexArgComment() {
	for l.peek() != '\n' && !l.isAtEnd() {
		l.advance()
	}

	value := strings.TrimSpace(string(l.source[l.start+1 : l.next]))
	l.addArgCommentLiteralToken(&value)
}

func (l *Lexer) lexFileHeader() {
	oneLiner := ""

	for !l.match('\n') {
		oneLiner = oneLiner + string(l.advance())
	}

	if oneLiner == "" {
		l.error("One-line description must not be empty")
	}

	if l.matchString("---") {
		l.addFileHeaderToken(&oneLiner, nil)
		return
	}

	l.expectNoEmit('\n', "Blank line must separate one-line description from multi-line description")
	for l.match('\n') {
		// skip blank lines
	}

	rest := ""
	for !l.matchString("\n---") {
		rest = rest + string(l.advance())
	}

	if rest == "" {
		l.addFileHeaderToken(&oneLiner, nil)
	} else {
		l.addFileHeaderToken(&oneLiner, &rest)
	}
}

func (l *Lexer) lexSpaceIndent() {
	numSpaces := 0
	for l.match(' ') {
		numSpaces++
	}
	if l.match('\t') {
		l.error("Mixing spaces and tabs for indentation is not allowed")
	}
	if l.match('\n') {
		// ignore blank lines
		return
	}
	l.emitIndentTokens(numSpaces, true)
	if l.next == l.start {
		// prior to going in here, we rewound to get the indentation parsing correct
		// if we're still at the same spot, it means we didn't have anything to parse and thus advance us forward,
		// so we just wanna undo the rewind
		l.next++
	}
}

func (l *Lexer) lexTabIndent() {
	l.rewind(1)
	numTabs := 0
	for l.match('\t') {
		numTabs++
	}
	if l.match(' ') {
		l.error("Mixing spaces and tabs for indentation is not allowed")
	}
	if l.match('\n') {
		// ignore blank lines
		return
	}
	l.emitIndentTokens(numTabs, false)
}

func (l *Lexer) currentLexeme() string {
	return string(l.source[l.start:l.next])
}

func (l *Lexer) addToken(tokenType TokenType) {
	lexeme := l.currentLexeme()
	token := NewToken(tokenType, lexeme, l.start, l.lineIndex, l.lineCharIndex)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addStringLiteralToken(literal string, followedByInlineExpr bool) {
	lexeme := l.currentLexeme()
	fullString := ""
	if l.inStringStartIndex != -1 {
		fullString = string(l.source[l.inStringStartIndex : l.next-1])
	}
	token := NewStringLiteralToken(STRING_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal, followedByInlineExpr, fullString)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addIntLiteralToken(literal int64) {
	lexeme := l.currentLexeme()
	token := NewIntLiteralToken(INT_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addFloatLiteralToken(literal float64) {
	lexeme := l.currentLexeme()
	token := NewFloatLiteralToken(FLOAT_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addBoolLiteralToken(literal bool) {
	lexeme := l.currentLexeme()
	token := NewBoolLiteralToken(BOOL_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addFileHeaderToken(oneLiner *string, rest *string) {
	lexeme := l.currentLexeme()
	token := NewFileHeaderToken(FILE_HEADER, lexeme, l.start, l.lineIndex, l.lineCharIndex, *oneLiner, rest)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addArgCommentLiteralToken(comment *string) {
	lexeme := l.currentLexeme()
	token := NewArgCommentToken(ARG_COMMENT, lexeme, l.start, l.lineIndex, l.lineCharIndex, comment)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) emitIndentTokens(numWhitespaces int, isSpaces bool) {
	if l.userUsingSpacesForTabs == nil {
		l.userUsingSpacesForTabs = &isSpaces
	}

	if *l.userUsingSpacesForTabs != isSpaces {
		l.error("Mixing spaces and tabs for indentation is not allowed")
	}

	if numWhitespaces == l.indentStack[len(l.indentStack)-1] {
		// no change
		return
	}

	if numWhitespaces > l.indentStack[len(l.indentStack)-1] {
		l.indentStack = append(l.indentStack, numWhitespaces)
		lexeme := l.currentLexeme()
		token := NewToken(INDENT, lexeme, l.start, l.lineIndex, l.lineCharIndex)
		l.Tokens = append(l.Tokens, token)
		return
	}

	for numWhitespaces < l.indentStack[len(l.indentStack)-1] {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		lexeme := l.currentLexeme()
		token := NewToken(DEDENT, lexeme, l.start, l.lineIndex, l.lineCharIndex)
		l.Tokens = append(l.Tokens, token)
	}

	expectedIndentationLevel := l.indentStack[len(l.indentStack)-1]
	if numWhitespaces != expectedIndentationLevel {
		l.error(fmt.Sprintf("Inconsistent indentation levels. Expected %d spaces/tabs, got %d",
			expectedIndentationLevel, numWhitespaces))
	}
}

func (l *Lexer) error(message string) {
	lexeme := l.currentLexeme()
	lexeme = strings.ReplaceAll(lexeme, "\n", "\\n") // todo, instead should maybe just write the last line?
	lineStart := l.lineCharIndex - (l.next - l.start - 1)
	RP.ErrorExit(fmt.Sprintf("Error at L%d/%d on '%s': %s\n", l.lineIndex, lineStart, lexeme, message))
}
