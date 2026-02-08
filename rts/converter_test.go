package rts_test

import (
	"testing"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Targeted converter tests ---

// Helper: parse source and convert to AST.
func convertSource(t *testing.T, src string) *rl.SourceFile {
	t.Helper()
	parser, err := rts.NewRadParser()
	require.NoError(t, err)
	defer parser.Close()
	tree := parser.Parse(src)
	return rts.ConvertCST(tree.Root(), src, "test.rad")
}

// Helper: get the first (and usually only) statement from a source file.
func firstStmt(t *testing.T, src string) rl.Node {
	t.Helper()
	sf := convertSource(t, src)
	require.Len(t, sf.Stmts, 1, "expected exactly 1 statement")
	return sf.Stmts[0]
}

// Helper: get the expression inside the first ExprStmt.
func firstExpr(t *testing.T, src string) rl.Node {
	t.Helper()
	stmt := firstStmt(t, src)
	exprStmt, ok := stmt.(*rl.ExprStmt)
	require.True(t, ok, "expected ExprStmt, got %T", stmt)
	return exprStmt.Expr
}

// --- Delegate chain collapsing ---

func TestConvert_DelegateChainCollapsing(t *testing.T) {
	// A bare identifier at top level becomes an ExprStmt, but the inner
	// expression should be a clean Identifier (all delegates collapsed).
	stmt := firstStmt(t, "x")
	exprStmt, ok := stmt.(*rl.ExprStmt)
	require.True(t, ok, "expected ExprStmt, got %T", stmt)
	id, ok := exprStmt.Expr.(*rl.Identifier)
	require.True(t, ok, "expected Identifier inside ExprStmt, got %T", exprStmt.Expr)
	assert.Equal(t, "x", id.Name)
}

func TestConvert_DelegateThroughExprStmt(t *testing.T) {
	// A bare literal at top level: delegates collapse to LitInt inside ExprStmt
	stmt := firstStmt(t, "42")
	exprStmt, ok := stmt.(*rl.ExprStmt)
	require.True(t, ok, "expected ExprStmt, got %T", stmt)
	lit, ok := exprStmt.Expr.(*rl.LitInt)
	require.True(t, ok, "expected LitInt inside ExprStmt, got %T", exprStmt.Expr)
	assert.Equal(t, int64(42), lit.Value)
}

// --- Leaf value pre-parsing ---

func TestConvert_LitInt(t *testing.T) {
	stmt := firstStmt(t, "a = 123")
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitInt)
	assert.Equal(t, int64(123), lit.Value)
}

func TestConvert_LitIntWithUnderscores(t *testing.T) {
	stmt := firstStmt(t, "a = 1_000_000")
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitInt)
	assert.Equal(t, int64(1_000_000), lit.Value)
}

func TestConvert_LitFloat(t *testing.T) {
	stmt := firstStmt(t, "a = 3.14")
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitFloat)
	assert.Equal(t, 3.14, lit.Value)
}

func TestConvert_LitBoolTrue(t *testing.T) {
	stmt := firstStmt(t, "a = true")
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitBool)
	assert.True(t, lit.Value)
}

func TestConvert_LitBoolFalse(t *testing.T) {
	stmt := firstStmt(t, "a = false")
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitBool)
	assert.False(t, lit.Value)
}

func TestConvert_LitNull(t *testing.T) {
	stmt := firstStmt(t, "a = null")
	assign := stmt.(*rl.Assign)
	_, ok := assign.Values[0].(*rl.LitNull)
	assert.True(t, ok, "expected LitNull")
}

func TestConvert_ScientificNotation_WholeNumber(t *testing.T) {
	stmt := firstStmt(t, "a = 1e3")
	assign := stmt.(*rl.Assign)
	// 1e3 = 1000 which is a whole number, should be LitInt
	lit, ok := assign.Values[0].(*rl.LitInt)
	require.True(t, ok, "expected LitInt for 1e3, got %T", assign.Values[0])
	assert.Equal(t, int64(1000), lit.Value)
}

func TestConvert_ScientificNotation_WholeFromDecimal(t *testing.T) {
	// 1.5e2 = 150.0, which is a whole number - converter produces LitInt
	stmt := firstStmt(t, "a = 1.5e2")
	assign := stmt.(*rl.Assign)
	lit, ok := assign.Values[0].(*rl.LitInt)
	require.True(t, ok, "expected LitInt for 1.5e2 (=150), got %T", assign.Values[0])
	assert.Equal(t, int64(150), lit.Value)
}

