package check

import (
	"fmt"
	"strings"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// reservedKeywords maps reserved keywords to their usage context.
// Note: rad/request/display are now context-sensitive in the grammar (GH #78 fix),
// so they can be used as identifiers and don't need error handling here.
var reservedKeywords = map[string]string{
	"args": "args blocks",
}

// missingNodeMessages maps MISSING node kinds to human-readable error messages.
// Uses constants from rl package where available; literal tokens (`:`, `)`, etc.)
// are used directly since they're tree-sitter's canonical names.
var missingNodeMessages = map[string]string{
	// Punctuation
	":":  "Expected ':'",
	")":  "Expected ')'",
	"(":  "Expected '('",
	"]":  "Expected ']'",
	"[":  "Expected '['",
	"}":  "Expected '}'",
	"{":  "Expected '{'",
	",":  "Expected ','",
	"=":  "Expected '='",
	"->": "Expected '->'",

	// Identifiers and expressions
	rl.K_IDENTIFIER: "Expected identifier",
	rl.K_EXPR:       "Expected expression",
	rl.K_VAR_PATH:   "Expected variable",

	// Literals
	rl.K_STRING: "Expected string",
	rl.K_INT:    "Expected integer",
	rl.K_FLOAT:  "Expected number",
	rl.K_BOOL:   "Expected boolean",
	rl.K_NULL:   "Expected null",

	// Types
	"type":           "Expected type",
	rl.K_STRING_TYPE: "Expected 'str' type",
	rl.K_INT_TYPE:    "Expected 'int' type",
	rl.K_FLOAT_TYPE:  "Expected 'float' type",
	rl.K_BOOL_TYPE:   "Expected 'bool' type",
	rl.K_VOID_TYPE:   "Expected 'void' type",
	rl.K_LIST_TYPE:   "Expected list type",
	rl.K_MAP_TYPE:    "Expected map type",
	rl.K_ANY_TYPE:    "Expected 'any' type",

	// Structure
	"newline": "Expected newline",
	"indent":  "Expected indented block",
	"dedent":  "Unexpected indentation",

	// Keywords
	"if":       "Expected 'if'",
	"else":     "Expected 'else'",
	"for":      "Expected 'for'",
	"while":    "Expected 'while'",
	"in":       "Expected 'in'",
	"fn":       "Expected 'fn'",
	"return":   "Expected 'return'",
	"break":    "Expected 'break'",
	"continue": "Expected 'continue'",
	"switch":   "Expected 'switch'",
	"case":     "Expected 'case'",
	"default":  "Expected 'default'",
	"defer":    "Expected 'defer'",
	"catch":    "Expected 'catch'",
	"pass":     "Expected 'pass'",
	"del":      "Expected 'del'",
}

// missingKindToErrorCode maps MISSING node kinds to specific error
// codes. Only the codes that empirically fire are mapped here -
// tree-sitter's error recovery emits ERROR nodes for most of the
// shapes that the retired 10003-10007/10010-10017/10019 codes
// were designed to catch, so those map entries would never be
// consulted. The active mappings stay because tree-sitter does
// produce MISSING nodes for them (`:` after an `if` header is the
// canonical case).
var missingKindToErrorCode = map[string]rl.Error{
	":":       rl.ErrMissingColon,
	"indent":  rl.ErrMissingIndent,
	"newline": rl.ErrInvalidSyntax,
}

// parentContextMessages provides more specific messages based on the parent node kind.
// Key is "parent_kind:missing_kind" or just "parent_kind" for general parent context.
var parentContextMessages = map[string]string{
	// Function definitions
	rl.K_FN_NAMED + ":" + rl.K_IDENTIFIER: "Expected function name",
	rl.K_FN_NAMED + ":->":                 "Expected '->' before return type",
	rl.K_FN_NAMED + ":type":               "Expected return type after '->'",
	rl.K_FN_NAMED + ":":                   "Expected ':' after function signature",
	rl.K_FN_LAMBDA + ":->":                "Expected '->' before return type",
	rl.K_FN_LAMBDA + ":type":              "Expected return type after '->'",

	// For loops
	rl.K_FOR_LOOP + ":" + rl.K_IDENTIFIER: "Expected loop variable name",
	rl.K_FOR_LOOP + ":in":                 "Expected 'in' keyword",
	rl.K_FOR_LOOP + ":" + rl.K_EXPR:       "Expected iterable expression",
	rl.K_FOR_LOOP + ":":                   "Expected ':' after for loop header",

	// While loops
	rl.K_WHILE_LOOP + ":" + rl.K_EXPR: "Expected condition expression",
	rl.K_WHILE_LOOP + ":":             "Expected ':' after while condition",

	// If statements
	rl.K_IF_STMT + ":" + rl.K_EXPR: "Expected condition expression",
	rl.K_IF_STMT + ":":             "Expected ':' after condition",

	// Assignments
	rl.K_ASSIGN + ":" + rl.K_EXPR: "Expected value after '='",
	rl.K_ASSIGN + ":=":            "Expected '=' for assignment",

	// Switch statements
	rl.K_SWITCH_STMT + ":" + rl.K_EXPR: "Expected value to switch on",
	rl.K_SWITCH_STMT + ":":             "Expected ':' after switch expression",

	// Arg block
	rl.K_ARG_BLOCK + ":" + rl.K_IDENTIFIER: "Expected argument name",
	rl.K_ARG_BLOCK + ":type":               "Expected argument type",

	// Function calls
	rl.K_CALL + ":(":            "Expected '(' after function name",
	rl.K_CALL + ":)":            "Expected ')' to close function call",
	rl.K_CALL + ":" + rl.K_EXPR: "Expected argument expression",

	// Lists
	rl.K_LIST + ":]":            "Expected ']' to close list",
	rl.K_LIST + ":" + rl.K_EXPR: "Expected list element",

	// Maps
	rl.K_MAP + ":}":            "Expected '}' to close map",
	rl.K_MAP + ":":             "Expected ':' between key and value",
	rl.K_MAP + ":" + rl.K_EXPR: "Expected map value",

	// Return statement
	rl.K_RETURN_STMT + ":" + rl.K_EXPR: "Expected return value",
}

// GenerateErrorMessage creates a specific error message for an invalid node.
// Returns the message, error code, and optional suggestion.
func GenerateErrorMessage(node *ts.Node, src string) (msg string, code rl.Error, suggestion *string) {
	if node.IsMissing() {
		return generateMissingNodeMessage(node)
	}
	if node.IsError() {
		return generateErrorNodeMessage(node, src)
	}
	return "Invalid syntax", rl.ErrInvalidSyntax, nil
}

// generateMissingNodeMessage generates an error message for a MISSING node.
func generateMissingNodeMessage(node *ts.Node) (string, rl.Error, *string) {
	kind := node.Kind()

	// Try parent-context-specific message first
	if parent := node.Parent(); parent != nil {
		parentKind := parent.Kind()

		// Try "parent:child" lookup
		contextKey := parentKind + ":" + kind
		if msg, ok := parentContextMessages[contextKey]; ok {
			code := getErrorCodeForMissingKind(kind)
			return msg, code, nil
		}
	}

	// Check explicit message map
	if msg, ok := missingNodeMessages[kind]; ok {
		code := getErrorCodeForMissingKind(kind)
		return msg, code, nil
	}

	// Fallback: humanize the node kind
	humanized := humanizeNodeKind(kind)
	return fmt.Sprintf("Expected %s", humanized), rl.ErrInvalidSyntax, nil
}

// generateErrorNodeMessage generates an error message for an ERROR node.
func generateErrorNodeMessage(node *ts.Node, src string) (string, rl.Error, *string) {
	// Extract the error content from source
	startByte := node.StartByte()
	endByte := node.EndByte()

	if startByte >= uint(len(src)) || endByte > uint(len(src)) {
		return "Invalid syntax", rl.ErrInvalidSyntax, nil
	}

	errorContent := src[startByte:endByte]
	trimmedContent := strings.TrimSpace(errorContent)

	// Heuristic: missing colon after a block-opening keyword. Fires
	// before the generic "unexpected token" so users see RAD10002
	// instead of RAD10009 for the very common shape.
	if msg, code, suggestion := checkMissingColon(errorContent); msg != "" {
		return msg, code, suggestion
	}

	// Heuristic: a block-opening header (`if x:`) followed by an
	// unindented next line. Tree-sitter doesn't emit a MISSING
	// "indent" for this shape, so we recognise it from the ERROR
	// content directly and surface RAD10018.
	if msg, code, suggestion := checkMissingIndent(errorContent); msg != "" {
		return msg, code, suggestion
	}

	// Heuristic 1: Check for unterminated string
	if msg, code, suggestion := checkUnterminatedString(errorContent, trimmedContent); msg != "" {
		return msg, code, suggestion
	}

	// Heuristic: bare quote ERROR token whose source position has no
	// matching close on the same line. checkUnterminatedString only
	// catches multi-byte quote tokens; this catches the case where
	// tree-sitter splits the open quote into a one-char ERROR and
	// the rest of the line into another ERROR.
	if msg, code, suggestion := checkUnterminatedQuoteToken(node, src, trimmedContent); msg != "" {
		return msg, code, suggestion
	}

	// Heuristic: `null` in a type union. Other-language users often
	// reach for `int|null` (TypeScript / Pyright spelling); Rad has a
	// single canonical form `int?`. Point them at it before they
	// hunt for a `null_type` declaration that doesn't exist.
	if msg, code, suggestion := checkNullInUnion(node, trimmedContent); msg != "" {
		return msg, code, suggestion
	}

	// Heuristic 2: Check for missing operator between values
	if msg, code, suggestion := checkMissingOperator(node, src, trimmedContent); msg != "" {
		return msg, code, suggestion
	}

	// Heuristic 3: Check for keyword in wrong context
	if msg, code, suggestion := checkKeywordInWrongContext(node, trimmedContent); msg != "" {
		return msg, code, suggestion
	}

	// Check for reserved keywords - extract first token from error content
	firstToken := strings.Fields(trimmedContent)
	if len(firstToken) > 0 {
		if context, isReserved := reservedKeywords[firstToken[0]]; isReserved {
			msg := fmt.Sprintf("'%s' is reserved (used in %s)", firstToken[0], context)
			suggestion := "use a different variable name"
			return msg, rl.ErrReservedKeyword, &suggestion
		}
	}

	// Try to get context from parent
	parent := node.Parent()
	if parent != nil {
		return generateContextualErrorMessage(node, parent, trimmedContent)
	}

	return "Invalid syntax", rl.ErrInvalidSyntax, nil
}

// checkUnterminatedString detects unterminated string literals.
// Pattern: string starts with quote but contains newline or doesn't end with matching quote.
// checkNullInUnion detects the `T | null` pattern, common in
// TypeScript / Pyright but invalid in Rad - the grammar doesn't list
// `null` as a leaf type. Rad has a single canonical spelling for
// nullable types: `T?`. Without this heuristic, users get
// "Unexpected '|null'" (fn_type case) or "Invalid function syntax"
// (fn_named / fn_lambda cases) and have to figure out the right form
// on their own; the suggestion turns that into "use 'T?' instead."
//
// Triggered when the ERROR node sits under any context where a type
// union is legal AND its content contains the bare word `null`:
//
//   - fn_param_or_return_type: the union site itself.
//   - typed_assign: parent above when the ERROR tails a typed-local
//     declaration (`x: int|null = ...`).
//   - fn_named, fn_lambda: parent above when the ERROR shows up in
//     a param-position annotation (`fn f(x: int|null):`,
//     `fn(x: int|null) <body>`). The parser bails before it can
//     wrap the union in a fn_param_or_return_type, so the ERROR
//     attaches to the fn node directly.
//   - fn_type: parent above when the ERROR shows up inside a
//     fn-type annotation (`fn(int|null) -> int`).
//
// We keep the parent-kind check strict so a stray `null` token
// elsewhere in source doesn't get this suggestion.
func checkNullInUnion(node *ts.Node, trimmed string) (string, rl.Error, *string) {
	// Walk up through any wrapping ERROR nodes - when the parser
	// misrecovers in a return-type position the whole `fn ... -> T:
	// body` shape collapses into an outer ERROR, so the immediate
	// parent of our content-bearing ERROR is itself an ERROR. Climb
	// until we hit something concrete, then test that.
	parent := node.Parent()
	climbed := false
	for parent != nil && parent.IsError() {
		parent = parent.Parent()
		climbed = true
	}
	if parent == nil {
		return "", "", nil
	}
	switch parent.Kind() {
	case "fn_param_or_return_type",
		rl.K_TYPED_ASSIGN,
		rl.K_FN_NAMED,
		rl.K_FN_LAMBDA,
		rl.K_FN_TYPE:
		// allowed contexts
	case "source_file":
		// Accept source_file only when we climbed through at least
		// one ERROR layer. That picks up the inner ERROR of a nested
		// fn-header collapse (`fn f() -> int|null:` shape) without
		// also firing on the surrounding outer ERROR - the outer one
		// would emit a duplicate hint on the whole header span.
		if !climbed {
			return "", "", nil
		}
	default:
		return "", "", nil
	}
	// Look for `null` as a standalone token adjacent to a `|`. We
	// can't just substring-match: identifiers like `nullify` should
	// not trigger this, and the ERROR content can drag in surrounding
	// source (newlines, body fragments) when the parser misrecovers.
	if !containsNullTypeToken(trimmed) {
		return "", "", nil
	}
	suggestion := "Rad spells nullable types as 'T?' (e.g. 'int?', not 'int|null')"
	return "'null' is not a valid type in a union",
		rl.ErrUnexpectedToken,
		&suggestion
}

// containsNullTypeToken reports whether `s` contains `null` used as
// a type token: `|null`, `null|`, or a bare `null` flanked by union
// punctuation / whitespace / type-position delimiters. We require an
// adjacent `|` so `null = 5` (the value literal) doesn't trigger the
// suggestion - only the union-position spelling does.
func containsNullTypeToken(s string) bool {
	for i := 0; i+4 <= len(s); i++ {
		if s[i:i+4] != "null" {
			continue
		}
		left := byte(0)
		if i > 0 {
			left = s[i-1]
		}
		right := byte(0)
		if i+4 < len(s) {
			right = s[i+4]
		}
		if !isTypeTokenBoundary(left) || !isTypeTokenBoundary(right) {
			continue
		}
		// Must be adjacent to a `|` somewhere - either immediately
		// before or after, or as the only token after stripping
		// trailing punctuation (the `null|` or `|null` shapes).
		if left == '|' || right == '|' {
			return true
		}
	}
	return false
}

// isTypeTokenBoundary reports whether `b` can separate `null` from
// surrounding text without making `null` part of an identifier. Word
// chars (letters, digits, underscore) glue identifiers together; any
// other byte (or zero for string edges) is a boundary.
func isTypeTokenBoundary(b byte) bool {
	if b == 0 {
		return true
	}
	if b >= 'a' && b <= 'z' {
		return false
	}
	if b >= 'A' && b <= 'Z' {
		return false
	}
	if b >= '0' && b <= '9' {
		return false
	}
	if b == '_' {
		return false
	}
	return true
}

// blockOpeningKeywords is the set of tokens that introduce a block
// terminated by a colon. Used by checkMissingColon /
// checkMissingIndent to recognise the shape of a half-formed
// header (`if x` without the trailing `:`).
var blockOpeningKeywords = map[string]bool{
	"if":     true,
	"elif":   true,
	"else":   true,
	"for":    true,
	"while":  true,
	"fn":     true,
	"switch": true,
	"case":   true,
	"defer":  true,
}

// checkMissingColon detects the `<block-keyword> <expr>` shape with
// no trailing `:`. Tree-sitter recovers from the missing colon by
// consuming following statements into a large ERROR rather than
// emitting MISSING ':' - so we look at the ERROR content directly.
// The user sees RAD10002 ("Missing colon...") instead of the
// generic RAD10009.
//
// Conservative: only fires when the ERROR content's first token is
// one of blockOpeningKeywords AND the first line (before any
// newline) contains no colon. False positives are rare in this
// narrow shape.
func checkMissingColon(raw string) (string, rl.Error, *string) {
	if raw == "" {
		return "", "", nil
	}
	firstLine := raw
	if i := strings.IndexByte(raw, '\n'); i >= 0 {
		firstLine = raw[:i]
	}
	trimmedLine := strings.TrimSpace(firstLine)
	if trimmedLine == "" {
		return "", "", nil
	}
	tokens := strings.Fields(trimmedLine)
	if len(tokens) == 0 || !blockOpeningKeywords[tokens[0]] {
		return "", "", nil
	}
	if strings.ContainsRune(trimmedLine, ':') {
		return "", "", nil
	}
	suggestion := "add ':' at the end of the line"
	msg := fmt.Sprintf("Missing ':' after '%s' header", tokens[0])
	return msg, rl.ErrMissingColon, &suggestion
}

// checkMissingIndent detects `<block-keyword> ... :\n<unindented>` -
// a header that opens a block but isn't followed by an indented
// body. Tree-sitter recovers by collapsing the header and the
// following stmt into one ERROR, so we recognise the shape from
// the content.
func checkMissingIndent(raw string) (string, rl.Error, *string) {
	nl := strings.IndexByte(raw, '\n')
	if nl < 0 {
		return "", "", nil
	}
	firstLine := strings.TrimRight(raw[:nl], " \t")
	if !strings.HasSuffix(firstLine, ":") {
		return "", "", nil
	}
	tokens := strings.Fields(strings.TrimSpace(firstLine))
	if len(tokens) == 0 || !blockOpeningKeywords[tokens[0]] {
		return "", "", nil
	}
	// The next non-empty line must NOT be more indented than the
	// header line. Find the header's indent first.
	rest := raw[nl+1:]
	for len(rest) > 0 {
		line := rest
		if i := strings.IndexByte(rest, '\n'); i >= 0 {
			line = rest[:i]
			rest = rest[i+1:]
		} else {
			rest = ""
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Header indent: count leading spaces/tabs of the original
		// first line (before TrimRight).
		headerIndent := leadingWhitespace(raw[:nl])
		bodyIndent := leadingWhitespace(line)
		if len(bodyIndent) <= len(headerIndent) {
			suggestion := "indent the body of the block"
			msg := fmt.Sprintf("Missing indent after '%s:' block header", tokens[0])
			return msg, rl.ErrMissingIndent, &suggestion
		}
		return "", "", nil
	}
	return "", "", nil
}

// leadingWhitespace returns the leading space/tab prefix of s.
func leadingWhitespace(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return s[:i]
		}
	}
	return s
}

