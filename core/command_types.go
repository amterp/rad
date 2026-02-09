package core

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// ScriptCommand represents a command defined in a Rad script's command: block
type ScriptCommand struct {
	Name             string // Internal (as written in script)
	ExternalName     string // External (hyphenated for CLI)
	Description      *string
	Args             []*ScriptArg // Command-specific arguments
	IsLambdaCallback bool
	CallbackName     *string    // For function reference callbacks
	CallbackLambda   *rl.Lambda // Eagerly converted AST lambda
}

func FromCmdBlock(cmdBlock *rl.CmdBlock, src string) (*ScriptCommand, error) {
	commandName := cmdBlock.Name
	externalName := rts.ToExternalName(commandName)

	args := make([]*ScriptArg, 0, len(cmdBlock.Decls))
	for _, decl := range cmdBlock.Decls {
		argName := decl.Name

		enumConstraint := cmdBlock.EnumConstraints[argName]
		regexConstraint := cmdBlock.RegexConstraints[argName]
		rangeConstraint := cmdBlock.RangeConstraints[argName]
		requiresConstraint := extractRelationsForArg(argName, cmdBlock.Requirements)
		excludesConstraint := extractRelationsForArg(argName, cmdBlock.Exclusions)

		scriptArg := FromArgDecl(
			decl,
			src,
			enumConstraint,
			regexConstraint,
			rangeConstraint,
			requiresConstraint,
			excludesConstraint,
		)
		args = append(args, scriptArg)
	}

	callback := cmdBlock.Callback
	return &ScriptCommand{
		Name:             commandName,
		ExternalName:     externalName,
		Description:      cmdBlock.Description,
		Args:             args,
		IsLambdaCallback: callback.IsLambda,
		CallbackName:     callback.IdentifierName,
		CallbackLambda:   callback.Lambda,
	}, nil
}

// extractRelationsForArg finds all related arg names for a given arg in a relation list.
// Returns external (CLI-visible) names since that's what Ra uses for constraint checking.
func extractRelationsForArg(argName string, relations []rl.ArgRelation) []string {
	for _, rel := range relations {
		if rel.Arg == argName {
			result := make([]string, len(rel.Related))
			for i, name := range rel.Related {
				result[i] = rts.ToExternalName(name)
			}
			return result
		}
	}
	return nil
}
