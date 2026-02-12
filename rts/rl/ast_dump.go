package rl

import (
	"fmt"
	"strings"
)

// AstDump produces a human-readable tree representation of an AST.
func AstDump(node Node) string {
	maxRow, maxCol := findASTMaxSpans(node)
	rowW := len(fmt.Sprintf("%d", maxRow))
	colW := len(fmt.Sprintf("%d", maxCol))
	fmtStr := fmt.Sprintf(
		"PS: [%%%dd, %%%dd] PE: [%%%dd, %%%dd] ",
		rowW, colW, rowW, colW)
	spacePad := strings.Repeat(" ", len(fmt.Sprintf(fmtStr, 0, 0, 0, 0)))
	var sb strings.Builder
	astDumpNode(&sb, fmtStr, spacePad, node, 0)
	return sb.String()
}

// findASTMaxSpans recursively walks the entire AST to find the true
// maximum row and column values. The root span's EndRow is always the
// overall max row, but max column can occur on any deeply nested node.
func findASTMaxSpans(node Node) (maxRow, maxCol int) {
	span := node.Span()
	maxRow = span.EndRow
	maxCol = max(span.StartCol, span.EndCol)

	var children []Node

	switch n := node.(type) {
	case *SourceFile:
		if n.Header != nil {
			children = append(children, n.Header)
		}
		if n.Args != nil {
			children = append(children, n.Args)
		}
		for _, cmd := range n.Cmds {
			children = append(children, cmd)
		}
		children = append(children, n.Stmts...)

	case *Assign:
		children = append(children, n.Targets...)
		children = append(children, n.Values...)
		if n.Catch != nil {
			children = append(children, n.Catch.Stmts...)
		}

	case *ExprStmt:
		children = append(children, n.Expr)
		if n.Catch != nil {
			children = append(children, n.Catch.Stmts...)
		}

	case *If:
		children = appendIfBranches(children, n.Branches)

	case *Switch:
		children = append(children, n.Discriminant)
		for _, c := range n.Cases {
			children = append(children, c.Keys...)
			children = append(children, c.Alt)
		}
		if n.Default != nil {
			children = append(children, n.Default.Alt)
		}

	case *SwitchCaseExpr:
		children = append(children, n.Values...)
	case *SwitchCaseBlock:
		children = append(children, n.Stmts...)

	case *ForLoop:
		children = append(children, n.Iter)
		children = append(children, n.Body...)
	case *WhileLoop:
		if n.Condition != nil {
			children = append(children, n.Condition)
		}
		children = append(children, n.Body...)

	case *Shell:
		children = append(children, n.Targets...)
		children = append(children, n.Cmd)
		if n.Catch != nil {
			children = append(children, n.Catch.Stmts...)
		}

	case *Del:
		children = append(children, n.Targets...)
	case *Defer:
		children = append(children, n.Body...)
	case *Return:
		children = append(children, n.Values...)
	case *Yield:
		children = append(children, n.Values...)

	case *FnDef:
		children = append(children, n.Body...)
	case *Lambda:
		children = append(children, n.Body...)

	case *OpBinary:
		children = append(children, n.Left, n.Right)
	case *OpUnary:
		children = append(children, n.Operand)
	case *Ternary:
		children = append(children, n.Condition, n.True, n.False)
	case *Fallback:
		children = append(children, n.Left, n.Right)
	case *CatchExpr:
		children = append(children, n.Left, n.Right)

	case *Call:
		children = append(children, n.Func)
		children = append(children, n.Args...)
		for _, na := range n.NamedArgs {
			children = append(children, na.Value)
		}

	case *VarPath:
		children = append(children, n.Root)
		for _, seg := range n.Segments {
			if seg.Index != nil {
				children = append(children, seg.Index)
			}
			if seg.Start != nil {
				children = append(children, seg.Start)
			}
			if seg.End != nil {
				children = append(children, seg.End)
			}
		}

	case *LitString:
		if !n.Simple {
			for _, seg := range n.Segments {
				if !seg.IsLiteral && seg.Expr != nil {
					children = append(children, seg.Expr)
				}
				if seg.Format != nil {
					if seg.Format.Padding != nil {
						children = append(children, seg.Format.Padding)
					}
					if seg.Format.Precision != nil {
						children = append(children, seg.Format.Precision)
					}
				}
			}
		}

	case *LitList:
		children = append(children, n.Elements...)
	case *LitMap:
		for _, e := range n.Entries {
			children = append(children, e.Key, e.Value)
		}

	case *ListComp:
		children = append(children, n.Expr, n.Iter)
		if n.Condition != nil {
			children = append(children, n.Condition)
		}

	case *RadBlock:
		if n.Source != nil {
			children = append(children, n.Source)
		}
		children = append(children, n.Stmts...)
	case *RadField:
		children = append(children, n.Identifiers...)
	case *RadFieldMod:
		children = append(children, n.Fields...)
		children = append(children, n.Args...)
	case *RadIf:
		children = appendIfBranches(children, n.Branches)

	case *JsonPath:
		for _, seg := range n.Segments {
			for _, idx := range seg.Indexes {
				if idx.Expr != nil {
					children = append(children, idx.Expr)
				}
			}
		}

	case *ArgBlock:
		for i := range n.Decls {
			children = append(children, &n.Decls[i])
		}
	case *ArgDecl:
		if n.Default != nil {
			children = append(children, n.Default)
		}
	case *CmdBlock:
		for i := range n.Decls {
			children = append(children, &n.Decls[i])
		}
		if n.Callback.Lambda != nil {
			children = append(children, n.Callback.Lambda)
		}

	// Leaf nodes: Break, Continue, Pass, Identifier, LitInt, LitFloat,
	// LitBool, LitNull, FileHeader, RadSort, simple LitString
	default:
		// no children to recurse into
	}

	for _, child := range children {
		childRow, childCol := findASTMaxSpans(child)
		maxRow = max(maxRow, childRow)
		maxCol = max(maxCol, childCol)
	}
	return
}

