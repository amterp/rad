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
		frame: NewFrame(),
	}
	// Seed SymbolTypes from any pre-pinned declared annotations the
	// binder recorded (function parameters with type annotations,
	// typed-local declarations). Reads in the body fall through
	// frame -> SymbolTypes and find the base type without a separate
	// "enter function scope" hook.
	//
	// We walk both resolved.Decls (covers typed-local declarations)
	// and the unique symbol set from resolved.Uses (covers params,
	// which the binder declares with a nil DefNode and so never
	// registers in Decls). The two sets overlap; the de-dup happens
	// implicitly because both write into the same map.
	for _, sym := range resolved.Decls {
		if sym.Declared != nil {
			tc.info.SymbolTypes[sym] = sym.Declared
		}
	}
	for _, sym := range resolved.Uses {
		if sym != nil && sym.Declared != nil {
			tc.info.SymbolTypes[sym] = sym.Declared
		}
	}
	tc.walkFile(file)
	return tc.info
}

// typeChecker carries state during a single TypeCheck invocation.
// Like the binder, it isn't safe for concurrent use; the public
// TypeCheck function constructs a fresh one per call.
//
// frame holds the flow-sensitive narrowings in effect at the current
// program point. It evolves through the walk: branches push/pop
// narrowings, assignments record the latest inferred / re-assigned
// type, and the if-join unions branches at the bottom. Reads of a
// symbol consult frame first, then info.SymbolTypes (the base
// declared / first-inferred type) as a fallback.
type typeChecker struct {
	resolved *Resolved
	info     *TypeInfo
	frame    *Frame
	// returnStack carries one frame per enclosing function/lambda
	// scope. `collected` accumulates return-value types for return
	// inference (used by synthLambda; reserved for hoisted-fn
	// inference). `expected` is the declared return type, used to
	// check each `return E` at the return-statement's own span;
	// nil means "no declared return, only inference applies."
	// Pushed on lambda/fn entry, popped on exit.
	returnStack []returnFrame
	// reassignedInScope is the position map of all reassignments
	// in the IMMEDIATELY enclosing fn / lambda body. Used by the
	// closure rule in synthLambda: a captured path that gets
	// reassigned AFTER the lambda's definition can't carry its
	// current narrowing into the lambda body (the lambda might
	// run after the reassignment invalidates it). Set on
	// fn / lambda entry (save & swap), restored on exit. Empty
	// map = no reassignments to worry about.
	reassignedInScope map[*Symbol][]rl.Span
}

type returnFrame struct {
	collected []rl.TypingT
	expected  rl.TypingT
}

func (tc *typeChecker) pushReturnFrame(expected rl.TypingT) {
	tc.returnStack = append(tc.returnStack, returnFrame{expected: expected})
}

func (tc *typeChecker) popReturnFrame() []rl.TypingT {
	n := len(tc.returnStack)
	out := tc.returnStack[n-1].collected
	tc.returnStack = tc.returnStack[:n-1]
	return out
}

// recordReturn appends a return-value type to the innermost fn
// scope's accumulator and, if that scope has a declared return
// type, checks the value against it at the return's own span. A
// return outside any fn (validation-error elsewhere) is a no-op.
func (tc *typeChecker) recordReturn(t rl.TypingT, retNode rl.Node) {
	n := len(tc.returnStack)
	if n == 0 {
		return
	}
	tc.returnStack[n-1].collected = append(tc.returnStack[n-1].collected, t)
	expected := tc.returnStack[n-1].expected
	if expected == nil || retNode == nil || t == nil {
		return
	}
	if isErrorType(t) || isDynamicLike(t) {
		return
	}
	if expected.IsAssignableFrom(t) {
		return
	}
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     retNode.Span(),
		Severity: IssueHint,
		Code:     rl.ErrTypeMismatch,
		Message: fmt.Sprintf("Return value of type '%s' is not assignable to declared return type '%s'",
			t.Name(), expected.Name()),
	})
}

func (tc *typeChecker) walkFile(file *rl.SourceFile) {
	// Pre-pass: walk top-level hoisted fn bodies in reverse-topo
	// SCC order so each fn body sees its callees' fully-inferred
	// return types. Mutually recursive fns within an SCC share a
	// placeholder return during the walk, which we resolve to the
	// union of all return statements across the SCC at the end.
	//
	// This is the Phase 2 plan's two-pass shape. Hoisted top-level
	// fns are processed here; their bodies are NOT re-walked in
	// the main pass below. Nested FnDef inference still happens
	// inline in the body walk via the same walkFnDef path and
	// uses the simple source-order strategy - mutual recursion
	// across nested fns is a rare enough corner case to defer
	// until it bites.
	tc.inferHoistedFnReturns(file)
	// Track reassignments in the file scope so lambdas defined at
	// top level get the closure-rule treatment for top-level
	// locals that get reassigned later.
	prevReassigned := tc.reassignedInScope
	tc.reassignedInScope = tc.scanReassignments(file.Stmts)
	defer func() { tc.reassignedInScope = prevReassigned }()
	for _, stmt := range file.Stmts {
		// Skip top-level FnDefs; the inference pre-pass already
		// walked their bodies and would otherwise double-emit
		// diagnostics.
		if _, isFn := stmt.(*rl.FnDef); isFn {
			continue
		}
		tc.walkStmt(stmt)
	}
	for _, cmd := range file.Cmds {
		tc.walkCmd(cmd)
	}
}

// inferHoistedFnReturns runs the SCC-ordered return-type inference
// pre-pass over top-level hoisted FnDefs. After this returns, every
// such fn's SymbolTypes entry carries its inferred (or declared)
// return type, so references-by-name see the right shape.
func (tc *typeChecker) inferHoistedFnReturns(file *rl.SourceFile) {
	fns := make([]*rl.FnDef, 0)
	for _, stmt := range file.Stmts {
		if fn, ok := stmt.(*rl.FnDef); ok && fn.Name != "" {
			fns = append(fns, fn)
		}
	}
	if len(fns) == 0 {
		return
	}

	// Build the dependency edges: fn -> set of fns it references in
	// its body (excluding nested FnDef bodies, which have their own
	// inference). The "references" are uses whose resolved Symbol
	// matches another top-level fn's symbol.
	fnSym := make(map[*rl.FnDef]*Symbol, len(fns))
	symToFn := make(map[*Symbol]*rl.FnDef, len(fns))
	for _, fn := range fns {
		sym := tc.resolved.Decls[fn]
		if sym == nil {
			continue
		}
		fnSym[fn] = sym
		symToFn[sym] = fn
	}

	deps := make(map[*rl.FnDef]map[*rl.FnDef]struct{}, len(fns))
	for _, fn := range fns {
		set := make(map[*rl.FnDef]struct{})
		tc.collectFnRefs(fn.Body, symToFn, set, fn)
		deps[fn] = set
	}

	// Compute SCCs in reverse-topological order (leaves first).
	sccs := tarjanSCC(fns, deps)

	for _, scc := range sccs {
		tc.processFnSCC(scc, fnSym)
	}
}

// collectFnRefs walks `nodes` looking for Identifier uses whose
// resolved Symbol is one of the hoisted fns in symToFn. Found refs
// (other than self) become edges in the dependency graph. Nested
// FnDef bodies are NOT descended into: those fns own their own
// inference; their references contribute to their own dependency
// edges, not the outer's.
func (tc *typeChecker) collectFnRefs(nodes []rl.Node, symToFn map[*Symbol]*rl.FnDef, set map[*rl.FnDef]struct{}, owner *rl.FnDef) {
	for _, n := range nodes {
		tc.collectFnRefsNode(n, symToFn, set, owner)
	}
}

func (tc *typeChecker) collectFnRefsNode(n rl.Node, symToFn map[*Symbol]*rl.FnDef, set map[*rl.FnDef]struct{}, owner *rl.FnDef) {
	if n == nil {
		return
	}
	switch v := n.(type) {
	case *rl.FnDef:
		// Don't recurse into nested FnDef bodies - they have their
		// own inference. (But do still capture the FnDef's name
		// as a binding; the surrounding stmt list is what we walk.)
		return
	case *rl.Identifier:
		if sym, ok := tc.resolved.Uses[v]; ok {
			if dep, ok := symToFn[sym]; ok && dep != owner {
				set[dep] = struct{}{}
			}
		}
		return
	case *rl.Lambda:
		// Lambdas DO contribute to dependency edges - their bodies
		// can reference outer fns, and synth'ing the lambda inside
		// the parent fn's walk reads those fns' planted types.
		tc.collectFnRefs(v.Body, symToFn, set, owner)
		return
	}
	for _, c := range n.Children() {
		tc.collectFnRefsNode(c, symToFn, set, owner)
	}
}

