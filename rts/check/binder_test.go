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

func TestResolve_FnDefParamsBindInFunctionScope(t *testing.T) {
	src := "fn greet(name):\n    print(name)\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	// Locate the function body's print(name) call's name argument.
	fn := file.Stmts[0].(*rl.FnDef)
	exprStmt := fn.Body[0].(*rl.ExprStmt)
	call := exprStmt.Expr.(*rl.Call)
	require.Len(t, call.Args, 1)
	nameUse := call.Args[0].(*rl.Identifier)

	sym, ok := r.Uses[nameUse]
	require.True(t, ok, "param use should resolve")
	assert.Equal(t, check.SymParam, sym.Kind)
	assert.Equal(t, "name", sym.Name)
	// The param scope should be a function scope owned by the FnDef.
	require.NotNil(t, sym.Scope)
	assert.Equal(t, check.ScopeFunction, sym.Scope.Kind)
	assert.Same(t, fn, sym.Scope.Owner)
}

func TestResolve_FnLocalShadowsFileScopeBinding(t *testing.T) {
	// `x = 2` inside the function introduces a function-scope local;
	// the inner use of x must resolve to that local, not to the
	// file-scope `x`. This matches Python and Rad runtime behavior.
	src := "x = 1\nfn f():\n    x = 2\n    print(x)\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	fileSym := r.File.Lookup("x")
	require.NotNil(t, fileSym)
	assert.Equal(t, r.File, fileSym.Scope)

	fn := file.Stmts[1].(*rl.FnDef)
	innerAssign := fn.Body[0].(*rl.Assign)
	innerTarget := innerAssign.Targets[0].(*rl.Identifier)
	innerSym, ok := r.Uses[innerTarget]
	require.True(t, ok)
	assert.NotSame(t, fileSym, innerSym, "inner x must be a new symbol, not the file-scope one")
	assert.Equal(t, check.ScopeFunction, innerSym.Scope.Kind)

	// The print(x) use should also resolve to the inner symbol.
	printCall := fn.Body[1].(*rl.ExprStmt).Expr.(*rl.Call)
	xUse := printCall.Args[0].(*rl.Identifier)
	assert.Same(t, innerSym, r.Uses[xUse])
}

func TestResolve_CompoundAssignRebindEnclosing(t *testing.T) {
	// `+=` and friends set UpdateEnclosing: they must NOT introduce a
	// new local. The compound-op needs an existing binding to operate
	// on, and creating a fresh local at the function scope would lose
	// every previous value.
	src := "x = 0\nfn f():\n    x += 1\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	fileSym := r.File.Lookup("x")
	require.NotNil(t, fileSym)

	fn := file.Stmts[1].(*rl.FnDef)
	compoundAssign := fn.Body[0].(*rl.Assign)
	require.True(t, compoundAssign.UpdateEnclosing, "compound assigns set UpdateEnclosing")
	target := compoundAssign.Targets[0].(*rl.Identifier)
	useSym, ok := r.Uses[target]
	require.True(t, ok, "compound-assign target should resolve to existing binding")
	assert.Same(t, fileSym, useSym)
}

func TestResolve_LambdaParamVisibleInBody(t *testing.T) {
	// A lambda used as a callback should have its params visible
	// inside the body. We use a simple list-map pattern.
	src := "f = fn(x) x + 1\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	assign := file.Stmts[0].(*rl.Assign)
	lambda := assign.Values[0].(*rl.Lambda)
	require.Len(t, lambda.Body, 1)
	// Expression-form lambda has the expression directly in Body[0],
	// not wrapped in an ExprStmt.
	add := lambda.Body[0].(*rl.OpBinary)
	xUse := add.Left.(*rl.Identifier)

	sym, ok := r.Uses[xUse]
	require.True(t, ok, "lambda body should see the param")
	assert.Equal(t, check.SymParam, sym.Kind)
	assert.Equal(t, check.ScopeLambda, sym.Scope.Kind)
}

func TestResolve_ParamDefaultEvaluatesInEnclosingScope(t *testing.T) {
	// `fn f(n = greeting)` - the default `greeting` reference must
	// resolve in the enclosing (file) scope, not against any param.
	// This matches Python's behavior and avoids surprise from
	// "later params shadow earlier defaults" rules.
	src := "greeting = \"hi\"\nfn f(n = greeting):\n    print(n)\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	fileSym := r.File.Lookup("greeting")
	require.NotNil(t, fileSym)

	fn := file.Stmts[1].(*rl.FnDef)
	require.NotNil(t, fn.Typing)
	require.Len(t, fn.Typing.Params, 1)
	dflt := fn.Typing.Params[0].DefaultAST
	require.NotNil(t, dflt)
	defaultIdent, ok := dflt.Node.(*rl.Identifier)
	require.True(t, ok)

	useSym, ok := r.Uses[defaultIdent]
	require.True(t, ok)
	assert.Same(t, fileSym, useSym, "default must resolve to enclosing scope binding")
}