// checkUnterminatedQuoteToken handles the case where tree-sitter
// records the opening quote of an unterminated string as a one-char
// ERROR token (separate from the rest of the line). The existing
// checkUnterminatedString already covers multi-char string ERRORs;
// this is the missing piece for the one-char shape.
//
// We look past the ERROR's end byte in `src`: if no matching close
// quote appears before the next newline, the string is unterminated.
func checkUnterminatedQuoteToken(node *ts.Node, src, trimmed string) (string, rl.Error, *string) {
	if len(trimmed) != 1 {
		return "", "", nil
	}
	quote := trimmed[0]
	if quote != '"' && quote != '\'' && quote != '`' {
		return "", "", nil
	}
	end := int(node.EndByte())
	for i := end; i < len(src); i++ {
		if src[i] == '\n' {
			break
		}
		if src[i] == quote {
			// found a matching close on the same line; not
			// unterminated.
			return "", "", nil
		}
	}
	quoteName := "double"
	switch quote {
	case '\'':
		quoteName = "single"
	case '`':
		quoteName = "backtick"
	}
	msg := fmt.Sprintf("Unterminated string literal (missing closing %s quote)", quoteName)
	suggestion := fmt.Sprintf("add closing %c before end of line", quote)
	return msg, rl.ErrUnterminatedString, &suggestion
}