// tarjanSCC returns the SCCs of the dependency graph in reverse
// topological order. Each inner slice is one SCC; the slice of
// slices is ordered so that an SCC's dependencies all appear
// earlier in the result. Standard Tarjan's algorithm.
func tarjanSCC(nodes []*rl.FnDef, deps map[*rl.FnDef]map[*rl.FnDef]struct{}) [][]*rl.FnDef {
	type state struct {
		index, lowlink int
		onStack        bool
		visited        bool
	}
	st := make(map[*rl.FnDef]*state, len(nodes))
	for _, n := range nodes {
		st[n] = &state{}
	}
	var stack []*rl.FnDef
	var sccs [][]*rl.FnDef
	idx := 0

	var strongconnect func(*rl.FnDef)
	strongconnect = func(v *rl.FnDef) {
		s := st[v]
		s.index = idx
		s.lowlink = idx
		s.visited = true
		idx++
		stack = append(stack, v)
		s.onStack = true

		for w := range deps[v] {
			ws, ok := st[w]
			if !ok {
				continue
			}
			if !ws.visited {
				strongconnect(w)
				if ws.lowlink < s.lowlink {
					s.lowlink = ws.lowlink
				}
			} else if ws.onStack {
				if ws.index < s.lowlink {
					s.lowlink = ws.index
				}
			}
		}

		if s.lowlink == s.index {
			var scc []*rl.FnDef
			for {
				n := len(stack) - 1
				w := stack[n]
				stack = stack[:n]
				st[w].onStack = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			sccs = append(sccs, scc)
		}
	}

	for _, n := range nodes {
		if !st[n].visited {
			strongconnect(n)
		}
	}
	// Tarjan emits SCCs in reverse-topo order naturally.
	return sccs
}

// processFnSCC walks every fn in the SCC, then resolves the
// shared inferred return type by unioning all collected returns
// across all members of the SCC.
//
// For a singleton SCC (no self-cycle, no mutual recursion), this
// reduces to: plant a placeholder if unannotated, walk the body,
// union the collected types, finalize. Placeholder is
// TypingNeverT - unionTypesForJoin drops Never on the way through,
// so any path that returns through the recursive call contributes
// nothing concrete and the non-recursive paths determine the
// final return type. If the SCC has no non-recursive return paths
// at all, the union collapses to Never and we fall back to
// Dynamic ("we couldn't infer") rather than surface Never to
// users.
func (tc *typeChecker) processFnSCC(scc []*rl.FnDef, fnSym map[*rl.FnDef]*Symbol) {
	// Step 1: plant placeholder return for every fn in the SCC
	// that doesn't already have a declared return. Fns with
	// declared returns keep their declaration; the body walk
	// uses it as `expected` and emits return-mismatch diagnostics.
	placeholder := rl.NewNeverType()
	needsInference := make(map[*rl.FnDef]bool, len(scc))
	for _, fn := range scc {
		sym := fnSym[fn]
		if sym == nil {
			continue
		}
		if fn.Typing == nil || fn.Typing.ReturnT == nil {
			needsInference[fn] = true
			// Replace SymbolTypes with a TypingFnT carrying the
			// placeholder return so recursive refs see something
			// concrete-looking during the body walk.
			params := []rl.TypingFnParam(nil)
			if fn.Typing != nil {
				params = fn.Typing.Params
			}
			var ret rl.TypingT = placeholder
			tc.info.SymbolTypes[sym] = &rl.TypingFnT{
				FnName:  fn.Name,
				Params:  params,
				ReturnT: &ret,
			}
		}
	}

	// Step 2: walk every body, collecting returns per fn.
	collected := make(map[*rl.FnDef][]rl.TypingT, len(scc))
	for _, fn := range scc {
		collected[fn] = tc.walkFnDefForInference(fn)
	}

	// Step 3: union ALL collected types across the SCC (the plan's
	// "unified with union of return statement types in the SCC").
	// Placeholder Never drops in unionTypesForJoin; recursive paths
	// thus contribute nothing concrete. We distinguish three cases
	// after the union:
	//
	//   - Empty (no returns at all in any SCC member): void. The
	//     fn legitimately has no return value.
	//   - Never (returns existed but all dropped as Never): all
	//     paths went recursive, no concrete exit. Fall back to
	//     Dynamic - we can't pin a type, but it's not "void"
	//     either since the call would in principle return
	//     something.
	//   - Anything else: that's the union of concrete returns.
	var all []rl.TypingT
	for _, fn := range scc {
		all = append(all, collected[fn]...)
	}
	var scope rl.TypingT
	if len(all) == 0 {
		scope = rl.NewVoidType()
	} else {
		scope = unionTypesForJoin(all)
		if _, isNever := scope.(*rl.TypingNeverT); isNever {
			scope = rl.NewDynamicType()
		}
	}

	// Step 4: write the resolved return type back into SymbolTypes
	// for every fn in the SCC that needed inference. Singletons get
	// the same answer they'd compute alone; multi-fn SCCs share the
	// SCC-wide union.
	for _, fn := range scc {
		if !needsInference[fn] {
			continue
		}
		sym := fnSym[fn]
		if sym == nil {
			continue
		}
		params := []rl.TypingFnParam(nil)
		if fn.Typing != nil {
			params = fn.Typing.Params
		}
		ret := scope
		tc.info.SymbolTypes[sym] = &rl.TypingFnT{
			FnName:  fn.Name,
			Params:  params,
			ReturnT: &ret,
		}
	}
}

// scanReassignments collects per-symbol assignment spans for every
// Identifier-target Assign inside `body`, stopping at nested fn /
// lambda boundaries (those have their own scope). The result is
// used by synthLambda to decide which captured narrowings to drop
// when entering a lambda body - Pyright's closure rule: a path
// reassigned after a lambda's definition can't safely carry its
// current narrowing inside the lambda body, since the lambda may
// run after that reassignment.
//
// Both fresh declarations (`x = 5`) and compound assigns
// (`x += 1`, `x++`) count. A param has 0 entries here (the binding
// is the parameter slot, not an Assign), so unmodified params
// stay narrowed inside captured lambdas, which is the intuitive
// behavior.
func (tc *typeChecker) scanReassignments(body []rl.Node) map[*Symbol][]rl.Span {
	out := make(map[*Symbol][]rl.Span)
	var visit func(rl.Node)
	visit = func(n rl.Node) {
		if n == nil {
			return
		}
		switch v := n.(type) {
		case *rl.FnDef, *rl.Lambda:
			// Don't recurse - inner fn / lambda bodies own their
			// own reassignment scope. Their own walks will populate
			// reassignedInScope when they fire.
			_ = v
			return
		case *rl.Assign:
			for _, t := range v.Targets {
				ident, ok := t.(*rl.Identifier)
				if !ok {
					continue
				}
				sym, ok := tc.resolved.Uses[ident]
				if !ok || sym == nil {
					continue
				}
				out[sym] = append(out[sym], ident.Span())
			}
			// Values still walked - they may contain nested
			// statements we care about (e.g. a fallback expression
			// with a side-effect, though Rad doesn't really have
			// embedded statements in exprs).
			for _, c := range n.Children() {
				visit(c)
			}
			return
		}
		for _, c := range n.Children() {
			visit(c)
		}
	}
	for _, s := range body {
		visit(s)
	}
	return out
}

// closureOverridesForLambda computes the SymbolTypes overrides a
// lambda body should layer on top of its captured frame, per
// Pyright's closure rule. For each symbol that:
//
//   - is narrowed in the enclosing frame at the lambda's position,
//   - is assigned anywhere AFTER the lambda's position in the
//     enclosing scope,
//
// we override it to its base type (info.SymbolTypes / sym.Declared
// fallback), effectively dropping the narrowing inside the lambda
// body. Symbols not reassigned (or reassigned only before the
// lambda) keep their narrowing.
func (tc *typeChecker) closureOverridesForLambda(lam *rl.Lambda) map[*Symbol]rl.TypingT {
	if len(tc.reassignedInScope) == 0 {
		return nil
	}
	lambdaStart := lam.Span().StartByte
	overrides := make(map[*Symbol]rl.TypingT)
	for sym, spans := range tc.reassignedInScope {
		reassignedAfter := false
		for _, s := range spans {
			if s.StartByte > lambdaStart {
				reassignedAfter = true
				break
			}
		}
		if !reassignedAfter {
			continue
		}
		// Only matters if the enclosing frame currently narrows
		// this symbol - otherwise there's nothing to drop.
		if _, narrowed := tc.frame.Lookup(sym); !narrowed {
			continue
		}
		var base rl.TypingT
		if t, ok := tc.info.SymbolTypes[sym]; ok {
			base = t
		} else if sym.Declared != nil {
			base = sym.Declared
		} else {
			base = rl.NewDynamicType()
		}
		overrides[sym] = base
	}
	if len(overrides) == 0 {
		return nil
	}
	return overrides
}

// prePlantLambdaSignature writes a placeholder TypingFnT (Never
// return) onto SymbolTypes for a target that's about to receive a
// lambda value. The body walk that follows can then resolve
// recursive references through callSignatureFor / synthIdentifier
// to that placeholder, instead of falling back to Dynamic. After
// synth completes, walkAssign overwrites SymbolTypes with the
// lambda's real inferred TypingFnT.
func (tc *typeChecker) prePlantLambdaSignature(sym *Symbol, lam *rl.Lambda) {
	params := []rl.TypingFnParam(nil)
	if lam.Typing != nil {
		params = lam.Typing.Params
	}
	var ret rl.TypingT = rl.NewNeverType()
	tc.info.SymbolTypes[sym] = &rl.TypingFnT{
		Params:  params,
		ReturnT: &ret,
	}
}

// walkFnDefForInference is walkFnDef but exposes the collected
// return-statement types so the SCC orchestrator can build the
// inferred return. Semantics are identical to walkFnDef otherwise.
func (tc *typeChecker) walkFnDefForInference(n *rl.FnDef) []rl.TypingT {
	saved := tc.frame
	if n.Typing != nil {
		for _, p := range n.Typing.Params {
			if p.DefaultAST != nil && p.DefaultAST.Node != nil {
				_ = tc.synth(p.DefaultAST.Node)
			}
		}
	}
	tc.frame = NewFrame()
	var declaredReturn rl.TypingT
	if n.Typing != nil && n.Typing.ReturnT != nil {
		declaredReturn = *n.Typing.ReturnT
	}
	tc.pushReturnFrame(declaredReturn)
	prevReassigned := tc.reassignedInScope
	tc.reassignedInScope = tc.scanReassignments(n.Body)
	tc.walkStmts(n.Body)
	tc.reassignedInScope = prevReassigned
	collected := tc.popReturnFrame()
	tc.frame = saved
	return collected
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
	case *rl.If:
		tc.walkIf(v)
	case *rl.Switch:
		tc.walkSwitch(v)
	case *rl.ForLoop:
		tc.walkForLoop(v)
	case *rl.WhileLoop:
		tc.walkWhileLoop(v)
	case *rl.FnDef:
		tc.walkFnDef(v)
	case *rl.Return:
		// Synth the return value(s) and feed them into the innermost
		// fn scope's return collector. Bare `return` is void; a
		// single value contributes its own type; multi-value
		// `return a, b` builds a tuple so the static return type
		// matches what the unpack-side `x, y = f()` would expect.
		var t rl.TypingT
		switch len(v.Values) {
		case 0:
			t = rl.NewVoidType()
		case 1:
			t = tc.synth(v.Values[0])
		default:
			elems := make([]rl.TypingT, len(v.Values))
			for i, val := range v.Values {
				elems[i] = tc.synth(val)
			}
			t = rl.NewTupleType(elems...)
		}
		tc.recordReturn(t, v)
	default:
		// Generic descent. Later sub-commits replace these with
		// kind-specific handlers (for loops, switch, return, etc.).
		for _, child := range n.Children() {
			tc.walkStmt(child)
		}
	}
}

