package core

import (
	"fmt"
	com "rad/core/common"

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
	RP.RadDebug("Tree dump:\n" + tree.Dump()) // todo should be lazy i.e. func

	var description *string
	fileHeader, err := tree.FindFileHeader()
	if err == nil && fileHeader != nil {
		description = &fileHeader.Contents
	}

	argBlock, err := tree.FindArgBlock()
	RP.RadDebug(fmt.Sprintf("Found arg block: %v", com.FlatStr(argBlock)))
	args := extractArgs(argBlock)

	return &ScriptData{
		ScriptName:  ScriptName,
		Args:        args,
		Description: description,
		Tree:        tree,
		Src:         src,
	}
}

func extractArgs(argBlock *rts.ArgBlock) []*ScriptArg {
	var args []*ScriptArg

	if argBlock == nil {
		return nil
	}

	for _, argDecl := range argBlock.Args {
		enumConstraint := argBlock.EnumConstraints[argDecl.Name.Name]
		regexConstraint := argBlock.RegexConstraints[argDecl.Name.Name]
		args = append(args, FromArgDecl(argDecl, enumConstraint, regexConstraint))
	}

	return args
}
