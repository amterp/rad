package check

import (
	"fmt"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// TypeInfo is the output of bidirectional type checking. It carries
// the type each symbol settled on, the type each expression node
// synthesized to, and any type-related diagnostics the checker
// recorded along the way.
//
// Like Resolved, it's a pure value over the (ast, resolved) inputs -
// no mutation of Symbol records, no dependence on source text. The
// LSP snapshot model can hold one alongside the matching Resolved
// and serve hover/goto-type queries lock-free.
//
// Phase 2a establishes the data model and handles literal + identifier
// synthesis only. Calls, operators, struct/list/map literals, and
// flow-sensitive narrowing come in subsequent commits.
type TypeInfo struct {
	// SymbolTypes maps a symbol to the type the checker has decided
	// it currently holds. For an annotated binding this is the
	// declared type; for an unannotated binding it's the type
	// synthesized from the most-recent assignment. Absent entries
	// mean "no information yet" (callers should treat as Dynamic).
	SymbolTypes map[*Symbol]rl.TypingT
	// ExprTypes maps an expression AST node to its synthesized type.
	// Useful for hover, and for narrowing passes that need to ask
	// "what was the static type of this sub-expression?"
	ExprTypes map[rl.Node]rl.TypingT
	// Issues are type-related findings: type mismatches, arg count
	// errors, and so on. The checker layer converts these to
	// Diagnostic. Empty in Phase 2a; populated as later sub-commits
	// add the real checks.
	Issues []BindIssue
}

// TypeCheck runs bidirectional type checking over the file using the
// pre-computed Resolved view. Returns nil if either input is nil.
//
// The pass is single-pass and AST-order: when it sees `x = expr`, it
// synthesizes the type of `expr`, then records that as the type of
// the symbol for `x`. Later references to `x` synth to that recorded
// type. Forward references (e.g. inside a lambda that's stored before
// it's called) get Dynamic - that's the right answer for a gradual
// system, and Phase 2e will refine it with Tarjan SCC ordering for
// genuine mutual recursion.
func TypeCheck(file *rl.SourceFile, resolved *Resolved) *TypeInfo {
	if file == nil || resolved == nil {
		return nil
	}
	tc := &typeChecker{
		resolved: resolved,
		info: &TypeInfo{
			SymbolTypes: map[*Symbol]rl.TypingT{},
			ExprTypes:   map[rl.Node]rl.TypingT{},
		},
	}
	tc.walkFile(file)
	return tc.info
}

// typeChecker carries state during a single TypeCheck invocation.
// Like the binder, it isn't safe for concurrent use; the public
// TypeCheck function constructs a fresh one per call.
type typeChecker struct {
	resolved *Resolved
	info     *TypeInfo
}

func (tc *typeChecker) walkFile(file *rl.SourceFile) {
	for _, stmt := range file.Stmts {
		tc.walkStmt(stmt)
	}
	for _, cmd := range file.Cmds {
		tc.walkCmd(cmd)
	}
}

// walkStmt dispatches on statement kind. Phase 2a recognizes only the
// shapes it can do something with; everything else is descended-into
// so identifier-uses still get their types recorded (for hover).
func (tc *typeChecker) walkStmt(n rl.Node) {
	if n == nil {
		return
	}
	switch v := n.(type) {
	case *rl.Assign:
		tc.walkAssign(v)
	case *rl.ExprStmt:
		_ = tc.synth(v.Expr)
	default:
		// Generic descent. Later sub-commits replace these with
		// kind-specific handlers (for loops, switch, return, etc.).
		for _, child := range n.Children() {
			tc.walkStmt(child)
		}
	}
}

func (tc *typeChecker) walkCmd(c *rl.CmdBlock) {
	// Cmd block defaults and inline-lambda body are still descended
	// for hover purposes; nothing actionable for Phase 2a.
	for i := range c.Decls {
		if c.Decls[i].Default != nil {
			_ = tc.synth(c.Decls[i].Default)
		}
	}
	if c.Callback.IsLambda && c.Callback.Lambda != nil {
		for _, stmt := range c.Callback.Lambda.Body {
			tc.walkStmt(stmt)
		}
	}
}

// walkAssign handles `x = expr` and `a, b = e1, e2`. For each
// target/value pair we synth the RHS and record its type as the
// type of the LHS symbol. Multi-value RHS aligns 1:1 with multi-
// target LHS at this stage; unpacking (where one RHS expression
// produces multiple values) is deferred.
func (tc *typeChecker) walkAssign(a *rl.Assign) {
	for i, val := range a.Values {
		valType := tc.synth(val)
		if i >= len(a.Targets) {
			continue
		}
		ident, ok := a.Targets[i].(*rl.Identifier)
		if !ok {
			continue
		}
		sym, ok := tc.resolved.Uses[ident]
		if !ok {
			continue
		}
		tc.info.SymbolTypes[sym] = valType
	}
}

// synth returns the static type of an expression node and records it
// on the ExprTypes index.
//
// Unhandled expression shapes return Dynamic, but their children are
// still walked recursively so that any identifier-uses or literals
// they contain still get their types recorded. That keeps hover and
// goto-type information useful for the sub-expressions of a construct
// even before its outer-shape handler is implemented; later
// sub-commits replace these generic descents with kind-specific
// synthesis (call return types, operator results, etc.).
func (tc *typeChecker) synth(n rl.Node) rl.TypingT {
	if n == nil {
		return rl.NewDynamicType()
	}
	switch v := n.(type) {
	case *rl.LitInt:
		return tc.record(v, rl.NewIntType())
	case *rl.LitFloat:
		return tc.record(v, rl.NewFloatType())
	case *rl.LitString:
		return tc.record(v, rl.NewStrType())
	case *rl.LitBool:
		return tc.record(v, rl.NewBoolType())
	case *rl.LitNull:
		// Rad models nullability via Optional<T>; a bare null literal
		// without context is best-typed as Dynamic until later
		// sub-commits give us a way to bubble the expected type into
		// synth (the "check" direction of bidirectional checking).
		return tc.record(v, rl.NewDynamicType())
	case *rl.Identifier:
		return tc.synthIdentifier(v)
	case *rl.Call:
		return tc.synthCall(v, 0)
	case *rl.VarPath:
		return tc.synthVarPath(v)
	case *rl.OpBinary:
		return tc.synthOpBinary(v)
	case *rl.OpUnary:
		return tc.synthOpUnary(v)
	case *rl.Ternary:
		return tc.synthTernary(v)
	case *rl.Fallback:
		return tc.synthFallback(v)
	case *rl.CatchExpr:
		return tc.synthCatchExpr(v)
	case *rl.LitList:
		return tc.synthLitList(v)
	case *rl.LitMap:
		return tc.synthLitMap(v)
	}
	for _, child := range n.Children() {
		_ = tc.synth(child)
	}
	return tc.record(n, rl.NewDynamicType())
}

// synthCall handles a function-call expression. For now this only
// works for calls whose callee is a direct Identifier resolving to a
// builtin - the binder marks those as SymBuiltin, and the runtime
// keeps a parsed TypingFnT for every builtin in FnSignaturesByName.
// User-defined function call-checking lands in Phase 2e once the
// Tarjan SCC pass populates signatures for hoisted fns.
//
// Phase 2b emits arity diagnostics only: too few, too many, missing
// required, unknown named arg, duplicated named arg. Per-arg type
// checking lands in the next sub-commit so this one stays focused on
// the arity-matching algorithm.
//
// implicitReceiverCount is non-zero for UFCS-style calls
// (`expr.method(args)`), where the receiver of the chain is the
// implicit first positional argument. The arity check then expects
// `len(args) + implicitReceiverCount` to match the formal signature.
func (tc *typeChecker) synthCall(call *rl.Call, implicitReceiverCount int) rl.TypingT {
	// Always synth the args themselves so identifier-uses get recorded.
	for _, a := range call.Args {
		_ = tc.synth(a)
	}
	for _, na := range call.NamedArgs {
		_ = tc.synth(na.Value)
	}
	// Also synth the callee. For UFCS calls the callee is just the
	// method-name Identifier (no enclosing Call.Func chain), but the
	// shape is the same either way.
	_ = tc.synth(call.Func)

	typing := tc.callSignatureFor(call.Func)
	if typing == nil {
		return tc.record(call, rl.NewDynamicType())
	}
	tc.checkCall(call, typing, implicitReceiverCount)

	if typing.ReturnT != nil {
		return tc.record(call, *typing.ReturnT)
	}
	return tc.record(call, rl.NewDynamicType())
}

// synthVarPath walks a chained path expression (e.g. `a.b[c].d`,
// `xs.sort()`). Non-UFCS segments are visited for hover/symbol
// purposes only - their actual semantics (field access, indexing,
// slicing) get static types in a later sub-commit.
//
// UFCS segments are calls whose first positional argument is the
// path's chain so far. We pull the Call out of the segment and
// type-check it with an implicit-receiver count of 1, so arity
// checks count the chain-receiver as the first arg.
func (tc *typeChecker) synthVarPath(v *rl.VarPath) rl.TypingT {
	_ = tc.synth(v.Root)
	for _, seg := range v.Segments {
		switch {
		case seg.IsUFCS:
			if call, ok := seg.Index.(*rl.Call); ok {
				_ = tc.synthCall(call, 1)
			}
		case seg.Index != nil:
			_ = tc.synth(seg.Index)
		case seg.IsSlice:
			if seg.Start != nil {
				_ = tc.synth(seg.Start)
			}
			if seg.End != nil {
				_ = tc.synth(seg.End)
			}
		}
	}
	return tc.record(v, rl.NewDynamicType())
}

// callSignatureFor returns the static signature of a call's callee,
// or nil when we can't determine one. Three resolution paths:
//
//  1. Builtin: ambient symbol; signature lives in FnSignaturesByName.
//     Fetched lazily here (the binder doesn't pre-populate every one).
//  2. Hoisted top-level user function: the FnDef's declared Typing.
//     This is what makes `add(1, 2)` type-check against the
//     `fn add(a: int, b: int) -> int:` declaration. Unannotated params
//     show up with a nil Type on TypingFnParam, which checkCall
//     already treats as "no constraint" (matching `any`).
//  3. Anything else (local-with-lambda, function-typed parameter,
//     getter expression) - returns nil today. Will be filled in once
//     SymbolTypes carries TypingFnT for these symbols.
//
// Returning nil makes the caller fall back to a Dynamic result with
// no arity / type-mismatch diagnostics, which is the right behavior
// for "we can't reason about this callee."
func (tc *typeChecker) callSignatureFor(callee rl.Node) *rl.TypingFnT {
	ident, ok := callee.(*rl.Identifier)
	if !ok {
		return nil
	}
	sym, ok := tc.resolved.Uses[ident]
	if !ok {
		return nil
	}
	switch sym.Kind {
	case SymBuiltin:
		sig, ok := rts.FnSignaturesByName[sym.Name]
		if !ok || sig.Typing == nil {
			return nil
		}
		return sig.Typing
	case SymHoistedFn:
		fn, ok := sym.DefNode.(*rl.FnDef)
		if !ok || fn.Typing == nil {
			return nil
		}
		return fn.Typing
	}
	return nil
}

// checkCall runs the call-matching algorithm: it pairs each explicit
// argument with a formal parameter, type-checks the match, and emits
// diagnostics for arity and assignability failures. Mirrors the
// runtime logic in core/type_fn.go but operates on AST positions and
// synthesized types rather than evaluated values.
//
//   - Positional args fill positional params left-to-right until a
//     NamedOnly param is reached. If the last positional param is
//     variadic, it absorbs every remaining positional arg.
//   - Named args must match a non-anonymous param by name and may not
//     duplicate something already filled positionally.
//   - After matching, every required param (no default, not variadic,
//     not optional) must have been seen.
//   - Each matched arg's synthesized type is checked against the
//     param's declared type via IsAssignableFrom; failures fire
//     ErrTypeMismatch.
//
// Two type-checks are intentionally deferred to a follow-on commit:
// the UFCS receiver's type against params[0], and variadic element
// types against the variadic's declared element type. Both need a
// bit more plumbing (receiver-type threading from synthVarPath,
// element-shape extraction from TypingListT) that doesn't belong on
// the arity path.
func (tc *typeChecker) checkCall(call *rl.Call, typing *rl.TypingFnT, implicitReceiverCount int) {
	params := typing.Params
	seen := make(map[string]bool, len(params))
	hasVariadic := len(params) > 0 && params[len(params)-1].IsVariadic

	// Account for any implicit first arg (UFCS receiver). The
	// receiver always fills params[0..implicitReceiverCount-1]. Type
	// of the receiver is not checked here yet; that lands in a
	// follow-on commit that threads chain types through synthVarPath.
	for i := 0; i < implicitReceiverCount && i < len(params); i++ {
		param := params[i]
		if param.IsVariadic {
			seen[param.Name] = true
			implicitReceiverCount = 0
			break
		}
		seen[param.Name] = true
	}

	idxBase := implicitReceiverCount
	totalArgs := implicitReceiverCount + len(call.Args)

	for argIdx, arg := range call.Args {
		paramIdx := idxBase + argIdx
		if paramIdx >= len(params) {
			if hasVariadic {
				break
			}
			tc.addCallIssue(call.Span(), rl.ErrWrongArgCount,
				fmt.Sprintf("Expected at most %d args, but was invoked with %d", len(params), totalArgs))
			break
		}
		param := params[paramIdx]
		if param.IsVariadic {
			// Variadic absorbs this and every later positional arg.
			// Per-element type-check deferred to a follow-on commit;
			// for now just record the param as filled.
			seen[param.Name] = true
			break
		}
		if param.NamedOnly {
			tc.addCallIssue(arg.Span(), rl.ErrWrongArgCount,
				"Too many positional args, remaining args are named-only")
			break
		}
		tc.checkArgType(arg, param)
		seen[param.Name] = true
	}

	byName := typing.ByName()
	for _, na := range call.NamedArgs {
		param, ok := byName[na.Name]
		if !ok {
			tc.addCallIssue(na.NameSpan, rl.ErrInvalidArgType,
				fmt.Sprintf("Unknown named argument '%s'", na.Name))
			continue
		}
		if param.AnonymousOnly() {
			tc.addCallIssue(na.NameSpan, rl.ErrInvalidArgType,
				fmt.Sprintf("Argument '%s' cannot be passed as named arg, only positionally", na.Name))
			continue
		}
		if seen[na.Name] {
			tc.addCallIssue(na.NameSpan, rl.ErrInvalidArgType,
				fmt.Sprintf("Argument '%s' already specified", na.Name))
			continue
		}
		tc.checkArgType(na.Value, param)
		seen[na.Name] = true
	}

	for _, param := range params {
		if seen[param.Name] {
			continue
		}
		if param.IsVariadic || param.IsOptional || param.DefaultAST != nil {
			continue
		}
		// A parameter typed as `T?` or `T|null` accepts null implicitly;
		// the runtime fills it with null when omitted. Don't flag those
		// as missing - matches the runtime's treatment.
		if param.Type != nil && (*param.Type).IsCompatibleWith(rl.NewNullSubject()) {
			continue
		}
		tc.addCallIssue(call.Span(), rl.ErrWrongArgCount,
			fmt.Sprintf("Missing required argument '%s'", param.Name))
	}
}

// checkArgType verifies an argument expression's type is assignable to
// the matched parameter's declared type. No-ops when the param is
// unannotated (nil declared type) - that's the "any" parameter case
// and we let the runtime catch any mismatch.
//
// Severity is intentionally Hint, not Error, while the static side
// lacks two pieces of fidelity it eventually needs: literal types
// (so a string-literal "seconds" can be checked against a string
// enum), and structural matching for function values. Until both
// exist, promoting to Error would short-circuit the runtime check
// path - which today produces richer "Value 'X' (Y) is not
// compatible..." errors that mention the actual value. Keeping
// this at Hint surfaces the issue in LSP and `rad check` while
// preserving runtime behavior for the value-aware cases.
func (tc *typeChecker) checkArgType(argNode rl.Node, param rl.TypingFnParam) {
	if param.Type == nil {
		return
	}
	expected := *param.Type
	argType := tc.synth(argNode)
	if argType == nil {
		return
	}
	if expected.IsAssignableFrom(argType) {
		return
	}
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     argNode.Span(),
		Severity: IssueHint,
		Code:     rl.ErrTypeMismatch,
		Message: fmt.Sprintf("Argument type '%s' is not assignable to expected type '%s'",
			argType.Name(), expected.Name()),
	})
}

