package rl

import (
	"fmt"
	"strings"
)

// AstDump produces a human-readable tree representation of an AST.
func AstDump(node Node) string {
	var sb strings.Builder
	astDumpNode(&sb, node, 0)
	return sb.String()
}

func astDumpNode(sb *strings.Builder, node Node, depth int) {
	indent := strings.Repeat("  ", depth)
	span := node.Span()

	switch n := node.(type) {
	case *SourceFile:
		fmt.Fprintf(sb, "%sSourceFile [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, stmt := range n.Stmts {
			astDumpNode(sb, stmt, depth+1)
		}

	case *Assign:
		unpack := ""
		if n.IsUnpacking {
			unpack = " unpack"
		}
		fmt.Fprintf(sb, "%sAssign%s [%d:%d-%d:%d]\n", indent, unpack, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, t := range n.Targets {
			astDumpNode(sb, t, depth+1)
		}
		for _, v := range n.Values {
			astDumpNode(sb, v, depth+1)
		}
		if n.Catch != nil {
			fmt.Fprintf(sb, "%s  Catch\n", indent)
			for _, s := range n.Catch.Stmts {
				astDumpNode(sb, s, depth+2)
			}
		}

	case *ExprStmt:
		fmt.Fprintf(sb, "%sExprStmt [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Expr, depth+1)

	case *If:
		fmt.Fprintf(sb, "%sIf [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for i, branch := range n.Branches {
			if branch.Condition != nil {
				fmt.Fprintf(sb, "%s  Branch %d (cond)\n", indent, i)
				astDumpNode(sb, branch.Condition, depth+2)
			} else {
				fmt.Fprintf(sb, "%s  Branch %d (else)\n", indent, i)
			}
			for _, s := range branch.Body {
				astDumpNode(sb, s, depth+2)
			}
		}

	case *Switch:
		fmt.Fprintf(sb, "%sSwitch [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Discriminant, depth+1)
		for _, c := range n.Cases {
			fmt.Fprintf(sb, "%s  Case\n", indent)
			for _, k := range c.Keys {
				astDumpNode(sb, k, depth+2)
			}
			astDumpNode(sb, c.Alt, depth+2)
		}
		if n.Default != nil {
			fmt.Fprintf(sb, "%s  Default\n", indent)
			astDumpNode(sb, n.Default.Alt, depth+2)
		}

	case *SwitchCaseExpr:
		fmt.Fprintf(sb, "%sCaseExpr\n", indent)
		for _, v := range n.Values {
			astDumpNode(sb, v, depth+1)
		}

	case *SwitchCaseBlock:
		fmt.Fprintf(sb, "%sCaseBlock\n", indent)
		for _, s := range n.Stmts {
			astDumpNode(sb, s, depth+1)
		}

	case *ForLoop:
		ctx := ""
		if n.Context != nil {
			ctx = fmt.Sprintf(" with %s", *n.Context)
		}
		fmt.Fprintf(sb, "%sForLoop vars=[%s]%s [%d:%d-%d:%d]\n", indent, strings.Join(n.Vars, ", "), ctx, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Iter, depth+1)
		for _, s := range n.Body {
			astDumpNode(sb, s, depth+1)
		}

	case *WhileLoop:
		fmt.Fprintf(sb, "%sWhileLoop [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		if n.Condition != nil {
			astDumpNode(sb, n.Condition, depth+1)
		}
		for _, s := range n.Body {
			astDumpNode(sb, s, depth+1)
		}

	case *Shell:
		fmt.Fprintf(sb, "%sShell quiet=%v confirm=%v [%d:%d-%d:%d]\n", indent, n.IsQuiet, n.IsConfirm, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Cmd, depth+1)

	case *Del:
		fmt.Fprintf(sb, "%sDel [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, t := range n.Targets {
			astDumpNode(sb, t, depth+1)
		}

	case *Defer:
		kind := "defer"
		if n.IsErrDefer {
			kind = "errdefer"
		}
		fmt.Fprintf(sb, "%sDefer %s [%d:%d-%d:%d]\n", indent, kind, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, s := range n.Body {
			astDumpNode(sb, s, depth+1)
		}

	case *Break:
		fmt.Fprintf(sb, "%sBreak [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	case *Continue:
		fmt.Fprintf(sb, "%sContinue [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	case *Pass:
		fmt.Fprintf(sb, "%sPass [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)

	case *Return:
		fmt.Fprintf(sb, "%sReturn [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, v := range n.Values {
			astDumpNode(sb, v, depth+1)
		}

	case *Yield:
		fmt.Fprintf(sb, "%sYield [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, v := range n.Values {
			astDumpNode(sb, v, depth+1)
		}

	case *FnDef:
		fmt.Fprintf(sb, "%sFnDef %q block=%v [%d:%d-%d:%d]\n", indent, n.Name, n.IsBlock, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, s := range n.Body {
			astDumpNode(sb, s, depth+1)
		}

	case *Lambda:
		fmt.Fprintf(sb, "%sLambda block=%v [%d:%d-%d:%d]\n", indent, n.IsBlock, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, s := range n.Body {
			astDumpNode(sb, s, depth+1)
		}

	case *OpBinary:
		fmt.Fprintf(sb, "%sOpBinary %s [%d:%d-%d:%d]\n", indent, n.Op, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Left, depth+1)
		astDumpNode(sb, n.Right, depth+1)

	case *OpUnary:
		fmt.Fprintf(sb, "%sOpUnary %s [%d:%d-%d:%d]\n", indent, n.Op, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Operand, depth+1)

	case *Ternary:
		fmt.Fprintf(sb, "%sTernary [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Condition, depth+1)
		astDumpNode(sb, n.True, depth+1)
		astDumpNode(sb, n.False, depth+1)

	case *Fallback:
		fmt.Fprintf(sb, "%sFallback [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Left, depth+1)
		astDumpNode(sb, n.Right, depth+1)

	case *Call:
		fmt.Fprintf(sb, "%sCall [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Func, depth+1)
		for _, a := range n.Args {
			astDumpNode(sb, a, depth+1)
		}
		for _, na := range n.NamedArgs {
			fmt.Fprintf(sb, "%s  Named %q\n", indent, na.Name)
			astDumpNode(sb, na.Value, depth+2)
		}

	case *VarPath:
		fmt.Fprintf(sb, "%sVarPath [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Root, depth+1)
		for _, seg := range n.Segments {
			if seg.Field != nil {
				fmt.Fprintf(sb, "%s  .%s\n", indent, *seg.Field)
			} else if seg.IsSlice {
				fmt.Fprintf(sb, "%s  [slice]\n", indent)
			} else {
				fmt.Fprintf(sb, "%s  [index]\n", indent)
				astDumpNode(sb, seg.Index, depth+2)
			}
		}

	case *Identifier:
		fmt.Fprintf(sb, "%sIdentifier %q [%d:%d-%d:%d]\n", indent, n.Name, span.StartRow, span.StartCol, span.EndRow, span.EndCol)

	case *LitInt:
		fmt.Fprintf(sb, "%sLitInt %d [%d:%d-%d:%d]\n", indent, n.Value, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	case *LitFloat:
		fmt.Fprintf(sb, "%sLitFloat %g [%d:%d-%d:%d]\n", indent, n.Value, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	case *LitBool:
		fmt.Fprintf(sb, "%sLitBool %v [%d:%d-%d:%d]\n", indent, n.Value, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	case *LitNull:
		fmt.Fprintf(sb, "%sLitNull [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)

	case *LitString:
		if n.Simple {
			fmt.Fprintf(sb, "%sLitString %q [%d:%d-%d:%d]\n", indent, n.Value, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		} else {
			fmt.Fprintf(sb, "%sLitString (interpolated, %d segments) [%d:%d-%d:%d]\n", indent, len(n.Segments), span.StartRow, span.StartCol, span.EndRow, span.EndCol)
			for _, seg := range n.Segments {
				if seg.IsLiteral {
					fmt.Fprintf(sb, "%s  Literal %q\n", indent, seg.Text)
				} else {
					fmt.Fprintf(sb, "%s  Interpolation\n", indent)
					astDumpNode(sb, seg.Expr, depth+2)
				}
			}
		}

	case *LitList:
		fmt.Fprintf(sb, "%sLitList (%d elements) [%d:%d-%d:%d]\n", indent, len(n.Elements), span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, e := range n.Elements {
			astDumpNode(sb, e, depth+1)
		}

	case *LitMap:
		fmt.Fprintf(sb, "%sLitMap (%d entries) [%d:%d-%d:%d]\n", indent, len(n.Entries), span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, e := range n.Entries {
			fmt.Fprintf(sb, "%s  Entry\n", indent)
			astDumpNode(sb, e.Key, depth+2)
			astDumpNode(sb, e.Value, depth+2)
		}

	case *ListComp:
		fmt.Fprintf(sb, "%sListComp vars=[%s] [%d:%d-%d:%d]\n", indent, strings.Join(n.Vars, ", "), span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		astDumpNode(sb, n.Expr, depth+1)
		astDumpNode(sb, n.Iter, depth+1)
		if n.Condition != nil {
			astDumpNode(sb, n.Condition, depth+1)
		}

	case *RadBlock:
		fmt.Fprintf(sb, "%sRadBlock %q [%d:%d-%d:%d]\n", indent, n.BlockType, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		if n.Source != nil {
			astDumpNode(sb, n.Source, depth+1)
		}
		for _, s := range n.Stmts {
			astDumpNode(sb, s, depth+1)
		}

	case *RadField:
		fmt.Fprintf(sb, "%sRadField [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, id := range n.Identifiers {
			astDumpNode(sb, id, depth+1)
		}
	case *RadSort:
		fmt.Fprintf(sb, "%sRadSort [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	case *RadFieldMod:
		fmt.Fprintf(sb, "%sRadFieldMod %q [%d:%d-%d:%d]\n", indent, n.ModType, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for _, f := range n.Fields {
			astDumpNode(sb, f, depth+1)
		}
		for _, a := range n.Args {
			astDumpNode(sb, a, depth+1)
		}
	case *RadIf:
		fmt.Fprintf(sb, "%sRadIf [%d:%d-%d:%d]\n", indent, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		for i, branch := range n.Branches {
			if branch.Condition != nil {
				fmt.Fprintf(sb, "%s  Branch %d (cond)\n", indent, i)
				astDumpNode(sb, branch.Condition, depth+2)
			} else {
				fmt.Fprintf(sb, "%s  Branch %d (else)\n", indent, i)
			}
			for _, s := range branch.Body {
				astDumpNode(sb, s, depth+2)
			}
		}

	default:
		fmt.Fprintf(sb, "%s<%T> [%d:%d-%d:%d]\n", indent, node, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	}
}
