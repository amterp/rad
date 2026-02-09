package core

import (
	"fmt"
	"strconv"

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

	// Validate syntax and get the AST
	ast := validateSyntax(src, tree, radTree)

	disableGlobalOpts := false
	disableArgsBlock := false
	var description *string
	if ast != nil && ast.Header != nil {
		description = &ast.Header.Contents
		if stashId, ok := ast.Header.MetadataEntries[MACRO_STASH_ID]; ok {
			RadHomeInst.SetStashId(stashId)
		}

		disableGlobalOpts = !defaultTruthyMacroToggle(ast.Header.MetadataEntries, MACRO_ENABLE_GLOBAL_OPTIONS)
		disableArgsBlock = !defaultTruthyMacroToggle(ast.Header.MetadataEntries, MACRO_ENABLE_ARGS_BLOCK)
	}

	var args []*ScriptArg
	if ast != nil && ast.Args != nil {
		RP.RadDebugf(fmt.Sprintf("Found arg block with %d declarations", len(ast.Args.Decls)))
		args = extractArgsFromAST(ast.Args, src)
	}

	var commands []*ScriptCommand
	if ast != nil && len(ast.Cmds) > 0 {
		RP.RadDebugf(fmt.Sprintf("Found %d command blocks", len(ast.Cmds)))
		commands = extractCommandsFromAST(ast.Cmds, src)
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

// tryConvertAST attempts CST-to-AST conversion, returning nil if the source
// has syntax errors that cause the converter to panic.
func tryConvertAST(tree *rts.RadTree, src string, file string) (ast *rl.SourceFile) {
	defer func() {
		if r := recover(); r != nil {
			ast = nil
		}
	}()
	return rts.ConvertCST(tree.Root(), src, file)
}

// validateSyntax checks for syntax errors and exits immediately if any are found.
// Returns the AST produced during validation so callers can extract metadata from it.
// This runs before argument parsing, so it only respects environment variables (like NO_COLOR),
// not command-line flags like --color=never.
func validateSyntax(src string, tree *rts.RadTree, parser *rts.RadParser) *rl.SourceFile {
	ast := tryConvertAST(tree, src, ScriptName)
	checker := check.NewCheckerWithTree(tree, parser, src, ast)
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

	return ast
}

func extractArgsFromAST(argBlock *rl.ArgBlock, src string) []*ScriptArg {
	if argBlock == nil {
		return nil
	}

	requires := buildRelationMap(argBlock.Requirements)
	excludes := buildRelationMap(argBlock.Exclusions)

	var args []*ScriptArg
	for _, decl := range argBlock.Decls {
		args = append(args, FromArgDecl(
			decl,
			src,
			argBlock.EnumConstraints[decl.Name],
			argBlock.RegexConstraints[decl.Name],
			argBlock.RangeConstraints[decl.Name],
			requires[decl.Name],
			excludes[decl.Name],
		))
	}
	return args
}

// buildRelationMap builds a map from internal arg name to related external
// (CLI-visible) arg names, expanding mutual relations in both directions.
// Keys are internal names (matching decl.Name for lookup), values are external
// names (matching how Ra registers constraints).
func buildRelationMap(relations []rl.ArgRelation) map[string][]string {
	result := make(map[string][]string)
	for _, rel := range relations {
		for _, related := range rel.Related {
			rightExternal := rts.ToExternalName(related)
			result[rel.Arg] = append(result[rel.Arg], rightExternal)
			if rel.IsMutual {
				leftExternal := rts.ToExternalName(rel.Arg)
				result[related] = append(result[related], leftExternal)
			}
		}
	}
	return result
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

func extractCommandsFromAST(cmdBlocks []*rl.CmdBlock, src string) []*ScriptCommand {
	commands := make([]*ScriptCommand, 0, len(cmdBlocks))

	for _, cmdBlock := range cmdBlocks {
		cmd, err := FromCmdBlock(cmdBlock, src)
		if err != nil {
			span := cmdBlock.Span()
			RP.CtxErrorExit(NewCtxFromSpan(src, span, fmt.Sprintf("Failed to extract command: %s", err.Error()), ""))
		}
		commands = append(commands, cmd)
	}

	return commands
}
