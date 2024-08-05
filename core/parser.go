package core

import "fmt"

const UNREACHABLE = "unreachable"

type Parser struct {
	tokens []Token
	next   int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		next:   0,
	}
}

func (p *Parser) Parse() []Stmt {
	var statements []Stmt
	for !p.isAtEnd() {
		s := p.statement()
		if _, ok := s.(*Empty); !ok {
			statements = append(statements, s)
		}
	}
	return statements
}

func (p *Parser) statement() Stmt {
	if p.match(NEWLINE) {
		return &Empty{}
	}

	return p.assignment()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().GetType() == EOF
}

func (p *Parser) peekType(tokenType TokenType) bool {
	return p.peek().GetType() == tokenType
}

func (p *Parser) peek() Token {
	return p.tokens[p.next]
}

func (p *Parser) match(tokenTypes ...TokenType) bool {
	for _, t := range tokenTypes {
		if p.peekType(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.next++
	}
	return p.previous()
}

func (p *Parser) previous() Token {
	return p.tokens[p.next-1]
}

func (p *Parser) consume(tokenType TokenType, errorMessageIfNotMatch string) Token {
	if p.peekType(tokenType) {
		return p.advance()
	}
	p.error(errorMessageIfNotMatch)
	panic("unreachable")
}

func (p *Parser) error(message string) {
	currentToken := p.tokens[p.next]
	panic(fmt.Sprintf("Error at L%d/%d on '%s': %s",
		currentToken.GetLine(), currentToken.GetCharLineStart(), currentToken.GetLexeme(), message))
}

func (p *Parser) assignment() Stmt {
	var names []Token
	names = append(names, p.identifier())

	for !p.match(EQUAL) {
		p.consume(COMMA, "Expected ',' between identifiers")
		names = append(names, p.identifier())
	}

	if len(names) > 1 {
		p.error("Multiple assignments not YET supported")
	}

	name := names[0]

	return p.primaryAssignment(name)
}

func (p *Parser) primaryAssignment(name Token) Stmt {
	initializer := p.expr()
	// note to self: i think i would need to update a map here that tracks variable names and types, if
	// i wanted to do 'static' type checking
	return &PrimaryAssign{name: name, initializer: initializer}
}

func (p *Parser) expr() Expr {
	if p.match(STRING_LITERAL) {
		return &StringLiteral{value: p.previous()}
	}

	if p.match(INT_LITERAL) {
		return &IntLiteral{value: p.previous()}
	}

	if p.match(BOOL_LITERAL) { // todo need to emit bool literal tokens
		return &BoolLiteral{value: p.previous()}
	}

	if p.match(IDENTIFIER) {
		return &Variable{name: p.previous()}
	}

	p.error("Expected identifier or string")
	panic(UNREACHABLE)
}

func (p *Parser) identifier() Token {
	if p.match(IDENTIFIER) {
		return p.previous()
	}
	p.error("Expected identifier")
	panic(UNREACHABLE)
}
