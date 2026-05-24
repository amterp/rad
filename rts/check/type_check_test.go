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

// hasIssue is a tiny helper for the arity tests below: did the
// type-check info include at least one diagnostic with the given
// error code?
func hasIssue(info *check.TypeInfo, code rl.Error) bool {
	for _, i := range info.Issues {
		if i.Code == code {
			return true
		}
	}
	return false
}

func TestTypeCheck_BuiltinCallReturnTypeRecorded(t *testing.T) {
	// `len(...)` returns int; the Call expression should synth to int
	// even when nested inside a larger expression.
	file, info, _ := typeInfoFromSrc(t, "x = len(\"hi\")\n")
	assign := file.Stmts[0].(*rl.Assign)
	call := assign.Values[0].(*rl.Call)
	got := info.ExprTypes[call]
	require.NotNil(t, got)
	assert.Equal(t, rl.T_INT, got.Name())
}

func TestTypeCheck_BuiltinTooFewArgsFiresWrongArgCount(t *testing.T) {
	// `replace(_original, _find, _replace)` requires 3 positional args.
	_, info, _ := typeInfoFromSrc(t, "x = replace(\"a\", \"b\")\n")
	assert.True(t, hasIssue(info, rl.ErrWrongArgCount),
		"expected ErrWrongArgCount for too few args")
}

func TestTypeCheck_BuiltinTooManyArgsFiresWrongArgCount(t *testing.T) {
	// `len` accepts exactly one positional arg.
	_, info, _ := typeInfoFromSrc(t, "x = len(\"a\", \"b\")\n")
	assert.True(t, hasIssue(info, rl.ErrWrongArgCount),
		"expected ErrWrongArgCount for too many args")
}

func TestTypeCheck_BuiltinVariadicAcceptsAnyCount(t *testing.T) {
	// `print(*_items, ...)` is variadic; calls with 0, 1, or N args
	// must all be accepted without firing the arity check.
	for _, src := range []string{
		"print()\n",
		"print(\"hi\")\n",
		"print(\"a\", \"b\", \"c\")\n",
	} {
		_, info, _ := typeInfoFromSrc(t, src)
		assert.False(t, hasIssue(info, rl.ErrWrongArgCount),
			"variadic call should not flag arity: %q", src)
	}
}

func TestTypeCheck_BuiltinUnknownNamedArg(t *testing.T) {
	// `print` accepts `sep` and `end` as named-only args; anything
	// else is an unknown-named-arg error.
	_, info, _ := typeInfoFromSrc(t, "print(\"hi\", bogus=1)\n")
	assert.True(t, hasIssue(info, rl.ErrInvalidArgType),
		"expected ErrInvalidArgType for unknown named arg")
}

func TestTypeCheck_BuiltinKnownNamedArgOK(t *testing.T) {
	_, info, _ := typeInfoFromSrc(t, "print(\"hi\", end=\"\")\n")
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_ArgTypeMismatchEmitsHint(t *testing.T) {
	// `len` expects str/list/map. Passing an int should surface a
	// type-mismatch hint - not an error, so the runtime still gets
	// to fire its richer value-aware message.
	_, info, _ := typeInfoFromSrc(t, "x = len(5)\n")
	require.NotEmpty(t, info.Issues)
	found := false
	for _, i := range info.Issues {
		if i.Code == rl.ErrTypeMismatch && i.Severity == check.IssueHint {
			found = true
			break
		}
	}
	assert.True(t, found, "expected a Hint-severity ErrTypeMismatch issue")
}

func TestTypeCheck_ArgTypeCorrectIsSilent(t *testing.T) {
	// `len("hi")` - str is a valid arg type for len. No diagnostics.
	_, info, _ := typeInfoFromSrc(t, "x = len(\"hi\")\n")
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrTypeMismatch, i.Code,
			"unexpected type-mismatch on a valid call: %s", i.Message)
	}
}

func TestTypeCheck_NamedArgTypeMismatchEmitsHint(t *testing.T) {
	// `print(... sep: str = ...)` - sep expects str. Passing an int
	// at the named-arg site should surface a type-mismatch hint.
	_, info, _ := typeInfoFromSrc(t, "print(\"hi\", sep=5)\n")
	found := false
	for _, i := range info.Issues {
		if i.Code == rl.ErrTypeMismatch && i.Severity == check.IssueHint {
			found = true
			break
		}
	}
	assert.True(t, found, "expected type-mismatch hint on named arg")
}

func TestTypeCheck_UFCSCallReceiverCountsAsFirstArg(t *testing.T) {
	// `"hi".upper()` desugars to `upper("hi")`. Without UFCS-awareness
	// the arity check would say "missing required arg". With it, the
	// receiver counts as the first positional arg and no diagnostic
	// fires.
	_, info, _ := typeInfoFromSrc(t, "x = \"hi\".upper()\n")
	assert.False(t, hasIssue(info, rl.ErrWrongArgCount),
		"UFCS receiver must count as the implicit first arg")
}

func TestTypeCheck_NoFalsePositivesOnSimpleAssignments(t *testing.T) {
	// Sanity check: type-correct trivial assignments should not
	// trigger any type-checker issues. Useful as a baseline before
	// more complex tests.
	_, info, _ := typeInfoFromSrc(t, "x = 5\ny = \"hi\"\n")
	assert.Empty(t, info.Issues)
}

