package core

import (
	"fmt"
	"strings"
)

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

	p.consumeNewlines()
	p.fileHeaderIfPresent(&statements)
	p.consumeNewlines()
	p.argBlockIfPresent(&statements)

	for !p.isAtEnd() {
		s := p.statement()
		p.consumeNewlines()
		if _, ok := s.(*Empty); !ok {
			statements = append(statements, s)
		}
	}
	return statements
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

func (p *Parser) peekTwoAhead() Token {
	return p.tokens[p.next+1]
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

func (p *Parser) tryConsume(tokenType TokenType) (Token, bool) {
	if p.peekType(tokenType) {
		return p.advance(), true
	}
	return nil, false
}

func (p *Parser) error(message string) {
	currentToken := p.tokens[p.next]
	lexeme := currentToken.GetLexeme()
	lexeme = strings.ReplaceAll(lexeme, "\n", "\\n") // todo, instead should maybe just write the last line?
	panic(fmt.Sprintf("Error at L%d/%d on '%s': %s",
		currentToken.GetLine(), currentToken.GetCharLineStart(), lexeme, message))
}

func (p *Parser) fileHeaderIfPresent(statements *[]Stmt) {
	if p.match(FILE_HEADER) {
		*statements = append(*statements, &FileHeader{fileHeaderToken: p.previous()})
	}
}

func (p *Parser) argBlockIfPresent(statements *[]Stmt) {
	if p.matchKeyword(ARGS, GLOBAL_KEYWORDS) {
		argsKeyword := p.previous()
		p.consume(COLON, "Expected ':' after 'args'")
		p.consumeNewlines()

		if !p.match(INDENT) {
			return
		}

		p.consumeNewlines()
		argsBlock := ArgBlock{argsKeyword: argsKeyword, argStmts: []ArgStmt{}}
		for !p.match(DEDENT) {
			s := p.argStatement()
			argsBlock.argStmts = append(argsBlock.argStmts, s)
			p.consumeNewlines()
		}
		*statements = append(*statements, &argsBlock)
	}
}

// argBlockConstraint         -> argStringRegexConstraint
//
//	| argIntRangeConstraint
//	| argOneWayReq
//	| argMutualExcl
//
// argStringRegexConstraint   -> IDENTIFIER ( "," IDENTIFIER )* "not"? "regex" REGEX
// argIntRangeConstraint      -> IDENTIFIER COMPARATORS INT
// argOneWayReq               -> IDENTIFIER "requires" IDENTIFIER
// argMutualExcl              -> "one_of" IDENTIFIER ( "," IDENTIFIER )+
func (p *Parser) argStatement() ArgStmt {
	if p.matchKeyword(ONE_OF, ARGS_BLOCK_KEYWORDS) {
		panic(NOT_IMPLEMENTED)
	}

	identifier := p.consume(IDENTIFIER, "Expected identifier or keyword")

	if p.peekType(STRING_LITERAL) ||
		p.peekType(IDENTIFIER) ||
		p.peekType(STRING) ||
		p.peekType(INT) ||
		p.peekType(BOOL) {

		return p.argDeclaration(identifier)
	}

	if p.matchKeyword(REQUIRES, ARGS_BLOCK_KEYWORDS) {
		panic(NOT_IMPLEMENTED)
	}

	// todo rest
	panic(NOT_IMPLEMENTED)
}

func (p *Parser) argDeclaration(identifier Token) ArgStmt {
	var stringLiteral Token
	if p.match(STRING_LITERAL) {
		stringLiteral = p.previous()
	}

	var flag Token
	if p.peekTwoAhead().GetType() == IDENTIFIER {
		flag = p.consume(IDENTIFIER, "Expected flag")
	}

	var argType Token
	var rslTypeEnum RslTypeEnum
	if p.matchKeyword(STRING, ARGS_BLOCK_KEYWORDS) {
		argType = p.previous()
		if p.match(BRACKETS) {
			rslTypeEnum = RslStringArray
		} else {
			rslTypeEnum = RslString
		}
	} else if p.matchKeyword(INT, ARGS_BLOCK_KEYWORDS) {
		argType = p.previous()
		if p.match(BRACKETS) {
			rslTypeEnum = RslIntArray
		} else {
			rslTypeEnum = RslInt
		}
	} else if p.matchKeyword(FLOAT, ARGS_BLOCK_KEYWORDS) {
		argType = p.previous()
		if p.match(BRACKETS) {
			rslTypeEnum = RslFloatArray
		} else {
			rslTypeEnum = RslFloat
		}
	} else if p.matchKeyword(BOOL, ARGS_BLOCK_KEYWORDS) {
		argType = p.previous()
		rslTypeEnum = RslBool
	} else {
		p.error("Expected arg type")
	}

	isOptional := false
	var defaultInit Expr
	if p.match(QUESTION) {
		isOptional = true
	} else if p.match(EQUAL) {
		defaultInit = p.expr()
	}

	argComment := p.consume(ARG_COMMENT, "Expected arg comment").(*ArgCommentToken)

	return &ArgDeclaration{
		identifier:  identifier,
		rename:      &stringLiteral,
		flag:        &flag,
		argType:     RslType{Token: argType, Type: rslTypeEnum},
		isOptional:  isOptional,
		defaultInit: &defaultInit,
		comment:     *argComment,
	}
}

func (p *Parser) statement() Stmt {
	if p.matchKeyword(RAD, GLOBAL_KEYWORDS) {
		return p.radBlock()
	}

	// todo for stmt

	// todo if stmt

	return p.assignment()
}

func (p *Parser) radBlock() *RadBlock {
	radToken := p.previous()
	var urlToken Expr
	if !p.peekType(COLON) {
		urlToken = p.expr()
	}
	p.consume(COLON, "Expecting ':' to start rad block")
	p.consumeNewlines()
	if !p.match(INDENT) {
		p.error("Expecting indented contents in rad block")
	}
	p.consumeNewlines()
	identifiers := []Token{}
	identifiers = append(identifiers, p.identifier())
	for !p.match(NEWLINE) {
		p.consume(COMMA, "Expected ',' between identifiers")
		identifiers = append(identifiers, p.identifier())
	}
	var radStatements []RadStmt
	radStatements = append(radStatements, &Fields{identifiers: identifiers})
	for !p.match(DEDENT) {
		p.consumeNewlines()
		radStatements = append(radStatements, p.radStatement())
	}
	return &RadBlock{radKeyword: radToken, url: &urlToken, radStmts: radStatements}
}

func (p *Parser) radStatement() RadStmt {
	// todo sort
	// todo modifier
	// todo table fmt
	// todo field fmt
	// todo filtering?
	panic(NOT_IMPLEMENTED)
}

func (p *Parser) assignment() Stmt {
	var identifiers []Token
	identifiers = append(identifiers, p.identifier())

	for !p.match(EQUAL) {
		p.consume(COMMA, "Expected ',' between identifiers")
		identifiers = append(identifiers, p.identifier())
	}

	if len(identifiers) > 1 {
		panic(NOT_IMPLEMENTED)
	}

	identifier := identifiers[0]

	if p.peekType(JSON_PATH_ELEMENT) {
		return p.jsonPathAssignment(identifier)
	} else {
		return p.primaryAssignment(identifier)
	}
}

func (p *Parser) jsonPathAssignment(identifier Token) Stmt {
	element := p.consume(JSON_PATH_ELEMENT, "Expected root json path element")
	var brackets Token
	if isArray := p.match(BRACKETS); isArray {
		brackets = p.previous()
	}
	elements := []JsonPathElement{{token: element, arrayToken: &brackets}}
	for !p.match(NEWLINE) {
		p.consume(DOT, "Expected '.' to separate json field elements")
		element = p.consume(JSON_PATH_ELEMENT, "Expected json path element after '.'")
		if p.match(BRACKETS) {
			brackets = p.previous()
		}
		elements = append(elements, JsonPathElement{token: element, arrayToken: &brackets})
	}
	return &JsonPathAssign{identifier: identifier, elements: elements}
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

	if p.match(FLOAT_LITERAL) {
		return &FloatLiteral{value: p.previous()}
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

// todo putting this everywhere isn't ideal... another way to handle insignificant newlines?
func (p *Parser) consumeNewlines() {
	for p.match(NEWLINE) {
		// throw away
	}
}

func (p *Parser) matchKeyword(tokenType TokenType, keywords map[string]TokenType) bool {
	next := p.peek()
	if next.GetType() != IDENTIFIER {
		return false
	}
	if keyword, ok := keywords[next.GetLexeme()]; ok {
		if keyword == tokenType {
			p.advance()
			return true
		}
	}
	return false
}
