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
	":":             "Expected ':'",
	rl.K_IDENTIFIER: "Expected identifier",
	rl.K_EXPR:       "Expected expression",
	")":             "Expected ')'",
	"(":             "Expected '('",
	"]":             "Expected ']'",
	"[":             "Expected '['",
	"}":             "Expected '}'",
	"{":             "Expected '{'",
	rl.K_STRING:     "Expected string",
	rl.K_INT:        "Expected integer",
	rl.K_FLOAT:      "Expected number",
	"type":          "Expected type",
	"newline":       "Expected newline",
	"indent":        "Expected indented block",
	"dedent":        "Unexpected indentation",
}

// missingKindToErrorCode maps MISSING node kinds to specific error codes.
var missingKindToErrorCode = map[string]rl.Error{
	":":             rl.ErrMissingColon,
	rl.K_IDENTIFIER: rl.ErrMissingIdentifier,
	rl.K_EXPR:       rl.ErrMissingExpression,
	")":             rl.ErrMissingCloseParen,
	"(":             rl.ErrMissingCloseParen,
	"]":             rl.ErrMissingCloseBracket,
	"[":             rl.ErrMissingCloseBracket,
	"}":             rl.ErrMissingCloseBrace,
	"{":             rl.ErrMissingCloseBrace,
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

	// Check explicit message map first
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
