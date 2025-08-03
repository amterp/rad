package core

import (
	"fmt"
	"strconv"

	com "github.com/amterp/rad/core/common"

	"github.com/samber/lo"

	"github.com/amterp/rad/rts"
)

type ScriptData struct {
	ScriptName        string
	Args              []*ScriptArg
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
	}

	tree := radTree.Parse(src)

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

	return &ScriptData{
		ScriptName:        ScriptName,
		Args:              args,
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
		for _, node := range invalidNodes {
			// TODO print all errors up front instead of exiting here
			RP.CtxErrorExit(NewCtx(sd.Src, node, "Invalid syntax", ""))
		}
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
		if !lo.Contains(argIdentifiers, reqLeft.Arg.Name) {
			errUndefinedArg(&reqLeft, reqLeft.Arg.Name)
		}

		for _, reqRight := range reqLeft.Required {
			if !lo.Contains(argIdentifiers, reqRight.Name) {
				errUndefinedArg(&reqRight, reqRight.Name)
			}

			requires[reqLeft.Arg.Name] = append(requires[reqLeft.Arg.Name], reqRight.Name)
			if reqLeft.IsMutual {
				requires[reqRight.Name] = append(requires[reqRight.Name], reqLeft.Arg.Name)
			}
		}
	}

	excludes := make(map[string][]string)
	for _, excludeLeft := range argBlock.Exclusions {
		if !lo.Contains(argIdentifiers, excludeLeft.Arg.Name) {
			errUndefinedArg(&excludeLeft, excludeLeft.Arg.Name)
		}

		for _, excludeRight := range excludeLeft.Excluded {
			if !lo.Contains(argIdentifiers, excludeRight.Name) {
				errUndefinedArg(&excludeRight, excludeRight.Name)
			}

			excludes[excludeLeft.Arg.Name] = append(excludes[excludeLeft.Arg.Name], excludeRight.Name)
			if excludeLeft.IsMutual {
				excludes[excludeRight.Name] = append(excludes[excludeRight.Name], excludeLeft.Arg.Name)
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

func errUndefinedArg(node rts.Node, name string) {
	RP.CtxErrorExit(NewCtx(node.CompleteSrc(), node.Node(), fmt.Sprintf("Undefined arg '%s'", name), ""))
}