// walkStmts walks a body of statements in order. Each statement may
// update tc.frame (assignments record new types, nested ifs join at
// the bottom of their sub-flow); subsequent statements see the
// post-update frame.
func (tc *typeChecker) walkStmts(stmts []rl.Node) {
	for _, s := range stmts {
		tc.walkStmt(s)
	}
}

// walkIf threads narrowings through an if/elif/else chain and joins
// the frames at the bottom.
//
// For each branch with a condition:
//   - The condition is interpreted against the accumulated falsy
//     frame (everything previous branches couldn'\''t match).
//   - The branch body walks with WhenTrue layered on the accumulated
//     frame; subsequent branches accumulate WhenFalse.
//   - If the branch falls off the end (doesn'\''t return/break/continue),
//     its exit frame contributes to the join.
//
// The else branch (no condition) walks with the final accumulated
// falsy frame. When there is no else, the "implicit else" - the path
// where no branch matched - is the accumulated falsy frame itself.
//
// Early-exit handling is the practical win: `if x == null: return`
// leaves the post-if frame as "x is non-null" because the null branch
// returned and the fall-through carries the falsy refinement forward.
func (tc *typeChecker) walkIf(n *rl.If) {
	initial := tc.frame
	acc := initial
	var branchFrames []*Frame
	hasElse := false
	for _, branch := range n.Branches {
		if branch.Condition == nil {
			hasElse = true
			tc.frame = acc
			tc.walkStmts(branch.Body)
			if !tc.branchExits(branch.Body) {
				branchFrames = append(branchFrames, tc.frame)
			}
			continue
		}
		// Synth the condition under the accumulated frame so hover
		// over a sub-expression in (e.g.) `elif x.foo:` sees the
		// previously-narrowed type of x.
		tc.frame = acc
		_ = tc.synth(branch.Condition)
		ref := tc.interpretCondition(branch.Condition, tc.frame)
		tc.frame = tc.frame.WithMany(ref.WhenTrue)
		tc.walkStmts(branch.Body)
		if !tc.branchExits(branch.Body) {
			branchFrames = append(branchFrames, tc.frame)
		}
		acc = acc.WithMany(ref.WhenFalse)
	}
	if !hasElse {
		// The fall-through path: no branch fired. Carries every
		// accumulated WhenFalse refinement.
		branchFrames = append(branchFrames, acc)
	}
	tc.frame = tc.joinFrames(initial, branchFrames)
}

// walkFnDef handles a function definition statement. The body opens
// a fresh frame so that narrowings active in the surrounding scope
// don'\''t leak into the function body - a function can be called
// from anywhere, and the call-site enclosing narrowings aren'\''t
// available at the body. Body-internal narrowings stay isolated too:
// after walking, tc.frame is restored to whatever the surrounding
// scope had before the function definition.
//
// Param default expressions are walked under the SURROUNDING frame
// (defaults are evaluated at call time but in the caller'\''s scope,
// not the callee'\''s - matching the runtime'\''s lazy default eval).
//
// Closure rule deferred: same caveat as synthLambda. Captured
// narrowings from the enclosing frame are NOT preserved. For a
// closure-friendly design we'\''d need Pyright'\''s reassignment-after-
// definition check. Today, nested functions just see Dynamic for
// outer locals. Conservative but sound.
func (tc *typeChecker) walkFnDef(n *rl.FnDef) {
	saved := tc.frame
	// Param defaults synth in the surrounding frame (not the body'\''s
	// fresh frame), matching Phase 2'\''s behavior - they'\''re evaluated
	// at the call site, which is in the caller'\''s scope.
	if n.Typing != nil {
		for _, p := range n.Typing.Params {
			if p.DefaultAST != nil && p.DefaultAST.Node != nil {
				_ = tc.synth(p.DefaultAST.Node)
			}
		}
	}
	tc.frame = NewFrame()
	// Push a return-collector frame so any `return E` inside the
	// body lands in this fn'\''s slot, not whatever lambda we may
	// be nested under at the call site. We feed the declared
	// return (if any) so body returns get checked against it at
	// their own span, same as lambdas.
	var declaredReturn rl.TypingT
	if n.Typing != nil && n.Typing.ReturnT != nil {
		declaredReturn = *n.Typing.ReturnT
	}
	tc.pushReturnFrame(declaredReturn)
	// Swap reassignedInScope to this fn body's scan, so lambdas
	// defined inside see the right enclosing-reassignments set.
	prevReassigned := tc.reassignedInScope
	tc.reassignedInScope = tc.scanReassignments(n.Body)
	tc.walkStmts(n.Body)
	tc.reassignedInScope = prevReassigned
	collected := tc.popReturnFrame()
	tc.frame = saved

	// Inference for nested fns. Top-level fns are handled in the
	// SCC pre-pass (inferHoistedFnReturns) which gets to walk
	// bodies in dependency order; nested fns just get walked here
	// in source order, with no SCC support. That means a nested
	// fn that mutually recurses with a sibling sees Any for
	// forward refs. Single-fn recursion is fine: the binder
	// planted fn.Typing on the symbol, but with nil ReturnT;
	// callSignatureFor now prefers SymbolTypes, so we plant a
	// Never placeholder before the walk and resolve after.
	//
	// We skip this work entirely when there's a declared return -
	// the symbol already carries the declared shape and the body
	// returns were checked against it via the pushReturnFrame
	// expected slot.
	if n.Name == "" || declaredReturn != nil {
		return
	}
	sym, ok := tc.resolved.Decls[n]
	if !ok || sym == nil {
		return
	}
	var ret rl.TypingT
	if len(collected) == 0 {
		ret = rl.NewVoidType()
	} else {
		ret = unionTypesForJoin(collected)
		if _, isNever := ret.(*rl.TypingNeverT); isNever {
			ret = rl.NewDynamicType()
		}
	}
	params := []rl.TypingFnParam(nil)
	if n.Typing != nil {
		params = n.Typing.Params
	}
	tc.info.SymbolTypes[sym] = &rl.TypingFnT{
		FnName:  n.Name,
		Params:  params,
		ReturnT: &ret,
	}
}

// walkForLoop handles `for vars in iter [with ctx]:`. The loop var
// is typed from the iterable'\''s element shape so reads inside the
// body see a real type (not Dynamic). After the loop the binding
// keeps that type - Rad'\''s runtime leaves loop variables in scope
// past the construct, and the static answer matches.
//
// Multi-var loops bind only the first var to the element type today.
// The binder records the LAST declared loop-var in resolved.Decls
// (overwriting earlier ones since they all key on the ForLoop node),
// which makes generalized multi-var typing fragile. When Rad adds a
// proper ForLoopVars map on Resolved we'\''ll cover unpacking shape
// (k, v) over maps and (i, v) over enumerated lists.
//
// Loop-body narrowings can persist past the construct because the
// runtime doesn'\''t open a new scope for loops. The frame after the
// body becomes the post-loop frame; a future "drop narrowings for
// vars assigned in body" (Sorbet rule, Phase 4i) lands separately.
func (tc *typeChecker) walkForLoop(n *rl.ForLoop) {
	iterType := tc.synth(n.Iter)

	if len(n.Vars) == 1 {
		if sym, ok := tc.resolved.Decls[n]; ok && sym != nil && sym.Kind == SymLoopVar {
			elem := loopElementType(iterType)
			tc.info.SymbolTypes[sym] = elem
			tc.frame = tc.frame.With(sym, elem)
		}
	}

	tc.walkStmts(n.Body)
}

// walkWhileLoop handles `while [cond]:`. Inside the body the
// condition'\''s WhenTrue applies (we entered because cond was
// truthy); after the loop the WhenFalse applies (we exited because
// cond finally turned falsy).
//
// Sorbet rule for body-assigned vars: before applying the condition
// narrowing, drop narrowings for any symbol the body reassigns. The
// body may run any number of times, so a "narrowed before entering
// the body" type isn'\''t sound for an iteration where the body
// already changed the var. Resetting to the base SymbolTypes /
// Declared is the conservative answer that lets the body'\''s own
// assignments re-narrow on each iteration as needed.
//
// We collect the assigned-symbol set from the body statements with
// a shallow walk (top-level Assigns only). Nested constructs
// (if/switch/loops) that reassign a var are missed by this pass; a
// deeper walk would catch them but pulls in more complexity than
// the common case needs. Worth revisiting if a real script hits the
// missed pattern.
func (tc *typeChecker) walkWhileLoop(n *rl.WhileLoop) {
	initial := tc.frame
	if n.Condition == nil {
		tc.walkStmts(n.Body)
		return
	}
	assigned := tc.collectAssignedSyms(n.Body)
	widened := tc.dropNarrowings(initial, assigned)

	tc.frame = widened
	_ = tc.synth(n.Condition)
	ref := tc.interpretCondition(n.Condition, widened)
	tc.frame = widened.WithMany(ref.WhenTrue)
	tc.walkStmts(n.Body)
	bodyExit := tc.frame

	// Post-loop has two possible paths:
	//   1. "Body never ran" - condition was false on entry.
	//   2. "Body ran one or more times, then exited" - either
	//      condition became falsy at the top OR a break inside
	//      the body jumped out.
	// Joining the two captures whatever the body mutated in its
	// frame: `while cond: x = "hello"` leaves x as the union of
	// its pre-body type (body never ran) and "hello" (body ran).
	//
	// We don'\''t special-case a "body always diverges" early exit
	// here. Distinguishing diverges-via-return (post-loop is dead)
	// from diverges-via-break (post-loop gets bodyExit at break
	// time) would need finer-grained reachability tracking than
	// branchExits provides. Joining both unconditionally is
	// safe - in the diverges-via-return case the post-loop is
	// unreachable anyway, and any spurious union arms there don'\''t
	// affect a real program path.
	//
	// WhenFalse is applied only on the fall-through path, since
	// that'\''s the only one where the loop exited "naturally" via
	// a falsy condition. The body-ran path may have exited via
	// break, where the condition'\''s truthy state still held.
	fallThrough := widened.WithMany(ref.WhenFalse)
	tc.frame = tc.joinFrames(initial, []*Frame{bodyExit, fallThrough})
}

