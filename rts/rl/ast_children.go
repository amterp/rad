package rl

// Children() implementations for all AST node types.
// Each concrete type declares its own children, keeping the walker generic.

func (n *SourceFile) Children() []Node {
	var c []Node
	if n.Header != nil {
		c = append(c, n.Header)
	}
	if n.Args != nil {
		c = append(c, n.Args)
	}
	for _, cmd := range n.Cmds {
		c = append(c, cmd)
	}
	c = append(c, n.Stmts...)
	return c
}

func (n *Assign) Children() []Node {
	c := make([]Node, 0, len(n.Targets)+len(n.Values))
	c = append(c, n.Targets...)
	c = append(c, n.Values...)
	if n.Catch != nil {
		c = append(c, n.Catch.Stmts...)
	}
	return c
}

func (n *ExprStmt) Children() []Node {
	c := []Node{n.Expr}
	if n.Catch != nil {
		c = append(c, n.Catch.Stmts...)
	}
	return c
}

func (n *If) Children() []Node {
	var c []Node
	for _, b := range n.Branches {
		if b.Condition != nil {
			c = append(c, b.Condition)
		}
		c = append(c, b.Body...)
	}
	return c
}

func (n *Switch) Children() []Node {
	c := []Node{n.Discriminant}
	for _, sc := range n.Cases {
		c = append(c, sc.Keys...)
		c = append(c, sc.Alt)
	}
	if n.Default != nil {
		c = append(c, n.Default.Alt)
	}
	return c
}

func (n *SwitchCaseExpr) Children() []Node  { return n.Values }
func (n *SwitchCaseBlock) Children() []Node { return n.Stmts }

func (n *ForLoop) Children() []Node {
	c := []Node{n.Iter}
	c = append(c, n.Body...)
	return c
}

func (n *WhileLoop) Children() []Node {
	var c []Node
	if n.Condition != nil {
		c = append(c, n.Condition)
	}
	c = append(c, n.Body...)
	return c
}

func (n *Shell) Children() []Node {
	c := make([]Node, 0, len(n.Targets)+1)
	c = append(c, n.Targets...)
	c = append(c, n.Cmd)
	if n.Catch != nil {
		c = append(c, n.Catch.Stmts...)
	}
	return c
}

func (n *Del) Children() []Node   { return n.Targets }
func (n *Defer) Children() []Node { return n.Body }

func (n *Return) Children() []Node { return n.Values }
func (n *Yield) Children() []Node  { return n.Values }

func (n *FnDef) Children() []Node {
	var c []Node
	if n.Typing != nil {
		for _, param := range n.Typing.Params {
			if param.DefaultAST != nil && param.DefaultAST.Node != nil {
				c = append(c, param.DefaultAST.Node)
			}
		}
	}
	c = append(c, n.Body...)
	return c
}

func (n *OpBinary) Children() []Node  { return []Node{n.Left, n.Right} }
func (n *OpUnary) Children() []Node   { return []Node{n.Operand} }
func (n *Ternary) Children() []Node   { return []Node{n.Condition, n.True, n.False} }
func (n *Fallback) Children() []Node  { return []Node{n.Left, n.Right} }
func (n *CatchExpr) Children() []Node { return []Node{n.Left, n.Right} }

func (n *Call) Children() []Node {
	c := []Node{n.Func}
	c = append(c, n.Args...)
	for _, na := range n.NamedArgs {
		c = append(c, na.Value)
	}
	return c
}

func (n *VarPath) Children() []Node {
	c := []Node{n.Root}
	for _, seg := range n.Segments {
		if seg.Index != nil {
			c = append(c, seg.Index)
		}
		if seg.Start != nil {
			c = append(c, seg.Start)
		}
		if seg.End != nil {
			c = append(c, seg.End)
		}
	}
	return c
}

func (n *Lambda) Children() []Node {
	var c []Node
	if n.Typing != nil {
		for _, param := range n.Typing.Params {
			if param.DefaultAST != nil && param.DefaultAST.Node != nil {
				c = append(c, param.DefaultAST.Node)
			}
		}
	}
	c = append(c, n.Body...)
	return c
}

func (n *LitString) Children() []Node {
	if n.Simple {
		return nil
	}
	var c []Node
	for _, seg := range n.Segments {
		if !seg.IsLiteral && seg.Expr != nil {
			c = append(c, seg.Expr)
		}
		if seg.Format != nil {
			if seg.Format.Padding != nil {
				c = append(c, seg.Format.Padding)
			}
			if seg.Format.Precision != nil {
				c = append(c, seg.Format.Precision)
			}
		}
	}
	return c
}

func (n *LitList) Children() []Node { return n.Elements }

func (n *LitMap) Children() []Node {
	c := make([]Node, 0, len(n.Entries)*2)
	for _, e := range n.Entries {
		c = append(c, e.Key, e.Value)
	}
	return c
}

func (n *ListComp) Children() []Node {
	c := []Node{n.Expr, n.Iter}
	if n.Condition != nil {
		c = append(c, n.Condition)
	}
	return c
}

func (n *RadBlock) Children() []Node {
	c := []Node{n.Source}
	c = append(c, n.Stmts...)
	return c
}

func (n *RadField) Children() []Node { return n.Identifiers }

func (n *RadFieldMod) Children() []Node {
	c := make([]Node, 0, len(n.Fields)+len(n.Args))
	c = append(c, n.Fields...)
	c = append(c, n.Args...)
	return c
}

func (n *RadIf) Children() []Node {
	var c []Node
	for _, b := range n.Branches {
		if b.Condition != nil {
			c = append(c, b.Condition)
		}
		c = append(c, b.Body...)
	}
	return c
}

func (n *JsonPath) Children() []Node {
	var c []Node
	for _, seg := range n.Segments {
		for _, idx := range seg.Indexes {
			if idx.Expr != nil {
				c = append(c, idx.Expr)
			}
		}
	}
	return c
}

func (n *ArgBlock) Children() []Node {
	var c []Node
	for i := range n.Decls {
		if n.Decls[i].Default != nil {
			c = append(c, n.Decls[i].Default)
		}
	}
	return c
}

func (n *ArgDecl) Children() []Node {
	if n.Default != nil {
		return []Node{n.Default}
	}
	return nil
}

func (n *CmdBlock) Children() []Node {
	var c []Node
	for i := range n.Decls {
		if n.Decls[i].Default != nil {
			c = append(c, n.Decls[i].Default)
		}
	}
	if n.Callback.Lambda != nil {
		c = append(c, n.Callback.Lambda)
	}
	return c
}

// Leaf nodes - no children.
func (n *Break) Children() []Node      { return nil }
func (n *Continue) Children() []Node   { return nil }
func (n *Pass) Children() []Node       { return nil }
func (n *Identifier) Children() []Node { return nil }
func (n *LitInt) Children() []Node     { return nil }
func (n *LitFloat) Children() []Node   { return nil }
func (n *LitBool) Children() []Node    { return nil }
func (n *LitNull) Children() []Node    { return nil }
func (n *RadSort) Children() []Node    { return nil }
func (n *FileHeader) Children() []Node { return nil }