// --- Operator tests ---------------------------------------------------
//
// These exercise the binary, unary, ternary, fallback, and catch
// handlers added in Phase 2c. The diagnostic-emitting cases all check
// for Hint severity (matching the precedent set by the per-arg check).

// hasOpIssue reports whether info recorded an ErrInvalidTypeForOp
// diagnostic at the expected severity. Used by the negative tests
// below to assert "the type checker flagged this op."
func hasOpIssue(info *check.TypeInfo) bool {
	for _, i := range info.Issues {
		if i.Code == rl.ErrInvalidTypeForOp && i.Severity == check.IssueHint {
			return true
		}
	}
	return false
}

// exprTypeOf returns the synthesized type for a top-level Assign's RHS.
// Most operator tests want to assert "the result type of `a + b` was X";
// having this in one place keeps each test to a couple of lines.
func exprTypeOf(t *testing.T, file *rl.SourceFile, info *check.TypeInfo) rl.TypingT {
	t.Helper()
	assign, ok := file.Stmts[0].(*rl.Assign)
	require.True(t, ok, "expected top-level Assign")
	require.NotEmpty(t, assign.Values)
	got, ok := info.ExprTypes[assign.Values[0]]
	require.True(t, ok, "ExprTypes should record the RHS type")
	return got
}