func checkUnterminatedString(raw, trimmed string) (string, rl.Error, *string) {
	if len(trimmed) == 0 {
		return "", "", nil
	}

	// Check for string that starts with a quote
	firstChar := trimmed[0]
	if firstChar != '"' && firstChar != '\'' && firstChar != '`' {
		return "", "", nil
	}

	// If the raw content contains a newline, it's likely unterminated
	if strings.Contains(raw, "\n") {
		quoteType := "double"
		if firstChar == '\'' {
			quoteType = "single"
		} else if firstChar == '`' {
			quoteType = "backtick"
		}
		msg := fmt.Sprintf("Unterminated string literal (missing closing %s quote)", quoteType)
		suggestion := fmt.Sprintf("add closing %c at end of string", firstChar)
		return msg, rl.ErrUnterminatedString, &suggestion
	}

	// Check if string doesn't end with matching quote
	if len(trimmed) > 1 {
		lastChar := trimmed[len(trimmed)-1]
		if lastChar != firstChar {
			msg := "Unterminated string literal"
			suggestion := fmt.Sprintf("add closing %c at end of string", firstChar)
			return msg, rl.ErrUnterminatedString, &suggestion
		}
	}

	return "", "", nil
}

// checkMissingOperator detects two consecutive values without an operator.
// Pattern: identifier/literal followed by another identifier/literal with no operator.
func checkMissingOperator(node *ts.Node, src, content string) (string, rl.Error, *string) {
	// Look for patterns like "foo bar" or "123 456" (two values with just whitespace)
	tokens := strings.Fields(content)
	if len(tokens) != 2 {
		return "", "", nil
	}

	// Check if both tokens look like values (identifiers, numbers, or strings)
	first, second := tokens[0], tokens[1]
	if looksLikeValue(first) && looksLikeValue(second) {
		msg := fmt.Sprintf("Missing operator between '%s' and '%s'", first, second)
		suggestion := "add an operator (+, -, *, /, ==, etc.) between values"
		return msg, rl.ErrMissingOperator, &suggestion
	}

	return "", "", nil
}