func (tc *typeChecker) addCallIssue(span rl.Span, code rl.Error, msg string) {
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     span,
		Severity: IssueError,
		Code:     code,
		Message:  msg,
	})
}

// synthIdentifier looks up the symbol an identifier refers to and
// returns whatever type the checker has decided it holds. If the
// symbol has no recorded type yet (forward reference, builtin
// without a synthesized signature, etc.), return Dynamic.
func (tc *typeChecker) synthIdentifier(ident *rl.Identifier) rl.TypingT {
	sym, ok := tc.resolved.Uses[ident]
	if !ok {
		return tc.record(ident, rl.NewDynamicType())
	}
	if t, known := tc.info.SymbolTypes[sym]; known {
		return tc.record(ident, t)
	}
	return tc.record(ident, rl.NewDynamicType())
}

// record stores the synthesized type for a node on ExprTypes and
// returns it, so call sites can write `return tc.record(n, t)` in one
// line.
func (tc *typeChecker) record(n rl.Node, t rl.TypingT) rl.TypingT {
	tc.info.ExprTypes[n] = t
	return t
}

// --- Operators ---------------------------------------------------------
//
// The operator handlers below mirror the runtime dispatch in
// core/expr_ops.go. Each one synthesizes the operand types, decides the
// result type from a fixed overload table, and emits a Hint-severity
// type-mismatch diagnostic when an operand pair isn't in the table.
//
// Severity choice: the runtime panics on a bad operator combination
// (ErrInvalidTypeForOp), so static promotion to Error would arguably be
// safe here. We keep this at Hint for now to match the precedent set
// in Phase 2b' (per-arg type-check) - the static side still lacks
// literal types and we want a single severity-migration pass to flip
// everything together once snapshots are updated.
//
// Operand types of `any`, `dynamic`, `<error>`, or `never` short-circuit
// the lookup: any/dynamic mean "we couldn't pin a type" and emitting
// here would nag users who deliberately wrote `any`; `<error>` is the
// poison marker that suppresses cascades.