func TestTypeCheck_IntPlusIntSynthesizesInt(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = 1 + 2\n")
	assert.Equal(t, rl.T_INT, exprTypeOf(t, file, info).Name())
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_IntPlusFloatSynthesizesFloat(t *testing.T) {
	// Mixed int/float arithmetic widens to float via the lone implicit
	// numeric widening rule.
	file, info, _ := typeInfoFromSrc(t, "x = 1 + 2.5\n")
	assert.Equal(t, rl.T_FLOAT, exprTypeOf(t, file, info).Name())
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_IntDivIntSynthesizesFloat(t *testing.T) {
	// Rad's `/` is true division, like Python 3 - int/int returns float,
	// not int. This is non-obvious and a likely source of bugs, so it's
	// worth a dedicated test.
	file, info, _ := typeInfoFromSrc(t, "x = 5 / 2\n")
	assert.Equal(t, rl.T_FLOAT, exprTypeOf(t, file, info).Name())
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_IntModIntSynthesizesInt(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = 5 % 2\n")
	assert.Equal(t, rl.T_INT, exprTypeOf(t, file, info).Name())
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_StrPlusStrSynthesizesStr(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = \"a\" + \"b\"\n")
	assert.Equal(t, rl.T_STR, exprTypeOf(t, file, info).Name())
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_StrTimesIntSynthesizesStr(t *testing.T) {
	// String repetition; both `"a" * 3` and `3 * "a"` are valid.
	file, info, _ := typeInfoFromSrc(t, "x = \"a\" * 3\n")
	assert.Equal(t, rl.T_STR, exprTypeOf(t, file, info).Name())
	assert.Empty(t, info.Issues)
}

func TestTypeCheck_IntPlusStrFlagsHint(t *testing.T) {
	// This is the migration case from v0.9 - `+` no longer coerces.
	// The runtime would emit ErrInvalidTypeForOp; we want the static
	// checker to surface it as a hint pre-execution.
	_, info, _ := typeInfoFromSrc(t, "x = 1 + \"hi\"\n")
	assert.True(t, hasOpIssue(info),
		"expected Hint-severity ErrInvalidTypeForOp for int + str")
}

func TestTypeCheck_LessThanReturnsBool(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = 1 < 2\n")
	assert.Equal(t, rl.T_BOOL, exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_LessThanRejectsMixedTypes(t *testing.T) {
	// `<`/`<=`/`>`/`>=` require numeric-vs-numeric or str-vs-str. int
	// vs str isn't well-defined and the runtime rejects it.
	_, info, _ := typeInfoFromSrc(t, "x = 1 < \"hi\"\n")
	assert.True(t, hasOpIssue(info),
		"expected hint for incompatible comparison")
}

func TestTypeCheck_EqualityAcceptsAnyOperands(t *testing.T) {
	// `==`/`!=` are total - the runtime can compare any two values,
	// even across types (the result is just "false"). No diagnostic
	// should fire on a type-mismatched equality.
	file, info, _ := typeInfoFromSrc(t, "x = 1 == \"hi\"\n")
	assert.Equal(t, rl.T_BOOL, exprTypeOf(t, file, info).Name())
	assert.False(t, hasOpIssue(info))
}

func TestTypeCheck_AndReturnsBool(t *testing.T) {
	// `and` ultimately boolifies the right operand (or returns false),
	// so the static result is always bool regardless of operand types.
	file, info, _ := typeInfoFromSrc(t, "x = true and 5\n")
	assert.Equal(t, rl.T_BOOL, exprTypeOf(t, file, info).Name())
	assert.False(t, hasOpIssue(info))
}

func TestTypeCheck_OrReturnsUnionOfOperands(t *testing.T) {
	// `or` returns the actual value of whichever operand wins, so the
	// result type is `int | str` here. Once narrowing lands we can
	// refine this to `(left - falsy) | right`.
	file, info, _ := typeInfoFromSrc(t, "x = 1 or \"fallback\"\n")
	got := exprTypeOf(t, file, info)
	assert.Equal(t, "int|str", got.Name())
	assert.False(t, hasOpIssue(info))
}

func TestTypeCheck_InRequiresContainerOnRight(t *testing.T) {
	// `in str` / `in list` / `in map` are the legal shapes; `in int`
	// is not. The static check shouldn't fire on the str-on-right
	// case but should on a numeric right.
	_, info, _ := typeInfoFromSrc(t, "x = \"a\" in \"abc\"\n")
	assert.False(t, hasOpIssue(info))

	_, info2, _ := typeInfoFromSrc(t, "x = 1 in 5\n")
	assert.True(t, hasOpIssue(info2))
}

func TestTypeCheck_InReturnsBool(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = \"a\" in \"abc\"\n")
	assert.Equal(t, rl.T_BOOL, exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_UnaryNotReturnsBool(t *testing.T) {
	// `not` accepts any truthy-able value (the runtime calls
	// TruthyFalsy) and always returns bool.
	file, info, _ := typeInfoFromSrc(t, "x = not 5\n")
	assert.Equal(t, rl.T_BOOL, exprTypeOf(t, file, info).Name())
	assert.False(t, hasOpIssue(info))
}

func TestTypeCheck_UnaryNegOnIntReturnsInt(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = -5\n")
	assert.Equal(t, rl.T_INT, exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_UnaryNegOnStrFlagsHint(t *testing.T) {
	// `-"hi"` is a runtime error; the static side flags it as a hint.
	_, info, _ := typeInfoFromSrc(t, "x = -\"hi\"\n")
	assert.True(t, hasOpIssue(info))
}

func TestTypeCheck_TernaryReturnsUnion(t *testing.T) {
	// `cond ? a : b` synthesizes the union of the branch types.
	file, info, _ := typeInfoFromSrc(t, "x = true ? 1 : \"hi\"\n")
	assert.Equal(t, "int|str", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_TernaryCollapsesIdenticalBranches(t *testing.T) {
	// When both branches have the same type, unionOf returns the
	// type itself rather than `int|int`. Keeps hover messages tidy.
	file, info, _ := typeInfoFromSrc(t, "x = true ? 1 : 2\n")
	assert.Equal(t, rl.T_INT, exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_FallbackReturnsUnion(t *testing.T) {
	// `?? ` is `(left - null) | right` once narrowing exists; for now
	// the safe over-approximation is `left | right`.
	file, info, _ := typeInfoFromSrc(t, "x = 1 ?? \"fallback\"\n")
	assert.Equal(t, "int|str", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_DynamicOperandDoesNotFireDiagnostic(t *testing.T) {
	// An identifier whose type is unknown (forward reference) synth
	// to Dynamic. `dynamic + int` should NOT fire the static check -
	// emitting on dynamic operands would nag users who deliberately
	// wrote `any` or have a value the checker just can't pin down.
	_, info, _ := typeInfoFromSrc(t, "y = unknown + 1\n")
	assert.False(t, hasOpIssue(info),
		"dynamic operand should suppress the type-mismatch hint")
}

// --- Collection literal tests ----------------------------------------

func TestTypeCheck_ListLiteralAllInt(t *testing.T) {
	// Homogeneous int list synthesizes to int[].
	file, info, _ := typeInfoFromSrc(t, "x = [1, 2, 3]\n")
	assert.Equal(t, "int[]", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_ListLiteralIntAndFloatWidensToFloat(t *testing.T) {
	// The plan's one and only implicit numeric widening: a mix of int
	// and float collapses to List<float> rather than List<int|float>.
	// Matches IsAssignableFrom (int flows into float at scalar slots).
	file, info, _ := typeInfoFromSrc(t, "x = [1, 2.5, 3]\n")
	assert.Equal(t, "float[]", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_ListLiteralMixedNonNumericProducesUnion(t *testing.T) {
	// Non-numeric mixes don't widen - the element type is a union.
	file, info, _ := typeInfoFromSrc(t, "x = [1, \"hi\"]\n")
	assert.Equal(t, "int|str[]", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_ListLiteralEmptyIsAnyList(t *testing.T) {
	// Empty literals fall back to the unparameterized form. No
	// "annotation required" nagging - a future look-around pass can
	// refine `xs = []` from later assignments / mutations.
	file, info, _ := typeInfoFromSrc(t, "x = []\n")
	assert.Equal(t, "list", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_MapLiteralStrIntEntries(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = {\"a\": 1, \"b\": 2}\n")
	assert.Equal(t, "{ str: int }", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_MapLiteralMixedValueTypesProducesUnion(t *testing.T) {
	// Non-widening mix on the value side: keys stay str, values
	// become int|str.
	file, info, _ := typeInfoFromSrc(t, "x = {\"a\": 1, \"b\": \"two\"}\n")
	got := exprTypeOf(t, file, info).Name()
	assert.Equal(t, "{ str: int|str }", got)
}

func TestTypeCheck_MapLiteralEmptyIsAnyMap(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = {}\n")
	assert.Equal(t, "map", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_NestedListPreservesInnerType(t *testing.T) {
	// `[[1, 2], [3]]` -> outer is List<List<int>>. Confirms that
	// element type propagates through recursive synth.
	file, info, _ := typeInfoFromSrc(t, "x = [[1, 2], [3]]\n")
	assert.Equal(t, "int[][]", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_ListLiteralWithErrorElementPoisons(t *testing.T) {
	// If any element is ErrorType (typically because its sub-expr
	// already failed), the whole literal becomes ErrorType so we
	// don't cascade diagnostics across the bad element's siblings.
	// Construction: `-"hi"` produces ErrorType, putting it in a list
	// poisons the list's element type.
	file, info, _ := typeInfoFromSrc(t, "x = [1, -\"hi\", 3]\n")
	assert.Equal(t, "<error>", exprTypeOf(t, file, info).Name())
}

// --- User-defined function call tests --------------------------------

func TestTypeCheck_TypedUserFnCallSynthesizesDeclaredReturnType(t *testing.T) {
	// `fn add(x: int, y: int) -> int` declared with full annotations.
	// Calling it should synth to int, the declared return type. This is
	// the headline win of Phase 2e: user fns now feed type info into
	// downstream synthesis the same way builtins always have.
	src := "fn add(x: int, y: int) -> int:\n    return x + y\nz = add(1, 2)\n"
	file, info, _ := typeInfoFromSrc(t, src)
	// `z`'s symbol-type should reflect the call's synthesized return.
	assign := file.Stmts[1].(*rl.Assign)
	call := assign.Values[0].(*rl.Call)
	got, ok := info.ExprTypes[call]
	require.True(t, ok)
	assert.Equal(t, rl.T_INT, got.Name())
}

func TestTypeCheck_UserFnCallWithTooFewArgsFlagsArity(t *testing.T) {
	// User-defined fns get the same arity check as builtins, since
	// callSignatureFor now returns the FnDef's Typing.
	src := "fn add(x: int, y: int) -> int: x + y\nz = add(1)\n"
	_, info, _ := typeInfoFromSrc(t, src)
	assert.True(t, hasIssue(info, rl.ErrWrongArgCount),
		"expected ErrWrongArgCount when user fn called with too few args")
}

func TestTypeCheck_UserFnCallWithWrongArgTypeFlagsHint(t *testing.T) {
	// Type mismatch on a user fn arg surfaces as a Hint, matching the
	// behavior on builtin calls. Hint severity keeps runtime's richer
	// value-aware error path in play.
	src := "fn add(x: int, y: int) -> int: x + y\nz = add(\"a\", 2)\n"
	_, info, _ := typeInfoFromSrc(t, src)
	found := false
	for _, i := range info.Issues {
		if i.Code == rl.ErrTypeMismatch && i.Severity == check.IssueHint {
			found = true
			break
		}
	}
	assert.True(t, found, "expected Hint-severity type-mismatch on user fn arg")
}

func TestTypeCheck_UnannotatedUserFnFallsBackToDynamic(t *testing.T) {
	// `fn add(x, y):` (no types, no return annotation). The signature
	// has nil ReturnT, so the call's synth result is Dynamic - the
	// same fallback used for builtins lacking signatures. Return-type
	// inference from `return` statements is a follow-on.
	src := "fn add(x, y):\n    return x + y\nz = add(1, 2)\n"
	file, info, _ := typeInfoFromSrc(t, src)
	assign := file.Stmts[1].(*rl.Assign)
	call := assign.Values[0].(*rl.Call)
	got, ok := info.ExprTypes[call]
	require.True(t, ok)
	assert.Equal(t, rl.T_DYNAMIC, got.Name())
}

func TestTypeCheck_UserFnCalledBeforeDefinitionWorks(t *testing.T) {
	// Hoisting: a fn can be called before it's defined in the source
	// (the binder pre-pass declares all top-level fns before stmt
	// walking). The call should still resolve to the declared return.
	src := "z = add(1, 2)\nfn add(x: int, y: int) -> int: x + y\n"
	file, info, _ := typeInfoFromSrc(t, src)
	call := file.Stmts[0].(*rl.Assign).Values[0].(*rl.Call)
	got, ok := info.ExprTypes[call]
	require.True(t, ok)
	assert.Equal(t, rl.T_INT, got.Name())
}

func TestTypeCheck_MutualRecursionDoesNotInfiniteLoop(t *testing.T) {
	// `fn f(): g()` and `fn g(): f()` mutually reference each other.
	// The Phase 2e wiring is signature-only (not body-driven return
	// inference), so this can't loop - both fns expose their declared
	// (or nil) return types statically. Sanity check that TypeCheck
	// completes and produces no spurious errors.
	src := "fn f():\n    g()\nfn g():\n    f()\nf()\n"
	_, info, _ := typeInfoFromSrc(t, src)
	// No arity / type-mismatch issues on these calls.
	assert.False(t, hasIssue(info, rl.ErrWrongArgCount))
	assert.False(t, hasIssue(info, rl.ErrTypeMismatch))
}

// --- ErrorType cascade tests -----------------------------------------
//
// These verify that a single poisoned sub-expression produces exactly
// one diagnostic, not a flood of derivative ones from each layer of
// the enclosing AST. Each test checks "the outer expression's synth
// type collapses to <error>" (cascade-suppression handshake) AND
// "the count of issues stays tight."

// countCode returns how many issues with the given code are present.
// Useful for asserting that a poisoned expression produces ONE
// diagnostic, not several as the poison flows up the AST.
func countCode(info *check.TypeInfo, code rl.Error) int {
	n := 0
	for _, i := range info.Issues {
		if i.Code == code {
			n++
		}
	}
	return n
}

func TestTypeCheck_NestedBinaryOpPoisonsToErrorType(t *testing.T) {
	// (1 + "hi") + 5 - the inner produces ErrorType + a diagnostic.
	// The outer must NOT emit a second "ErrorType + int is invalid"
	// hint; anyIsErrorType short-circuits the lookup.
	file, info, _ := typeInfoFromSrc(t, "x = (1 + \"hi\") + 5\n")
	got := exprTypeOf(t, file, info).Name()
	assert.Equal(t, "<error>", got)
	// Exactly one invalid-op diagnostic - from the inner expr only.
	assert.Equal(t, 1, countCode(info, rl.ErrInvalidTypeForOp))
}

func TestTypeCheck_TernaryBranchErrorCollapsesToErrorType(t *testing.T) {
	// cond ? bad : 5 - bad is ErrorType. Before this commit unionOf
	// would have returned `<error>|int`, and the next operator usage
	// would have fired a fresh diagnostic. Now the ternary collapses
	// to ErrorType so downstream operators short-circuit cleanly.
	file, info, _ := typeInfoFromSrc(t, "x = true ? (1 + \"hi\") : 5\n")
	assert.Equal(t, "<error>", exprTypeOf(t, file, info).Name())
	assert.Equal(t, 1, countCode(info, rl.ErrInvalidTypeForOp))
}

func TestTypeCheck_TernaryConditionErrorCollapsesToErrorType(t *testing.T) {
	// A bad condition propagates - we don't know which branch fires,
	// so we can't trust either branch's type to characterize the
	// result. Better to mark the whole ternary as poisoned.
	file, info, _ := typeInfoFromSrc(t, "x = (1 + \"hi\") ? 1 : 2\n")
	assert.Equal(t, "<error>", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_FallbackWithErrorOperandCollapses(t *testing.T) {
	// x ?? y where x is poisoned. Whole expression is uncertain.
	file, info, _ := typeInfoFromSrc(t, "x = (1 + \"hi\") ?? 5\n")
	assert.Equal(t, "<error>", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_CatchWithErrorOperandCollapses(t *testing.T) {
	file, info, _ := typeInfoFromSrc(t, "x = (1 + \"hi\") catch 5\n")
	assert.Equal(t, "<error>", exprTypeOf(t, file, info).Name())
}

func TestTypeCheck_PoisonedArgDoesNotEmitArgTypeMismatch(t *testing.T) {
	// `len(1 + "hi")` - the arg is poisoned. checkArgType sees
	// ErrorType, which IsAssignableFrom anything (poisoned arm of
	// gradual consistency), so no spurious arg-mismatch hint fires.
	// The inner op error is the only diagnostic the user sees.
	_, info, _ := typeInfoFromSrc(t, "x = len(1 + \"hi\")\n")
	assert.Equal(t, 1, countCode(info, rl.ErrInvalidTypeForOp))
	assert.Equal(t, 0, countCode(info, rl.ErrTypeMismatch),
		"poisoned arg should not produce an additional type-mismatch hint")
}

// --- Suggestion / help-line tests ------------------------------------

func TestTypeCheck_StrPlusIntAttachesV09MigrationHint(t *testing.T) {
	// `+` no longer coerces types in v0.9. The runtime attaches a
	// help line pointing at the migration doc when it fires this
	// error; the static check needs to surface the same hint so LSP
	// and `rad check` users get the same actionable follow-up.
	_, info, _ := typeInfoFromSrc(t, "x = \"hi\" + 5\n")
	found := false
	for _, i := range info.Issues {
		if i.Code == rl.ErrInvalidTypeForOp && i.Suggestion != "" {
			found = true
			assert.Contains(t, i.Suggestion, "string interpolation",
				"hint should point users at string interpolation")
			break
		}
	}
	assert.True(t, found, "expected a Suggestion on the str+int hint")
}

func TestTypeCheck_StrPlusStrDoesNotAttachMigrationHint(t *testing.T) {
	// str+str is valid (concat) - no hint, no Suggestion. Sanity
	// check that we only attach the migration tip in the case where
	// pre-v0.9 code would have coerced.
	_, info, _ := typeInfoFromSrc(t, "x = \"a\" + \"b\"\n")
	for _, i := range info.Issues {
		assert.Empty(t, i.Suggestion,
			"valid str+str should not attach any Suggestion: %s", i.Message)
	}
}

func TestTypeCheck_IntPlusListDoesNotAttachStrMigrationHint(t *testing.T) {
	// The migration hint is specific to str+coercible operand
	// combinations - it's the wrong follow-up for, say, int+list,
	// which has its own gotcha.
	_, info, _ := typeInfoFromSrc(t, "x = 1 + [1, 2]\n")
	for _, i := range info.Issues {
		if i.Code == rl.ErrInvalidTypeForOp {
			assert.Empty(t, i.Suggestion,
				"int+list should not attach the str migration hint")
		}
	}
}

// --- Typed local declaration tests (Phase 3) -------------------------

func TestTypeCheck_TypedLocalRecordsDeclaredType(t *testing.T) {
	// `x: int = 5` - the symbol's recorded type should be the
	// declared `int`, not the RHS's synth result (which happens to
	// also be int here, but the point is the binder set Declared and
	// the checker used it).
	file, info, resolved := typeInfoFromSrc(t, "x: int = 5\n")
	target := file.Stmts[0].(*rl.Assign).Targets[0].(*rl.Identifier)
	sym := resolved.Uses[target]
	require.NotNil(t, sym)
	require.NotNil(t, sym.Declared, "binder should populate Declared")
	assert.Equal(t, rl.T_INT, sym.Declared.Name())
	assert.Equal(t, rl.T_INT, info.SymbolTypes[sym].Name())
}

func TestTypeCheck_TypedLocalAcceptsAssignableRHS(t *testing.T) {
	// int literal flows into `: int` slot without diagnostic.
	_, info, _ := typeInfoFromSrc(t, "x: int = 5\n")
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrTypeMismatch, i.Code,
			"valid typed local should not produce a type-mismatch")
	}
}

func TestTypeCheck_TypedLocalRejectsIncompatibleRHS(t *testing.T) {
	// str literal can't flow into `: int`. Hint severity matches the
	// rest of Phase 2's assignability precedent.
	_, info, _ := typeInfoFromSrc(t, "x: int = \"hi\"\n")
	found := false
	for _, i := range info.Issues {
		if i.Code == rl.ErrTypeMismatch && i.Severity == check.IssueHint {
			found = true
			break
		}
	}
	assert.True(t, found,
		"expected a Hint-severity type-mismatch when RHS isn't assignable to declared")
}

func TestTypeCheck_TypedLocalDeclaredFloatAcceptsInt(t *testing.T) {
	// The single implicit widening: int flows into float. So
	// `x: float = 5` is fine, no diagnostic.
	_, info, _ := typeInfoFromSrc(t, "x: float = 5\n")
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrTypeMismatch, i.Code)
	}
}

func TestTypeCheck_TypedLocalReassignChecksAgainstDeclared(t *testing.T) {
	// After `x: int = 5`, a later `x = "hi"` should still be flagged
	// against the original declared type. The annotation is sticky
	// for the binding's lifetime.
	_, info, _ := typeInfoFromSrc(t, "x: int = 5\nx = \"hi\"\n")
	count := 0
	for _, i := range info.Issues {
		if i.Code == rl.ErrTypeMismatch {
			count++
		}
	}
	assert.Equal(t, 1, count,
		"reassignment with incompatible type should fire one type-mismatch")
}

func TestTypeCheck_TypedLocalDeclaredAnyAcceptsAnything(t *testing.T) {
	// `: any` is the user-opt-in escape hatch - every type flows in.
	_, info, _ := typeInfoFromSrc(t, "x: any = 5\nx = \"hi\"\n")
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrTypeMismatch, i.Code,
			"any-typed local should accept any RHS")
	}
}

func TestTypeCheck_UntypedLocalUnchanged(t *testing.T) {
	// Sanity check that the introduction of typed locals doesn't
	// disturb the existing untyped path. `x = 5; x = "hi"` should
	// still rebind freely (no declared annotation means no
	// reassignment constraint).
	_, info, _ := typeInfoFromSrc(t, "x = 5\nx = \"hi\"\n")
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrTypeMismatch, i.Code,
			"untyped local should accept any reassignment")
	}
}

func TestTypeCheck_TypedLocalWithRadBlockKeywordName(t *testing.T) {
	// Variables named `rad` / `request` / `display` (the rad_block
	// keywords) can ALSO be typed-locals. The external scanner's
	// BLOCK_COLON peek distinguishes a typed-assign's ':' from a
	// rad_block's block-colon by what follows it, so `rad: int = 5`
	// is unambiguously a typed-local declaration regardless of the
	// keyword overlap. This test pins that down.
	file, info, resolved := typeInfoFromSrc(t, "rad: int = 5\n")
	require.NotEmpty(t, file.Stmts, "should produce one stmt")
	assign, ok := file.Stmts[0].(*rl.Assign)
	require.True(t, ok, "expected Assign, got %T", file.Stmts[0])
	require.NotNil(t, assign.DeclaredType,
		"typed_assign with `rad` LHS should carry DeclaredType")
	ident := assign.Targets[0].(*rl.Identifier)
	assert.Equal(t, "rad", ident.Name)
	sym := resolved.Uses[ident]
	require.NotNil(t, sym)
	assert.Equal(t, rl.T_INT, sym.Declared.Name())
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrTypeMismatch, i.Code,
			"int RHS into `rad: int` should not produce a mismatch")
	}
}

// --- Phase 4e: if/elif/else flow + join logic ------------------------

func TestTypeCheck_IfNotNull_NarrowsTruthyToNonNull(t *testing.T) {
	// Inside the then-branch, x is int (not int?).
	src := `x: int? = 5
if x != null:
    y = x
`
	file, info, resolved := typeInfoFromSrc(t, src)
	ifStmt := file.Stmts[1].(*rl.If)
	yAssign := ifStmt.Branches[0].Body[0].(*rl.Assign)
	xUse := yAssign.Values[0].(*rl.Identifier)
	gotXUse := info.ExprTypes[xUse]
	require.NotNil(t, gotXUse)
	assert.Equal(t, rl.T_INT, gotXUse.Name(),
		"x inside `if x != null:` body should narrow to int")

	ySym := resolved.Uses[yAssign.Targets[0].(*rl.Identifier)]
	require.NotNil(t, ySym)
	assert.Equal(t, rl.T_INT, info.SymbolTypes[ySym].Name(),
		"y inferred from narrowed x should be int")
}

func TestTypeCheck_IfReturnNarrowsFallthrough(t *testing.T) {
	// Early-exit pattern: `if x == null: return` leaves x narrowed to
	// non-null on the path past the if.
	src := `fn f(x: int?):
    if x == null:
        return
    y = x
`
	file, info, resolved := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	// fn.Body: [If, Assign(y=x)]
	yAssign := fn.Body[1].(*rl.Assign)
	xUse := yAssign.Values[0].(*rl.Identifier)
	gotXUse := info.ExprTypes[xUse]
	require.NotNil(t, gotXUse)
	assert.Equal(t, rl.T_INT, gotXUse.Name(),
		"x after `if x == null: return` should narrow to int")

	ySym := resolved.Uses[yAssign.Targets[0].(*rl.Identifier)]
	require.NotNil(t, ySym)
	assert.Equal(t, rl.T_INT, info.SymbolTypes[ySym].Name())
}

func TestTypeCheck_IfElseUnionsBranches(t *testing.T) {
	// Different branches assign different types to the same untyped
	// local; the join produces a union.
	src := `if true:
    x = 5
else:
    x = "hi"
y = x
`
	file, info, resolved := typeInfoFromSrc(t, src)
	yAssign := file.Stmts[1].(*rl.Assign)
	ySym := resolved.Uses[yAssign.Targets[0].(*rl.Identifier)]
	require.NotNil(t, ySym)
	got := info.SymbolTypes[ySym]
	require.NotNil(t, got)
	// Union dedupes by Name; order isn'\''t guaranteed because the
	// joined symbol map iterates Go maps. Accept either ordering.
	name := got.Name()
	assert.Contains(t, []string{"int|str", "str|int"}, name,
		"y after if/else with int/str assignments should be int|str")
}

func TestTypeCheck_IfNoElseLeavesBaseOnFallthrough(t *testing.T) {
	// `if cond: x = "hi"` (no else) leaves x as int|str after - the
	// fall-through path keeps x at int from before the if.
	src := `x = 5
if true:
    x = "hi"
y = x
`
	file, info, resolved := typeInfoFromSrc(t, src)
	yAssign := file.Stmts[2].(*rl.Assign)
	ySym := resolved.Uses[yAssign.Targets[0].(*rl.Identifier)]
	require.NotNil(t, ySym)
	got := info.SymbolTypes[ySym]
	require.NotNil(t, got)
	name := got.Name()
	assert.Contains(t, []string{"int|str", "str|int"}, name)
}

func TestTypeCheck_NestedIfNarrowingThreads(t *testing.T) {
	// type_of narrows a union; the nested if can rely on the outer
	// narrowing being still in scope.
	src := `fn f(x: int|str):
    if type_of(x) == "int":
        if x > 0:
            y = x
`
	file, info, resolved := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	outerIf := fn.Body[0].(*rl.If)
	innerIf := outerIf.Branches[0].Body[0].(*rl.If)
	yAssign := innerIf.Branches[0].Body[0].(*rl.Assign)
	xUse := yAssign.Values[0].(*rl.Identifier)
	gotXUse := info.ExprTypes[xUse]
	require.NotNil(t, gotXUse)
	assert.Equal(t, rl.T_INT, gotXUse.Name(),
		"x in nested narrowed scope should still be int")
	ySym := resolved.Uses[yAssign.Targets[0].(*rl.Identifier)]
	require.NotNil(t, ySym)
	assert.Equal(t, rl.T_INT, info.SymbolTypes[ySym].Name())
}

func TestTypeCheck_IfElseAccumulatesFalsy(t *testing.T) {
	// The else branch should see falsy refinement from the matching
	// if condition. Without `elif` in Rad, we nest a second if in
	// the else to exercise the same logic.
	src := `fn f(x: int|str|bool):
    if type_of(x) == "int":
        a = x
    else:
        if type_of(x) == "str":
            b = x
        else:
            c = x
`
	file, info, _ := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	outer := fn.Body[0].(*rl.If)

	// Branch 0 body: a = x -> x narrows to int
	aUse := outer.Branches[0].Body[0].(*rl.Assign).Values[0].(*rl.Identifier)
	assert.Equal(t, rl.T_INT, info.ExprTypes[aUse].Name(),
		"x in then-branch should be int")

	inner := outer.Branches[1].Body[0].(*rl.If)
	// Inner then: b = x -> x is str (int excluded by outer falsy).
	bUse := inner.Branches[0].Body[0].(*rl.Assign).Values[0].(*rl.Identifier)
	assert.Equal(t, rl.T_STR, info.ExprTypes[bUse].Name(),
		"x in nested-elif (str) should be str (int excluded)")

	// Inner else: c = x -> x is bool (int and str both excluded).
	cUse := inner.Branches[1].Body[0].(*rl.Assign).Values[0].(*rl.Identifier)
	assert.Equal(t, rl.T_BOOL, info.ExprTypes[cUse].Name(),
		"x in deepest else should be bool (only remaining arm)")
}

// --- Phase 4g: while/for loop rules ----------------------------------

func TestTypeCheck_ForLoopVarFromListElement(t *testing.T) {
	src := `for i in [1, 2, 3]:
    print(i)
`
	file, info, resolved := typeInfoFromSrc(t, src)
	loop := file.Stmts[0].(*rl.ForLoop)
	iSym := resolved.Decls[loop]
	require.NotNil(t, iSym, "binder should record the loop var symbol")
	got, ok := info.SymbolTypes[iSym]
	require.True(t, ok, "loop var should get a symbol type")
	assert.Equal(t, rl.T_INT, got.Name(), "loop var over int[] should be int")
}

func TestTypeCheck_ForLoopVarFromStringIsStr(t *testing.T) {
	src := `for c in "hello":
    print(c)
`
	file, info, resolved := typeInfoFromSrc(t, src)
	loop := file.Stmts[0].(*rl.ForLoop)
	cSym := resolved.Decls[loop]
	require.NotNil(t, cSym)
	assert.Equal(t, rl.T_STR, info.SymbolTypes[cSym].Name())
}

func TestTypeCheck_ForLoopVarDynamicForUnknownIter(t *testing.T) {
	// A function call we can'\''t resolve yields Dynamic.
	src := `for v in unknown_fn():
    print(v)
`
	file, info, resolved := typeInfoFromSrc(t, src)
	loop := file.Stmts[0].(*rl.ForLoop)
	vSym := resolved.Decls[loop]
	require.NotNil(t, vSym)
	assert.Equal(t, rl.T_DYNAMIC, info.SymbolTypes[vSym].Name())
}

func TestTypeCheck_WhileNarrowsBody(t *testing.T) {
	src := `fn f(x: int?):
    while x != null:
        y = x
`
	file, info, _ := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	whileL := fn.Body[0].(*rl.WhileLoop)
	yAssign := whileL.Body[0].(*rl.Assign)
	xUse := yAssign.Values[0].(*rl.Identifier)
	assert.Equal(t, rl.T_INT, info.ExprTypes[xUse].Name(),
		"x inside while body should narrow to non-null")
}

// --- Phase 4f: switch + exhaustiveness -------------------------------

func TestTypeCheck_SwitchNarrowsDiscriminantPerCase(t *testing.T) {
	src := `fn f(name: ["alice", "bob", "charlie"]):
    switch name:
        case "alice":
            a = name
        case "bob":
            b = name
        case "charlie":
            c = name
`
	file, info, _ := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	sw := fn.Body[0].(*rl.Switch)

	aUse := sw.Cases[0].Alt.(*rl.SwitchCaseBlock).Stmts[0].(*rl.Assign).Values[0].(*rl.Identifier)
	bUse := sw.Cases[1].Alt.(*rl.SwitchCaseBlock).Stmts[0].(*rl.Assign).Values[0].(*rl.Identifier)
	cUse := sw.Cases[2].Alt.(*rl.SwitchCaseBlock).Stmts[0].(*rl.Assign).Values[0].(*rl.Identifier)

	assert.Equal(t, `["alice"]`, info.ExprTypes[aUse].Name())
	assert.Equal(t, `["bob"]`, info.ExprTypes[bUse].Name())
	assert.Equal(t, `["charlie"]`, info.ExprTypes[cUse].Name())
}

func TestTypeCheck_SwitchExhaustiveOverEnumNoDiagnostic(t *testing.T) {
	src := `fn f(name: ["a", "b"]):
    switch name:
        case "a":
            x = 1
        case "b":
            x = 2
`
	_, info, _ := typeInfoFromSrc(t, src)
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrNonExhaustiveSwitch, i.Code,
			"exhaustive switch should produce no non-exhaustive diagnostic")
	}
}

func TestTypeCheck_SwitchNonExhaustiveOverEnumFires(t *testing.T) {
	src := `fn f(name: ["a", "b", "c"]):
    switch name:
        case "a":
            x = 1
        case "b":
            x = 2
`
	_, info, _ := typeInfoFromSrc(t, src)
	found := false
	for _, i := range info.Issues {
		if i.Code == rl.ErrNonExhaustiveSwitch {
			found = true
			assert.Contains(t, i.Message, "c",
				"diagnostic should name the missing case value")
		}
	}
	assert.True(t, found, "non-exhaustive switch over closed enum should fire")
}

func TestTypeCheck_SwitchWithDefaultSuppressesDiagnostic(t *testing.T) {
	src := `fn f(name: ["a", "b", "c"]):
    switch name:
        case "a":
            x = 1
        default:
            x = 99
`
	_, info, _ := typeInfoFromSrc(t, src)
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrNonExhaustiveSwitch, i.Code,
			"default branch should suppress non-exhaustive diagnostic")
	}
}

func TestTypeCheck_SwitchNonEnumDiscriminantSkipsExhaustiveness(t *testing.T) {
	// Plain str (open type) - no exhaustiveness check applies.
	src := `fn f(name: str):
    switch name:
        case "a":
            x = 1
`
	_, info, _ := typeInfoFromSrc(t, src)
	for _, i := range info.Issues {
		assert.NotEqual(t, rl.ErrNonExhaustiveSwitch, i.Code,
			"open-typed discriminant should never trigger exhaustiveness")
	}
}

func TestTypeCheck_IfDoesNotLeakNarrowingAfter(t *testing.T) {
	// After a non-exiting if, the narrowing should not persist into
	// subsequent statements at the same scope. Use a fn param so the
	// base type stays int? - a top-level `x: int? = 5` would narrow
	// x to int on assignment, and the if'\''s null-predicate would have
	// nothing to strip.
	src := `fn f(x: int?):
    if x != null:
        y = x
    z = x
`
	file, info, _ := typeInfoFromSrc(t, src)
	fn := file.Stmts[0].(*rl.FnDef)
	zAssign := fn.Body[1].(*rl.Assign)
	xUse := zAssign.Values[0].(*rl.Identifier)
	gotXUse := info.ExprTypes[xUse]
	require.NotNil(t, gotXUse)
	// Branches contribute int (truthy exit) and int? (acc = base).
	// The dedupe-by-Name union keeps them as int|int?; semantically
	// equivalent to int? (since int <: int?), but a tighter
	// representation would need union-arm subsumption that we don'\''t
	// have today.
	name := gotXUse.Name()
	assert.Contains(t, []string{"int?", "int|int?", "int?|int"}, name,
		"x after if (no else) should not stay narrowed to int")
}
