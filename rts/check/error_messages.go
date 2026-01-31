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
	errorContent = strings.TrimSpace(errorContent)

	// Check for reserved keywords - extract first token from error content
	firstToken := strings.Fields(errorContent)
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
		return generateContextualErrorMessage(node, parent, errorContent)
	}

	return "Invalid syntax", rl.ErrInvalidSyntax, nil
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
