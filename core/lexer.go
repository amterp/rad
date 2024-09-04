package core

import (
	"fmt"
	"strconv"
	"strings"
)

type Lexer struct {
	source                 string
	start                  int   // index of start of the current lexeme (0 indexed)
	next                   int   // index of next character to be read (0 indexed)
	lineIndex              int   // current line number (1 indexed)
	lineCharIndex          int   // index of latest parsed char in the current line (1 indexed)
	indentStack            []int // stack of indents to, to emit indent/dedent tokens
	userUsingSpacesForTabs *bool // nil until we see the first case of a space indent
	Tokens                 []Token
}

func NewLexer(source string) *Lexer {
	return &Lexer{
		source:                 source,
		start:                  0,
		next:                   0,
		lineIndex:              1,
		lineCharIndex:          0,
		indentStack:            []int{0},
		userUsingSpacesForTabs: nil,
		Tokens:                 []Token{},
	}
}

func (l *Lexer) Lex() []Token {
	for !l.isAtEnd() {
		l.scanToken()
	}

	// emit dedent tokens for any remaining indents
	for i := len(l.indentStack) - 1; i > 0; i-- {
		lexeme := l.source[l.next:l.next]
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
		l.addToken(PLUS)
	case '-':
		l.addToken(MINUS)
	case '@':
		l.addToken(AT)
	case '#':
		if l.match('!') && l.lineCharIndex == 2 {
			l.lexShebang()
		} else {
			l.lexArgComment()
		}
	case '"':
		if l.match('"') {
			if l.match('"') {
				if !l.match('\n') {
					l.error("Expected newline after triple quote")
				} else {
					l.lexFileHeader()
				}
			} else {
				l.addStringLiteralToken("")
			}
		} else {
			l.lexStringLiteral()
		}
	case 'j':
		if l.matchString("son") {
			l.lexJsonPath()
		}
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		} else {
			l.addToken(SLASH)
		}
	case '*':
		l.addToken(STAR)
	case ' ', '\t':
		// ignore whitespace if not at start of line
	default:
		if isDigit(c) {
			l.lexNumber()
		} else if isAlpha(c) {
			l.lexIdentifier()
		} else {
			l.error("Unexpected character")
		}
	}
}

func (l *Lexer) advance() rune {
	r := rune(l.source[l.next])
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

	nextRune := rune(l.source[l.next])
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

func (l *Lexer) lexStringLiteral() {
	value := ""
	for !l.match('"') && !l.isAtEnd() {
		value = value + string(l.advance())
	}
	l.addStringLiteralToken(value)
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

	lexeme := l.source[l.start:l.next]

	if isFloat {
		literal, err := strconv.ParseFloat(lexeme, 64)
		if err != nil {
			l.error("Invalid float")
		}
		l.addFloatLiteralToken(literal)
	} else {
		// int
		literal, err := strconv.Atoi(lexeme)
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

	text := l.source[l.start:l.next]

	if text == "true" {
		l.addBoolLiteralToken(true)
	} else if text == "false" {
		l.addBoolLiteralToken(false)
	} else {
		l.addToken(IDENTIFIER)
	}
}

func (l *Lexer) lexJsonPath() {
	isArray := l.matchString("[]")
	l.addJsonPathElementToken("json", isArray)

	for l.peek() != '\n' && !l.isAtEnd() {
		l.start = l.next
		l.expectAndEmit('.', DOT, "Expected '.' to preface next json path element")
		l.lexJsonPathElement()
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

	value := strings.TrimSpace(l.source[l.start+1 : l.next])
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

	if l.matchString("\"\"\"") {
		l.addFileHeaderToken(&oneLiner, nil)
		return
	}

	l.expectNoEmit('\n', "Blank line must separate one-line description from multi-line description")
	for l.match('\n') {
		// skip blank lines
	}

	rest := ""
	for !l.matchString("\n\"\"\"") {
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

func (l *Lexer) lexJsonPathElement() {
	value := ""
	escaping := false
	for ((l.peek() != '.' && l.peek() != '[') || escaping) && l.peek() != '\n' && !l.isAtEnd() {
		if l.peek() == '\\' {
			escaping = true
			l.advance()
		} else {
			if escaping {
				escaping = false
			}
			value = value + string(l.advance())
		}
	}
	includesBrackets := l.matchString("[]")
	l.addJsonPathElementToken(value, includesBrackets)
}

func (l *Lexer) addToken(tokenType TokenType) {
	lexeme := l.source[l.start:l.next]
	token := NewToken(tokenType, lexeme, l.start, l.lineIndex, l.lineCharIndex)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addStringLiteralToken(literal string) {
	lexeme := l.source[l.start:l.next]
	token := NewStringLiteralToken(STRING_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addIntLiteralToken(literal int) {
	lexeme := l.source[l.start:l.next]
	token := NewIntLiteralToken(INT_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addFloatLiteralToken(literal float64) {
	lexeme := l.source[l.start:l.next]
	token := NewFloatLiteralToken(FLOAT_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addBoolLiteralToken(literal bool) {
	lexeme := l.source[l.start:l.next]
	token := NewBoolLiteralToken(BOOL_LITERAL, lexeme, l.start, l.lineIndex, l.lineCharIndex, literal)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addFileHeaderToken(oneLiner *string, rest *string) {
	lexeme := l.source[l.start:l.next]
	token := NewFileHeaderToken(FILE_HEADER, lexeme, l.start, l.lineIndex, l.lineCharIndex, *oneLiner, rest)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) addArgCommentLiteralToken(comment *string) {
	lexeme := l.source[l.start:l.next]
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
		lexeme := l.source[l.start:l.next]
		token := NewToken(INDENT, lexeme, l.start, l.lineIndex, l.lineCharIndex)
		l.Tokens = append(l.Tokens, token)
		return
	}

	for numWhitespaces < l.indentStack[len(l.indentStack)-1] {
		l.indentStack = l.indentStack[:len(l.indentStack)-1]
		lexeme := l.source[l.start:l.next]
		token := NewToken(DEDENT, lexeme, l.start, l.lineIndex, l.lineCharIndex)
		l.Tokens = append(l.Tokens, token)
	}

	expectedIndentationLevel := l.indentStack[len(l.indentStack)-1]
	if numWhitespaces != expectedIndentationLevel {
		l.error(fmt.Sprintf("Inconsistent indentation levels. Expected %d spaces/tabs, got %d",
			expectedIndentationLevel, numWhitespaces))
	}
}

func (l *Lexer) addJsonPathElementToken(jsonPathElement string, isArray bool) {
	lexeme := l.source[l.start:l.next]
	token := NewJsonPathElementToken(JSON_PATH_ELEMENT, lexeme, l.start, l.lineIndex, l.lineCharIndex, jsonPathElement, isArray)
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) error(message string) {
	lexeme := l.source[l.start:l.next]
	lexeme = strings.ReplaceAll(lexeme, "\n", "\\n") // todo, instead should maybe just write the last line?
	lineStart := l.lineCharIndex - (l.next - l.start - 1)
	panic(fmt.Sprintf("Error at L%d/%d on '%s': %s", l.lineIndex, lineStart, lexeme, message))
}
