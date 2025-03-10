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

	for _, argDecl := range argBlock.Args {
		enumConstraint := argBlock.EnumConstraints[argDecl.Name.Name]
		regexConstraint := argBlock.RegexConstraints[argDecl.Name.Name]
		rangeConstraint := argBlock.RangeConstraints[argDecl.Name.Name]
		args = append(args, FromArgDecl(argDecl, enumConstraint, regexConstraint, rangeConstraint))
	}

	return args
}
