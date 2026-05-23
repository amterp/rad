package check

import "github.com/amterp/rad/rts/rl"

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
	}
	for _, child := range n.Children() {
		_ = tc.synth(child)
	}
	return tc.record(n, rl.NewDynamicType())
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
