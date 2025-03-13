package core

import (
	"fmt"
	com "rad/core/common"

	"github.com/samber/lo"

	"github.com/amterp/rts"
)

type ScriptData struct {
	ScriptName  string
	Args        []*ScriptArg
	Description *string
	Tree        *rts.RslTree
	Src         string
}

func ExtractMetadata(src string) *ScriptData {
	rslTree, err := rts.NewRslParser()
	if err != nil {
		RP.RadErrorExit("Failed to create RSL tree sitter: " + err.Error())
	}

	tree := rslTree.Parse(src)
	RP.RadDebugf("Tree dump:\n" + tree.Dump()) // todo should be lazy i.e. func

	var description *string
	fileHeader, ok := tree.FindFileHeader()
	if ok {
		description = &fileHeader.Contents
	}

	argBlock, ok := tree.FindArgBlock()
	RP.RadDebugf(fmt.Sprintf("Found arg block: %v", com.FlatStr(argBlock)))
	args := extractArgs(argBlock)

	return &ScriptData{
		ScriptName:  ScriptName,
		Args:        args,
		Description: description,
		Tree:        tree,
		Src:         src,
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

func errUndefinedArg(node rts.Node, name string) {
	RP.CtxErrorExit(NewCtx(node.CompleteSrc(), node.Node(), fmt.Sprintf("Undefined arg '%s'", name), ""))
}