func TestResolve_ForLoopVarVisibleInBodyOnly(t *testing.T) {
	src := "items = [1,2,3]\nfor x in items:\n    print(x)\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	// The loop var is visible inside the body and remains visible
	// after - Rad's runtime writes loop vars into the enclosing env,
	// so `for i in range(3): pass; print(i)` is valid Rad. The binder
	// matches that semantics.
	forLoop := file.Stmts[1].(*rl.ForLoop)
	call := forLoop.Body[0].(*rl.ExprStmt).Expr.(*rl.Call)
	xUse := call.Args[0].(*rl.Identifier)
	sym, ok := r.Uses[xUse]
	require.True(t, ok)
	assert.Equal(t, check.SymLoopVar, sym.Kind)
	assert.Equal(t, check.ScopeFile, sym.Scope.Kind)
	assert.Same(t, sym, r.File.Lookup("x"))
}

func TestResolve_ForLoopIterEvaluatedInEnclosingScope(t *testing.T) {
	// `items` is a file-level binding; the iter source must resolve
	// against the file scope, not against the (not-yet-existing)
	// loop scope.
	src := "items = [1,2,3]\nfor x in items:\n    pass\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	fileItems := r.File.Lookup("items")
	require.NotNil(t, fileItems)

	forLoop := file.Stmts[1].(*rl.ForLoop)
	iterIdent := forLoop.Iter.(*rl.Identifier)
	assert.Same(t, fileItems, r.Uses[iterIdent])
}

func TestResolve_ForLoopMultiVarsAllDeclared(t *testing.T) {
	// `for i, v in enumerate(xs):` binds both i and v in the body.
	src := "xs = [1,2,3]\nfor i, v in xs:\n    print(i)\n    print(v)\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	forLoop := file.Stmts[1].(*rl.ForLoop)
	require.Equal(t, []string{"i", "v"}, forLoop.Vars)

	iUse := forLoop.Body[0].(*rl.ExprStmt).Expr.(*rl.Call).Args[0].(*rl.Identifier)
	vUse := forLoop.Body[1].(*rl.ExprStmt).Expr.(*rl.Call).Args[0].(*rl.Identifier)
	require.NotNil(t, r.Uses[iUse])
	require.NotNil(t, r.Uses[vUse])
	assert.Equal(t, "i", r.Uses[iUse].Name)
	assert.Equal(t, "v", r.Uses[vUse].Name)
}

func TestResolve_WhileLoopBodyShareEnclosingScope(t *testing.T) {
	// Rad has no per-block scope: variables introduced inside a while
	// body remain visible after the loop. The binder must reflect that
	// so a future undefined-variable check doesn't falsely flag
	// `print(tmp)` after the loop.
	src := "i = 0\nwhile i < 3:\n    tmp = i\n    i += 1\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	whileLoop := file.Stmts[1].(*rl.WhileLoop)
	tmpAssign := whileLoop.Body[0].(*rl.Assign)
	tmpDecl := tmpAssign.Targets[0].(*rl.Identifier)
	tmpSym := r.Uses[tmpDecl]
	require.NotNil(t, tmpSym)
	assert.Equal(t, check.ScopeFile, tmpSym.Scope.Kind)
	assert.Same(t, tmpSym, r.File.Lookup("tmp"))

	// `i += 1` (compound assign) rebinds the file-scope i as before.
	iAssign := whileLoop.Body[1].(*rl.Assign)
	require.True(t, iAssign.UpdateEnclosing)
	iTarget := iAssign.Targets[0].(*rl.Identifier)
	assert.Same(t, r.File.Lookup("i"), r.Uses[iTarget])
}