func (tc *typeChecker) synthOpBinary(n *rl.OpBinary) rl.TypingT {
	left := tc.synth(n.Left)
	right := tc.synth(n.Right)

	if anyIsErrorType(left, right) {
		return tc.record(n, rl.NewErrorTypeType())
	}
	if anyIsDynamicLike(left, right) {
		// Equality and the boolean ops still have a known result type
		// even when one operand is dynamic; everything else falls back
		// to dynamic so we don't speculate.
		switch n.Op {
		case rl.OpEq, rl.OpNeq, rl.OpAnd, rl.OpLt, rl.OpLte, rl.OpGt, rl.OpGte,
			rl.OpIn, rl.OpNotIn:
			return tc.record(n, rl.NewBoolType())
		}
		return tc.record(n, rl.NewDynamicType())
	}

	result, ok := binaryOpResult(n.Op, left, right)
	if !ok {
		tc.addOpIssue(n.Span(), n.Op, left, right, n.IsCompound)
		return tc.record(n, rl.NewErrorTypeType())
	}
	return tc.record(n, result)
}

func (tc *typeChecker) synthOpUnary(n *rl.OpUnary) rl.TypingT {
	operand := tc.synth(n.Operand)
	if isErrorType(operand) {
		return tc.record(n, rl.NewErrorTypeType())
	}
	switch n.Op {
	case rl.OpNot:
		// `not <anything>` is always bool - the runtime calls TruthyFalsy
		// regardless of operand type, so any operand is acceptable.
		return tc.record(n, rl.NewBoolType())
	case rl.OpNeg, rl.OpAdd:
		// Unary - and unary + require a numeric operand.
		if isDynamicLike(operand) {
			return tc.record(n, rl.NewDynamicType())
		}
		if isInt(operand) {
			return tc.record(n, rl.NewIntType())
		}
		if isFloat(operand) {
			return tc.record(n, rl.NewFloatType())
		}
		tc.addUnaryOpIssue(n.Span(), n.Op, operand)
		return tc.record(n, rl.NewErrorTypeType())
	}
	return tc.record(n, rl.NewDynamicType())
}