func TestConvert_ScientificNotation_Fractional(t *testing.T) {
	// 1.5e-1 = 0.15, which is NOT a whole number - converter produces LitFloat
	stmt := firstStmt(t, "a = 1.5e-1")
	assign := stmt.(*rl.Assign)
	lit, ok := assign.Values[0].(*rl.LitFloat)
	require.True(t, ok, "expected LitFloat for 1.5e-1, got %T", assign.Values[0])
	assert.Equal(t, 0.15, lit.Value)
}

// --- String escape sequences ---

func TestConvert_String_DoubleQuote_Escapes(t *testing.T) {
	stmt := firstStmt(t, `a = "hello\nworld"`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, "hello\nworld", lit.Value)
}

func TestConvert_String_DoubleQuote_EscapedQuote(t *testing.T) {
	stmt := firstStmt(t, `a = "say \"hi\""`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, `say "hi"`, lit.Value)
}

func TestConvert_String_SingleQuote_EscapedQuote(t *testing.T) {
	stmt := firstStmt(t, `a = 'it\'s'`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, "it's", lit.Value)
}

func TestConvert_String_SingleQuote_UnescapedDoubleQuote(t *testing.T) {
	// Inside single quotes, \" should stay as literal \" (not just ")
	stmt := firstStmt(t, `a = 'hello\"world'`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, `hello\"world`, lit.Value)
}

func TestConvert_String_Tab(t *testing.T) {
	stmt := firstStmt(t, `a = "col1\tcol2"`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, "col1\tcol2", lit.Value)
}

func TestConvert_String_EscapedBackslash(t *testing.T) {
	stmt := firstStmt(t, `a = "path\\file"`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, `path\file`, lit.Value)
}

func TestConvert_String_Empty(t *testing.T) {
	stmt := firstStmt(t, `a = ""`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, "", lit.Value)
}

// --- String interpolation ---

func TestConvert_String_Interpolation(t *testing.T) {
	stmt := firstStmt(t, `a = "hello {name}"`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.False(t, lit.Simple)
	require.Len(t, lit.Segments, 2)
	assert.True(t, lit.Segments[0].IsLiteral)
	assert.Equal(t, "hello ", lit.Segments[0].Text)
	assert.False(t, lit.Segments[1].IsLiteral)
	id, ok := lit.Segments[1].Expr.(*rl.Identifier)
	require.True(t, ok)
	assert.Equal(t, "name", id.Name)
}

func TestConvert_String_EscapedBrace(t *testing.T) {
	stmt := firstStmt(t, `a = "use \{braces}"`)
	assign := stmt.(*rl.Assign)
	lit := assign.Values[0].(*rl.LitString)
	assert.True(t, lit.Simple)
	assert.Equal(t, "use {braces}", lit.Value)
}

// --- Compound assign desugaring ---

func TestConvert_CompoundAssign_Plus(t *testing.T) {
	stmt := firstStmt(t, "x += 3")
	assign := stmt.(*rl.Assign)
	assert.Len(t, assign.Targets, 1)
	assert.Len(t, assign.Values, 1)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpAdd, bin.Op)
}

func TestConvert_CompoundAssign_Minus(t *testing.T) {
	stmt := firstStmt(t, "x -= 3")
	assign := stmt.(*rl.Assign)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpSub, bin.Op)
}

func TestConvert_CompoundAssign_Star(t *testing.T) {
	stmt := firstStmt(t, "x *= 3")
	assign := stmt.(*rl.Assign)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpMul, bin.Op)
}

func TestConvert_CompoundAssign_Slash(t *testing.T) {
	stmt := firstStmt(t, "x /= 3")
	assign := stmt.(*rl.Assign)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpDiv, bin.Op)
}

func TestConvert_CompoundAssign_Percent(t *testing.T) {
	stmt := firstStmt(t, "x %= 3")
	assign := stmt.(*rl.Assign)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpMod, bin.Op)
}

// --- Increment/decrement desugaring ---