// looksLikeValue returns true if the token looks like a value (identifier, number, string).
func looksLikeValue(token string) bool {
	if len(token) == 0 {
		return false
	}

	// Looks like a number
	if isDigit(token[0]) || (token[0] == '-' && len(token) > 1 && isDigit(token[1])) {
		return true
	}

	// Looks like an identifier (starts with letter or underscore)
	if isLetter(token[0]) || token[0] == '_' {
		// But not if it's a keyword
		if isKeyword(token) {
			return false
		}
		return true
	}

	// Looks like a string literal
	if token[0] == '"' || token[0] == '\'' || token[0] == '`' {
		return true
	}

	// Looks like a boolean or null
	if token == "true" || token == "false" || token == "null" {
		return true
	}

	return false
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isKeyword returns true if the token is a Rad keyword.
func isKeyword(token string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true, "in": true,
		"fn": true, "return": true, "break": true, "continue": true,
		"switch": true, "case": true, "default": true, "defer": true,
		"catch": true, "pass": true, "del": true, "args": true,
		"rad": true,
		"not": true, "and": true, "or": true,
	}
	return keywords[token]
}

// checkKeywordInWrongContext detects keywords used where they don't belong.
func checkKeywordInWrongContext(node *ts.Node, content string) (string, rl.Error, *string) {
	tokens := strings.Fields(content)
	if len(tokens) == 0 {
		return "", "", nil
	}

	firstToken := tokens[0]

	// Check specific keyword misuse patterns
	switch firstToken {
	case "else":
		// "else" without preceding "if"
		if parent := node.Parent(); parent != nil {
			if parent.Kind() != rl.K_IF_STMT {
				msg := "'else' without matching 'if'"
				suggestion := "add an 'if' statement before 'else'"
				return msg, rl.ErrKeywordMisuse, &suggestion
			}
		}
	case "case", "default":
		// "case"/"default" outside switch
		if !isInsideSwitch(node) {
			msg := fmt.Sprintf("'%s' outside of switch statement", firstToken)
			suggestion := "use 'case' and 'default' only inside 'switch'"
			return msg, rl.ErrKeywordMisuse, &suggestion
		}
	case "break", "continue":
		// These are already handled elsewhere, but we can provide a better message
		if !isInsideLoop(node) {
			msg := fmt.Sprintf("'%s' outside of loop", firstToken)
			suggestion := fmt.Sprintf("'%s' can only be used inside 'for' or 'while' loops", firstToken)
			return msg, rl.ErrKeywordMisuse, &suggestion
		}
	case "return", "yield":
		if !isInsideFunction(node) {
			msg := fmt.Sprintf("'%s' outside of function", firstToken)
			suggestion := fmt.Sprintf("'%s' can only be used inside function definitions", firstToken)
			return msg, rl.ErrKeywordMisuse, &suggestion
		}
	}

	return "", "", nil
}

