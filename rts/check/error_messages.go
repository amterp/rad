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

// missingKindToErrorCode maps MISSING node kinds to specific error codes.
var missingKindToErrorCode = map[string]rl.Error{
	":":             rl.ErrMissingColon,
	rl.K_IDENTIFIER: rl.ErrMissingIdentifier,
	rl.K_EXPR:       rl.ErrMissingExpression,
	rl.K_VAR_PATH:   rl.ErrMissingExpression,
	")":             rl.ErrMissingCloseParen,
	"(":             rl.ErrMissingOpenParen,
	"]":             rl.ErrMissingCloseBracket,
	"[":             rl.ErrMissingOpenBracket,
	"}":             rl.ErrMissingCloseBrace,
	"{":             rl.ErrMissingOpenBrace,
	",":             rl.ErrMissingComma,
	"=":             rl.ErrMissingEquals,
	"->":            rl.ErrMissingArrow,
	"type":          rl.ErrMissingType,
	"newline":       rl.ErrMissingNewline,
	"indent":        rl.ErrMissingIndent,
	"dedent":        rl.ErrUnexpectedIndent,
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
	rl.K_CALL + ":(":          "Expected '(' after function name",
	rl.K_CALL + ":)":          "Expected ')' to close function call",
	rl.K_CALL + ":" + rl.K_EXPR: "Expected argument expression",

	// Lists
	rl.K_LIST + ":]":          "Expected ']' to close list",
	rl.K_LIST + ":" + rl.K_EXPR: "Expected list element",

	// Maps
	rl.K_MAP + ":}":          "Expected '}' to close map",
	rl.K_MAP + ":":           "Expected ':' between key and value",
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

	// Heuristic 1: Check for unterminated string
	if msg, code, suggestion := checkUnterminatedString(errorContent, trimmedContent); msg != "" {
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
		"rad": true, "request": true, "display": true,
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