// synthTernary handles `cond ? whenTrue : whenFalse`. The condition
// can be any truthy-able value; the result type is the union of the
// two branch types (or just one of them if they're identical).
func (tc *typeChecker) synthTernary(n *rl.Ternary) rl.TypingT {
	_ = tc.synth(n.Condition)
	whenTrue := tc.synth(n.True)
	whenFalse := tc.synth(n.False)
	return tc.record(n, unionOf(whenTrue, whenFalse))
}

// synthFallback handles `left ?? right`. Today we approximate the
// result as `union(left, right)` - once narrowing lands we can refine
// this to `(left - null) | right`, but the union is a safe
// over-approximation that doesn't lose any valid programs.
func (tc *typeChecker) synthFallback(n *rl.Fallback) rl.TypingT {
	left := tc.synth(n.Left)
	right := tc.synth(n.Right)
	return tc.record(n, unionOf(left, right))
}

// synthCatchExpr handles `expr catch fallback`. Same shape as Fallback
// but catches errors rather than null. Result is union(left, right);
// narrowing will later subtract `error` from the left branch.
func (tc *typeChecker) synthCatchExpr(n *rl.CatchExpr) rl.TypingT {
	left := tc.synth(n.Left)
	right := tc.synth(n.Right)
	return tc.record(n, unionOf(left, right))
}

