package core

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

const (
	onlyOneReturnValueAllowed = "Binary operators are only allowed in expressions with one return value"
)

type Parser struct {
	tokens              []Token
	next                int
	nestedForBlockLevel int
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
	return p.lookback(1)
}

func (p *Parser) lookback(num int) Token {
	return p.tokens[p.next-num]
}

func (p *Parser) consume(tokenType TokenType, errorMessageIfNotMatch string) Token {
	if p.peekType(tokenType) {
		return p.advance()
	}
	p.error(errorMessageIfNotMatch)
	panic(UNREACHABLE)
}

func (p *Parser) tryConsume(tokenType TokenType) (Token, bool) {
	if p.peekType(tokenType) {
		return p.advance(), true
	}
	return nil, false
}

// todo this func is susceptible to pointing at an uninformative token
func (p *Parser) error(message string) {
	currentToken := p.tokens[p.next]
	RP.TokenErrorExit(currentToken, message+"\n")
}

func (p *Parser) fileHeaderIfPresent(statements *[]Stmt) {
	if p.matchAny(FILE_HEADER) {
		previous := p.previous()
		fh := previous.(*FilerHeaderToken)
		*statements = append(*statements, &FileHeader{FhToken: *fh})
	}
}

func (p *Parser) argBlockIfPresent(statements *[]Stmt) {
	if p.matchKeyword(GLOBAL_KEYWORDS, ARGS) {
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
	if p.matchKeyword(ARGS_BLOCK_KEYWORDS, ONE_OF) {
		panic(NOT_IMPLEMENTED)
	}

	identifier := p.consume(IDENTIFIER, "Expected Identifier or keyword")

	if p.peekType(STRING_LITERAL) ||
		p.peekType(IDENTIFIER) ||
		p.peekType(STRING) ||
		p.peekType(INT_LITERAL) ||
		p.peekType(BOOL) {

		return p.argDeclaration(identifier)
	}

	if p.matchKeyword(ARGS_BLOCK_KEYWORDS, REQUIRES) {
		panic(NOT_IMPLEMENTED)
	}

	// todo rest
	panic(NOT_IMPLEMENTED)
}

func (p *Parser) argDeclaration(identifier Token) ArgStmt {
	var renameLiteral Token
	if p.matchAny(STRING_LITERAL) {
		renameLiteral = p.previous()
	}

	var flag Token
	if p.peekTwoAhead().GetType() == IDENTIFIER {
		if p.peekType(IDENTIFIER) {
			// non-int flag
			flag = p.consume(IDENTIFIER, "Expected Flag")
		} else if p.peekType(INT_LITERAL) {
			// int flag
			flag = p.consume(INT_LITERAL, "Expected Flag")
		}
	}

	rslType := p.rslArgType()

	isOptional := false
	var defaultLiteral LiteralOrArray
	if p.matchAny(QUESTION) {
		isOptional = true
	} else if p.matchAny(EQUAL) {
		isOptional = true
		rslTypeEnum := rslType.Type
		defaultLiteralIfPresent, ok := p.literalOrArray(&rslTypeEnum)
		if !ok {
			p.error("Expected default value")
		} else {
			defaultLiteral = defaultLiteralIfPresent
		}
	}

	var argComment *ArgCommentToken
	if p.matchAny(ARG_COMMENT) {
		argComment = p.previous().(*ArgCommentToken)
	}

	return &ArgDeclaration{
		Identifier: identifier,
		Rename:     &renameLiteral,
		Flag:       &flag,
		ArgType:    rslType,
		IsOptional: isOptional,
		Default:    &defaultLiteral,
		Comment:    argComment,
	}
}

func (p *Parser) statement() Stmt {
	if p.matchKeyword(GLOBAL_KEYWORDS, RAD) {
		return p.radBlock(Rad)
	}

	if p.matchKeyword(GLOBAL_KEYWORDS, REQUEST) {
		return p.radBlock(Request)
	}

	if p.matchKeyword(GLOBAL_KEYWORDS, DISPLAY) {
		return p.radBlock(Display)
	}

	if p.peekKeyword(GLOBAL_KEYWORDS, IF) {
		return p.ifStmt()
	}

	if p.peekKeyword(GLOBAL_KEYWORDS, FOR) {
		return p.forStmt()
	}

	if p.peekKeyword(GLOBAL_KEYWORDS, BREAK) {
		if p.nestedForBlockLevel == 0 {
			p.error("Break statement must be inside a for loop")
		}
		return &BreakStmt{BreakToken: p.consumeKeyword(GLOBAL_KEYWORDS, BREAK)}
	}

	if p.peekKeyword(GLOBAL_KEYWORDS, CONTINUE) {
		if p.nestedForBlockLevel == 0 {
			p.error("Continue statement must be inside a for loop")
		}
		return &ContinueStmt{ContinueToken: p.consumeKeyword(GLOBAL_KEYWORDS, CONTINUE)}
	}

	if p.peekKeyword(GLOBAL_KEYWORDS, DELETE) {
		return p.deleteStmt()
	}

	if p.peekKeyword(GLOBAL_KEYWORDS, DEFER) {
		return p.deferStmt()
	}

	if p.isShellCmdNext() {
		return p.shellCmd([]Token{})
	}

	if p.peekTypeSeries(IDENTIFIER, LEFT_PAREN) {
		return p.functionCallStmt()
	}

	return p.assignment()
}

func (p *Parser) radBlock(radType RadBlockType) *RadBlock {
	radToken := p.previous()

	var srcToken *Expr
	if radType == Request || radType == Rad {
		if p.peekType(COLON) {
			p.error(fmt.Sprintf("Expecting url or other source for %v statement", radType))
		}
		expr := p.expr(1)
		srcToken = &expr
	} else {
		p.consume(COLON, fmt.Sprintf("Expecting ':' to immediately follow %q, preceding indented block", radType))
	}

	radStatements := p.radBlockStmts(radType)
	// todo: we should validate, including if a field stmt is not listed but should be (based on other statements),
	//  or if *too many* are listed. When we re-visit static analysis of rad blocks, specifically if-statements,
	//  we add this.
	radBlock := &RadBlock{RadKeyword: radToken, RadType: radType, Source: srcToken, Stmts: radStatements}
	return radBlock
}

func (p *Parser) radBlockStmts(radType RadBlockType) []RadStmt {
	var radStatements []RadStmt
	if p.peekType(COLON) || p.nextNonNewLineTokenIs(INDENT) { // todo i think this breaks if there's a newline between the colon and indent
		if radType != Display {
			p.consume(COLON, fmt.Sprintf("Expecting ':' to precede indented %v block", radType))
		}
		p.consumeNewlines()
		if !p.matchAny(INDENT) {
			p.error(fmt.Sprintf("Expecting indented contents in %s block", radType))
		}
		p.consumeNewlines()
		for !p.matchAny(DEDENT, EOF) {
			radStatements = append(radStatements, p.radStatement(radType))
			p.consumeNewlines()
		}
	}
	return radStatements
}

func (p *Parser) radStatement(radType RadBlockType) RadStmt {
	if p.matchKeyword(RAD_BLOCK_KEYWORDS, FIELDS) {
		return p.radFieldsStatement()
	}

	if p.matchKeyword(RAD_BLOCK_KEYWORDS, SORT) {
		if radType == Request {
			// note: maybe we allow this if we add explicit 'display' statements into request blocks?
			p.error("Sort statement is not allowed in a request block")
		}
		return p.radSortStatement()
	}

	if p.peekKeyword(RAD_BLOCK_KEYWORDS, IF) {
		var cases []RadIfCase
		cases = append(cases, p.radIfCase(radType))
		var elseBlock *[]RadStmt
		for p.peekKeyword(RAD_BLOCK_KEYWORDS, ELSE) {
			p.consumeKeyword(RAD_BLOCK_KEYWORDS, ELSE)
			if p.peekKeyword(RAD_BLOCK_KEYWORDS, IF) {
				cases = append(cases, p.radIfCase(radType))
			} else {
				p.consume(COLON, "Expected ':' after 'else'")
				radStmts := p.radBlockStmts(radType)
				elseBlock = &radStmts
			}
		}
		return RadIfStmt{Cases: cases, ElseBlock: elseBlock}
	}

	// todo modifier
	// todo table fmt
	// todo field fmt
	// todo filtering?

	return p.radFieldMods()
}

func (p *Parser) radIfCase(radType RadBlockType) RadIfCase {
	ifToken := p.consumeKeyword(RAD_BLOCK_KEYWORDS, IF)
	condition := p.expr(1)
	p.consume(COLON, "Expected ':' after rad if condition")
	block := p.radBlockStmts(radType)
	return RadIfCase{IfToken: ifToken, Condition: condition, Body: block}
}

func (p *Parser) radFieldMods() RadStmt {
	identifiers := p.commaSeparatedIdentifiers()
	p.consume(COLON, "Expected ':' to begin field modifier block")
	p.consumeNewlines()
	p.consume(INDENT, "Expected indented block after colon")

	var mods []RadFieldModStmt
	for !p.matchAny(DEDENT, EOF) {
		p.consumeNewlines()
		if p.matchKeyword(RAD_BLOCK_KEYWORDS, TRUNCATE) {
			mods = append(mods, p.truncStmt())
		}
		if p.matchKeyword(RAD_BLOCK_KEYWORDS, COLOR) {
			mods = append(mods, p.colorStmt())
		}
		// todo other field mod stmts
	}
	return &FieldMods{Identifiers: identifiers, Mods: mods}
}

func (p *Parser) truncStmt() RadFieldModStmt {
	truncateToken := p.previous()
	return &Truncate{TruncToken: truncateToken, Value: p.expr(1)}
}

func (p *Parser) colorStmt() RadFieldModStmt {
	colorToken := p.previous()
	return &Color{ColorToken: colorToken, ColorValue: p.expr(1), Regex: p.expr(1)}
}

func (p *Parser) radFieldsStatement() RadStmt {
	var identifiers []Token
	identifiers = append(identifiers, p.identifier())
	for !p.matchAny(NEWLINE, DEDENT, EOF) {
		p.consume(COMMA, "Expected ',' between identifiers for rad fields")
		identifiers = append(identifiers, p.identifier())
	}
	return &Fields{Identifiers: identifiers}
}

func (p *Parser) radSortStatement() RadStmt {
	sortToken := p.previous()
	asc := Asc
	desc := Desc

	if p.matchAny(NEWLINE) {
		return &Sort{SortToken: sortToken, Identifiers: []Token{}, Directions: []SortDir{}, GeneralSort: &asc}
	}

	nextMatchesAsc := p.matchKeyword(RAD_BLOCK_KEYWORDS, ASC)
	nextMatchesDesc := p.matchKeyword(RAD_BLOCK_KEYWORDS, DESC)
	if nextMatchesAsc || nextMatchesDesc {
		p.consume(NEWLINE, "Expected newline after general sort direction")
		dir := lo.Ternary(nextMatchesAsc, asc, desc)
		return &Sort{SortToken: sortToken, Identifiers: []Token{}, Directions: []SortDir{}, GeneralSort: &dir}
	}

	var identifiers []Token
	var directions []SortDir

	for !p.matchAny(NEWLINE, EOF) {
		identifiers = append(identifiers, p.identifier())
		nextMatchesAsc = p.matchKeyword(RAD_BLOCK_KEYWORDS, ASC)
		nextMatchesDesc = p.matchKeyword(RAD_BLOCK_KEYWORDS, DESC)
		if nextMatchesAsc || nextMatchesDesc {
			dir := lo.Ternary(nextMatchesAsc, asc, desc)
			directions = append(directions, dir)
		} else {
			directions = append(directions, asc)
		}
		if !p.peekType(NEWLINE) {
			p.consume(COMMA, "Expected ',' between sort fields")
		}
	}
	return &Sort{SortToken: sortToken, Identifiers: identifiers, Directions: directions, GeneralSort: nil}
}

func (p *Parser) commaSeparatedIdentifiers() []Token {
	var identifiers []Token
	identifiers = append(identifiers, p.identifier())
	for p.matchAny(COMMA) {
		identifiers = append(identifiers, p.identifier())
	}
	return identifiers
}

func (p *Parser) ifStmt() IfStmt {
	var cases []IfCase
	cases = append(cases, p.ifCase())
	var elseBlock *Block
	for p.peekKeyword(GLOBAL_KEYWORDS, ELSE) {
		p.consumeKeyword(GLOBAL_KEYWORDS, ELSE)
		if p.peekKeyword(GLOBAL_KEYWORDS, IF) {
			cases = append(cases, p.ifCase())
		} else {
			p.consume(COLON, "Expected ':' after 'else'")
			block := p.block()
			elseBlock = &block
		}
	}
	return IfStmt{Cases: cases, ElseBlock: elseBlock}
}

func (p *Parser) ifCase() IfCase {
	ifToken := p.consumeKeyword(GLOBAL_KEYWORDS, IF)
	condition := p.expr(1)
	p.consume(COLON, "Expected ':' after if condition")
	block := p.block()
	return IfCase{IfToken: ifToken, Condition: condition, Body: block}
}

func (p *Parser) block() Block {
	var stmts []Stmt
	p.consumeNewlines()
	p.consume(INDENT, "Expected indented block")
	for !p.matchAny(DEDENT) {
		stmts = append(stmts, p.statement())
		p.consumeNewlines()
	}
	return Block{Stmts: stmts}
}

func (p *Parser) forStmt() ForStmt {
	forToken := p.consumeKeyword(GLOBAL_KEYWORDS, FOR)
	identifier1 := p.consume(IDENTIFIER, "Expected identifier after 'for'")
	var identifier2 *Token
	if p.matchAny(COMMA) {
		i := p.consume(IDENTIFIER, "Expected identifier after ','")
		identifier2 = &i
	}
	p.consumeKeyword(GLOBAL_KEYWORDS, IN)
	rangeExpr := p.expr(1)
	p.consume(COLON, "Expected ':' after range expression")
	p.nestedForBlockLevel += 1
	block := p.block()
	p.nestedForBlockLevel -= 1
	return ForStmt{ForToken: forToken, Identifier1: identifier1, Identifier2: identifier2, Range: rangeExpr, Body: block}
}

func (p *Parser) deleteStmt() Stmt {
	deleteToken := p.consumeKeyword(GLOBAL_KEYWORDS, DELETE)
	var vars []VarPath
	vars = append(vars, p.varPath())
	for p.matchAny(COMMA) {
		vars = append(vars, p.varPath())
	}
	return &DeleteStmt{DeleteToken: deleteToken, Vars: vars}
}

func (p *Parser) deferStmt() Stmt {
	deferToken := p.consumeKeyword(GLOBAL_KEYWORDS, DEFER)
	if p.matchAny(COLON) {
		block := p.block()
		return &DeferStmt{DeferToken: deferToken, DeferredStmt: nil, DeferredBlock: &block}
	} else {
		stmt := p.statement()
		return &DeferStmt{DeferToken: deferToken, DeferredStmt: &stmt, DeferredBlock: nil}
	}
}

func (p *Parser) varPath() VarPath {
	identifier := p.consume(IDENTIFIER, "Expected identifier")
	var keys []Expr
	for p.matchAny(LEFT_BRACKET) {
		keys = append(keys, p.expr(1))
		p.consume(RIGHT_BRACKET, "Expected ']' after collection key")
	}
	return VarPath{Identifier: identifier, Keys: keys}
}

func (p *Parser) functionCallStmt() Stmt {
	functionCall := p.functionCall(NO_NUM_RETURN_VALUES_CONSTRAINT)
	return &FunctionStmt{Call: functionCall}
}

func (p *Parser) assignment() Stmt {
	var identifiers []Token
	identifiers = append(identifiers, p.identifier())

	if p.matchAny(LEFT_BRACKET) {
		return p.collectionEntryAssignment(identifiers[0])
	}

	if p.matchAny(PLUS_EQUAL, MINUS_EQUAL, STAR_EQUAL, SLASH_EQUAL) {
		return p.compoundAssignment(identifiers[0], p.previous())
	}

	for !p.matchAny(EQUAL) {
		p.consume(COMMA, "Expected ',' between identifiers")
		identifiers = append(identifiers, p.identifier())
	}

	// finished matching left side of equal sign, now parse right side

	if p.matchKeyword(GLOBAL_KEYWORDS, SWITCH) {
		block := p.switchBlock(identifiers)
		return &SwitchAssignment{Identifiers: identifiers, Block: block}
	}

	if p.isShellCmdNext() {
		return p.shellCmd(identifiers)
	}

	if len(identifiers) == 1 && p.peekType(JSON_PATH_ELEMENT) {
		identifier := identifiers[0]
		return p.jsonPathAssignment(identifier)
	}

	return p.primaryAssignment(identifiers)
}

func (p *Parser) compoundAssignment(identifier Token, operator Token) Stmt {
	expr := p.expr(1)
	switch operator.GetType() {
	case PLUS_EQUAL:
		return &CompoundAssign{Name: identifier, Operator: operator, Value: expr}
	case MINUS_EQUAL:
		return &CompoundAssign{Name: identifier, Operator: operator, Value: expr}
	case STAR_EQUAL:
		return &CompoundAssign{Name: identifier, Operator: operator, Value: expr}
	case SLASH_EQUAL:
		return &CompoundAssign{Name: identifier, Operator: operator, Value: expr}
	default:
		p.error("Invalid compound assignment operator")
		panic(UNREACHABLE)
	}
}

func (p *Parser) collectionEntryAssignment(identifier Token) Stmt {
	// just consumed left bracket
	key := p.expr(1)
	p.consume(RIGHT_BRACKET, "Expected ']' after collection key")
	// todo technically it should not be illegal to have e.g. a[0] as a standalone 'statement',
	//  but we don't allow it here
	operator := p.consumeAny("Expected one of the following operators: [=, +=, -=, *=, /=]",
		EQUAL, PLUS_EQUAL, MINUS_EQUAL, STAR_EQUAL, SLASH_EQUAL)
	value := p.expr(1)
	return &CollectionEntryAssign{Identifier: identifier, Key: key, Operator: operator, Value: value}
}

func (p *Parser) jsonPathAssignment(identifier Token) Stmt {
	element := p.consume(JSON_PATH_ELEMENT, "Expected root json path element").(*JsonPathElementToken)
	var brackets Token
	if isArray := p.matchAny(BRACKETS); isArray {
		brackets = p.previous()
	}
	elements := []JsonPathElement{{token: *element, arrayToken: &brackets}}
	for !p.matchAny(NEWLINE) {
		p.consume(DOT, "Expected '.' to separate json field elements")
		element = p.consume(JSON_PATH_ELEMENT, "Expected json path element after '.'").(*JsonPathElementToken)
		if p.matchAny(BRACKETS) {
			brackets = p.previous()
		}
		elements = append(elements, JsonPathElement{token: *element, arrayToken: &brackets})
	}
	return &JsonPathAssign{Identifier: identifier, Path: JsonPath{elements: elements}}
}

func (p *Parser) switchBlock(identifiers []Token) SwitchBlock {
	switchToken := p.previous()
	var discriminator Token
	if !p.matchAny(COLON) {
		discriminator = p.consume(IDENTIFIER, "Expected discriminator or colon after switch")
		p.consume(COLON, "Expected ':' after switch discriminator")
	} else if len(identifiers) == 0 {
		// this is a switch block without assignment
		p.error("Switch assignments must have a discriminator")
	}

	p.consumeNewlinesMin(1)
	p.consume(INDENT, "Expected indented block after switch")

	var stmts []SwitchStmt
	for !p.matchAny(DEDENT) {
		stmts = append(stmts, p.switchStmt(discriminator != nil, len(identifiers)))
		p.consumeNewlines()
	}
	return SwitchBlock{SwitchToken: switchToken, Discriminator: &discriminator, Stmts: stmts}
}

func (p *Parser) switchStmt(hasDiscriminator bool, expectedNumReturnValues int) SwitchStmt {
	if p.matchKeyword(SWITCH_BLOCK_KEYWORDS, CASE) {
		return p.caseStmt(hasDiscriminator, expectedNumReturnValues)
	}

	if p.matchKeyword(SWITCH_BLOCK_KEYWORDS, DEFAULT) {
		return p.switchDefaultStmt(expectedNumReturnValues)
	}

	p.error("Expected 'case' or 'default' in switch block")
	panic(UNREACHABLE)
}

func (p *Parser) caseStmt(hasDiscriminator bool, expectedNumReturnValues int) SwitchStmt {
	var keys []StringLiteral
	if hasDiscriminator {
		keys = append(keys, p.stringLiteral())
		for !p.matchAny(COLON) {
			p.consume(COMMA, "Expected ',' between case keys")
			keys = append(keys, p.stringLiteral())
		}
	} else {
		p.consume(COLON, "Expected ':' after 'case' when no discriminator")
	}

	var values []Expr
	values = append(values, p.expr(expectedNumReturnValues))
	for !p.matchAny(NEWLINE) {
		p.consume(COMMA, "Expected ',' between case values")
		values = append(values, p.expr(expectedNumReturnValues))
	}

	if len(values) != expectedNumReturnValues {
		// todo technically redundant due to expr taking expectedNumReturnValues?
		p.error(fmt.Sprintf("Expected %d return values, got %d", expectedNumReturnValues, len(values)))
	}

	return &SwitchCase{CaseKeyword: p.previous(), Keys: keys, Values: values}
}

func (p *Parser) switchDefaultStmt(expectedNumReturnValues int) SwitchStmt {
	p.consume(COLON, "Expected ':' after 'default'")
	var values []Expr
	values = append(values, p.expr(expectedNumReturnValues))
	for !p.matchAny(NEWLINE) {
		p.consume(COMMA, "Expected ',' between default values")
		values = append(values, p.expr(expectedNumReturnValues))
	}

	if len(values) != expectedNumReturnValues {
		// todo technically redundant due to expr taking expectedNumReturnValues?
		p.error(fmt.Sprintf("Expected %d return values, got %d", expectedNumReturnValues, len(values)))
	}

	return &SwitchDefault{DefaultKeyword: p.previous(), Values: values}
}

func (p *Parser) primaryAssignment(identifiers []Token) Stmt {
	initializer := p.expr(len(identifiers))
	return &PrimaryAssign{Identifiers: identifiers, Initializer: initializer}
}

func (p *Parser) expr(numExpectedReturnValues int) Expr {
	return p.ternary(numExpectedReturnValues)
}

func (p *Parser) ternary(numExpectedReturnValues int) Expr {
	expr := p.or(numExpectedReturnValues)

	if p.matchAny(QUESTION) {
		questionMark := p.previous()
		if numExpectedReturnValues != 1 {
			// todo technically, there's no reason to now allow multiple return from e.g a function
			p.error(onlyOneReturnValueAllowed)
		}
		trueBranch := p.expr(1)
		p.consume(COLON, "Expected ':' after true branch of ternary operator")
		falseBranch := p.expr(1)
		expr = &Ternary{Condition: expr, QuestionMark: questionMark, True: trueBranch, False: falseBranch}
	}

	return expr
}

func (p *Parser) or(numExpectedReturnValues int) Expr {
	expr := p.and(numExpectedReturnValues)

	for p.matchKeyword(ALL_KEYWORDS, OR) {
		if numExpectedReturnValues != 1 {
			p.error(onlyOneReturnValueAllowed)
		}
		operator := p.previous()
		right := p.and(1)
		expr = &Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) and(numExpectedReturnValues int) Expr {
	expr := p.equality(numExpectedReturnValues)

	for p.matchKeyword(ALL_KEYWORDS, AND) {
		if numExpectedReturnValues != 1 {
			p.error(onlyOneReturnValueAllowed)
		}
		operator := p.previous()
		right := p.equality(1)
		expr = &Logical{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) equality(numExpectedReturnValues int) Expr {
	expr := p.membership(numExpectedReturnValues)

	for p.matchAny(NOT_EQUAL, EQUAL_EQUAL) {
		if numExpectedReturnValues != 1 {
			p.error(onlyOneReturnValueAllowed)
		}
		operator := p.previous()
		right := p.membership(1)
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) membership(numExpectedReturnValues int) Expr {
	expr := p.comparison(numExpectedReturnValues)

	for p.peekKeyword(GLOBAL_KEYWORDS, NOT) || p.peekKeyword(GLOBAL_KEYWORDS, IN) {
		if p.matchKeywordSeries(GLOBAL_KEYWORDS, NOT, IN) {
			if numExpectedReturnValues != 1 {
				p.error(onlyOneReturnValueAllowed)
			}
			notToken := p.lookback(2)

			opToken := NewToken(NOT_IN, "not in", notToken.GetCharStart(), notToken.GetLine(), notToken.GetCharLineStart())
			right := p.comparison(1)
			expr = &Binary{Left: expr, Operator: opToken, Right: right}
		} else if p.matchKeyword(GLOBAL_KEYWORDS, IN) {
			if numExpectedReturnValues != 1 {
				p.error(onlyOneReturnValueAllowed)
			}
			inIdentifierToken := p.previous()
			operator := NewToken(IN, "in", inIdentifierToken.GetCharStart(), inIdentifierToken.GetLine(), inIdentifierToken.GetCharLineStart())
			right := p.comparison(1)
			expr = &Binary{Left: expr, Operator: operator, Right: right}
		}
		// we're here, then we must've matched just 'not' without 'in'.
		// must not be a membership expr, so skip this level.
	}

	return expr
}

func (p *Parser) comparison(numExpectedReturnValues int) Expr {
	expr := p.term(numExpectedReturnValues)

	for p.matchAny(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		if numExpectedReturnValues != 1 {
			p.error(onlyOneReturnValueAllowed)
		}
		operator := p.previous()
		right := p.term(1)
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) term(numExpectedReturnValues int) Expr {
	expr := p.factor(numExpectedReturnValues)

	for p.matchAny(MINUS, PLUS) {
		if numExpectedReturnValues != 1 {
			p.error(onlyOneReturnValueAllowed)
		}
		operator := p.previous()
		right := p.factor(1)
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) factor(numExpectedReturnValues int) Expr {
	expr := p.unary(numExpectedReturnValues)

	for p.matchAny(SLASH, STAR) {
		if numExpectedReturnValues != 1 {
			p.error(onlyOneReturnValueAllowed)
		}
		operator := p.previous()
		right := p.unary(1)
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) unary(numExpectedReturnValues int) Expr {
	if p.matchAny(MINUS, PLUS) || p.matchKeyword(GLOBAL_KEYWORDS, NOT) {
		operator := p.previous()
		right := p.unary(1)
		return &Unary{Operator: operator, Right: right}
	}

	return p.primary(numExpectedReturnValues)
}

func (p *Parser) primary(numExpectedReturnValues int) Expr {
	var expr Expr

	if p.matchAny(LEFT_PAREN) {
		expr = p.expr(1)
		p.consume(RIGHT_PAREN, "Expected ')' after expression")
		expr = &Grouping{Value: expr}
	} else if literal, ok := p.literal(nil); ok {
		expr = ExprLoa{Value: &LoaLiteral{Value: literal}}
	} else if arrayExpr, ok := p.arrayExpr(); ok {
		expr = arrayExpr
	} else if mapExpr, ok := p.mapExpr(); ok {
		expr = mapExpr
	} else if p.matchAny(IDENTIFIER) {
		identifier := p.previous()
		if p.peekType(LEFT_PAREN) {
			// ( after an identifier -> function call
			p.rewind()
			expr = p.functionCall(numExpectedReturnValues)
		} else {
			expr = &Variable{Name: identifier}
		}
	} else {
		p.error("Expected expression")
	}

	for p.peekType(LEFT_BRACKET) {
		openBracket := p.advance()
		if p.matchAny(COLON) {
			if p.matchAny(RIGHT_BRACKET) {
				// a[:]
				expr = &SliceAccess{ListOrString: expr, ColonToken: openBracket, OpenBracketToken: openBracket}
			} else {
				// a[:end]
				endExpr := p.expr(1)
				colonToken := p.consume(RIGHT_BRACKET, "Expected ']' after slice")
				expr = &SliceAccess{ListOrString: expr, Start: nil, ColonToken: colonToken, End: &endExpr, OpenBracketToken: openBracket}
			}
		} else {
			firstExpr := p.expr(1)
			if p.matchAny(RIGHT_BRACKET) {
				// a[idx]
				expr = &CollectionAccess{Collection: expr, Key: firstExpr, OpenBracketToken: openBracket}
			} else if p.matchAny(COLON) {
				if p.matchAny(RIGHT_BRACKET) {
					// a[start:]
					expr = &SliceAccess{ListOrString: expr, Start: &firstExpr, ColonToken: openBracket, OpenBracketToken: openBracket}
				} else {
					// a[start:end]
					endExpr := p.expr(1)
					colonToken := p.consume(RIGHT_BRACKET, "Expected ']' after collection access")
					expr = &SliceAccess{ListOrString: expr, Start: &firstExpr, ColonToken: colonToken, End: &endExpr, OpenBracketToken: openBracket}
				}
			} else {
				p.error("Expected ']' for collection access or ':' for slice")
			}
		}
	}

	return expr
}

func (p *Parser) functionCall(numExpectedReturnValues int) FunctionCall {
	function := p.consume(IDENTIFIER, "Expected function name")
	p.consume(LEFT_PAREN, "Expected '(' after function name")
	var args []Expr
	var namedArgs []NamedArg
	if !p.matchAny(RIGHT_PAREN) {
		args = append(args, p.expr(1))
		for p.matchAny(COMMA) {
			args = append(args, p.expr(1))
		}

		if p.matchAny(EQUAL) {
			// the latest parsed "arg" must be a named arg prior to '='. remove it and re-interpret.
			lastArg := args[len(args)-1]
			args = args[:len(args)-1]
			argToReinterpret, ok := lastArg.(*Variable)
			if !ok {
				p.error("Expected variable for named argument")
			}
			firstNamedArg := NamedArg{Arg: argToReinterpret.Name, Value: p.expr(1)}

			namedArgs = append(namedArgs, firstNamedArg)
			for p.matchAny(COMMA) { // todo would like to support multi-line formatting
				namedArgs = append(namedArgs, p.namedArg())
			}
		}

		p.consume(RIGHT_PAREN, "Expected ')' after function arguments")
	}
	return FunctionCall{Function: function, Args: args, NamedArgs: namedArgs, NumExpectedReturnValues: numExpectedReturnValues}
}

func (p *Parser) namedArg() NamedArg {
	argName := p.consume(IDENTIFIER, "Bug! Expected named argument")
	p.consume(EQUAL, "Expected '=' after named argument")
	argValue := p.expr(1)
	return NamedArg{Arg: argName, Value: argValue}
}

func (p *Parser) arrayExpr() (Expr, bool) {
	if p.matchAny(BRACKETS) {
		return &ArrayExpr{Values: []Expr{}}, true
	}

	if !p.matchAny(LEFT_BRACKET) {
		return nil, false
	}

	if p.matchAny(RIGHT_BRACKET) {
		// technically redundant as it should be one BRACKETS token, but w/e
		return &ArrayExpr{Values: []Expr{}}, true
	}

	expr := p.expr(1)
	if p.peekKeyword(GLOBAL_KEYWORDS, FOR) {
		return p.listComprehension(expr)
	}

	values := []Expr{expr}
	for !p.matchAny(RIGHT_BRACKET) {
		p.consume(COMMA, "Expected ',' between array elements")
		values = append(values, p.expr(1))
	}

	return &ArrayExpr{Values: values}, true
}

func (p *Parser) mapExpr() (Expr, bool) {
	if !p.matchAny(LEFT_BRACE) {
		return nil, false
	}
	openBrace := p.previous()

	if p.matchAny(RIGHT_BRACE) {
		return &MapExpr{Keys: []Expr{}, Values: []Expr{}, OpenBraceToken: openBrace}, true
	}

	keys := []Expr{p.expr(1)}
	p.consume(COLON, "Expected ':' between map key and value")
	values := []Expr{p.expr(1)}
	for !p.matchAny(RIGHT_BRACE) {
		p.consume(COMMA, "Expected ',' between map elements")
		keys = append(keys, p.expr(1))
		p.consume(COLON, "Expected ':' between map key and value")
		values = append(values, p.expr(1))
	}

	return &MapExpr{Keys: keys, Values: values, OpenBraceToken: openBrace}, true
}

func (p *Parser) listComprehension(expr Expr) (Expr, bool) {
	forToken := p.consumeKeyword(GLOBAL_KEYWORDS, FOR)
	identifier1 := p.consume(IDENTIFIER, "Expected identifier after 'for'")
	var identifier2 *Token
	if p.matchAny(COMMA) {
		i := p.consume(IDENTIFIER, "Expected identifier after ','")
		identifier2 = &i
	}
	p.consumeKeyword(GLOBAL_KEYWORDS, IN)
	rangeExpr := p.expr(1)

	if p.matchKeyword(GLOBAL_KEYWORDS, IF) {
		condition := p.expr(1)
		p.consume(RIGHT_BRACKET, "Expected ']' after list comprehension")
		return &ListComprehension{Expression: expr, For: forToken, Identifier1: identifier1, Identifier2: identifier2, Range: rangeExpr, Condition: &condition}, true
	}

	p.consume(RIGHT_BRACKET, "Expected ']' after list comprehension")
	return &ListComprehension{Expression: expr, For: forToken, Identifier1: identifier1, Identifier2: identifier2, Range: rangeExpr, Condition: nil}, true
}

func (p *Parser) literalOrArray(expectedType *RslArgTypeT) (LiteralOrArray, bool) {
	if literal, ok := p.literal(expectedType); ok {
		return &LoaLiteral{Value: literal}, true
	}

	arrayLiteral, ok := p.arrayLiteral(expectedType)
	if ok {
		return &LoaArray{Value: arrayLiteral}, true
	}

	return nil, false
}

func (p *Parser) literal(expectedType *RslArgTypeT) (Literal, bool) {
	if p.peekType(STRING_LITERAL) {
		if expectedType != nil && *expectedType != ArgStringT {
			p.error(fmt.Sprintf("Expected %s literal, got string", expectedType.AsString()))
		}
		return p.stringLiteral(), true
	}

	// todo need to emit bool literal tokens
	if p.peekType(BOOL_LITERAL) {
		if expectedType != nil && *expectedType != ArgBoolT {
			p.error(fmt.Sprintf("Expected %s literal, got bool", expectedType.AsString()))
		}
		return p.boolLiteral(), true
	}

	numMinuses := 0

	for p.peekType(MINUS) || p.peekType(PLUS) {
		if p.matchAny(MINUS) {
			numMinuses += 1
		} else {
			p.matchAny(PLUS)
		}
	}

	isNegative := numMinuses%2 == 1

	if p.peekType(INT_LITERAL) {
		if expectedType != nil && *expectedType != ArgIntT && *expectedType != ArgFloatT {
			p.error(fmt.Sprintf("Expected %s literal, got int", expectedType.AsString()))
		}
		return p.intLiteral(isNegative), true
	}

	if p.peekType(FLOAT_LITERAL) {
		if expectedType != nil && *expectedType != ArgFloatT {
			p.error(fmt.Sprintf("Expected %s literal, got float", expectedType.AsString()))
		}
		return p.floatLiteral(isNegative), true
	}

	return nil, false
}

func (p *Parser) stringLiteral() StringLiteral {
	var stringLiteralTokens []StringLiteralToken
	var inlineExprs []InlineExpr

	for {
		literal := p.consume(STRING_LITERAL, "Expected string literal").(*StringLiteralToken)
		stringLiteralTokens = append(stringLiteralTokens, *literal)
		if literal.FollowedByInlineExpr {
			inlineExprs = append(inlineExprs, p.inlineExpr())
		} else {
			break
		}
	}

	return StringLiteral{Value: stringLiteralTokens, InlineExprs: inlineExprs}
}

// todo support 0 padding e.g. {:010}
func (p *Parser) inlineExpr() InlineExpr {
	expr := p.expr(1)
	rslFormatting := strings.Builder{}
	goFormatting := strings.Builder{}
	builtRslFormat := ""
	builtGoFormat := ""
	isFloatFormat := false
	if p.matchAny(COLON) {
		// imagine we're parsing <10.2
		goFormatting.WriteString("%")

		if p.matchAny(LESS) {
			rslFormatting.WriteString("<")
			goFormatting.WriteString("-")
		} else if p.matchAny(GREATER) {
			rslFormatting.WriteString(">")
			// not required in Go, it defaults to right align
		}

		if p.peekType(FLOAT_LITERAL) {
			width := p.floatLiteral(false)
			rslFormatting.WriteString(width.Value.GetLexeme())
			goFormatting.WriteString(width.Value.GetLexeme())

			isFloatFormat = true
		}

		if p.previous().GetType() != FLOAT_LITERAL {
			if p.matchAny(DOT) {
				isFloatFormat = true
				rslFormatting.WriteString(".")
				goFormatting.WriteString(".")
			}

			if p.peekType(INT_LITERAL) {
				// could be padding width or precision, if dot preceded
				intLiteral := p.intLiteral(false)
				rslFormatting.WriteString(intLiteral.Value.GetLexeme())
				goFormatting.WriteString(intLiteral.Value.GetLexeme())
			} else if isFloatFormat {
				p.error("Expected precision int literal after dot for inline formatting")
			}
		}

		builtRslFormat = rslFormatting.String()
		if len(builtRslFormat) == 0 {
			// nothing was specified
			p.error("Expected formatting after colon in inline expression")
		}

		if !p.peekType(STRING_LITERAL) {
			p.error(fmt.Sprintf("Unexpected token %q after inline expression formatting %q",
				p.peek().GetLexeme(), builtRslFormat))
		}

		builtGoFormat = goFormatting.String()
	}
	if builtRslFormat == "" {
		return InlineExpr{Expression: expr, Formatting: nil}
	} else {
		return InlineExpr{
			Expression: expr,
			Formatting: &InlineExprFormat{
				RslFormat: builtRslFormat, GoFormat: builtGoFormat, IsFloatFormat: isFloatFormat,
			},
		}
	}
}

func (p *Parser) intLiteral(isNegativeViaOtherTokens bool) IntLiteral {
	literal := p.consume(INT_LITERAL, "Expected int literal").(*IntLiteralToken)
	return IntLiteral{Value: *literal, IsNegative: isNegativeViaOtherTokens}
}

func (p *Parser) floatLiteral(isNegativeViaOtherTokens bool) FloatLiteral {
	literal := p.consume(FLOAT_LITERAL, "Expected float literal").(*FloatLiteralToken)
	return FloatLiteral{Value: *literal, IsNegative: isNegativeViaOtherTokens}
}

func (p *Parser) boolLiteral() BoolLiteral {
	literal := p.consume(BOOL_LITERAL, "Expected bool literal").(*BoolLiteralToken)
	return BoolLiteral{Value: *literal}
}

func (p *Parser) arrayLiteral(expectedType *RslArgTypeT) (ArrayLiteral, bool) {
	if p.peekType(BRACKETS) {
		return p.mixedArrayLiteral(nil), true
	}

	if !p.matchAny(LEFT_BRACKET) {
		return nil, false
	}
	p.rewind() // rewind the left bracket

	if expectedType == nil || *expectedType == ArgMixedArrayT {
		return p.mixedArrayLiteral(nil), true
	}

	var unwrappedType RslArgTypeT
	switch *expectedType {
	case ArgStringArrayT:
		unwrappedType = ArgStringT
	case ArgIntArrayT:
		unwrappedType = ArgIntT
	case ArgFloatArrayT:
		unwrappedType = ArgFloatT
	case ArgBoolArrayT:
		unwrappedType = ArgBoolT
	default:
		p.error("Invalid array type " + expectedType.AsString())
		panic(UNREACHABLE)
	}

	return p.mixedArrayLiteral(&unwrappedType), true
}

func (p *Parser) mixedArrayLiteral(expectedType *RslArgTypeT) MixedArrayLiteral {
	if p.matchAny(BRACKETS) {
		return MixedArrayLiteral{Values: []LiteralOrArray{}}
	}

	p.consume(LEFT_BRACKET, "Expected '[' to start array")

	var values []LiteralOrArray
	for !p.matchAny(RIGHT_BRACKET) {
		literal, ok := p.literalOrArray(expectedType)
		if !ok {
			p.error("Expected literalOrArray in mixed array")
		}
		values = append(values, literal)
		if !p.peekType(RIGHT_BRACKET) {
			p.consume(COMMA, "Expected ',' between array elements")
		}
	}
	return MixedArrayLiteral{Values: values}
}

func (p *Parser) rslArgType() RslArgType {
	var argType Token
	var rslTypeEnum RslArgTypeT
	if p.matchKeyword(ARGS_BLOCK_KEYWORDS, ARRAY) { // todo technically this is used for typing in non-arg contexts too
		argType = p.previous()
		if p.matchAny(BRACKETS) {
			p.error("Brackets cannot follow 'array' type, just use 'array'.")
		} else {
			rslTypeEnum = ArgMixedArrayT
		}
	} else if p.matchKeyword(ARGS_BLOCK_KEYWORDS, STRING) {
		argType = p.previous()
		if p.matchAny(BRACKETS) {
			rslTypeEnum = ArgStringArrayT
		} else {
			rslTypeEnum = ArgStringT
		}
	} else if p.matchKeyword(ARGS_BLOCK_KEYWORDS, INT) {
		argType = p.previous()
		if p.matchAny(BRACKETS) {
			rslTypeEnum = ArgIntArrayT
		} else {
			rslTypeEnum = ArgIntT
		}
	} else if p.matchKeyword(ARGS_BLOCK_KEYWORDS, FLOAT) {
		argType = p.previous()
		if p.matchAny(BRACKETS) {
			rslTypeEnum = ArgFloatArrayT
		} else {
			rslTypeEnum = ArgFloatT
		}
	} else if p.matchKeyword(ARGS_BLOCK_KEYWORDS, BOOL) {
		argType = p.previous()
		if p.matchAny(BRACKETS) {
			rslTypeEnum = ArgBoolArrayT
		} else {
			rslTypeEnum = ArgBoolT
		}
	} else {
		p.error("Expected arg type")
	}
	return RslArgType{Token: argType, Type: rslTypeEnum}
}

func (p *Parser) isShellCmdNext() bool {
	return p.peekKeyword(GLOBAL_KEYWORDS, UNSAFE) || p.peekType(DOLLAR)
}

func (p *Parser) shellCmd(identifiers []Token) Stmt {
	var unsafeToken *Token
	if p.matchKeyword(GLOBAL_KEYWORDS, UNSAFE) {
		token := p.previous()
		unsafeToken = &token
	}
	dollarToken := p.consume(DOLLAR, "Expected '$' to start shell command")

	var bangToken *Token
	if p.matchAny(EXCLAMATION) {
		token := p.previous()
		bangToken = &token
	}

	shellCmdExpr := p.expr(1)

	p.consumeNewlines()

	var failBlock *Block
	if p.matchKeyword(GLOBAL_KEYWORDS, FAIL) {
		p.consume(COLON, "Expected ':' after fail keyword")
		p.consumeNewlines()
		if p.peekType(INDENT) {
			b := p.block()
			failBlock = &b
		}
	}

	var recoverBlock *Block
	if p.matchKeyword(GLOBAL_KEYWORDS, RECOVER) {
		p.consume(COLON, "Expected ':' after recover keyword")
		p.consumeNewlines()
		if p.peekType(INDENT) {
			b := p.block()
			recoverBlock = &b
		}
	}

	if bangToken != nil && failBlock != nil {
		p.error("Critical shell command (!) cannot have a 'fail' block")
	}

	if bangToken != nil && recoverBlock != nil {
		p.error("Critical shell command (!) cannot have a 'recover' block")
	}

	if bangToken != nil && unsafeToken != nil {
		p.error("Critical shell command (!) cannot also be 'unsafe'")
	}

	if unsafeToken != nil && failBlock != nil {
		p.error("unsafe shell command cannot have a 'fail' block")
	}

	if unsafeToken != nil && recoverBlock != nil {
		p.error("unsafe shell command cannot have a 'recover' block")
	}

	if failBlock != nil && recoverBlock != nil {
		p.error("Cannot have both 'fail' and 'recover' blocks for shell command")
	}

	if bangToken == nil && failBlock == nil && recoverBlock == nil && unsafeToken == nil {
		p.error("Expected unsafe shell command to have either a 'fail' or a 'recover' block")
	}

	return &ShellCmd{
		Identifiers:  identifiers,
		Unsafe:       unsafeToken,
		Dollar:       dollarToken,
		CmdExpr:      shellCmdExpr,
		Bang:         bangToken,
		FailBlock:    failBlock,
		RecoverBlock: recoverBlock,
	}
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
	p.consumeNewlinesMin(0)
}

func (p *Parser) consumeNewlinesMin(min int) {
	matched := 0
	for !p.isAtEnd() && p.matchAny(NEWLINE) {
		// throw away
		matched++
	}
	if matched < min && !p.isAtEnd() {
		p.error("Expected newline")
	}
}

func (p *Parser) matchKeyword(keywords map[string]TokenType, tokenType TokenType) bool {
	return p.matchKeywordSeries(keywords, tokenType)
}

func (p *Parser) matchKeywordSeries(keywords map[string]TokenType, tokenType ...TokenType) bool {
	next := p.peek()
	if next.GetType() != IDENTIFIER {
		return false
	}
	for i, t := range tokenType {
		if keyword, ok := keywords[p.peek().GetLexeme()]; ok {
			if keyword != t {
				for j := 0; j < i; j++ {
					p.rewind()
				}
				return false
			}
			p.advance()
		} else {
			for j := 0; j < i; j++ {
				p.rewind()
			}
			return false
		}
	}
	return true
}

func (p *Parser) consumeKeyword(keywords map[string]TokenType, tokenType TokenType) Token {
	if !p.matchKeyword(keywords, tokenType) {
		p.error(fmt.Sprintf("Expected keyword %s", tokenType))
		panic(UNREACHABLE)
	}
	return p.previous()
}

func (p *Parser) consumeAny(errMsg string, expected ...TokenType) Token {
	if p.matchAny(expected...) {
		return p.previous()
	}
	p.error(errMsg)
	panic(UNREACHABLE)
}

func (p *Parser) peekKeyword(keywords map[string]TokenType, expectedKeyword TokenType) bool {
	if p.matchKeyword(keywords, expectedKeyword) {
		p.rewind()
		return true
	}
	return false
}

func (p *Parser) nextNonNewLineTokenIs(expected TokenType) bool {
	for !p.isAtEnd() && p.matchAny(NEWLINE) {
		// throw away
	}
	return p.peekType(expected)
}
