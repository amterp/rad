package core

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// ScriptCommand represents a command defined in a Rad script's command: block
type ScriptCommand struct {
	Name           string // Internal (as written in script)
	ExternalName   string // External (hyphenated for CLI)
	Description    *string
	Args           []*ScriptArg // Command-specific arguments
	CallbackType   rts.CallbackType
	CallbackName   *string  // For function reference callbacks (rts.CallbackIdentifier)
	CallbackLambda *rl.Lambda // Eagerly converted AST lambda
}

func FromCmdBlock(cmdBlock *rts.CmdBlock) (*ScriptCommand, error) {
	// Extract command name
	commandName := cmdBlock.Name.Name
	externalName := rts.ToExternalName(commandName)

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
		ExternalName:   externalName,
		Description:    description,
		Args:           args,
		CallbackType:   callback.Type,
		CallbackName:   callback.IdentifierName,
		CallbackLambda: callback.LambdaAST,
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
