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

// findASTMaxSpans returns the maximum row and column values from the root node's span.
// Since AST node spans encompass all their children, we only need to check the root.
func findASTMaxSpans(node Node) (maxRow, maxCol int) {
	span := node.Span()
	maxCol = span.EndCol
	if span.StartCol > maxCol {
		maxCol = span.StartCol
	}
	return span.EndRow, maxCol
}

func astDumpNode(sb *strings.Builder, fmtStr string, spacePad string, node Node, depth int) {
	indent := strings.Repeat("  ", depth)
	span := node.Span()

	switch n := node.(type) {
	case *SourceFile:
		fmt.Fprintf(sb, fmtStr, span.StartRow, span.StartCol, span.EndRow, span.EndCol)
		fmt.Fprintf(sb, "%sSourceFile\n", indent)
		for _, stmt := range n.Stmts {
			astDumpNode(sb, fmtStr, spacePad, stmt, depth+1)
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