// appendIfBranches collects all child nodes from IfBranch slices,
// shared by both If and RadIf.
func appendIfBranches(children []Node, branches []IfBranch) []Node {
	for _, b := range branches {
		if b.Condition != nil {
			children = append(children, b.Condition)
		}
		children = append(children, b.Body...)
	}
	return children
}

func astDumpNode(sb *strings.Builder, fmtStr string, spacePad string, node Node, depth int) {
	indent := strings.Repeat("  ", depth)
	span := node.Span()

	switch n := node.(type) {
	case *SourceFile:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sSourceFile\n", indent)
		if n.Header != nil {
			astDumpNode(sb, fmtStr, spacePad, n.Header, depth+1)
		}
		if n.Args != nil {
			astDumpNode(sb, fmtStr, spacePad, n.Args, depth+1)
		}
		for _, cmd := range n.Cmds {
			astDumpNode(sb, fmtStr, spacePad, cmd, depth+1)
		}
		for _, stmt := range n.Stmts {
			astDumpNode(sb, fmtStr, spacePad, stmt, depth+1)
		}

	case *FileHeader:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		desc := n.Contents
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Fprintf(sb, "%sFileHeader %q\n", indent, desc)
		for k, v := range n.MetadataEntries {
			fmt.Fprintf(sb, "%s%s@%s = %s\n", spacePad, indent, k, v)
		}

	case *ArgBlock:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sArgBlock\n", indent)
		for i := range n.Decls {
			astDumpNode(sb, fmtStr, spacePad, &n.Decls[i], depth+1)
		}

	case *ArgDecl:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		flags := ""
		if n.IsOptional {
			flags += " optional"
		}
		if n.IsVariadic {
			flags += " variadic"
		}
		fmt.Fprintf(sb, "%sArgDecl %q type=%s%s\n", indent, n.Name, n.TypeName, flags)
		if n.Default != nil {
			fmt.Fprintf(sb, "%s%sDefault:\n", spacePad, indent)
			astDumpNode(sb, fmtStr, spacePad, n.Default, depth+2)
		}

	case *CmdBlock:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sCmdBlock %q\n", indent, n.Name)
		if n.Description != nil {
			fmt.Fprintf(sb, "%s%sDescription: %q\n", spacePad, indent, *n.Description)
		}
		for i := range n.Decls {
			astDumpNode(sb, fmtStr, spacePad, &n.Decls[i], depth+1)
		}
		cb := n.Callback
		if cb.IdentifierName != nil {
			fmt.Fprintf(sb, "%s%sCallback: %s\n", spacePad, indent, *cb.IdentifierName)
		} else if cb.Lambda != nil {
			fmt.Fprintf(sb, "%s%sCallback: lambda\n", spacePad, indent)
			astDumpNode(sb, fmtStr, spacePad, cb.Lambda, depth+2)
		}

	case *Assign:
		unpack := ""
		if n.IsUnpacking {
			unpack = " unpack"
		}
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sAssign%s\n", indent, unpack)
		for _, t := range n.Targets {
			astDumpNode(sb, fmtStr, spacePad, t, depth+1)
		}
		for _, v := range n.Values {
			astDumpNode(sb, fmtStr, spacePad, v, depth+1)
		}
		if n.Catch != nil {
			fmt.Fprintf(sb, "%s%sCatch\n", spacePad, indent)
			for _, s := range n.Catch.Stmts {
				astDumpNode(sb, fmtStr, spacePad, s, depth+2)
			}
		}

	case *ExprStmt:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sExprStmt\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Expr, depth+1)

	case *If:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sIf\n", indent)
		for i, branch := range n.Branches {
			if branch.Condition != nil {
				fmt.Fprintf(sb, "%s%sBranch %d (cond)\n", spacePad, indent, i)
				astDumpNode(sb, fmtStr, spacePad, branch.Condition, depth+2)
			} else {
				fmt.Fprintf(sb, "%s%sBranch %d (else)\n", spacePad, indent, i)
			}
			for _, s := range branch.Body {
				astDumpNode(sb, fmtStr, spacePad, s, depth+2)
			}
		}

	case *Switch:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sSwitch\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Discriminant, depth+1)
		for _, c := range n.Cases {
			fmt.Fprintf(sb, "%s%sCase\n", spacePad, indent)
			for _, k := range c.Keys {
				astDumpNode(sb, fmtStr, spacePad, k, depth+2)
			}
			astDumpNode(sb, fmtStr, spacePad, c.Alt, depth+2)
		}
		if n.Default != nil {
			fmt.Fprintf(sb, "%s%sDefault\n", spacePad, indent)
			astDumpNode(sb, fmtStr, spacePad, n.Default.Alt, depth+2)
		}

	case *SwitchCaseExpr:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sCaseExpr\n", indent)
		for _, v := range n.Values {
			astDumpNode(sb, fmtStr, spacePad, v, depth+1)
		}

	case *SwitchCaseBlock:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sCaseBlock\n", indent)
		for _, s := range n.Stmts {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *ForLoop:
		ctx := ""
		if n.Context != nil {
			ctx = fmt.Sprintf(" with %s", *n.Context)
		}
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sForLoop vars=[%s]%s\n", indent, strings.Join(n.Vars, ", "), ctx)
		astDumpNode(sb, fmtStr, spacePad, n.Iter, depth+1)
		for _, s := range n.Body {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *WhileLoop:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sWhileLoop\n", indent)
		if n.Condition != nil {
			astDumpNode(sb, fmtStr, spacePad, n.Condition, depth+1)
		}
		for _, s := range n.Body {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *Shell:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sShell quiet=%v confirm=%v\n", indent, n.IsQuiet, n.IsConfirm)
		astDumpNode(sb, fmtStr, spacePad, n.Cmd, depth+1)

	case *Del:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sDel\n", indent)
		for _, t := range n.Targets {
			astDumpNode(sb, fmtStr, spacePad, t, depth+1)
		}

	case *Defer:
		kind := "defer"
		if n.IsErrDefer {
			kind = "errdefer"
		}
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sDefer %s\n", indent, kind)
		for _, s := range n.Body {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *Break:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sBreak\n", indent)
	case *Continue:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sContinue\n", indent)
	case *Pass:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sPass\n", indent)

	case *Return:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sReturn\n", indent)
		for _, v := range n.Values {
			astDumpNode(sb, fmtStr, spacePad, v, depth+1)
		}

	case *Yield:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sYield\n", indent)
		for _, v := range n.Values {
			astDumpNode(sb, fmtStr, spacePad, v, depth+1)
		}

	case *FnDef:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sFnDef %q block=%v\n", indent, n.Name, n.IsBlock)
		for _, s := range n.Body {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *Lambda:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLambda block=%v\n", indent, n.IsBlock)
		for _, s := range n.Body {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *OpBinary:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sOpBinary %s\n", indent, n.Op)
		astDumpNode(sb, fmtStr, spacePad, n.Left, depth+1)
		astDumpNode(sb, fmtStr, spacePad, n.Right, depth+1)

	case *OpUnary:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sOpUnary %s\n", indent, n.Op)
		astDumpNode(sb, fmtStr, spacePad, n.Operand, depth+1)

	case *Ternary:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sTernary\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Condition, depth+1)
		astDumpNode(sb, fmtStr, spacePad, n.True, depth+1)
		astDumpNode(sb, fmtStr, spacePad, n.False, depth+1)

	case *Fallback:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sFallback\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Left, depth+1)
		astDumpNode(sb, fmtStr, spacePad, n.Right, depth+1)

	case *CatchExpr:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sCatchExpr\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Left, depth+1)
		astDumpNode(sb, fmtStr, spacePad, n.Right, depth+1)

	case *Call:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sCall\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Func, depth+1)
		for _, a := range n.Args {
			astDumpNode(sb, fmtStr, spacePad, a, depth+1)
		}
		for _, na := range n.NamedArgs {
			fmt.Fprintf(sb, "%s%sNamed %q\n", spacePad, indent, na.Name)
			astDumpNode(sb, fmtStr, spacePad, na.Value, depth+2)
		}

	case *VarPath:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sVarPath\n", indent)
		astDumpNode(sb, fmtStr, spacePad, n.Root, depth+1)
		for _, seg := range n.Segments {
			if seg.Field != nil {
				fmt.Fprintf(sb, "%s%s.%s\n", spacePad, indent, *seg.Field)
			} else if seg.IsSlice {
				fmt.Fprintf(sb, "%s%s[slice]\n", spacePad, indent)
			} else {
				fmt.Fprintf(sb, "%s%s[index]\n", spacePad, indent)
				astDumpNode(sb, fmtStr, spacePad, seg.Index, depth+2)
			}
		}

	case *Identifier:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sIdentifier %q\n", indent, n.Name)

	case *LitInt:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLitInt %d\n", indent, n.Value)
	case *LitFloat:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLitFloat %g\n", indent, n.Value)
	case *LitBool:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLitBool %v\n", indent, n.Value)
	case *LitNull:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLitNull\n", indent)

	case *LitString:
		if n.Simple {
			fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
			fmt.Fprintf(sb, "%sLitString %q\n", indent, n.Value)
		} else {
			fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
			fmt.Fprintf(sb, "%sLitString (interpolated, %d segments)\n", indent, len(n.Segments))
			for _, seg := range n.Segments {
				if seg.IsLiteral {
					fmt.Fprintf(sb, "%s%sLiteral %q\n", spacePad, indent, seg.Text)
				} else {
					fmt.Fprintf(sb, "%s%sInterpolation\n", spacePad, indent)
					astDumpNode(sb, fmtStr, spacePad, seg.Expr, depth+2)
				}
			}
		}

	case *LitList:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLitList (%d elements)\n", indent, len(n.Elements))
		for _, e := range n.Elements {
			astDumpNode(sb, fmtStr, spacePad, e, depth+1)
		}

	case *LitMap:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sLitMap (%d entries)\n", indent, len(n.Entries))
		for _, e := range n.Entries {
			fmt.Fprintf(sb, "%s%sEntry\n", spacePad, indent)
			astDumpNode(sb, fmtStr, spacePad, e.Key, depth+2)
			astDumpNode(sb, fmtStr, spacePad, e.Value, depth+2)
		}

	case *ListComp:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sListComp vars=[%s]\n", indent, strings.Join(n.Vars, ", "))
		astDumpNode(sb, fmtStr, spacePad, n.Expr, depth+1)
		astDumpNode(sb, fmtStr, spacePad, n.Iter, depth+1)
		if n.Condition != nil {
			astDumpNode(sb, fmtStr, spacePad, n.Condition, depth+1)
		}

	case *RadBlock:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sRadBlock %q\n", indent, n.BlockType)
		if n.Source != nil {
			astDumpNode(sb, fmtStr, spacePad, n.Source, depth+1)
		}
		for _, s := range n.Stmts {
			astDumpNode(sb, fmtStr, spacePad, s, depth+1)
		}

	case *RadField:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sRadField\n", indent)
		for _, id := range n.Identifiers {
			astDumpNode(sb, fmtStr, spacePad, id, depth+1)
		}
	case *RadSort:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sRadSort\n", indent)
	case *RadFieldMod:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sRadFieldMod %q\n", indent, n.ModType)
		for _, f := range n.Fields {
			astDumpNode(sb, fmtStr, spacePad, f, depth+1)
		}
		for _, a := range n.Args {
			astDumpNode(sb, fmtStr, spacePad, a, depth+1)
		}
	case *JsonPath:
		var parts []string
		for _, seg := range n.Segments {
			s := seg.Key
			for _, idx := range seg.Indexes {
				if idx.Expr != nil {
					s += "[expr]"
				} else {
					s += "[]"
				}
			}
			parts = append(parts, s)
		}
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sJsonPath %s\n", indent, strings.Join(parts, "."))

	case *RadIf:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sRadIf\n", indent)
		for i, branch := range n.Branches {
			if branch.Condition != nil {
				fmt.Fprintf(sb, "%s%sBranch %d (cond)\n", spacePad, indent, i)
				astDumpNode(sb, fmtStr, spacePad, branch.Condition, depth+2)
			} else {
				fmt.Fprintf(sb, "%s%sBranch %d (else)\n", spacePad, indent, i)
			}
			for _, s := range branch.Body {
				astDumpNode(sb, fmtStr, spacePad, s, depth+2)
			}
		}

	default:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%s<%T>\n", indent, node)
	}
}
