package core

import (
	"fmt"
	"strconv"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

type ScriptData struct {
	ScriptName        string
	Args              []*ScriptArg
	Commands          []*ScriptCommand
	Description       *string
	Tree              *rts.RadTree
	Src               string
	DisableGlobalOpts bool
	DisableArgsBlock  bool
}

func ExtractMetadata(src string) *ScriptData {
	radTree, err := rts.NewRadParser()
	if err != nil {
		RP.RadErrorExit("Failed to create Rad tree sitter: " + err.Error())
		panic(UNREACHABLE)
	}

	tree := radTree.Parse(src)

	// Validate syntax before attempting to extract metadata
	validateSyntax(src, tree, radTree)

	disableGlobalOpts := false
	disableArgsBlock := false
	var description *string
	if fileHeader, ok := tree.FindFileHeader(); ok {
		description = &fileHeader.Contents
		if stashId, ok := fileHeader.MetadataEntries[MACRO_STASH_ID]; ok {
			RadHomeInst.SetStashId(stashId)
		}

		disableGlobalOpts = !defaultTruthyMacroToggle(fileHeader.MetadataEntries, MACRO_ENABLE_GLOBAL_OPTIONS)
		disableArgsBlock = !defaultTruthyMacroToggle(fileHeader.MetadataEntries, MACRO_ENABLE_ARGS_BLOCK)
	}

	var args []*ScriptArg
	if argBlock, ok := tree.FindArgBlock(); ok {
		RP.RadDebugf(fmt.Sprintf("Found arg block: %v", com.Dump(argBlock)))
		args = extractArgs(argBlock)
	}

	var commands []*ScriptCommand
	if cmdBlocks, ok := tree.FindCmdBlocks(); ok {
		RP.RadDebugf(fmt.Sprintf("Found %d command blocks", len(cmdBlocks)))
		commands = extractCommands(cmdBlocks)
	}

	return &ScriptData{
		ScriptName:        ScriptName,
		Args:              args,
		Commands:          commands,
		Description:       description,
		Tree:              tree,
		Src:               src,
		DisableGlobalOpts: disableGlobalOpts,
		DisableArgsBlock:  disableArgsBlock,
	}
}

func (sd *ScriptData) ValidateNoErrors() {
	invalidNodes := sd.Tree.FindInvalidNodes()
	if len(invalidNodes) > 0 {
		renderer := NewDiagnosticRenderer(RIo.StdErr)
		for _, node := range invalidNodes {
			span := NewSpanFromNode(node, sd.ScriptName)
			diag := NewDiagnostic(SeverityError, rl.ErrInvalidSyntax, "Invalid syntax", sd.Src, span)
			renderer.Render(diag)
		}
		RExit.Exit(1)
	}
}

// validateSyntax checks for syntax errors and exits immediately if any are found.
// This runs before argument parsing, so it only respects environment variables (like NO_COLOR),
// not command-line flags like --color=never.
func validateSyntax(src string, tree *rts.RadTree, parser *rts.RadParser) {
	// AST is nil here: validation runs before AST conversion, so CST checks suffice.
	checker := check.NewCheckerWithTree(tree, parser, src, nil)
	result, err := checker.CheckDefault()
	if err != nil {
		RP.RadErrorExit("Failed to validate syntax: " + err.Error())
		panic(UNREACHABLE)
	}

	// Collect all Error severity diagnostics
	var errors []Diagnostic
	for _, diag := range result.Diagnostics {
		if diag.Severity == check.Error {
			// Convert to core.Diagnostic using the new format
			coreDiag := NewDiagnosticFromCheck(diag, ScriptName)
			errors = append(errors, coreDiag)
		}
	}

	if len(errors) > 0 {
		// Before showing errors, check if user requested inspection flags
		handleGlobalInspectionFlagsOnInvalidSyntax(src, tree)

		// Render all errors using the new Rust-style renderer
		renderer := NewDiagnosticRenderer(RIo.StdErr)
		for _, diag := range errors {
			renderer.Render(diag)
		}
		RExit.Exit(1)
	}
}

func extractArgs(argBlock *rts.ArgBlock) []*ScriptArg {
	var args []*ScriptArg

	if argBlock == nil {
		return nil
	}

	argIdentifiers := make([]string, 0)
	for _, argDecl := range argBlock.Args {
		argIdentifiers = append(argIdentifiers, argDecl.Name.Name)
	}

	requires := make(map[string][]string)
	for _, reqLeft := range argBlock.Requirements {
		for _, reqRight := range reqLeft.Required {
			// Ra will be given external names, so transform constraint names to match
			leftExternal := rts.ToExternalName(reqLeft.Arg.Name)
			rightExternal := rts.ToExternalName(reqRight.Name)
			requires[leftExternal] = append(requires[leftExternal], rightExternal)
			if reqLeft.IsMutual {
				requires[rightExternal] = append(requires[rightExternal], leftExternal)
			}
		}
	}

	excludes := make(map[string][]string)
	for _, excludeLeft := range argBlock.Exclusions {
		for _, excludeRight := range excludeLeft.Excluded {
			// Ra will be given external names, so transform constraint names to match
			leftExternal := rts.ToExternalName(excludeLeft.Arg.Name)
			rightExternal := rts.ToExternalName(excludeRight.Name)
			excludes[leftExternal] = append(excludes[leftExternal], rightExternal)
			if excludeLeft.IsMutual {
				excludes[rightExternal] = append(excludes[rightExternal], leftExternal)
			}
		}
	}

	for _, argDecl := range argBlock.Args {
		enumConstraint := argBlock.EnumConstraints[argDecl.Name.Name]
		regexConstraint := argBlock.RegexConstraints[argDecl.Name.Name]
		rangeConstraint := argBlock.RangeConstraints[argDecl.Name.Name]
		requiresConstraint := requires[argDecl.Name.Name]
		excludesConstraint := excludes[argDecl.Name.Name]
		args = append(args, FromArgDecl(
			argDecl,
			enumConstraint,
			regexConstraint,
			rangeConstraint,
			requiresConstraint,
			excludesConstraint,
		))
	}

	return args
}

func defaultTruthyMacroToggle(macroMap map[string]string, macro string) bool {
	val, ok := macroMap[macro]
	if !ok {
		return true
	}

	var radVal RadValue
	if i64, err := strconv.ParseInt(val, 10, 64); err == nil {
		radVal = newRadValueInt64(i64)
	} else if f64, err := strconv.ParseFloat(val, 64); err == nil {
		radVal = newRadValueFloat64(f64)
	} else if b, err := strconv.ParseBool(val); err == nil {
		radVal = newRadValueBool(b)
	} else {
		radVal = newRadValueStr(val)
	}

	return radVal.TruthyFalsy()
}

func extractCommands(cmdBlocks []*rts.CmdBlock) []*ScriptCommand {
	commands := make([]*ScriptCommand, 0, len(cmdBlocks))

	for _, cmdBlock := range cmdBlocks {
		cmd, err := FromCmdBlock(cmdBlock)
		if err != nil {
			RP.CtxErrorExit(NewCtxFromRtsNode(cmdBlock, fmt.Sprintf("Failed to extract command: %s", err.Error())))
		}
		commands = append(commands, cmd)
	}

	return commands
}
