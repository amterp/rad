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
		p.consume(COLON, "Expected ':' after 'Args'")
		p.consumeNewlines()

		if !p.match(INDENT) {
			return
		}

		p.consumeNewlines()
		argsBlock := ArgBlock{argsKeyword: argsKeyword, ArgStmts: []ArgStmt{}}
		for !p.match(DEDENT) {
			s := p.argStatement()
			argsBlock.ArgStmts = append(argsBlock.ArgStmts, s)
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

	identifier := p.consume(IDENTIFIER, "Expected Identifier or keyword")

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
		flag = p.consume(IDENTIFIER, "Expected Flag")
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
	var defaultLiteral *LiteralOrArray
	if p.match(QUESTION) {
		isOptional = true
	} else if p.match(EQUAL) {
		defaultLiteral = p.literalOrArray(rslTypeEnum)
	}

	argComment := p.consume(ARG_COMMENT, "Expected arg Comment").(*ArgCommentToken)

	return &ArgDeclaration{
		Identifier: identifier,
		Rename:     &stringLiteral,
		Flag:       &flag,
		ArgType:    RslType{Token: argType, Type: rslTypeEnum},
		IsOptional: isOptional,
		Default:    defaultLiteral,
		Comment:    *argComment,
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
	urlToken := p.expr()
	p.consume(COLON, "Expecting ':' to start rad block")
	p.consumeNewlines()
	if !p.match(INDENT) {
		p.error("Expecting indented contents in rad block")
	}
	p.consumeNewlines()
	var radStatements []RadStmt
	for !p.match(DEDENT) {
		p.consumeNewlines()
		radStatements = append(radStatements, p.radStatement())
	}
	radBlock := &RadBlock{radKeyword: radToken, url: &urlToken, radStmts: radStatements}
	p.validateRadBlock(radBlock)
	return radBlock
}

func (p *Parser) radStatement() RadStmt {
	if p.matchKeyword(FIELDS, RAD_BLOCK_KEYWORDS) {
		return p.radFieldsStatement()
	}
	// todo sort
	// todo modifier
	// todo table fmt
	// todo field fmt
	// todo filtering?
	panic(NOT_IMPLEMENTED)
}

func (p *Parser) validateRadBlock(radBlock *RadBlock) {
	hasFieldsStmt := false
	for _, stmt := range radBlock.radStmts {
		if _, ok := stmt.(*Fields); ok {
			if hasFieldsStmt {
				p.error("Only one 'fields' statement is allowed in a rad block")
			}
			hasFieldsStmt = true
		}
	}
	if !hasFieldsStmt {
		p.error("A rad block must contain a 'fields' statement")
	}
}

func (p *Parser) radFieldsStatement() RadStmt {
	var identifiers []Token
	identifiers = append(identifiers, p.identifier())
	for !p.match(NEWLINE) {
		p.consume(COMMA, "Expected ',' between identifiers")
		identifiers = append(identifiers, p.identifier())
	}
	return &Fields{identifiers: identifiers}
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
	p.literalOrArray()

	if p.match(IDENTIFIER) {
		return &Variable{Name: p.previous()}
	}

	p.error("Expected Identifier or string")
	panic(UNREACHABLE)
}

func (p *Parser) literalOrArray(expectedType RslTypeEnum) LiteralOrArray {
	literal, ok := p.literal(&expectedType)
	if ok {
		return &LiteralOrArrayHolder{LiteralVal: &literal}
	}

	array := p.arrayLiteral(expectedType)
	return &LiteralOrArrayHolder{ArrayVal: &array}
}

func (p *Parser) literal(expectedType *RslTypeEnum) (Literal, bool) {
	if p.match(STRING_LITERAL) {
		if expectedType != nil && *expectedType != RslString {
			p.error("Expected string literal")
		}
		return &StringLiteral{Value: p.previous()}, true
	}

	if p.match(INT_LITERAL) {
		if expectedType != nil && *expectedType != RslInt {
			p.error("Expected int literal")
		}
		return &IntLiteral{Value: p.previous()}, true
	}

	if p.match(FLOAT_LITERAL) {
		if expectedType != nil && *expectedType != RslFloat {
			p.error("Expected float literal")
		}
		return &FloatLiteral{Value: p.previous()}, true
	}

	// todo need to emit bool literal tokens
	if p.match(BOOL_LITERAL) {
		if expectedType != nil && *expectedType != RslBool {
			p.error("Expected bool literal")
		}
		return &BoolLiteral{Value: p.previous()}, true
	}

	return nil, false
}

func (p *Parser) arrayLiteral(expectedType RslTypeEnum) ArrayLiteral {
	switch expectedType {
	case RslString:
		return p.stringArrayLiteral()
	case RslInt:
		return p.intArrayLiteral()
	case RslFloat:
		return p.floatArrayLiteral()
	case RslBool:
		return p.boolArrayLiteral()
	default:
		p.error(fmt.Sprintf("Unknown array of type %v", expectedType))
		panic(UNREACHABLE)
	}
}

func (p *Parser) stringArrayLiteral() ArrayLiteral {
	var values []StringLiteral
	expectedType := RslString
	for !p.match(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(StringLiteral))
		if !p.match(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return StringArrayLiteral{Values: values}
}

func (p *Parser) intArrayLiteral() ArrayLiteral {
	var values []IntLiteral
	expectedType := RslInt
	for !p.match(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(IntLiteral))
		if !p.match(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return IntArrayLiteral{Values: values}
}

func (p *Parser) floatArrayLiteral() ArrayLiteral {
	var values []FloatLiteral
	expectedType := RslFloat
	for !p.match(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(FloatLiteral))
		if !p.match(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return FloatArrayLiteral{Values: values}
}

func (p *Parser) boolArrayLiteral() ArrayLiteral {
	var values []BoolLiteral
	expectedType := RslBool
	for !p.match(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(BoolLiteral))
		if !p.match(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return BoolArrayLiteral{Values: values}
}

func (p *Parser) identifier() Token {
	if p.match(IDENTIFIER) {
		return p.previous()
	}
	p.error("Expected Identifier")
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