func TestConvert_Increment(t *testing.T) {
	stmt := firstStmt(t, "x++")
	assign := stmt.(*rl.Assign)
	assert.Len(t, assign.Targets, 1)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpAdd, bin.Op)
	one := bin.Right.(*rl.LitInt)
	assert.Equal(t, int64(1), one.Value)
}

func TestConvert_Decrement(t *testing.T) {
	stmt := firstStmt(t, "x--")
	assign := stmt.(*rl.Assign)
	bin := assign.Values[0].(*rl.OpBinary)
	assert.Equal(t, rl.OpSub, bin.Op)
	one := bin.Right.(*rl.LitInt)
	assert.Equal(t, int64(1), one.Value)
}

// --- VarPath construction ---

func TestConvert_VarPath_DotAccess(t *testing.T) {
	stmt := firstStmt(t, "a = x.y.z")
	assign := stmt.(*rl.Assign)
	vp := assign.Values[0].(*rl.VarPath)
	root := vp.Root.(*rl.Identifier)
	assert.Equal(t, "x", root.Name)
	require.Len(t, vp.Segments, 2)
	assert.NotNil(t, vp.Segments[0].Field)
	assert.Equal(t, "y", *vp.Segments[0].Field)
	assert.NotNil(t, vp.Segments[1].Field)
	assert.Equal(t, "z", *vp.Segments[1].Field)
}

func TestConvert_VarPath_BracketIndex(t *testing.T) {
	stmt := firstStmt(t, "a = x[0]")
	assign := stmt.(*rl.Assign)
	vp := assign.Values[0].(*rl.VarPath)
	root := vp.Root.(*rl.Identifier)
	assert.Equal(t, "x", root.Name)
	require.Len(t, vp.Segments, 1)
	assert.False(t, vp.Segments[0].IsSlice)
	idx := vp.Segments[0].Index.(*rl.LitInt)
	assert.Equal(t, int64(0), idx.Value)
}

func TestConvert_VarPath_Mixed(t *testing.T) {
	stmt := firstStmt(t, "a = x[0].y")
	assign := stmt.(*rl.Assign)
	vp := assign.Values[0].(*rl.VarPath)
	require.Len(t, vp.Segments, 2)
	// First segment: bracket index
	idx := vp.Segments[0].Index.(*rl.LitInt)
	assert.Equal(t, int64(0), idx.Value)
	// Second segment: dot access
	assert.NotNil(t, vp.Segments[1].Field)
	assert.Equal(t, "y", *vp.Segments[1].Field)
}

// --- For loop ---

func TestConvert_ForLoop_Unpacking(t *testing.T) {
	src := `for k, v in mymap:
    pass`
	stmt := firstStmt(t, src)
	loop := stmt.(*rl.ForLoop)
	assert.Equal(t, []string{"k", "v"}, loop.Vars)
	_, ok := loop.Iter.(*rl.Identifier)
	assert.True(t, ok)
}

// --- List comprehension ---

func TestConvert_ListComp_WithCondition(t *testing.T) {
	stmt := firstStmt(t, "a = [x * 2 for x in items if x > 0]")
	assign := stmt.(*rl.Assign)
	lc := assign.Values[0].(*rl.ListComp)
	assert.Equal(t, []string{"x"}, lc.Vars)
	assert.NotNil(t, lc.Condition)
	// The expression should be a binary op (x * 2)
	_, ok := lc.Expr.(*rl.OpBinary)
	assert.True(t, ok)
	// The condition should be a binary op (x > 0)
	cond, ok := lc.Condition.(*rl.OpBinary)
	assert.True(t, ok)
	assert.Equal(t, rl.OpGt, cond.Op)
}

// --- Switch ---

func TestConvert_Switch_WithDefault(t *testing.T) {
	src := `switch x:
    case 1 -> "one"
    case 2, 3 -> "few"
    default -> "other"`
	stmt := firstStmt(t, src)
	sw := stmt.(*rl.Switch)
	assert.NotNil(t, sw.Discriminant)
	require.Len(t, sw.Cases, 2)
	assert.Len(t, sw.Cases[0].Keys, 1)
	assert.Len(t, sw.Cases[1].Keys, 2)
	assert.NotNil(t, sw.Default)
}

// --- Binary operators ---

