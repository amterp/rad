package rts

import (
	"fmt"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type CallbackType int

const (
	CallbackIdentifier CallbackType = iota
	CallbackLambda
)

type CmdBlock struct {
	BaseNode
	Name        CmdName
	Description *CmdDescription
	Args        []ArgDecl
	// Constraints (reuse from ArgBlock)
	EnumConstraints  map[string]*ArgEnumConstraint
	RegexConstraints map[string]*ArgRegexConstraint
	RangeConstraints map[string]*ArgRangeConstraint
	Requirements     []ArgRequirement
	Exclusions       []ArgExclusion
	// Callback
	Callback CmdCallback
}

type CmdName struct {
	BaseNode
	Name string
}

type CmdDescription struct {
	BaseNode
	Contents string
}

type CmdCallback struct {
	BaseNode
	Type            CallbackType
	IdentifierName  *string    // For function reference callbacks (CallbackIdentifier)
	IdentifierSpan  *rl.Span   // Span of the identifier node (for diagnostics)
	LambdaAST       *rl.Lambda // Eagerly converted AST for lambda callbacks
}

// newCmdBlock constructs a CmdBlock from a tree-sitter node
func newCmdBlock(src string, node *ts.Node) (*CmdBlock, bool) {
	// Extract command name
	nameNode := node.ChildByFieldName(rl.F_NAME)
	if nameNode == nil {
		return nil, false
	}
	name := CmdName{
		BaseNode: newBaseNode(src, nameNode),
		Name:     src[nameNode.StartByte():nameNode.EndByte()],
	}

	// Extract optional description
	var description *CmdDescription
	descNode := node.ChildByFieldName(rl.F_DESCRIPTION)
	if descNode != nil {
		contentsNode := descNode.ChildByFieldName(rl.F_CONTENTS)
		if contentsNode != nil {
			contents := src[contentsNode.StartByte():contentsNode.EndByte()]
			description = &CmdDescription{
				BaseNode: newBaseNode(src, contentsNode),
				Contents: NormalizeIndentedText(contents),
			}
		}
	}

	// Extract arguments and constraints (reuse arg parsing logic)
	args := findArgDeclarations(src, node)
	enumConstraints := findArgEnumConstraints(src, node)
	regexConstraints := findArgRegexConstraints(src, node)
	rangeConstraints := findArgRangeConstraints(src, node)
	requirements := findArgRequirements(src, node)
	exclusions := findArgExclusions(src, node)

	// Extract callback
	callback := extractCmdCallback(src, node)

	return &CmdBlock{
		BaseNode:         newBaseNode(src, node),
		Name:             name,
		Description:      description,
		Args:             args,
		EnumConstraints:  enumConstraints,
		RegexConstraints: regexConstraints,
		RangeConstraints: rangeConstraints,
		Requirements:     requirements,
		Exclusions:       exclusions,
		Callback:         callback,
	}, true
}

// extractCmdCallback extracts the callback from a cmd_calls node
func extractCmdCallback(src string, node *ts.Node) CmdCallback {
	callsNode := node.ChildByFieldName(rl.F_CALLS)
	if callsNode == nil {
		panic(fmt.Sprintf("Bug! Command block missing 'calls' field at byte %d", node.StartByte()))
	}

	// Check for identifier callback
	identifierNode := callsNode.ChildByFieldName(rl.F_CALLBACK_IDENTIFIER)
	if identifierNode != nil {
		identifierName := src[identifierNode.StartByte():identifierNode.EndByte()]
		identifierSpan := rl.Span{
			StartByte: int(identifierNode.StartByte()),
			EndByte:   int(identifierNode.EndByte()),
			StartRow:  int(identifierNode.StartPosition().Row),
			StartCol:  int(identifierNode.StartPosition().Column),
			EndRow:    int(identifierNode.EndPosition().Row),
			EndCol:    int(identifierNode.EndPosition().Column),
		}
		return CmdCallback{
			BaseNode:       newBaseNode(src, callsNode),
			Type:           CallbackIdentifier,
			IdentifierName: &identifierName,
			IdentifierSpan: &identifierSpan,
		}
	}

	// Check for lambda callback
	lambdaNode := callsNode.ChildByFieldName(rl.F_CALLBACK_LAMBDA)
	if lambdaNode != nil {
		lambdaAST := safeConvertLambda(lambdaNode, src)
		return CmdCallback{
			BaseNode:  newBaseNode(src, callsNode),
			Type:      CallbackLambda,
			LambdaAST: lambdaAST,
		}
	}

	panic(fmt.Sprintf("Bug! Command callback has neither identifier nor lambda at byte %d", callsNode.StartByte()))
}

// safeConvertLambda converts a lambda CST node to AST, recovering from panics
// caused by malformed syntax. Returns nil if conversion fails.
func safeConvertLambda(lambdaNode *ts.Node, src string) (result *rl.Lambda) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
		}
	}()
	return ConvertLambda(lambdaNode, src, "")
}