// collectAssignedSyms scans a body of statements for assignment
// targets and returns the set of Symbols that get reassigned. Used
// by the while-loop Sorbet rule to widen narrowings before
// re-applying the condition's refinement.
//
// Recurses into nested control-flow blocks (if/elif/else,
// switch/case/default, inner for/while bodies, catch blocks),
// since a reassignment buried inside any of those can still fire
// on some iteration. Stops at FnDef and Lambda boundaries: nested
// function bodies have their own narrowing context, and any
// reassignment inside one is scoped to that inner frame, not the
// enclosing loop's variables.
func (tc *typeChecker) collectAssignedSyms(stmts []rl.Node) map[*Symbol]bool {
	out := map[*Symbol]bool{}
	for _, s := range stmts {
		tc.collectAssignedSymsIn(s, out)
	}
	return out
}

func (tc *typeChecker) collectAssignedSymsIn(n rl.Node, out map[*Symbol]bool) {
	if n == nil {
		return
	}
	switch v := n.(type) {
	case *rl.Assign:
		for _, target := range v.Targets {
			ident, ok := target.(*rl.Identifier)
			if !ok {
				continue
			}
			if sym, ok := tc.resolved.Uses[ident]; ok && sym != nil {
				out[sym] = true
			}
		}
		if v.Catch != nil {
			for _, s := range v.Catch.Stmts {
				tc.collectAssignedSymsIn(s, out)
			}
		}
	case *rl.If:
		for _, b := range v.Branches {
			for _, s := range b.Body {
				tc.collectAssignedSymsIn(s, out)
			}
		}
	case *rl.Switch:
		for _, c := range v.Cases {
			tc.collectAssignedSymsIn(c.Alt, out)
		}
		if v.Default != nil {
			tc.collectAssignedSymsIn(v.Default.Alt, out)
		}
	case *rl.SwitchCaseBlock:
		for _, s := range v.Stmts {
			tc.collectAssignedSymsIn(s, out)
		}
	case *rl.ForLoop:
		for _, s := range v.Body {
			tc.collectAssignedSymsIn(s, out)
		}
	case *rl.WhileLoop:
		for _, s := range v.Body {
			tc.collectAssignedSymsIn(s, out)
		}
	case *rl.ExprStmt:
		if v.Catch != nil {
			for _, s := range v.Catch.Stmts {
				tc.collectAssignedSymsIn(s, out)
			}
		}
	case *rl.Shell:
		for _, target := range v.Targets {
			ident, ok := target.(*rl.Identifier)
			if !ok {
				continue
			}
			if sym, ok := tc.resolved.Uses[ident]; ok && sym != nil {
				out[sym] = true
			}
		}
		if v.Catch != nil {
			for _, s := range v.Catch.Stmts {
				tc.collectAssignedSymsIn(s, out)
			}
		}
		// FnDef, Lambda: intentionally not recursed - inner function
		// bodies have their own narrowing context.
	}
}

// dropNarrowings returns a frame where each symbol in syms has its
// narrowing reset back to the base type (SymbolTypes or Declared).
// Implemented as a child frame that re-pins those symbols to base -
// the parent chain is unchanged so other narrowings keep flowing.
func (tc *typeChecker) dropNarrowings(frame *Frame, syms map[*Symbol]bool) *Frame {
	if len(syms) == 0 {
		return frame
	}
	overrides := make(map[*Symbol]rl.TypingT, len(syms))
	for sym := range syms {
		if t, ok := tc.info.SymbolTypes[sym]; ok {
			overrides[sym] = t
			continue
		}
		if sym.Declared != nil {
			overrides[sym] = sym.Declared
		}
		// No known base type. We don'\''t add an override here, so a
		// prior narrowing from an ancestor frame keeps applying via
		// Frame.Lookup'\''s parent walk. That'\''s probably wrong for
		// the Sorbet rule - the whole point is to widen vars the
		// body may reassign - but the fix is to track base types
		// for every symbol, not to special-case this branch. Until
		// then, this is the conservative under-widening: we may keep
		// a narrowing we should have dropped, but we don'\''t introduce
		// spurious narrowings.
	}
	return frame.WithMany(overrides)
}

// loopElementType returns the element type for a single-var loop over
// the given iterable. Lists yield their element type; maps yield the
// key type (matches the Rad runtime, which iterates map keys); strings
// yield str (each iteration produces a one-char string). Anything else
// falls back to Dynamic.
func loopElementType(iter rl.TypingT) rl.TypingT {
	if iter == nil {
		return rl.NewDynamicType()
	}
	switch v := iter.(type) {
	case *rl.TypingListT:
		return v.Elem()
	case *rl.TypingAnyListT:
		return rl.NewAnyType()
	case *rl.TypingStrT:
		return rl.NewStrType()
	case *rl.TypingMapT:
		return v.KeyT()
	case *rl.TypingAnyMapT:
		return rl.NewAnyType()
	}
	return rl.NewDynamicType()
}

// walkSwitch handles `switch <disc>:` with per-case narrowing of the
// discriminant. For each case the body walks with the discriminant
// symbol narrowed to the case'\''s match type; the discriminant'\''s
// residual (= base minus all peeled cases) flows into the default
// branch.
//
// String-enum exhaustiveness: when the discriminant is a closed type
// (string-enum) and no default arm catches the rest, the residual
// after peeling all explicit cases should be Never. If it'\''s
// non-Never, the switch is missing variants - emit an
// ErrNonExhaustiveSwitch diagnostic naming the unmatched values.
//
// Non-enum discriminants get the normal walk (case bodies don'\''t
// narrow) and no exhaustiveness check. The mechanism is in place;
// extending it to int / int-enum / sealed unions is straightforward
// when those types arrive.
func (tc *typeChecker) walkSwitch(n *rl.Switch) {
	discType := tc.synth(n.Discriminant)
	var discSym *Symbol
	if id, ok := n.Discriminant.(*rl.Identifier); ok {
		if sym, ok := tc.resolved.Uses[id]; ok {
			discSym = sym
		}
	}

	initial := tc.frame
	residual := discType
	var branchFrames []*Frame

	for _, c := range n.Cases {
		caseType := tc.matchTypeForCaseKeys(c.Keys)
		var branchFrame *Frame
		if discSym != nil && caseType != nil && !isErrorType(caseType) {
			branchFrame = initial.With(discSym, caseType)
		} else {
			branchFrame = initial
		}
		tc.frame = branchFrame
		exitsEarly := tc.walkSwitchAlt(c.Alt)
		if !exitsEarly {
			branchFrames = append(branchFrames, tc.frame)
		}
		residual = subtractEnumType(residual, caseType)
	}

	if n.Default != nil {
		var defFrame *Frame
		if discSym != nil && residual != nil {
			defFrame = initial.With(discSym, residual)
		} else {
			defFrame = initial
		}
		tc.frame = defFrame
		if !tc.walkSwitchAlt(n.Default.Alt) {
			branchFrames = append(branchFrames, tc.frame)
		}
	} else {
		// No default - the unmatched values fall through past the
		// switch unchanged. If the residual is non-Never on a closed
		// (string-enum) discriminant type, emit a non-exhaustive
		// diagnostic; the runtime would error at the unmatched
		// value, and surfacing it statically is the whole point.
		if !isNeverType(residual) && isClosedDiscriminant(discType) {
			tc.emitNonExhaustiveSwitch(n, residual)
		}
		var fallFrame *Frame
		if discSym != nil && residual != nil {
			fallFrame = initial.With(discSym, residual)
		} else {
			fallFrame = initial
		}
		branchFrames = append(branchFrames, fallFrame)
	}

	tc.frame = tc.joinFrames(initial, branchFrames)
}

// walkSwitchAlt walks the body of a switch case (block form) or
// synths the expressions (expression form), returning whether the
// case unconditionally diverts (return/break/continue at the end of
// the block form).
//
// Expression-form alts never exit early: they evaluate to a value
// and fall through to the post-switch flow.
func (tc *typeChecker) walkSwitchAlt(alt rl.Node) bool {
	switch a := alt.(type) {
	case *rl.SwitchCaseBlock:
		tc.walkStmts(a.Stmts)
		return tc.branchExits(a.Stmts)
	case *rl.SwitchCaseExpr:
		for _, v := range a.Values {
			_ = tc.synth(v)
		}
		return false
	}
	return false
}

// matchTypeForCaseKeys turns a list of case keys into the type that
// the discriminant takes inside that case'\''s body. For string
// literals we build a string-enum naming the matched values; for
// anything else we currently return nil (no narrowing applied).
//
// Mixed string + non-string keys disqualify the partition; nil is
// safer than guessing. Future work: int-enum once Rad gains them.
func (tc *typeChecker) matchTypeForCaseKeys(keys []rl.Node) rl.TypingT {
	if len(keys) == 0 {
		return nil
	}
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		s, ok := simpleStringValue(k)
		if !ok {
			return nil
		}
		values = append(values, s)
	}
	return rl.NewStrEnumType(values...)
}