func TestConvert_BinaryOps(t *testing.T) {
	tests := []struct {
		src string
		op  rl.Operator
	}{
		{"a = 1 + 2", rl.OpAdd},
		{"a = 1 - 2", rl.OpSub},
		{"a = 1 * 2", rl.OpMul},
		{"a = 1 / 2", rl.OpDiv},
		{"a = 1 % 2", rl.OpMod},
		{"a = 1 == 2", rl.OpEq},
		{"a = 1 != 2", rl.OpNeq},
		{"a = 1 < 2", rl.OpLt},
		{"a = 1 <= 2", rl.OpLte},
		{"a = 1 > 2", rl.OpGt},
		{"a = 1 >= 2", rl.OpGte},
		{"a = true and false", rl.OpAnd},
		{"a = true or false", rl.OpOr},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			stmt := firstStmt(t, tt.src)
			assign := stmt.(*rl.Assign)
			bin := assign.Values[0].(*rl.OpBinary)
			assert.Equal(t, tt.op, bin.Op)
		})
	}
}

// --- Unary operators ---

func TestConvert_UnaryNeg(t *testing.T) {
	stmt := firstStmt(t, "a = -5")
	assign := stmt.(*rl.Assign)
	unary := assign.Values[0].(*rl.OpUnary)
	assert.Equal(t, rl.OpNeg, unary.Op)
}

func TestConvert_UnaryNot(t *testing.T) {
	stmt := firstStmt(t, "a = not true")
	assign := stmt.(*rl.Assign)
	unary := assign.Values[0].(*rl.OpUnary)
	assert.Equal(t, rl.OpNot, unary.Op)
}

// --- Span accuracy ---

func TestConvert_SpanAccuracy_Identifier(t *testing.T) {
	// Bare expression at top level becomes ExprStmt; check inner Identifier's span
	stmt := firstStmt(t, "hello")
	exprStmt := stmt.(*rl.ExprStmt)
	id := exprStmt.Expr.(*rl.Identifier)
	span := id.Span()
	assert.Equal(t, "test.rad", span.File)
	assert.Equal(t, 0, span.StartRow)
	assert.Equal(t, 0, span.StartCol)
	assert.Equal(t, 0, span.EndRow)
	assert.Equal(t, 5, span.EndCol)
	assert.Equal(t, 0, span.StartByte)
	assert.Equal(t, 5, span.EndByte)
}

func TestConvert_SpanAccuracy_SecondLine(t *testing.T) {
	src := "x = 1\ny = 2"
	sf := convertSource(t, src)
	require.Len(t, sf.Stmts, 2)

	// Second assignment
	secondAssign := sf.Stmts[1].(*rl.Assign)
	span := secondAssign.Span()
	assert.Equal(t, 1, span.StartRow)
	assert.Equal(t, 0, span.StartCol)
	assert.Equal(t, 1, span.EndRow)
	assert.Equal(t, 5, span.EndCol)
}

// --- Nested lambdas ---

func TestConvert_NestedLambda(t *testing.T) {
	src := `a = fn() fn() 42`
	stmt := firstStmt(t, src)
	assign := stmt.(*rl.Assign)
	outer, ok := assign.Values[0].(*rl.Lambda)
	require.True(t, ok, "expected outer Lambda, got %T", assign.Values[0])
	require.Len(t, outer.Body, 1)
	inner, ok := outer.Body[0].(*rl.Lambda)
	require.True(t, ok, "expected inner Lambda, got %T", outer.Body[0])
	require.Len(t, inner.Body, 1)
	_, ok = inner.Body[0].(*rl.LitInt)
	assert.True(t, ok, "expected LitInt at innermost level")
}

// --- Ternary ---

func TestConvert_Ternary(t *testing.T) {
	stmt := firstStmt(t, "a = cond ? x : y")
	assign := stmt.(*rl.Assign)
	tern := assign.Values[0].(*rl.Ternary)
	assert.NotNil(t, tern.Condition)
	assert.NotNil(t, tern.True)
	assert.NotNil(t, tern.False)
}

// --- Fallback ---

func TestConvert_Fallback(t *testing.T) {
	stmt := firstStmt(t, "a = x ?? y")
	assign := stmt.(*rl.Assign)
	fb := assign.Values[0].(*rl.Fallback)
	assert.NotNil(t, fb.Left)
	assert.NotNil(t, fb.Right)
}

// --- Lists and maps ---

