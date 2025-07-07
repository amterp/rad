package rts

import (
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type CmdBlock struct {
	BaseNode
	Cmds []CmdDecl
}

func (b CmdBlock) FindCmd(name string) *CmdDecl {
	for _, cmd := range b.Cmds {
		if cmd.Name.Str == name {
			return &cmd
		}
	}
	return nil
}

type CmdDecl struct {
	BaseNode
	Name CmdDeclToken
	Path CmdDeclToken
}

type CmdDeclToken struct {
	BaseNode
	Str string
}

func newCmdBlock(src string, node *ts.Node) (*CmdBlock, bool) {
	cmds := node.ChildrenByFieldName("cmd", node.Walk())
	cmdDecls := make([]CmdDecl, 0, len(cmds))
	for _, cmd := range cmds {
		nameNode := cmd.ChildByFieldName("name")
		pathNode := cmd.ChildByFieldName("path")
		if nameNode == nil || pathNode == nil {
			continue // malformed command block
		}

		cmdDecls = append(cmdDecls, CmdDecl{
			BaseNode: newBaseNode(src, &cmd),
			Name: CmdDeclToken{
				BaseNode: newBaseNode(src, nameNode),
				Str:      rl.GetSrc(nameNode, src),
			},
			Path: CmdDeclToken{
				BaseNode: newBaseNode(src, pathNode),
				Str:      extractString(src, pathNode),
			},
		})
	}

	return &CmdBlock{
		BaseNode: newBaseNode(src, node),
		Cmds:     cmdDecls,
	}, false
}