// subtractEnumType removes the values of caseType from base. Returns
// base unchanged if either side isn'\''t a string-enum. For exhausted
// enums (all values peeled), returns Never.
func subtractEnumType(base, caseType rl.TypingT) rl.TypingT {
	if base == nil {
		return nil
	}
	bEnum, bOK := base.(*rl.TypingStrEnumT)
	cEnum, cOK := caseType.(*rl.TypingStrEnumT)
	if !bOK || !cOK {
		return base
	}
	caseSet := make(map[string]bool, len(cEnum.Values()))
	for _, v := range cEnum.Values() {
		caseSet[v] = true
	}
	remaining := make([]string, 0, len(bEnum.Values()))
	for _, v := range bEnum.Values() {
		if !caseSet[v] {
			remaining = append(remaining, v)
		}
	}
	if len(remaining) == 0 {
		return rl.NewNeverType()
	}
	return rl.NewStrEnumType(remaining...)
}

// isClosedDiscriminant reports whether a discriminant'\''s static type
// is a closed set the checker can exhaustively analyze. Today: just
// string-enums. Bool would be a near-term extension (true / false
// being the closed set), if we add an exhaustive-bool predicate.
func isClosedDiscriminant(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingStrEnumT)
	return ok
}

// emitNonExhaustiveSwitch records a diagnostic naming the unmatched
// values of a closed discriminant. The message lists the values so
// the user can read the fix off the diagnostic without inspecting
// the discriminant'\''s declared type.
func (tc *typeChecker) emitNonExhaustiveSwitch(n *rl.Switch, residual rl.TypingT) {
	msg := "Switch is not exhaustive"
	if enum, ok := residual.(*rl.TypingStrEnumT); ok {
		msg = fmt.Sprintf("Switch is not exhaustive; missing case for %s", enum.Name())
	}
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     n.Span(),
		Severity: IssueWarning,
		Code:     rl.ErrNonExhaustiveSwitch,
		Message:  msg,
	})
}

// branchExitsEarly is a deep "does this body always diverge?" check
// suitable for if/switch branch joins. Phase 4e'\''s "last statement
// is a return/break/continue" version was too shallow: a common
// guard pattern like
//
//	if err:
//	    log_it()
//	    return
//
// has return as the second-to-last statement of the if body, but
// branchExitsEarly only inspected the LAST statement (log_it()), so
// the surrounding flow didn'\''t pick up the divergence and the
// "after-the-if" narrowing was wrong.
//
// The new predicate walks the body, recursing through if (every
// branch including an explicit else must diverge), switch (every
// case + default OR exhaustive over a closed enum), and while-true
// (no top-level break in the body). It also recognizes calls to
// no-return builtins like exit().
//
// Reachability via type-system: a body whose entry frame has any
// symbol typed Never is unreachable, hence divergent. The frame
// argument is optional - pass nil to skip that check. For the
// branch-join use today we pass nil since the frame is already
// authoritative at the call site.
//
// insideLoop controls whether break/continue count as divergence:
// only inside an actual loop do they transfer control past the
// enclosing construct. The current call sites (walkIf, walkSwitch)
// don'\''t themselves carry a loop indicator yet, so they default to
// true to preserve the prior behavior - break/continue have always
// been treated as "exits this branch" in practice because
// branchExitsEarly accepted them unconditionally.
// branchExits is the typeChecker method form. The receiver lets the
// divergence walk consult sym info (noReturn builtins, closed-enum
// switch exhaustiveness) that the standalone helper can'\''t reach.
// Callers from walkIf / walkSwitch use this method form.
func (tc *typeChecker) branchExits(body []rl.Node) bool {
	return bodyDiverges(body, divergeCtx{insideLoop: true}, tc)
}

// divergeCtx threads context through the recursive divergence walk.
// Today it only tracks the loop scope; future fields (e.g. function-
// return-type) can fit alongside without rewiring the call sites.
type divergeCtx struct {
	insideLoop bool
}

// bodyDiverges is the entry point: ANY divergent statement in the
// body means the whole body diverges, because everything after it
// is unreachable.
func bodyDiverges(stmts []rl.Node, ctx divergeCtx, tc *typeChecker) bool {
	for _, s := range stmts {
		if stmtDiverges(s, ctx, tc) {
			return true
		}
	}
	return false
}

// stmtDiverges is the syntactic recursion on a single statement.
// Mirrors the structure laid out in mypy / Pyright / TypeScript:
// terminal control-flow nodes diverge directly; composite nodes
// diverge only when every alternative does.
func stmtDiverges(n rl.Node, ctx divergeCtx, tc *typeChecker) bool {
	if n == nil {
		return false
	}
	switch s := n.(type) {
	case *rl.Return, *rl.Yield:
		return true
	case *rl.Break, *rl.Continue:
		return ctx.insideLoop
	case *rl.ExprStmt:
		if call, ok := s.Expr.(*rl.Call); ok {
			return callIsNoReturn(call, tc)
		}
		return false
	case *rl.If:
		return ifDiverges(s, ctx, tc)
	case *rl.Switch:
		return switchDiverges(s, ctx, tc)
	case *rl.WhileLoop:
		return whileDiverges(s, tc)
	case *rl.ForLoop:
		// For loops may iterate zero times; the body might never
		// run. Even if the body diverges, the whole loop doesn'\''t.
		return false
	case *rl.Assign:
		// Bare assignment never diverges. An assignment with a
		// catch block is also non-divergent: catch is an alternate
		// continuation, and the success path always falls through.
		return false
	}
	return false
}

// ifDiverges: every branch must diverge AND there must be an explicit
// else. Without an else, the "no branch matched" path is an implicit
// fall-through, so the if as a whole doesn'\''t diverge.
func ifDiverges(n *rl.If, ctx divergeCtx, tc *typeChecker) bool {
	hasElse := false
	for _, branch := range n.Branches {
		if branch.Condition == nil {
			hasElse = true
		}
		if !bodyDiverges(branch.Body, ctx, tc) {
			return false
		}
	}
	return hasElse
}

// switchDiverges: every explicit case must diverge, AND either a
// default arm is present and also diverges, OR the discriminant'\''s
// residual after peeling all cases is Never (i.e., exhaustive on a
// closed type).
func switchDiverges(n *rl.Switch, ctx divergeCtx, tc *typeChecker) bool {
	for _, c := range n.Cases {
		if !altDiverges(c.Alt, ctx, tc) {
			return false
		}
	}
	if n.Default != nil {
		return altDiverges(n.Default.Alt, ctx, tc)
	}
	// No default - require exhaustiveness on a closed discriminant.
	if tc == nil {
		return false
	}
	discType := tc.synth(n.Discriminant)
	residual := discType
	for _, c := range n.Cases {
		caseType := tc.matchTypeForCaseKeys(c.Keys)
		residual = subtractEnumType(residual, caseType)
	}
	return isNeverType(residual)
}

// altDiverges: case bodies are statement lists, case expressions
// always fall through (they produce a value).
func altDiverges(alt rl.Node, ctx divergeCtx, tc *typeChecker) bool {
	switch a := alt.(type) {
	case *rl.SwitchCaseBlock:
		return bodyDiverges(a.Stmts, ctx, tc)
	case *rl.SwitchCaseExpr:
		return false
	}
	return false
}

// whileDiverges: an infinite `while:` (no condition) diverges iff
// the body has no top-level break that targets this loop. A
// conditional `while cond:` may execute zero times - non-divergent.
// (Detecting constant-true conditions would require literal folding;
// we don'\''t do it today.)
func whileDiverges(n *rl.WhileLoop, tc *typeChecker) bool {
	if n.Condition != nil {
		return false
	}
	return !bodyHasTopLevelBreak(n.Body)
}

// bodyHasTopLevelBreak scans a body for `break` statements whose
// target is the enclosing loop. Recurses into if/switch (their
// breaks still target the enclosing loop) but NOT into nested
// ForLoop / WhileLoop (their breaks target themselves).
func bodyHasTopLevelBreak(stmts []rl.Node) bool {
	for _, s := range stmts {
		if stmtHasTopLevelBreak(s) {
			return true
		}
	}
	return false
}

func stmtHasTopLevelBreak(n rl.Node) bool {
	switch s := n.(type) {
	case *rl.Break:
		return true
	case *rl.If:
		for _, b := range s.Branches {
			if bodyHasTopLevelBreak(b.Body) {
				return true
			}
		}
	case *rl.Switch:
		for _, c := range s.Cases {
			if altHasTopLevelBreak(c.Alt) {
				return true
			}
		}
		if s.Default != nil && altHasTopLevelBreak(s.Default.Alt) {
			return true
		}
	}
	return false
}

func altHasTopLevelBreak(alt rl.Node) bool {
	if a, ok := alt.(*rl.SwitchCaseBlock); ok {
		return bodyHasTopLevelBreak(a.Stmts)
	}
	return false
}

// callIsNoReturn reports whether call is to a builtin function
// whose runtime semantics never return - exit(), today. Future
// shape: check the builtin signature'\''s return type against Never.
func callIsNoReturn(call *rl.Call, tc *typeChecker) bool {
	if tc == nil {
		return false
	}
	ident, ok := call.Func.(*rl.Identifier)
	if !ok {
		return false
	}
	sym, ok := tc.resolved.Uses[ident]
	if !ok || sym == nil || sym.Kind != SymBuiltin {
		return false
	}
	switch sym.Name {
	case "exit":
		return true
	}
	return false
}