// binaryOpResult is the static overload table. Returns (result, true)
// for supported operand combinations and (_, false) for combinations
// the runtime would reject. Mirrors core/expr_ops.go.
func binaryOpResult(op rl.Operator, l, r rl.TypingT) (rl.TypingT, bool) {
	switch op {
	case rl.OpEq, rl.OpNeq:
		// Equality is total - any type can be compared to any other.
		// The runtime's RadValue.Equals handles every combination.
		return rl.NewBoolType(), true
	case rl.OpAnd:
		// `and` returns false on a falsy left and the bool-coercion of
		// the right otherwise. Always bool.
		return rl.NewBoolType(), true
	case rl.OpOr:
		// `or` returns the actual value of whichever operand wins, so
		// the static result is the union of the two branches. Once
		// narrowing lands this can become `(left - falsy) | right`.
		return unionOf(l, r), true
	case rl.OpIn, rl.OpNotIn:
		// Result is always bool; the right side must be a container
		// (str / list / map). Left can be anything.
		if isStr(r) || isList(r) || isMap(r) {
			return rl.NewBoolType(), true
		}
		return nil, false
	case rl.OpLt, rl.OpLte, rl.OpGt, rl.OpGte:
		// Numeric or string-on-string comparison.
		if isNumeric(l) && isNumeric(r) {
			return rl.NewBoolType(), true
		}
		if isStr(l) && isStr(r) {
			return rl.NewBoolType(), true
		}
		return nil, false
	case rl.OpAdd:
		// int+int -> int, with int->float widening.
		if isInt(l) && isInt(r) {
			return rl.NewIntType(), true
		}
		if isNumeric(l) && isNumeric(r) {
			return rl.NewFloatType(), true
		}
		// str+str (concat). Error operands also flow through as str.
		if (isStr(l) || isError(l)) && (isStr(r) || isError(r)) {
			return rl.NewStrType(), true
		}
		// list+list -> any-list. Element typing for list+list is left
		// to a later sub-commit (it needs union-of-element-types and
		// invariance-aware widening; AnyList is safe for now).
		if isList(l) && isList(r) {
			return rl.NewAnyListType(), true
		}
		return nil, false
	case rl.OpSub:
		if isInt(l) && isInt(r) {
			return rl.NewIntType(), true
		}
		if isNumeric(l) && isNumeric(r) {
			return rl.NewFloatType(), true
		}
		return nil, false
	case rl.OpMul:
		// int*int -> int, mixed numeric -> float.
		if isInt(l) && isInt(r) {
			return rl.NewIntType(), true
		}
		if isNumeric(l) && isNumeric(r) {
			return rl.NewFloatType(), true
		}
		// String repetition: str*int and int*str both yield str.
		if (isStr(l) && isInt(r)) || (isInt(l) && isStr(r)) {
			return rl.NewStrType(), true
		}
		return nil, false
	case rl.OpDiv:
		// Rad's `/` is true division: int/int yields float, not int.
		// The runtime intentionally returns float64 from int/int.
		if isNumeric(l) && isNumeric(r) {
			return rl.NewFloatType(), true
		}
		return nil, false
	case rl.OpMod:
		// int%int -> int; any other numeric mix widens to float
		// (mirroring math.Mod behavior in core/expr_ops.go).
		if isInt(l) && isInt(r) {
			return rl.NewIntType(), true
		}
		if isNumeric(l) && isNumeric(r) {
			return rl.NewFloatType(), true
		}
		return nil, false
	}
	return nil, false
}

