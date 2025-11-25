package core

import (
	"github.com/amterp/rad/rts"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// ScriptCommand represents a command defined in a Rad script's command: block
type ScriptCommand struct {
	Name           string
	Description    *string
	Args           []*ScriptArg // Command-specific arguments
	CallbackType   rts.CallbackType
	CallbackName   *string  // For function reference callbacks (rts.CallbackIdentifier)
	CallbackLambda *ts.Node // For inline lambda callbacks (rts.CallbackLambda)
}

func FromCmdBlock(cmdBlock *rts.CmdBlock) (*ScriptCommand, error) {
	// Extract command name
	commandName := cmdBlock.Name.Name

	// Extract optional description
	var description *string
	if cmdBlock.Description != nil {
		description = &cmdBlock.Description.Contents
	}

	// Convert arguments using the same logic as args: block
	args := make([]*ScriptArg, 0, len(cmdBlock.Args))
	for _, argDecl := range cmdBlock.Args {
		argName := argDecl.Name.Name

		// Look up constraints for this arg
		enumConstraint := cmdBlock.EnumConstraints[argName]
		regexConstraint := cmdBlock.RegexConstraints[argName]
		rangeConstraint := cmdBlock.RangeConstraints[argName]

		// Convert Requirements and Exclusions
		requiresConstraint := extractRequiresForArg(argName, cmdBlock.Requirements)
		excludesConstraint := extractExcludesForArg(argName, cmdBlock.Exclusions)

		// Convert to ScriptArg
		scriptArg := FromArgDecl(
			argDecl,
			enumConstraint,
			regexConstraint,
			rangeConstraint,
			requiresConstraint,
			excludesConstraint,
		)
		args = append(args, scriptArg)
	}

	// Extract callback information
	callback := cmdBlock.Callback
	return &ScriptCommand{
		Name:           commandName,
		Description:    description,
		Args:           args,
		CallbackType:   callback.Type,
		CallbackName:   callback.IdentifierName,
		CallbackLambda: callback.LambdaNode,
	}, nil
}

// extractRequiresForArg finds all "requires" constraints for a given arg name
func extractRequiresForArg(argName string, requirements []rts.ArgRequirement) []string {
	for _, req := range requirements {
		if req.Arg.Name == argName {
			result := make([]string, len(req.Required))
			for i, required := range req.Required {
				result[i] = required.Name
			}
			return result
		}
	}
	return nil
}

// extractExcludesForArg finds all "excludes" constraints for a given arg name
func extractExcludesForArg(argName string, exclusions []rts.ArgExclusion) []string {
	for _, excl := range exclusions {
		if excl.Arg.Name == argName {
			result := make([]string, len(excl.Excluded))
			for i, excluded := range excl.Excluded {
				result[i] = excluded.Name
			}
			return result
		}
	}
	return nil
}