// joinFrames computes the post-branch frame from a set of branch-exit
// frames. For each symbol narrowed by any branch (relative to
// initial), the joined type is the union of each branch'\''s effective
// type for that symbol. Branches that didn'\''t narrow the symbol
// contribute the base type from info.SymbolTypes - or the initial
// frame'\''s narrowing if present, via Frame.Lookup'\''s parent walk.
//
// Returns initial unchanged when no branch narrowed anything (or
// when only one branch survived early-exit filtering).
func (tc *typeChecker) joinFrames(initial *Frame, branches []*Frame) *Frame {
	if len(branches) == 0 {
		return initial
	}
	if len(branches) == 1 {
		return branches[0]
	}
	narrowed := map[*Symbol]bool{}
	for _, branch := range branches {
		for cur := branch; cur != nil && cur != initial; cur = cur.parent {
			for s := range cur.bindings {
				narrowed[s] = true
			}
		}
	}
	if len(narrowed) == 0 {
		return initial
	}
	joined := make(map[*Symbol]rl.TypingT, len(narrowed))
	for sym := range narrowed {
		types := make([]rl.TypingT, 0, len(branches))
		for _, branch := range branches {
			if t, ok := branch.Lookup(sym); ok {
				types = append(types, t)
				continue
			}
			if t, ok := tc.info.SymbolTypes[sym]; ok {
				types = append(types, t)
			}
		}
		if len(types) > 0 {
			joined[sym] = unionTypesForJoin(types)
		}
	}
	return initial.WithMany(joined)
}

// unionTypesForJoin collapses a slice of branch types into a single
// type, applying the type-lattice join semantics so that semantically
// equivalent unions collapse to their tightest representation.
//
// Pipeline (in order):
//
//  1. Flatten: hoist nested Union arms so the result is a single-
//     level union.
//  2. Drop Never (unreachable arm) and ErrorType (poison marker).
//     Both contribute nothing to a join; unioning them with X gives X.
//  3. Structural merge: collapse multiple StrEnum arms into one with
//     the unioned value set. (Future: do the same for Optional arms
//     with compatible inners.) Preserves the first occurrence'\''s
//     position in the result for diagnostic order.
//  4. Pairwise subsumption: if some arm B is assignable-from another
//     arm A (B is a super-type of A), drop A. This is what makes
//     int|int? collapse to int? - TypingOptionalT.IsAssignableFrom
//     already accepts non-optional T, so the general subsumption
//     handles it without a special case. Skip gradual arms (any /
//     dynamic / error_type) on the subsumer side to prevent them
//     from swallowing concrete information.
//  5. Dedupe by Name as a final guard.
//
// When everything collapses to nothing, return Never (the empty
// join). Single result returned bare. Multiple distinct arms wrapped
// in a Union.
func unionTypesForJoin(types []rl.TypingT) rl.TypingT {
	flat := flattenUnion(types)
	flat = dropNeverAndErrorType(flat)
	flat = mergeStructuralByKind(flat)
	flat = applySubsumption(flat)
	flat = dedupeByName(flat)

	switch len(flat) {
	case 0:
		return rl.NewNeverType()
	case 1:
		return flat[0]
	default:
		return rl.NewUnionType(flat...)
	}
}

// UnionJoinForTest exposes unionTypesForJoin for testing. Tests in
// the check_test package can'\''t reach the unexported helper; this
// thin wrapper keeps the test interface explicit (not "everything in
// the package is exported"). Not for production use.
func UnionJoinForTest(types []rl.TypingT) rl.TypingT {
	return unionTypesForJoin(types)
}

// flattenUnion produces a single-level slice from the input: each
// TypingUnionT arm is replaced by its constituent arms, nils are
// dropped. Single-level invariant simplifies the downstream passes.
func flattenUnion(types []rl.TypingT) []rl.TypingT {
	out := make([]rl.TypingT, 0, len(types))
	for _, t := range types {
		if t == nil {
			continue
		}
		if u, ok := t.(*rl.TypingUnionT); ok {
			out = append(out, u.Types()...)
			continue
		}
		out = append(out, t)
	}
	return out
}

// dropNeverAndErrorType filters out the two "empty contribution"
// markers. Never means "this arm was proven unreachable"; ErrorType
// is the checker'\''s poison sentinel - both should disappear in a
// join (unioning Never or ErrorType with X yields X).
func dropNeverAndErrorType(types []rl.TypingT) []rl.TypingT {
	out := make([]rl.TypingT, 0, len(types))
	for _, t := range types {
		switch t.(type) {
		case *rl.TypingNeverT, *rl.TypingErrorTypeT:
			continue
		}
		out = append(out, t)
	}
	return out
}

// mergeStructuralByKind collapses arms of the same "merge kind" into
// a single arm. Today we merge string-enums: multiple StrEnum arms
// fold into one with the unioned value set, so a if/elif/else over
// closed-enum cases produces `StrEnum<"a","b","c">` instead of
// `StrEnum<"a">|StrEnum<"b">|StrEnum<"c">`.
//
// The first StrEnum'\''s position in the input is preserved as the
// position of the merged result. This keeps diagnostic ordering
// stable when other arms are interleaved.
func mergeStructuralByKind(types []rl.TypingT) []rl.TypingT {
	// Collect string-enum values in input order; dedupe.
	seenVals := map[string]bool{}
	var mergedVals []string
	firstEnumPos := -1
	for i, t := range types {
		e, ok := t.(*rl.TypingStrEnumT)
		if !ok {
			continue
		}
		if firstEnumPos == -1 {
			firstEnumPos = i
		}
		for _, v := range e.Values() {
			if seenVals[v] {
				continue
			}
			seenVals[v] = true
			mergedVals = append(mergedVals, v)
		}
	}
	if firstEnumPos == -1 {
		return types
	}
	out := make([]rl.TypingT, 0, len(types))
	emitted := false
	for _, t := range types {
		if _, ok := t.(*rl.TypingStrEnumT); ok {
			if !emitted {
				out = append(out, rl.NewStrEnumType(mergedVals...))
				emitted = true
			}
			continue
		}
		out = append(out, t)
	}
	return out
}

// applySubsumption drops arms that are subsumed by another arm via
// IsAssignableFrom. The algorithm is the "kept-set with bidirectional
// pruning" pattern from Pyright: walk candidates left-to-right, for
// each one (a) drop it if anything already kept subsumes it; (b)
// otherwise, drop any kept arms it subsumes, then add it.
//
// Gradual types (Any, Dynamic, ErrorType) don'\''t subsume on the
// subsumer side: we don'\''t want `int | any` to collapse to `any`,
// because the concrete int arm carries real static info that a
// downstream narrowing might exploit. Matching Pyright'\''s rule
// keeps concrete arms alive alongside gradual ones.
//
// O(n^2) but n is tiny (typically 2-5 arms); correctness over speed.
func applySubsumption(types []rl.TypingT) []rl.TypingT {
	var kept []rl.TypingT
	for _, candidate := range types {
		// Gradual candidates can'\''t be subsumed - we always want to
		// preserve an explicit `any` in the join. The asymmetry
		// matters because gradual consistency makes IsAssignableFrom
		// return true in both directions for any-vs-T, which would
		// otherwise let concrete arms "subsume" gradual ones and
		// silently erase them.
		subsumed := false
		if !isGradualLike(candidate) {
			for _, k := range kept {
				if isGradualLike(k) {
					continue
				}
				if k.IsAssignableFrom(candidate) {
					subsumed = true
					break
				}
			}
		}
		if subsumed {
			continue
		}
		// Candidate keeps; drop any kept arms it subsumes - again
		// skip when the candidate is gradual so concretes survive.
		if !isGradualLike(candidate) {
			pruned := kept[:0]
			for _, k := range kept {
				if isGradualLike(k) {
					pruned = append(pruned, k)
					continue
				}
				if candidate.IsAssignableFrom(k) {
					continue
				}
				pruned = append(pruned, k)
			}
			kept = pruned
		}
		kept = append(kept, candidate)
	}
	return kept
}

// dedupeByName guards against duplicate Name() outputs slipping
// through (e.g. two distinct instances of TypingIntT). Subsumption
// catches most of these, but a final Name-based pass is cheap
// insurance.
func dedupeByName(types []rl.TypingT) []rl.TypingT {
	seen := map[string]bool{}
	out := make([]rl.TypingT, 0, len(types))
	for _, t := range types {
		n := t.Name()
		if seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, t)
	}
	return out
}