// --- Type predicates --------------------------------------------------
//
// Small helpers that keep the overload table readable. They unwrap
// Optional<T> on the way in (`int?` should behave like `int` for these
// classifications); operating on null is a runtime concern handled by
// narrowing, not by the static op tables.

func isInt(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingIntT)
	return ok
}

func isFloat(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingFloatT)
	return ok
}

func isNumeric(t rl.TypingT) bool { return isInt(t) || isFloat(t) }

func isStr(t rl.TypingT) bool {
	switch t.(type) {
	case *rl.TypingStrT, *rl.TypingStrEnumT:
		return true
	}
	return false
}

func isError(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingErrorT)
	return ok
}

func isList(t rl.TypingT) bool {
	switch t.(type) {
	case *rl.TypingAnyListT, *rl.TypingListT, *rl.TypingTupleT:
		return true
	}
	return false
}

func isMap(t rl.TypingT) bool {
	switch t.(type) {
	case *rl.TypingAnyMapT, *rl.TypingMapT, *rl.TypingStructT:
		return true
	}
	return false
}

func isErrorType(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingErrorTypeT)
	return ok
}

// isDynamicLike covers any/dynamic - the "we don't know" types - but
// not never or <error>, which have their own handling. Used to short-
// circuit operator dispatch with "result is dynamic, no diagnostic".
func isDynamicLike(t rl.TypingT) bool {
	switch t.(type) {
	case *rl.TypingAnyT, *rl.TypingDynamicT:
		return true
	}
	return false
}

