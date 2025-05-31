package check

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/raderr"
)

// todo be able to check scripts with different versions of Rad?

type RadChecker interface {
	UpdateSrc(src string)
	CheckDefault() (Result, error)
	Check(Opts) (Result, error)
}

type RadCheckerImpl struct {
	parser *rts.RadParser
	tree   *rts.RadTree
	src    string
}

func NewChecker() (RadChecker, error) {
	parser, err := rts.NewRadParser()
	if err != nil {
		return nil, err
	}
	tree := parser.Parse("")
	return NewCheckerWithTree(tree, parser, ""), nil
}

func NewCheckerWithTree(tree *rts.RadTree, parser *rts.RadParser, src string) RadChecker {
	return &RadCheckerImpl{
		parser: parser,
		tree:   tree,
		src:    src,
	}
}

func (c *RadCheckerImpl) UpdateSrc(src string) {
	if c.tree == nil {
		c.tree = c.parser.Parse(src)
	} else {
		c.tree.Update(src)
	}
	c.src = src
}

func (c *RadCheckerImpl) CheckDefault() (Result, error) {
	return c.Check(NewOpts())
}

// todo use opts
func (c *RadCheckerImpl) Check(opts Opts) (Result, error) {
	diagnostics := make([]Diagnostic, 0)
	c.addInvalidNodes(&diagnostics)
	return Result{
		Diagnostics: diagnostics,
	}, nil
}

func (c *RadCheckerImpl) addInvalidNodes(d *[]Diagnostic) {
	nodes := c.tree.FindInvalidNodes()
	for _, node := range nodes {
		*d = append(*d, NewDiagnosticError(node, c.src, "Invalid syntax", raderr.ErrInvalidSyntax))
	}
}