// isGradualLike identifies types the static checker uses to signal
// "we couldn'\''t pin a type" (Any, Dynamic) or "an error is
// suppressing cascades" (ErrorType). They should not subsume
// concrete arms in a union join.
func isGradualLike(t rl.TypingT) bool {
	switch t.(type) {
	case *rl.TypingAnyT, *rl.TypingDynamicT, *rl.TypingErrorTypeT:
		return true
	}
	return false
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
//
// For typed locals (`x: int = 5`, with sym.Declared set) the RHS
// must be assignable to the declared type, and the recorded symbol
// type stays Declared rather than the RHS-derived value. Subsequent
// `x = something` reassignments are checked against the same
// Declared, so the annotation acts as a stable contract for every
// later read of the binding.
func (tc *typeChecker) walkAssign(a *rl.Assign) {
	for i, val := range a.Values {
		// Recursive-lambda support: when a lambda RHS is assigned
		// to an identifier whose symbol has no declared type, the
		// recursive reference inside the lambda body would otherwise
		// see Dynamic (SymbolTypes hasn't been populated yet). Plant
		// a tentative TypingFnT with a Never return placeholder so
		// the recursive callsite synths to Never (which propagates
		// trivially and drops in unionTypesForJoin, leaving non-
		// recursive paths to determine the inferred return). This
		// mirrors the hoisted-fn SCC pre-pass strategy for the
		// local-binding case. After synth completes, the lambda's
		// real TypingFnT overwrites the placeholder via the normal
		// SymbolTypes update below.
		if i < len(a.Targets) {
			if lam, ok := val.(*rl.Lambda); ok {
				if ident, ok := a.Targets[i].(*rl.Identifier); ok {
					if sym, ok := tc.resolved.Uses[ident]; ok && sym.Declared == nil {
						tc.prePlantLambdaSignature(sym, lam)
					}
				}
			}
		}
		valType := tc.synth(val)
		if i >= len(a.Targets) {
			continue
		}
		// Indexed assignment (`xs[i] = v`, `m[k] = v`): the target is
		// a VarPath whose last segment is a bracket-index. The
		// Identifier branch below doesn't fire, so we handle it here
		// and continue. We deliberately don't update SymbolTypes or
		// frame for these: indexed assign mutates the container's
		// contents, not the binding, so the symbol's type is unchanged.
		if vp, ok := a.Targets[i].(*rl.VarPath); ok {
			_ = tc.synth(vp) // walk children for hover/use tracking
			tc.checkIndexedAssignTarget(vp, val, valType)
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
		if sym.Declared != nil {
			tc.checkAssignAgainstDeclared(val, valType, sym.Declared)
			tc.info.SymbolTypes[sym] = sym.Declared
			// The frame tracks the RHS-derived type within the
			// current flow: a typed local can have a stricter
			// type than Declared at a given point (after `x =
			// 5` inside a branch, x is int even though declared
			// int|str). On a mismatch the runtime would error;
			// the frame stays at Declared so downstream reads
			// don'\''t cascade.
			actual := sym.Declared
			if sym.Declared.IsAssignableFrom(valType) {
				actual = valType
			}
			tc.frame = tc.frame.With(sym, actual)
			continue
		}
		// Unannotated: SymbolTypes records the base/initial inferred
		// type for hover and for join'\''s "branch didn'\''t narrow this"
		// fallback. The frame holds the flow-sensitive override,
		// which gets unioned at branch joins.
		tc.info.SymbolTypes[sym] = valType
		tc.frame = tc.frame.With(sym, valType)
	}
	// Walk the catch block (if any) with assignment targets narrowed
	// to their error component. The catch runs when the RHS errored,
	// so each target currently holds an error value - inside the
	// catch body, reads of the target should see `error`, not the
	// non-error narrowing walkAssign just installed in the frame.
	//
	// For a typed local `x: int|error = parse_int(...)` the override
	// is the error arm of the declared type. For an unannotated
	// local where the RHS synthed to int|error, we pick out the
	// error arm. Falls back to a bare TypingErrorT when we can'\''t
	// extract one (the runtime guarantee is "RHS errored," so error
	// is always sound).
	if a.Catch != nil {
		savedFrame := tc.frame
		overrides := make(map[*Symbol]rl.TypingT, len(a.Targets))
		for i, target := range a.Targets {
			ident, ok := target.(*rl.Identifier)
			if !ok {
				continue
			}
			sym, ok := tc.resolved.Uses[ident]
			if !ok || sym == nil {
				continue
			}
			var rhsType rl.TypingT
			if i < len(a.Values) {
				rhsType = tc.synth(a.Values[i])
			}
			errArm := extractErrorFrom(rhsType)
			if errArm == nil {
				errArm = rl.NewErrorType()
			}
			overrides[sym] = errArm
		}
		tc.frame = tc.frame.WithMany(overrides)
		tc.walkStmts(a.Catch.Stmts)
		tc.frame = savedFrame
	}
}

// checkAssignAgainstDeclared emits a type-mismatch when the assigned
// value can't flow into the declared slot. Severity matches Phase 2's
// per-arg precedent: Hint, not Error - the runtime still produces a
// richer value-aware message when the script runs, and we only
// promote once literal types fill the missing fidelity (a one-pass
// severity migration covers all assignability checks at once).
//
// ErrorType / Dynamic short-circuit: a poisoned RHS already produced
// a diagnostic and any-likes are universally assignable, so no extra
// diagnostic fires.
func (tc *typeChecker) checkAssignAgainstDeclared(valNode rl.Node, valType, declared rl.TypingT) {
	if valType == nil || isErrorType(valType) || isDynamicLike(valType) {
		return
	}
	if declared.IsAssignableFrom(valType) {
		return
	}
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     valNode.Span(),
		Severity: IssueHint,
		Code:     rl.ErrTypeMismatch,
		Message: fmt.Sprintf("Value of type '%s' is not assignable to declared type '%s'",
			valType.Name(), declared.Name()),
	})
}

// checkIndexedAssignTarget fires the ErrCollectionElementMismatch
// diagnostic when an indexed assignment like `xs[i] = v` or
// `m[k] = v` would put a value of the wrong type into a statically
// typed collection.
//
// The check is opt-in via annotation. We require the root symbol to
// carry an explicit Declared type (`xs: int[]`, `m: { str: int }`,
// or a typed param), and we skip when the user wrote the open
// `list` / `map` annotations - those are the gradual-typing escape
// hatch. Untyped locals (`xs = [1, 2]`) also skip: the literal does
// pin an inferred element type, but firing on indexed-assign through
// it would be intrusive for scripts that intentionally build a
// list with mixed elements over time.
//
// The element type itself comes from the frame (so narrowing through
// `if xs != null: xs[0] = ...` on a `xs: int[]?` declaration still
// catches mismatches), but only after the declared-type gate above
// has admitted us. Concretely: we want the most precise statically
// known element type, gated by the user having opted in.
//
// Severity is Hint, matching checkAssignAgainstDeclared: the runtime
// still enforces declared element types when the script runs, and
// this is the static "heads up" version. Other forms of mutation
// (struct field assignment, slice assignment) aren't checked here.
func (tc *typeChecker) checkIndexedAssignTarget(target *rl.VarPath, valNode rl.Node, valType rl.TypingT) {
	if valType == nil || isErrorType(valType) || isDynamicLike(valType) {
		return
	}
	segs := target.Segments
	if len(segs) == 0 {
		return
	}
	last := segs[len(segs)-1]
	if last.Index == nil || last.IsSlice || last.IsUFCS || last.Field != nil {
		return
	}
	// Today we only check the simple `<identifier>[i] = v` shape.
	// Chained targets (`m["a"][0] = v`, `obj.field[i] = v`) need a
	// chain-prefix type walker that we haven't built yet. Skipping
	// chains keeps Phase 6 sound: we just don't catch some mismatches
	// we could've caught.
	if len(segs) != 1 {
		return
	}
	rootIdent, ok := target.Root.(*rl.Identifier)
	if !ok {
		return
	}
	sym, ok := tc.resolved.Uses[rootIdent]
	if !ok || sym == nil || sym.Declared == nil {
		// Untyped or unresolved: opt-in only.
		return
	}
	// Open `list` / `map` annotations are the explicit "any element
	// goes" form. Respect them even when the frame happens to hold a
	// tighter type from a literal RHS.
	switch sym.Declared.(type) {
	case *rl.TypingAnyListT, *rl.TypingAnyMapT:
		return
	}
	rootType := tc.synthIdentifier(rootIdent)
	expected, ok := containerElementType(rootType)
	if !ok {
		return
	}
	if isErrorType(expected) || isDynamicLike(expected) {
		return
	}
	if expected.IsAssignableFrom(valType) {
		return
	}
	tc.info.Issues = append(tc.info.Issues, BindIssue{
		Span:     valNode.Span(),
		Severity: IssueHint,
		Code:     rl.ErrCollectionElementMismatch,
		Message: fmt.Sprintf("Value of type '%s' is not assignable to element type '%s'",
			valType.Name(), expected.Name()),
	})
}