func anyIsErrorType(types ...rl.TypingT) bool {
	for _, t := range types {
		if isErrorType(t) {
			return true
		}
	}
	return false
}

func anyIsDynamicLike(types ...rl.TypingT) bool {
	for _, t := range types {
		if isDynamicLike(t) {
			return true
		}
	}
	return false
}

// unionOf produces a static union of two types, collapsing duplicates
// by Name(). Used by `or`, `??`, `catch`, and `?:` to express "result
// is one of the two operand types".
func unionOf(a, b rl.TypingT) rl.TypingT {
	if a == nil && b == nil {
		return rl.NewDynamicType()
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.Name() == b.Name() {
		return a
	}
	return rl.NewUnionType(a, b)
}

// --- Operator diagnostics --------------------------------------------

func (tc *typeChecker) addOpIssue(span rl.Span, op rl.Operator, left, right rl.TypingT, isCompound bool) {
	opStr := op.String()
	if isCompound {
		opStr += "="
	}
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     span,
		Severity: IssueHint,
		Code:     rl.ErrInvalidTypeForOp,
		Message: fmt.Sprintf("Invalid operand types: cannot do '%s %s %s'",
			left.Name(), opStr, right.Name()),
	})
}

func (tc *typeChecker) addUnaryOpIssue(span rl.Span, op rl.Operator, operand rl.TypingT) {
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     span,
		Severity: IssueHint,
		Code:     rl.ErrInvalidTypeForOp,
		Message: fmt.Sprintf("Invalid operand type '%s' for unary '%s'",
			operand.Name(), op.String()),
	})
}

