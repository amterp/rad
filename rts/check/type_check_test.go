package check_test

import (
	"testing"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// typeInfoFromSrc is the typical entry point for type-checker tests:
// parse the source, run the binder, run the type checker. Returns
// both the parsed file and the type info so tests can index into
// AST nodes for ExprTypes lookups.
func typeInfoFromSrc(t *testing.T, src string) (*rl.SourceFile, *check.TypeInfo, *check.Resolved) {
	t.Helper()
	file := parseFile(t, src)
	resolved := check.Resolve(file)
	require.NotNil(t, resolved)
	info := check.TypeCheck(file, resolved)
	require.NotNil(t, info)
	return file, info, resolved
}

func TestTypeCheck_NilInputsReturnNil(t *testing.T) {
	assert.Nil(t, check.TypeCheck(nil, nil))
}

func TestTypeCheck_IntLiteralSynthesizesInt(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = 5\n")
	assign := file.Stmts[0].(*rl.Assign)
	lit := assign.Values[0].(*rl.LitInt)
	gotExpr, ok := info.ExprTypes[lit]
	require.True(t, ok, "ExprTypes should record the literal's type")
	assert.Equal(t, rl.T_INT, gotExpr.Name())
}

func TestTypeCheck_FloatLiteralSynthesizesFloat(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = 1.5\n")
	lit := file.Stmts[0].(*rl.Assign).Values[0].(*rl.LitFloat)
	assert.Equal(t, rl.T_FLOAT, info.ExprTypes[lit].Name())
}

func TestTypeCheck_StringLiteralSynthesizesStr(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = \"hi\"\n")
	lit := file.Stmts[0].(*rl.Assign).Values[0].(*rl.LitString)
	assert.Equal(t, rl.T_STR, info.ExprTypes[lit].Name())
}

func TestTypeCheck_BoolLiteralSynthesizesBool(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = true\n")
	lit := file.Stmts[0].(*rl.Assign).Values[0].(*rl.LitBool)
	assert.Equal(t, rl.T_BOOL, info.ExprTypes[lit].Name())
}

func TestTypeCheck_AssignmentRecordsSymbolType(t *testing.T) {
	// After `x = 5`, the symbol's recorded type must be int. This is
	// what enables a downstream synth(`y = x`) to know y is int too.
	file, info, resolved := typeInfoFromSrc(t, "x = 5\n")
	target := file.Stmts[0].(*rl.Assign).Targets[0].(*rl.Identifier)
	sym := resolved.Uses[target]
	require.NotNil(t, sym)
	got, ok := info.SymbolTypes[sym]
	require.True(t, ok, "symbol type should be recorded after assignment")
	assert.Equal(t, rl.T_INT, got.Name())
}

func TestTypeCheck_IdentifierReadsRecordedSymbolType(t *testing.T) {
	// `x = 5; y = x` - the use of x in the RHS of the second assign
	// should synth to int (the type recorded by the first assign),
	// and y should inherit that.
	file, info, resolved := typeInfoFromSrc(t, "x = 5\ny = x\n")
	secondAssign := file.Stmts[1].(*rl.Assign)
	xUse := secondAssign.Values[0].(*rl.Identifier)
	gotUse, ok := info.ExprTypes[xUse]
	require.True(t, ok)
	assert.Equal(t, rl.T_INT, gotUse.Name())

	ySym := resolved.Uses[secondAssign.Targets[0].(*rl.Identifier)]
	require.NotNil(t, ySym)
	assert.Equal(t, rl.T_INT, info.SymbolTypes[ySym].Name())
}

func TestTypeCheck_ForwardReferenceFallsBackToDynamic(t *testing.T) {
	// Inside a function body, referring to a name whose type hasn't
	// been recorded yet (here: the function's own name during its
	// own body) yields Dynamic. Phase 2e will revisit this for
	// genuine mutual-recursion via Tarjan SCC.
	src := "fn f():\n    g()\n"
	file, info, _ := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	call := fn.Body[0].(*rl.ExprStmt).Expr.(*rl.Call)
	callee := call.Func.(*rl.Identifier)
	// `g` is undefined; the binder doesn't put it in Uses, so synth
	// returns Dynamic (the fallback for unresolved names).
	got := info.ExprTypes[callee]
	require.NotNil(t, got)
	assert.Equal(t, rl.T_DYNAMIC, got.Name())
}

func TestTypeCheck_NoIssuesEmittedYet(t *testing.T) {
	// Sanity check: Phase 2a doesn't emit any type diagnostics. A
	// type-mismatched assignment like `x: int = "hi"` would be
	// caught by Phase 2b/c, not this commit.
	_, info, _ := typeInfoFromSrc(t, "x = 5\ny = \"hi\"\n")
	assert.Empty(t, info.Issues)
}