func TestConvert_LitList(t *testing.T) {
	stmt := firstStmt(t, "a = [1, 2, 3]")
	assign := stmt.(*rl.Assign)
	list := assign.Values[0].(*rl.LitList)
	assert.Len(t, list.Elements, 3)
}

func TestConvert_LitMap(t *testing.T) {
	stmt := firstStmt(t, `a = {"x": 1, "y": 2}`)
	assign := stmt.(*rl.Assign)
	m := assign.Values[0].(*rl.LitMap)
	assert.Len(t, m.Entries, 2)
}

// --- Call ---

func TestConvert_Call(t *testing.T) {
	expr := firstExpr(t, `print("hello", 42)`)
	call := expr.(*rl.Call)
	fn := call.Func.(*rl.Identifier)
	assert.Equal(t, "print", fn.Name)
	assert.Len(t, call.Args, 2)
}

func TestConvert_Call_NamedArg(t *testing.T) {
	expr := firstExpr(t, `foo(1, bar=2)`)
	call := expr.(*rl.Call)
	assert.Len(t, call.Args, 1)
	require.Len(t, call.NamedArgs, 1)
	assert.Equal(t, "bar", call.NamedArgs[0].Name)
}

// --- If ---

func TestConvert_If_ElseIf_Else(t *testing.T) {
	src := `if x:
    1
else if y:
    2
else:
    3`
	stmt := firstStmt(t, src)
	ifn := stmt.(*rl.If)
	require.Len(t, ifn.Branches, 3)
	assert.NotNil(t, ifn.Branches[0].Condition) // if
	assert.NotNil(t, ifn.Branches[1].Condition) // else if
	assert.Nil(t, ifn.Branches[2].Condition)    // else
}

// --- While ---

func TestConvert_WhileLoop(t *testing.T) {
	src := `while x > 0:
    x--`
	stmt := firstStmt(t, src)
	w := stmt.(*rl.WhileLoop)
	assert.NotNil(t, w.Condition)
	assert.Len(t, w.Body, 1)
}

func TestConvert_WhileInfinite(t *testing.T) {
	src := `while:
    break`
	stmt := firstStmt(t, src)
	w := stmt.(*rl.WhileLoop)
	assert.Nil(t, w.Condition, "infinite while should have nil condition")
}

// --- Defer ---

func TestConvert_Defer(t *testing.T) {
	src := `defer:
    print("done")`
	stmt := firstStmt(t, src)
	d := stmt.(*rl.Defer)
	assert.False(t, d.IsErrDefer)
	assert.Len(t, d.Body, 1)
}

func TestConvert_ErrDefer(t *testing.T) {
	src := `errdefer:
    print("error")`
	stmt := firstStmt(t, src)
	d := stmt.(*rl.Defer)
	assert.True(t, d.IsErrDefer)
}

// --- Return and Yield ---

func TestConvert_Return(t *testing.T) {
	src := `fn foo():
    return 1, 2`
	stmt := firstStmt(t, src)
	fn := stmt.(*rl.FnDef)
	ret := fn.Body[0].(*rl.Return)
	assert.Len(t, ret.Values, 2)
}

func TestConvert_Yield(t *testing.T) {
	src := `fn foo():
    yield 42`
	stmt := firstStmt(t, src)
	fn := stmt.(*rl.FnDef)
	y := fn.Body[0].(*rl.Yield)
	assert.Len(t, y.Values, 1)
}

// --- Del ---

func TestConvert_Del(t *testing.T) {
	stmt := firstStmt(t, "del x, y")
	d := stmt.(*rl.Del)
	assert.Len(t, d.Targets, 2)
}

// --- Unpacking assign ---

func TestConvert_UnpackingAssign(t *testing.T) {
	stmt := firstStmt(t, "a, b = 1, 2")
	assign := stmt.(*rl.Assign)
	assert.True(t, assign.IsUnpacking)
	assert.Len(t, assign.Targets, 2)
	assert.Len(t, assign.Values, 2)
}

// --- Comments and shebang are skipped ---

func TestConvert_CommentsSkipped(t *testing.T) {
	src := `// this is a comment
x = 1
// another comment`
	sf := convertSource(t, src)
	require.Len(t, sf.Stmts, 1)
	_, ok := sf.Stmts[0].(*rl.Assign)
	assert.True(t, ok)
}
