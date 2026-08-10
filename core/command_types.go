package core

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// ScriptCommand represents a command defined in a Rad script's command: block.
// A command is either a leaf (HasCallback, dispatches) or a namespace (SubCmds,
// routes); args declared on a namespace are shared with every descendant.
type ScriptCommand struct {
	Name             string // Internal (as written in script)
	ExternalName     string // External (hyphenated for CLI)
	Description      *string
	Args             []*ScriptArg // Command-specific arguments
	HasCallback      bool
	IsLambdaCallback bool
	CallbackName     *string    // For function reference callbacks
	CallbackLambda   *rl.Lambda // Eagerly converted AST lambda
	SubCmds          []*ScriptCommand
	Parent           *ScriptCommand // nil at the top level
	// IsDefault marks the command its level runs when none is named.
	IsDefault bool
}

// IsNamespace reports whether this command routes to sub-commands rather than
// dispatching to a callback of its own.
func (c *ScriptCommand) IsNamespace() bool { return len(c.SubCmds) > 0 }

// pathNames returns the external command names from the top level down to this
// command - the tokens a user types to reach it.
func (c *ScriptCommand) pathNames() []string {
	var names []string
	for cur := c; cur != nil; cur = cur.Parent {
		names = append([]string{cur.ExternalName}, names...)
	}
	return names
}

// argChain returns this command's own args followed by those inherited from
// each namespace above it, nearest first. Order is deliberate: the command the
// user named is the context they're in, so its args are asked about first.
func (c *ScriptCommand) argChain() []*ScriptArg {
	var args []*ScriptArg
	for cur := c; cur != nil; cur = cur.Parent {
		args = append(args, cur.Args...)
	}
	return args
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
		lenConstraint := cmdBlock.LenConstraints[argName]
		requiresConstraint := extractRelationsForArg(argName, cmdBlock.Requirements)
		excludesConstraint := extractRelationsForArg(argName, cmdBlock.Exclusions)

		scriptArg := FromArgDecl(
			decl,
			src,
			enumConstraint,
			regexConstraint,
			rangeConstraint,
			lenConstraint,
			requiresConstraint,
			excludesConstraint,
		)
		args = append(args, scriptArg)
	}

	cmd := &ScriptCommand{
		Name:         commandName,
		ExternalName: externalName,
		Description:  cmdBlock.Description,
		Args:         args,
		IsDefault:    cmdBlock.IsDefault(),
	}

	// A namespace has no callback of its own; only the terminal command runs.
	if callback := cmdBlock.Callback; callback != nil {
		cmd.HasCallback = true
		cmd.IsLambdaCallback = callback.IsLambda
		cmd.CallbackLambda = callback.Lambda
		if callback.Identifier != nil {
			cmd.CallbackName = &callback.Identifier.Name
		}
	}

	for _, subBlock := range cmdBlock.SubCmds {
		sub, err := FromCmdBlock(subBlock, src)
		if err != nil {
			return nil, err
		}
		sub.Parent = cmd
		cmd.SubCmds = append(cmd.SubCmds, sub)
	}

	return cmd, nil
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