// isInsideSwitch checks if the node is inside a switch statement.
func isInsideSwitch(node *ts.Node) bool {
	for n := node.Parent(); n != nil; n = n.Parent() {
		if n.Kind() == rl.K_SWITCH_STMT {
			return true
		}
	}
	return false
}

// isInsideLoop checks if the node is inside a loop.
func isInsideLoop(node *ts.Node) bool {
	for n := node.Parent(); n != nil; n = n.Parent() {
		kind := n.Kind()
		if kind == rl.K_FOR_LOOP || kind == rl.K_WHILE_LOOP {
			return true
		}
	}
	return false
}

// isInsideFunction checks if the node is inside a function definition.
func isInsideFunction(node *ts.Node) bool {
	for n := node.Parent(); n != nil; n = n.Parent() {
		kind := n.Kind()
		if kind == rl.K_FN_NAMED || kind == rl.K_FN_LAMBDA {
			return true
		}
	}
	return false
}

// generateContextualErrorMessage generates an error message based on parent context.
func generateContextualErrorMessage(node, parent *ts.Node, content string) (string, rl.Error, *string) {
	parentKind := parent.Kind()

	switch parentKind {
	case rl.K_ARG_BLOCK:
		return "Invalid argument declaration", rl.ErrInvalidSyntax, nil
	case rl.K_FOR_LOOP:
		return "Invalid for loop syntax", rl.ErrInvalidSyntax, nil
	case rl.K_IF_STMT:
		return "Invalid if statement", rl.ErrInvalidSyntax, nil
	case rl.K_FN_NAMED, rl.K_FN_LAMBDA:
		return "Invalid function syntax", rl.ErrInvalidSyntax, nil
	case rl.K_RAD_BLOCK:
		return "Invalid rad block syntax", rl.ErrInvalidSyntax, nil
	}

	// Default to unexpected token with content if short enough
	if len(content) <= 20 && len(content) > 0 {
		return fmt.Sprintf("Unexpected '%s'", content), rl.ErrUnexpectedToken, nil
	}

	return "Invalid syntax", rl.ErrInvalidSyntax, nil
}

// getErrorCodeForMissingKind returns the appropriate error code for a MISSING node kind.
func getErrorCodeForMissingKind(kind string) rl.Error {
	if code, ok := missingKindToErrorCode[kind]; ok {
		return code
	}
	return rl.ErrInvalidSyntax
}

// humanizeNodeKind converts a node kind to a human-readable string.
func humanizeNodeKind(kind string) string {
	// Replace underscores with spaces
	result := strings.ReplaceAll(kind, "_", " ")

	// Handle common patterns
	result = strings.TrimSuffix(result, " type")

	return result
}
