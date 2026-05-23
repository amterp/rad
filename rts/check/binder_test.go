package check_test

import (
	"testing"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseFile is a thin helper around the parser+converter for binder tests.
// We keep these tests at the check package level so they exercise the
// public Resolve API the way real callers (the checker and LSP) will.
func parseFile(t *testing.T, src string) *rl.SourceFile {
	t.Helper()
	parser, err := rts.NewRadParser()
	require.NoError(t, err)
	defer parser.Close()
	tree := parser.Parse(src)
	return rts.ConvertCST(tree.Root(), src, "binder_test.rad")
}

func TestResolve_NilFileReturnsNil(t *testing.T) {
	assert.Nil(t, check.Resolve(nil))
}

func TestResolve_EmptyFileHasFileAndBuiltinScopes(t *testing.T) {
	r := check.Resolve(parseFile(t, ""))
	require.NotNil(t, r)
	require.NotNil(t, r.File)
	require.NotNil(t, r.Builtin)
	assert.Equal(t, check.ScopeFile, r.File.Kind)
	assert.Equal(t, check.ScopeBuiltin, r.Builtin.Kind)
	// File scope chains up through the builtin scope - this is what
	// makes ambient names like `print` reachable from lookups.
	assert.Same(t, r.Builtin, r.File.Parent)
}

func TestResolve_TopLevelAssignmentDeclaresLocal(t *testing.T) {
	r := check.Resolve(parseFile(t, "x = 1\n"))
	require.NotNil(t, r)
	sym := r.File.Lookup("x")
	require.NotNil(t, sym, "x should be declared at file scope")
	assert.Equal(t, "x", sym.Name)
	assert.Equal(t, check.SymLocal, sym.Kind)
	assert.Same(t, r.File, sym.Scope)
}

func TestResolve_HoistedFunctionVisibleBeforeDecl(t *testing.T) {
	// Calling greet() before its definition must still resolve to it -
	// this is the whole point of hoisting top-level functions.
	src := "greet()\n\nfn greet():\n    print(\"hi\")\n"
	r := check.Resolve(parseFile(t, src))
	require.NotNil(t, r)

	sym := r.File.Lookup("greet")
	require.NotNil(t, sym)
	assert.Equal(t, check.SymHoistedFn, sym.Kind)
}

func TestResolve_BuiltinSymbolSynthesizedOnFirstUse(t *testing.T) {
	r := check.Resolve(parseFile(t, "print(\"hi\")\n"))
	require.NotNil(t, r)

	sym := r.Builtin.Symbols["print"]
	require.NotNil(t, sym, "print should be in the builtin scope after use")
	assert.Equal(t, check.SymBuiltin, sym.Kind)
	assert.Same(t, r.Builtin, sym.Scope)
}

func TestResolve_IdentifierUseLinksToDeclaration(t *testing.T) {
	// Both the LHS-of-second-assign and the second-line RHS identifier
	// should resolve to the *same* symbol introduced by line 1.
	src := "x = 1\ny = x\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	declSym := r.File.Lookup("x")
	require.NotNil(t, declSym)

	// Walk to the second statement's RHS identifier and check it
	// resolved to declSym. We pull it out structurally rather than by
	// span scan so the test stays robust to whitespace.
	require.GreaterOrEqual(t, len(file.Stmts), 2)
	secondAssign, ok := file.Stmts[1].(*rl.Assign)
	require.True(t, ok)
	require.Len(t, secondAssign.Values, 1)
	ident, ok := secondAssign.Values[0].(*rl.Identifier)
	require.True(t, ok)

	useSym, ok := r.Uses[ident]
	require.True(t, ok, "RHS identifier should be recorded in Uses")
	assert.Same(t, declSym, useSym)
}

func TestResolve_RebindingDoesNotCreateNewSymbol(t *testing.T) {
	// A second assignment to the same name re-binds, it doesn't shadow.
	// Both assignments share one Symbol so the LSP can find every
	// reference and the type checker has one place to record narrowing.
	src := "x = 1\nx = 2\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	sym := r.File.Lookup("x")
	require.NotNil(t, sym)

	// Second statement's target identifier should be recorded as a use
	// of the same symbol, not as a new declaration.
	secondAssign := file.Stmts[1].(*rl.Assign)
	target := secondAssign.Targets[0].(*rl.Identifier)
	assert.Same(t, sym, r.Uses[target])
	// The first assignment's target is the canonical decl node.
	firstAssign := file.Stmts[0].(*rl.Assign)
	assert.Same(t, sym, r.Decls[firstAssign.Targets[0]])
}