// containerElementType returns the assignable-into element type of a
// statically typed collection. Returns (nil, false) for any container
// that doesn't have a checkable element type - untyped collections
// (AnyList/AnyMap), structs (we don't check field-assign here), and
// anything else (Dynamic, unions, etc.).
func containerElementType(c rl.TypingT) (rl.TypingT, bool) {
	switch v := c.(type) {
	case *rl.TypingListT:
		return v.Elem(), true
	case *rl.TypingMapT:
		return v.ValT(), true
	}
	return nil, false
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
//
// ExprTypes doubles as a memoization cache. Several call paths visit
// the same arg node more than once (synthCall walks args, then
// checkArgType re-synths each to compare against the formal type);
// without memoization, any diagnostic emitted during synth would fire
// once per visit. Caching here makes synth idempotent so each AST
// node produces at most one diagnostic regardless of visit count.
func (tc *typeChecker) synth(n rl.Node) rl.TypingT {
	if n == nil {
		return rl.NewDynamicType()
	}
	if t, ok := tc.info.ExprTypes[n]; ok {
		return t
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
	case *rl.Lambda:
		return tc.synthLambda(v)
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
		// Prefer the SymbolTypes-stored TypingFnT when available -
		// inferHoistedFnReturns writes the inferred return type
		// there during SCC processing, and during the SCC walk the
		// stored type carries the Never placeholder we want
		// recursive call sites to see. The raw fn.Typing on the
		// AST has nil ReturnT for unannotated fns, which would
		// collapse to Dynamic and defeat the placeholder strategy.
		if t, ok := tc.info.SymbolTypes[sym]; ok {
			if fnT, ok := t.(*rl.TypingFnT); ok {
				return fnT
			}
		}
		fn, ok := sym.DefNode.(*rl.FnDef)
		if !ok || fn.Typing == nil {
			return nil
		}
		return fn.Typing
	}
	// Fall-through for non-builtin / non-hoisted symbols: locals
	// bound to lambdas, function-typed params, etc. When their
	// SymbolTypes entry carries a TypingFnT (planted by typed-
	// local declarations, param annotations, or the recursive-
	// lambda pre-plant in walkAssign), return it so the call site
	// gets real shape-check + return-type resolution. Without
	// this branch, every fn-value call through a local fell back
	// to Dynamic.
	if t, ok := tc.info.SymbolTypes[sym]; ok {
		if fnT, ok := t.(*rl.TypingFnT); ok {
			return fnT
		}
	}
	// Frame may carry a tighter narrowed type for this symbol;
	// prefer it when it's a TypingFnT (e.g. a future narrowing
	// rule that proves the value is callable).
	if t, ok := tc.frame.Lookup(sym); ok {
		if fnT, ok := t.(*rl.TypingFnT); ok {
			return fnT
		}
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
// returns the type that holds at this program point: the frame
// narrowing if any, the base SymbolTypes entry otherwise. Forward
// references and unknown symbols fall back to Dynamic.
func (tc *typeChecker) synthIdentifier(ident *rl.Identifier) rl.TypingT {
	sym, ok := tc.resolved.Uses[ident]
	if !ok {
		return tc.record(ident, rl.NewDynamicType())
	}
	if t, known := tc.frame.Lookup(sym); known {
		return tc.record(ident, t)
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
	// Never operands suppress operand-type diagnostics and propagate
	// Never. This matters during the hoisted-fn SCC inference walk:
	// recursive calls inside an SCC return the Never placeholder, and
	// expressions like `factorial(n-1) * n` would otherwise emit
	// "Invalid operand: never * int" before the placeholder gets
	// resolved. Semantically Never is the bottom type - a value of
	// type Never can't exist, so any operation on it is vacuous.
	if anyIsNever(left, right) {
		return tc.record(n, rl.NewNeverType())
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
//
// A poisoned condition propagates: if cond synths to ErrorType, we
// can't tell which branch fires, and pretending the result is
// `whenTrue | whenFalse` would silently absorb the bad condition's
// failure into a plausible-looking type that downstream code uses
// without realizing the upstream is broken.
func (tc *typeChecker) synthTernary(n *rl.Ternary) rl.TypingT {
	cond := tc.synth(n.Condition)
	whenTrue := tc.synth(n.True)
	whenFalse := tc.synth(n.False)
	if isErrorType(cond) {
		return tc.record(n, rl.NewErrorTypeType())
	}
	return tc.record(n, unionOf(whenTrue, whenFalse))
}

// synthFallback handles `left ?? right`. The fallback fires when
// left is null, so the result is `(left - null) | right` - the
// non-null portion of left, unioned with the fallback expression.
//
// When left has no null component, the fallback can never fire, but
// we still union the two arms: the user wrote the fallback for a
// reason and excluding right entirely would surprise them if our
// nullability inference is wrong (and gradual typing means it
// sometimes is).
func (tc *typeChecker) synthFallback(n *rl.Fallback) rl.TypingT {
	left := tc.synth(n.Left)
	right := tc.synth(n.Right)
	if nonNull := stripNullFrom(left); nonNull != nil {
		return tc.record(n, unionOf(nonNull, right))
	}
	return tc.record(n, unionOf(left, right))
}

// synthCatchExpr handles `expr catch fallback`. Catch fires when
// left evaluates to an error, so the result is
// `(left - error) | right`.
//
// Same conservative tilt as synthFallback: when left has no error
// component, we still union the two arms rather than dropping right.
func (tc *typeChecker) synthCatchExpr(n *rl.CatchExpr) rl.TypingT {
	left := tc.synth(n.Left)
	right := tc.synth(n.Right)
	if nonErr := stripErrorFrom(left); nonErr != nil {
		return tc.record(n, unionOf(nonErr, right))
	}
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

func isNever(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingNeverT)
	return ok
}

func anyIsNever(types ...rl.TypingT) bool {
	for _, t := range types {
		if isNever(t) {
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
//
// ErrorType in either operand collapses the whole union to ErrorType.
// Returning `int | <error>` would leak the poison into downstream
// operator and assignment checks - the invariant collection /
// assignability rules can't reason about a union containing a
// poisoned arm, so each downstream use would re-fire a fresh
// diagnostic. Aggressive cascade prevention is the contract every
// other handler already follows; unionOf needs to honor it too.
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
	if isErrorType(a) || isErrorType(b) {
		return rl.NewErrorTypeType()
	}
	if a.Name() == b.Name() {
		return a
	}
	return rl.NewUnionType(a, b)
}

// --- Operator diagnostics --------------------------------------------

// strPlusMigrationHint is the help-line the runtime attaches when a
// str+non-str / non-str+str operation hits ErrInvalidTypeForOp.
// Surface the same text statically so LSP and `rad check` users see
// the same actionable follow-up as `rad <script>` users do.
const strPlusMigrationHint = "In v0.9, + no longer coerces types. Use string interpolation instead. See: https://amterp.dev/rad/migrations/v0.9/"

func (tc *typeChecker) addOpIssue(span rl.Span, op rl.Operator, left, right rl.TypingT, isCompound bool) {
	opStr := op.String()
	if isCompound {
		opStr += "="
	}
	issue := BindIssue{
		Span:     span,
		Severity: IssueHint,
		Code:     rl.ErrInvalidTypeForOp,
		Message: fmt.Sprintf("Invalid operand types: cannot do '%s %s %s'",
			left.Name(), opStr, right.Name()),
	}
	if op == rl.OpAdd && isStrPlusCoercible(left, right) {
		issue.Suggestion = strPlusMigrationHint
	}
	tc.info.Issues = append(tc.info.Issues, issue)
}

// isStrPlusCoercible mirrors core/expr_ops.go's isStrPlusNonStr: do we
// have a str on one side and an int/float/bool on the other? Those
// are the combinations users likely intended pre-v0.9 - the migration
// hint helps far more on those than on, say, list + int.
func isStrPlusCoercible(left, right rl.TypingT) bool {
	coercible := func(t rl.TypingT) bool {
		return isInt(t) || isFloat(t) || isBool(t)
	}
	return (isStr(left) && coercible(right)) || (coercible(left) && isStr(right))
}

func isBool(t rl.TypingT) bool {
	_, ok := t.(*rl.TypingBoolT)
	return ok
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

// synthLambda walks the lambda'\''s body and synthesizes a structural
// TypingFnT. Params come from the grammar annotation (l.Typing.Params,
// possibly with nil per-param Type for unannotated args - that maps
// to `any`). The return type is the headline work here:
//
//   - Declared annotation (`fn(x) -> int: ...`): use it verbatim.
//   - Block-form lambda: union the types of every `return E` we
//     encountered while walking the body. Bare `return` contributes
//     void. Empty (no returns at all) is void.
//   - Expression-form lambda: the body is a single ExprStmt and the
//     expression IS the return value, so we synth it directly and
//     skip the walk-and-collect dance.
//
// Closure rule deferred: same caveat as walkFnDef. Captured-path
// narrowings from the enclosing frame are NOT preserved. For a
// closure-friendly design we'\''d need Pyright'\''s reassignment-
// after-definition check. Today, the body opens a fresh frame so
// outer locals appear unnarrowed (read as their base type).
//
// Recursion: anonymous lambdas can'\''t self-reference. A named
// recursive lambda bound to a local (`f = fn(x) f(x-1)`) would synth
// the recursive `f` to Dynamic because SymbolTypes[f] isn'\''t set
// during the walk. Sound but lossy - the inferred return would
// collapse to Dynamic via any-like subsumption rules. Fixing this
// needs the Tarjan SCC + placeholder return type machinery the plan
// already describes for hoisted fns.
func (tc *typeChecker) synthLambda(n *rl.Lambda) rl.TypingT {
	enclosing := tc.frame
	// Lambdas inherit the enclosing frame'\''s narrowings - that'\''s
	// closure semantics. Pyright'\''s closure rule: drop narrowings
	// on any captured path that gets reassigned AFTER the lambda'\''s
	// definition in the enclosing scope. The lambda may run after
	// such a reassignment, at which point the captured narrowing
	// no longer holds. The overrides layer on top of the enclosing
	// frame, restoring captured-but-reassigned paths to their base
	// type for the body walk.
	if overrides := tc.closureOverridesForLambda(n); overrides != nil {
		tc.frame = tc.frame.WithMany(overrides)
	}

	// Feed the declared return (if any) into the return frame so
	// each `return E` in the body gets checked against the
	// declaration at its own span. Lambdas without a declared
	// return have nothing to check against; the inferred return
	// flows out via popReturnFrame.
	var declaredReturn rl.TypingT
	if n.Typing != nil && n.Typing.ReturnT != nil {
		declaredReturn = *n.Typing.ReturnT
	}
	tc.pushReturnFrame(declaredReturn)

	// Swap reassignedInScope so any lambda nested inside this one
	// sees this body'\''s reassignments as its enclosing scope.
	prevReassigned := tc.reassignedInScope
	tc.reassignedInScope = tc.scanReassignments(n.Body)

	// Expression-form lambdas (`fn(x) x + 1`) put the expression
	// node directly in Body (verified against the converter output -
	// it does not wrap in ExprStmt). Each body entry is the value to
	// return; synth it to feed the inferred return type. We pass
	// the body node itself as the diagnostic span - it'\''s the
	// implicit return for these. Block-form lambdas walk via
	// walkStmts and rely on the `*rl.Return` case to populate
	// returnStack.
	if !n.IsBlock {
		for _, stmt := range n.Body {
			if stmt == nil {
				continue
			}
			tc.recordReturn(tc.synth(stmt), stmt)
		}
	} else {
		tc.walkStmts(n.Body)
	}

	tc.reassignedInScope = prevReassigned
	collected := tc.popReturnFrame()
	tc.frame = enclosing

	// Honor an explicit return annotation; otherwise infer. Same
	// shape as hoisted-fn inference: empty returns → void, all-
	// dropped (purely recursive) → Dynamic, else the union.
	var returnT rl.TypingT
	if n.Typing != nil && n.Typing.ReturnT != nil {
		returnT = *n.Typing.ReturnT
	} else if len(collected) == 0 {
		returnT = rl.NewVoidType()
	} else {
		returnT = unionTypesForJoin(collected)
		if _, isNever := returnT.(*rl.TypingNeverT); isNever {
			returnT = rl.NewDynamicType()
		}
	}

	// Construct a fresh TypingFnT so we don'\''t mutate the parsed
	// l.Typing (other readers - signature display, hover - rely on
	// the AST being immutable). Params are shared by value; the
	// caller never writes through ReturnT, so a fresh pointer is
	// safe.
	params := []rl.TypingFnParam(nil)
	if n.Typing != nil {
		params = n.Typing.Params
	}
	fn := &rl.TypingFnT{
		Params:  params,
		ReturnT: &returnT,
	}
	return tc.record(n, fn)
}

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