func TestResolve_ListCompVarsShareEnclosingScope(t *testing.T) {
	// `[x * 2 for x in xs]` - the comprehension's loop var lives in the
	// enclosing scope at runtime (executeForLoop SetVars into the
	// current env), so the binder declares it there too. Side effect:
	// post-comp references to x resolve, matching what the user sees.
	src := "xs = [1,2,3]\nresult = [x * 2 for x in xs]\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	resultAssign := file.Stmts[1].(*rl.Assign)
	comp := resultAssign.Values[0].(*rl.ListComp)
	mult := comp.Expr.(*rl.OpBinary)
	xUse := mult.Left.(*rl.Identifier)
	sym := r.Uses[xUse]
	require.NotNil(t, sym)
	assert.Equal(t, check.ScopeFile, sym.Scope.Kind)
	assert.Same(t, sym, r.File.Lookup("x"))
}

func TestResolve_SwitchCaseBodySharesEnclosingScope(t *testing.T) {
	// Switch case bodies run in the enclosing env. A local declared in
	// one case body's first assignment binds in that enclosing scope;
	// a same-named assignment in a sibling case rebinds the same
	// symbol (control flow ensures only one runs per dispatch).
	src := "x = 1\nswitch x:\n    case 1:\n        tmp = \"a\"\n    case 2:\n        tmp = \"b\"\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	tmpSym := r.File.Lookup("tmp")
	require.NotNil(t, tmpSym, "case-body assignment binds in the enclosing scope")

	sw := file.Stmts[1].(*rl.Switch)
	require.GreaterOrEqual(t, len(sw.Cases), 2)
	caseA := sw.Cases[0].Alt.(*rl.SwitchCaseBlock)
	caseB := sw.Cases[1].Alt.(*rl.SwitchCaseBlock)
	tmpA := caseA.Stmts[0].(*rl.Assign).Targets[0].(*rl.Identifier)
	tmpB := caseB.Stmts[0].(*rl.Assign).Targets[0].(*rl.Identifier)
	// Both cases bind the SAME symbol (one logical variable, one
	// runtime slot, even if the case bodies are mutually exclusive).
	assert.Same(t, tmpSym, r.Uses[tmpA])
	assert.Same(t, tmpSym, r.Uses[tmpB])
}

func TestResolve_SwitchDiscriminantInEnclosingScope(t *testing.T) {
	src := "x = 1\nswitch x:\n    case 1:\n        pass\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	fileX := r.File.Lookup("x")
	require.NotNil(t, fileX)
	sw := file.Stmts[1].(*rl.Switch)
	discIdent := sw.Discriminant.(*rl.Identifier)
	assert.Same(t, fileX, r.Uses[discIdent])
}

func TestResolve_DeferBodySharesEnclosingScope(t *testing.T) {
	// A defer body runs later in the enclosing function's env (the
	// interpreter uses runBlock, not a child env), so any local it
	// declares becomes visible to the rest of the function from the
	// point of declaration onward.
	src := "fn f():\n    defer:\n        tmp = 1\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	fn := file.Stmts[0].(*rl.FnDef)
	defer_ := fn.Body[0].(*rl.Defer)
	tmpDecl := defer_.Body[0].(*rl.Assign).Targets[0].(*rl.Identifier)
	tmpSym := r.Uses[tmpDecl]
	require.NotNil(t, tmpSym)
	// The enclosing scope is the function body, not a per-defer scope.
	assert.Equal(t, check.ScopeFunction, tmpSym.Scope.Kind)
	assert.Same(t, fn, tmpSym.Scope.Owner)
}

func TestResolve_CmdBlockArgsVisibleAtFileScope(t *testing.T) {
	// Cmd args are populated by the runtime into the file env before
	// the callback runs. The binder mirrors that: args become bindings
	// at file scope (with kind SymCmdArg so LSP can still point users
	// at the cmd_block's decl).
	src := "command greet:\n    name str\n    calls fn():\n        print(name)\n"
	file := parseFile(t, src)
	require.NotNil(t, file)
	require.Len(t, file.Cmds, 1, "expected one cmd_block")
	r := check.Resolve(file)
	require.NotNil(t, r)

	nameSym := r.File.Lookup("name")
	require.NotNil(t, nameSym, "cmd args live at file scope")
	assert.Equal(t, check.SymCmdArg, nameSym.Kind)

	cmd := file.Cmds[0]
	require.True(t, cmd.Callback.IsLambda, "expected inline-lambda callback")
	require.NotNil(t, cmd.Callback.Lambda)
	cb := cmd.Callback.Lambda
	require.GreaterOrEqual(t, len(cb.Body), 1)
	call := cb.Body[0].(*rl.ExprStmt).Expr.(*rl.Call)
	nameUse := call.Args[0].(*rl.Identifier)
	assert.Same(t, nameSym, r.Uses[nameUse], "lambda body sees cmd arg via file scope")
}

