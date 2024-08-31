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
	p.consumeNewlines()

	for !p.isAtEnd() {
		s := p.statement()
		if _, ok := s.(*Empty); !ok {
			statements = append(statements, s)
		}
		p.consumeNewlines()
	}
	return statements
}

func (p *Parser) isAtEnd() bool {
	return p.peek().GetType() == EOF
}

func (p *Parser) peekType(tokenType TokenType) bool {
	return p.peek().GetType() == tokenType
}

func (p *Parser) peekTypeSeries(tokenType ...TokenType) bool {
	for i, t := range tokenType {
		token := p.advance()
		if token.GetType() != t {
			p.next -= i + 1
			return false
		}
	}
	p.next -= len(tokenType)
	return true
}

func (p *Parser) peek() Token {
	return p.tokens[p.next]
}

func (p *Parser) peekTwoAhead() Token {
	return p.tokens[p.next+1]
}

func (p *Parser) matchAny(tokenTypes ...TokenType) bool {
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

func (p *Parser) rewind() {
	p.next--
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
	if p.matchAny(FILE_HEADER) {
		*statements = append(*statements, &FileHeader{FileHeaderToken: p.previous()})
	}
}

func (p *Parser) argBlockIfPresent(statements *[]Stmt) {
	if p.matchKeyword(ARGS, GLOBAL_KEYWORDS) {
		argsKeyword := p.previous()
		p.consume(COLON, "Expected ':' after 'args'")
		p.consumeNewlines()

		if !p.matchAny(INDENT) {
			return
		}

		p.consumeNewlines()
		argsBlock := ArgBlock{ArgsKeyword: argsKeyword, Stmts: []ArgStmt{}}
		for !p.matchAny(DEDENT) {
			s := p.argStatement()
			argsBlock.Stmts = append(argsBlock.Stmts, s)
			p.consumeNewlines()
		}
		*statements = append(*statements, &argsBlock)
	}
}

// argBlockConstraint       -> argStringRegexConstraint
//
//	| argIntRangeConstraint
//	| argOneWayReq
//	| argMutualExcl
//
// argStringRegexConstraint -> IDENTIFIER ( "," IDENTIFIER )* "not"? "regex" REGEX
// argIntRangeConstraint    -> IDENTIFIER COMPARATORS INT
// argOneWayReq             -> IDENTIFIER "requires" IDENTIFIER
// argMutualExcl            -> "one_of" IDENTIFIER ( "," IDENTIFIER )+
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
	if p.matchAny(STRING_LITERAL) {
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
		if p.matchAny(BRACKETS) {
			rslTypeEnum = RslStringArray
		} else {
			rslTypeEnum = RslString
		}
	} else if p.matchKeyword(INT, ARGS_BLOCK_KEYWORDS) {
		argType = p.previous()
		if p.matchAny(BRACKETS) {
			rslTypeEnum = RslIntArray
		} else {
			rslTypeEnum = RslInt
		}
	} else if p.matchKeyword(FLOAT, ARGS_BLOCK_KEYWORDS) {
		argType = p.previous()
		if p.matchAny(BRACKETS) {
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
	var defaultLiteral LiteralOrArray
	if p.matchAny(QUESTION) {
		isOptional = true
	} else if p.matchAny(EQUAL) {
		defaultLiteralIfPresent, ok := p.literalOrArray(&rslTypeEnum)
		if !ok {
			p.error("Expected default value")
		} else {
			defaultLiteral = defaultLiteralIfPresent
		}
	}

	argComment := p.consume(ARG_COMMENT, "Expected arg Comment").(*ArgCommentToken)

	return &ArgDeclaration{
		Identifier: identifier,
		Rename:     &stringLiteral,
		Flag:       &flag,
		ArgType:    RslType{Token: argType, Type: rslTypeEnum},
		IsOptional: isOptional,
		Default:    &defaultLiteral,
		Comment:    *argComment,
	}
}

func (p *Parser) statement() Stmt {
	if p.matchKeyword(RAD, GLOBAL_KEYWORDS) {
		return p.radBlock()
	}

	// todo all keywords
	// todo for stmt
	// todo if stmt

	if p.peekTypeSeries(IDENTIFIER, LEFT_PAREN) {
		return p.functionCallStmt()
	}

	return p.assignment()
}

func (p *Parser) radBlock() *RadBlock {
	radToken := p.previous()
	urlToken := p.expr()
	p.consume(COLON, "Expecting ':' to start rad block")
	p.consumeNewlines()
	if !p.matchAny(INDENT) {
		p.error("Expecting indented contents in rad block")
	}
	p.consumeNewlines()
	var radStatements []RadStmt
	for !p.matchAny(DEDENT) {
		p.consumeNewlines()
		radStatements = append(radStatements, p.radStatement())
	}
	radBlock := &RadBlock{RadKeyword: radToken, Url: &urlToken, RadStmts: radStatements}
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
	for _, stmt := range radBlock.RadStmts {
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
	for !p.matchAny(NEWLINE) {
		p.consume(COMMA, "Expected ',' between identifiers")
		identifiers = append(identifiers, p.identifier())
	}
	return &Fields{Identifiers: identifiers}
}

func (p *Parser) functionCallStmt() Stmt {
	functionCall := p.functionCall()
	return &ExprStmt{Expression: functionCall}
}

func (p *Parser) assignment() Stmt {
	var identifiers []Token
	identifiers = append(identifiers, p.identifier())

	for !p.matchAny(EQUAL) {
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
	if isArray := p.matchAny(BRACKETS); isArray {
		brackets = p.previous()
	}
	elements := []JsonPathElement{{token: element, arrayToken: &brackets}}
	for !p.matchAny(NEWLINE) {
		p.consume(DOT, "Expected '.' to separate json field elements")
		element = p.consume(JSON_PATH_ELEMENT, "Expected json path element after '.'")
		if p.matchAny(BRACKETS) {
			brackets = p.previous()
		}
		elements = append(elements, JsonPathElement{token: element, arrayToken: &brackets})
	}
	return &JsonPathAssign{Identifier: identifier, Elements: elements}
}

func (p *Parser) primaryAssignment(name Token) Stmt {
	initializer := p.expr()
	return &PrimaryAssign{Name: name, Initializer: initializer}
}

func (p *Parser) expr() Expr {
	return p.or()
}

func (p *Parser) or() Expr {
	expr := p.and()

	for p.matchKeyword(OR, ALL_KEYWORDS) {
		operator := p.previous()
		right := p.and()
		expr = &Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) and() Expr {
	expr := p.equality()

	for p.matchKeyword(AND, ALL_KEYWORDS) {
		operator := p.previous()
		right := p.equality()
		expr = &Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) equality() Expr {
	expr := p.comparison()

	for p.matchAny(NOT_EQUAL, EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()

	for p.matchAny(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()

	for p.matchAny(MINUS, PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()

	for p.matchAny(SLASH, STAR) {
		operator := p.previous()
		right := p.unary()
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) unary() Expr {
	if p.matchAny(NOT, MINUS, PLUS) {
		operator := p.previous()
		right := p.unary()
		return &Unary{Operator: operator, Right: right}
	}

	return p.primary()
}

func (p *Parser) primary() Expr {
	if p.matchAny(LEFT_PAREN) {
		expr := p.expr()
		p.consume(RIGHT_PAREN, "Expected ')' after expression")
		return &Grouping{Value: expr}
	}

	if loa, ok := p.literalOrArray(nil); ok {
		return &ExprLoa{Value: loa}
	}

	if arrayExpr, ok := p.arrayExpr(); ok {
		return arrayExpr
	}

	if p.matchAny(IDENTIFIER) {
		identifier := p.previous()
		if p.matchAny(LEFT_BRACKET) {
			array := p.expr()
			p.consume(RIGHT_BRACKET, "Expected ']' after array expression")
			return &ArrayAccess{Array: identifier, Index: array}
		}
		if p.peekType(LEFT_PAREN) {
			p.rewind()
			return p.functionCall()
		}
		return &Variable{Name: identifier}
	}

	p.error("Expected expression")
	panic(UNREACHABLE)
}

func (p *Parser) functionCall() Expr {
	function := p.consume(IDENTIFIER, "Expected function name")
	p.consume(LEFT_PAREN, "Expected '(' after function name")
	var args []Expr
	if !p.matchAny(RIGHT_PAREN) {
		args = append(args, p.expr())
		for !p.matchAny(RIGHT_PAREN) {
			p.consume(COMMA, "Expected ',' between function arguments")
			args = append(args, p.expr())
		}
	}
	return &FunctionCall{Function: function, Args: args}
}

func (p *Parser) arrayExpr() (Expr, bool) {
	if !p.matchAny(LEFT_BRACKET) {
		return nil, false
	}

	if p.matchAny(RIGHT_BRACKET) {
		return &ArrayExpr{Values: []Expr{}}, true
	}

	values := []Expr{p.expr()}
	for !p.matchAny(RIGHT_BRACKET) {
		p.consume(COMMA, "Expected ',' between array elements")
		values = append(values, p.expr())
	}

	return &ArrayExpr{Values: values}, true
}

func (p *Parser) literalOrArray(expectedType *RslTypeEnum) (LiteralOrArray, bool) {
	if literal, ok := p.literal(expectedType); ok {
		return &LoaLiteral{Value: literal}, true
	}

	arrayLiteral, ok := p.arrayLiteral(expectedType)
	if ok {
		return &LoaArray{Value: arrayLiteral}, true
	}

	return nil, false
}

func (p *Parser) literal(expectedType *RslTypeEnum) (Literal, bool) {
	if p.matchAny(STRING_LITERAL) {
		if expectedType != nil && *expectedType != RslString {
			p.error("Expected string literal")
		}
		return &StringLiteral{Value: p.previous()}, true
	}

	if p.matchAny(INT_LITERAL) {
		if expectedType != nil && *expectedType != RslInt {
			p.error("Expected int literal")
		}
		return &IntLiteral{Value: p.previous()}, true
	}

	if p.matchAny(FLOAT_LITERAL) {
		if expectedType != nil && *expectedType != RslFloat {
			p.error("Expected float literal")
		}
		return &FloatLiteral{Value: p.previous()}, true
	}

	// todo need to emit bool literal tokens
	if p.matchAny(BOOL_LITERAL) {
		if expectedType != nil && *expectedType != RslBool {
			p.error("Expected bool literal")
		}
		return &BoolLiteral{Value: p.previous()}, true
	}

	return nil, false
}

func (p *Parser) arrayLiteral(expectedType *RslTypeEnum) (ArrayLiteral, bool) {
	if !p.matchAny(LEFT_BRACKET) {
		return nil, false
	}

	if p.matchAny(RIGHT_BRACKET) {
		return &UnknownArrayLiteral{}, true
	}

	if literal, ok := p.literal(expectedType); ok {
		switch literal.(type) {
		case StringLiteral:
			stringArray := p.stringArrayLiteral().(StringArrayLiteral)
			stringArray.Values = append([]StringLiteral{literal.(StringLiteral)}, stringArray.Values...)
			return stringArray, true
		case IntLiteral:
			intArray := p.intArrayLiteral().(IntArrayLiteral)
			intArray.Values = append([]IntLiteral{literal.(IntLiteral)}, intArray.Values...)
			return intArray, true
		case FloatLiteral:
			floatArray := p.floatArrayLiteral().(FloatArrayLiteral)
			floatArray.Values = append([]FloatLiteral{literal.(FloatLiteral)}, floatArray.Values...)
			return floatArray, true
		case BoolLiteral: // todo technically not part of the arg_types.go handling
			boolArray := p.boolArrayLiteral().(BoolArrayLiteral)
			boolArray.Values = append([]BoolLiteral{literal.(BoolLiteral)}, boolArray.Values...)
			return boolArray, true
		}
	}

	return nil, false
}

func (p *Parser) stringArrayLiteral() ArrayLiteral {
	var values []StringLiteral
	expectedType := RslString
	for !p.matchAny(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(StringLiteral))
		if !p.matchAny(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return StringArrayLiteral{Values: values}
}

func (p *Parser) intArrayLiteral() ArrayLiteral {
	var values []IntLiteral
	expectedType := RslInt
	for !p.matchAny(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(IntLiteral))
		if !p.matchAny(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return IntArrayLiteral{Values: values}
}

func (p *Parser) floatArrayLiteral() ArrayLiteral {
	var values []FloatLiteral
	expectedType := RslFloat
	for !p.matchAny(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(FloatLiteral))
		if !p.matchAny(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return FloatArrayLiteral{Values: values}
}

func (p *Parser) boolArrayLiteral() ArrayLiteral {
	var values []BoolLiteral
	expectedType := RslBool
	for !p.matchAny(RIGHT_BRACKET) {
		literal, ok := p.literal(&expectedType)
		if !ok {
			p.error("Expected literal in array")
		}
		values = append(values, literal.(BoolLiteral))
		if !p.matchAny(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return BoolArrayLiteral{Values: values}
}

func (p *Parser) identifier() Token {
	if p.matchAny(IDENTIFIER) {
		return p.previous()
	}
	p.error("Expected Identifier")
	panic(UNREACHABLE)
}

// todo putting this everywhere isn't ideal... another way to handle insignificant newlines?
func (p *Parser) consumeNewlines() {
	for p.matchAny(NEWLINE) {
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