// --- Collection literals ---------------------------------------------
//
// List and map literals synthesize to a parameterized collection type
// derived from their elements. Empty literals fall back to the
// unparameterized AnyList/AnyMap rather than erroring: that's the
// gradual-typing choice. A future "look-around" pass can refine
// `xs = []` followed by `xs.append(1)` to `List<int>`, but the safe
// over-approximation lets every existing program type-check today.

func (tc *typeChecker) synthLitList(n *rl.LitList) rl.TypingT {
	if len(n.Elements) == 0 {
		return tc.record(n, rl.NewAnyListType())
	}
	elemTypes := make([]rl.TypingT, 0, len(n.Elements))
	for _, e := range n.Elements {
		elemTypes = append(elemTypes, tc.synth(e))
	}
	widened := widenElementTypes(elemTypes)
	// Bare-ErrorType element poisons the whole literal. Wrapping it in
	// List<ErrorType> would cascade: invariant collection assignability
	// rejects List<X>.IsAssignableFrom(List<ErrorType>) for every X.
	if isErrorType(widened) {
		return tc.record(n, rl.NewErrorTypeType())
	}
	return tc.record(n, rl.NewListType(widened))
}

func (tc *typeChecker) synthLitMap(n *rl.LitMap) rl.TypingT {
	if len(n.Entries) == 0 {
		return tc.record(n, rl.NewAnyMapType())
	}
	keyTypes := make([]rl.TypingT, 0, len(n.Entries))
	valTypes := make([]rl.TypingT, 0, len(n.Entries))
	for _, e := range n.Entries {
		keyTypes = append(keyTypes, tc.synth(e.Key))
		valTypes = append(valTypes, tc.synth(e.Value))
	}
	keyT := widenElementTypes(keyTypes)
	valT := widenElementTypes(valTypes)
	if isErrorType(keyT) || isErrorType(valT) {
		return tc.record(n, rl.NewErrorTypeType())
	}
	return tc.record(n, rl.NewMapType(keyT, valT))
}

// widenElementTypes computes the static element type for a collection
// from its individual element types. The rules:
//
//   - If any element is ErrorType, return ErrorType so cascading
//     diagnostics stay suppressed.
//   - If any element is any/dynamic, return Dynamic - we can't pin a
//     useful element type and AnyList/AnyMap is a more honest answer.
//   - Apply the lone implicit numeric widening: a mix of int and float
//     collapses to float (matching IsAssignableFrom), not int|float.
//     Otherwise unique types form a union; identical types collapse.
//
// Returns Dynamic for an empty slice as a defensive fallback; callers
// handle the truly-empty case (LitList{}, LitMap{}) before getting
// here.
func widenElementTypes(types []rl.TypingT) rl.TypingT {
	if len(types) == 0 {
		return rl.NewDynamicType()
	}
	allNumeric := true
	hasFloat := false
	for _, t := range types {
		if isErrorType(t) {
			return rl.NewErrorTypeType()
		}
		if isDynamicLike(t) {
			return rl.NewDynamicType()
		}
		if !isNumeric(t) {
			allNumeric = false
		}
		if isFloat(t) {
			hasFloat = true
		}
	}
	if allNumeric && hasFloat {
		return rl.NewFloatType()
	}
	// Deduplicate by Name(). Order-preserving so a list literal of all
	// `str` stays `List<str>` rather than getting reordered.
	seen := map[string]bool{}
	unique := make([]rl.TypingT, 0, len(types))
	for _, t := range types {
		name := t.Name()
		if seen[name] {
			continue
		}
		seen[name] = true
		unique = append(unique, t)
	}
	if len(unique) == 1 {
		return unique[0]
	}
	return rl.NewUnionType(unique...)
}