func TestResolve_CmdBlockArgsVisibleToNamedCallback(t *testing.T) {
	// A `calls handler` callback is just a hoisted function. The
	// runtime populates cmd args into the file env before invoking
	// handler, so handler's body resolves `name` against file scope.
	// The binder must mirror this: cmd args at file scope.
	src := "command greet:\n    name str\n    calls handler\n\nfn handler():\n    print(name)\n"
	file := parseFile(t, src)
	require.Len(t, file.Cmds, 1)
	r := check.Resolve(file)
	require.NotNil(t, r)

	nameSym := r.File.Lookup("name")
	require.NotNil(t, nameSym)

	handler := file.Stmts[0].(*rl.FnDef)
	call := handler.Body[0].(*rl.ExprStmt).Expr.(*rl.Call)
	nameUse := call.Args[0].(*rl.Identifier)
	assert.Same(t, nameSym, r.Uses[nameUse], "handler resolves cmd arg via file scope")
}

func TestResolve_NamedCallbackTargetIsResolvable(t *testing.T) {
	// `calls handler` is a function reference: the binder records the
	// callback identifier in Resolved.Uses, bound to the same symbol as
	// the `fn handler` definition. That use index is what powers LSP
	// goto-def / find-refs / rename / hover on the callback site.
	src := "command run:\n    calls handler\n\nfn handler():\n    pass\n"
	file := parseFile(t, src)
	require.NotNil(t, file)
	r := check.Resolve(file)
	require.NotNil(t, r)

	handlerSym := r.File.Lookup("handler")
	require.NotNil(t, handlerSym, "callback target must be visible at file scope")
	assert.Equal(t, check.SymHoistedFn, handlerSym.Kind)

	require.NotEmpty(t, file.Cmds)
	cbIdent := file.Cmds[0].Callback.Identifier
	require.NotNil(t, cbIdent, "named callback should be an Identifier node")
	assert.Same(t, handlerSym, r.Uses[cbIdent],
		"callback identifier must resolve to the fn's symbol")
}

func TestResolve_DuplicateFnParamEmitsIssue(t *testing.T) {
	src := "fn add(x, x):\n    return x\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	// Exactly one duplicate-parameter issue should fire.
	dupes := 0
	for _, issue := range r.Issues {
		if issue.Code == rl.ErrDuplicateParameter {
			dupes++
		}
	}
	assert.Equal(t, 1, dupes, "expected one duplicate-param issue")
}

func TestResolve_DuplicateLambdaParamEmitsIssue(t *testing.T) {
	src := "f = fn(a, a) a\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	dupes := 0
	for _, issue := range r.Issues {
		if issue.Code == rl.ErrDuplicateParameter {
			dupes++
		}
	}
	assert.Equal(t, 1, dupes)
}

func TestResolve_ShadowingNotADuplicate(t *testing.T) {
	// `fn f(x)` where x exists at file scope is *shadowing*, not a
	// duplicate param. The check must only fire for two params in the
	// same parameter list.
	src := "x = 1\nfn f(x):\n    return x\n"
	file := parseFile(t, src)
	r := check.Resolve(file)
	require.NotNil(t, r)

	for _, issue := range r.Issues {
		assert.NotEqual(t, rl.ErrDuplicateParameter, issue.Code,
			"shadowing must not produce a duplicate-param error")
	}
}

func TestCheck_CmdCallbackBuiltinAlone_NoFalseWarning(t *testing.T) {
	// Regression: a `calls <builtin>` callback whose builtin is never
	// otherwise referenced must not be flagged as undefined. The binder
	// resolves a named callback like any other identifier, including
	// against the builtin set, so `calls print` is clean even when
	// nothing else in the script mentions `print`.
	src := "command run:\n    calls print\n"
	parser, err := rts.NewRadParser()
	require.NoError(t, err)
	defer parser.Close()
	tree := parser.Parse(src)
	chk := check.NewCheckerWithTree(tree, parser, src, rts.ConvertCST(tree.Root(), src, "test.rad"))
	res, err := chk.Check()
	require.NoError(t, err)
	for _, d := range res.Diagnostics {
		if d.Code != nil && (*d.Code == rl.ErrUnknownFunction || *d.Code == rl.ErrUndefinedVariable) {
			t.Fatalf("unexpected undefined-callback diagnostic: %s", d.Message)
		}
	}
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
