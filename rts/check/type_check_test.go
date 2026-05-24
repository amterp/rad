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
