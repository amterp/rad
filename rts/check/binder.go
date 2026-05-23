package check

import (
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
)

// Resolve performs name resolution on a parsed script and returns a
// Resolved view: a scope tree plus per-node maps from uses to symbols
// and from declarations to symbols.
//
// The result is a pure value over the input AST. It holds no references
// to source text and never mutates the AST, so callers can share it
// freely (the LSP relies on this for snapshot-based reads).
//
// This is the *binder* phase in the Pyright sense: eager, single-pass,
// no type information. The type checker (Phase 2) consumes Resolved and
// populates Declared/Inferred on each Symbol.
//
// NOTE: Phase 1a establishes the data structures and the file-level
// binding pass (hoisted functions, args block, top-level assignments,
// identifier resolution). Per-construct scoping for function bodies,
// lambdas, loops, switch, and defer lands in Phase 1b.
func Resolve(file *rl.SourceFile) *Resolved {
	if file == nil {
		return nil
	}
	b := newBinder()
	b.bindFile(file)
	return b.resolved
}

// binder carries the mutable state required during a single resolution
// pass. It is not safe for concurrent use; the public Resolve function
// constructs a fresh one per call.
type binder struct {
	resolved *Resolved
	current  *Scope
	// builtins is the singleton set of runtime-provided function names.
	// We hold it once to avoid lookups through the heavier FunctionSet
	// API for every identifier we visit.
	builtins *rts.FunctionSet
}

func newBinder() *binder {
	builtinScope := &Scope{Kind: ScopeBuiltin, Symbols: map[string]*Symbol{}}
	return &binder{
		resolved: &Resolved{
			Builtin: builtinScope,
			Uses:    map[rl.Node]*Symbol{},
			Decls:   map[rl.Node]*Symbol{},
		},
		current:  builtinScope,
		builtins: rts.GetBuiltInFunctions(),
	}
}

// pushScope creates a child scope of the current scope and makes it
// active. Returns the new scope so callers can hold a reference (e.g.
// to associate it with an AST node later).
func (b *binder) pushScope(kind ScopeKind, owner rl.Node) *Scope {
	s := &Scope{
		Parent:  b.current,
		Kind:    kind,
		Owner:   owner,
		Symbols: map[string]*Symbol{},
	}
	b.current = s
	return s
}

func (b *binder) popScope() {
	if b.current != nil {
		b.current = b.current.Parent
	}
}

// declare introduces a symbol in the current scope. If a symbol with
// that name already exists locally, the existing one is returned -
// duplicate-declaration diagnostics are emitted later (Phase 1c), and
// for now we want a stable identity per name in a scope.
func (b *binder) declare(name string, kind SymbolKind, span rl.Span, node rl.Node) *Symbol {
	if existing, ok := b.current.Symbols[name]; ok {
		return existing
	}
	sym := &Symbol{
		Name:     name,
		Kind:     kind,
		DeclSpan: span,
		DefNode:  node,
		Scope:    b.current,
	}
	b.current.Symbols[name] = sym
	if node != nil {
		b.resolved.Decls[node] = sym
	}
	return sym
}

// ensureBuiltin returns (or lazily creates) the ambient Symbol for a
// runtime-provided name. The Symbol lives in the builtin scope and is
// reused for every reference, which gives LSP find-references a single
// identity per builtin.
func (b *binder) ensureBuiltin(name string) *Symbol {
	if sym, ok := b.resolved.Builtin.Symbols[name]; ok {
		return sym
	}
	sym := &Symbol{
		Name:  name,
		Kind:  SymBuiltin,
		Scope: b.resolved.Builtin,
	}
	b.resolved.Builtin.Symbols[name] = sym
	return sym
}

// resolveIdentifier records the use of `ident` and returns the Symbol
// it resolves to, or nil if the name is unknown. Unresolved names are
// not yet diagnosed - that lands in Phase 1c.
func (b *binder) resolveIdentifier(ident *rl.Identifier) *Symbol {
	if sym := b.current.Lookup(ident.Name); sym != nil {
		b.resolved.Uses[ident] = sym
		return sym
	}
	if b.builtins != nil && b.builtins.Contains(ident.Name) {
		sym := b.ensureBuiltin(ident.Name)
		b.resolved.Uses[ident] = sym
		return sym
	}
	return nil
}

// bindFile is the entry point. It opens the file scope, hoists named
// functions and arg-block declarations (both visible across the entire
// file body), then walks every top-level statement.
func (b *binder) bindFile(file *rl.SourceFile) {
	fileScope := b.pushScope(ScopeFile, nil)
	b.resolved.File = fileScope
	defer b.popScope()

	// Hoist top-level functions so calls earlier in the file can refer
	// to definitions later in the file. Existing checks already rely on
	// this order (see addUnknownFunctionHintsAST); we replicate it here
	// as the source of truth for name resolution.
	for _, stmt := range file.Stmts {
		if fn, ok := stmt.(*rl.FnDef); ok {
			b.declare(fn.Name, SymHoistedFn, fn.DefSpan, fn)
		}
	}

	// Arg-block declarations become ambient locals at file scope. The
	// runtime populates them from CLI flags before any user statement
	// runs, so by the time control reaches the body they exist.
	if file.Args != nil {
		for i := range file.Args.Decls {
			decl := &file.Args.Decls[i]
			b.declare(decl.Name, SymArg, decl.Span(), decl)
		}
	}

	// Walk statements. Per-construct scoping for function bodies,
	// loops, etc. lands in Phase 1b; for now visit() falls through to
	// children so identifier uses at file scope still resolve.
	for _, stmt := range file.Stmts {
		b.visit(stmt)
	}
}

// visit walks one AST node, dispatching on node kind. For nodes that
// introduce bindings or scopes the binder records the binding before
// descending. For everything else it just walks children.
func (b *binder) visit(n rl.Node) {
	if n == nil {
		return
	}
	switch v := n.(type) {
	case *rl.Identifier:
		b.resolveIdentifier(v)
	case *rl.Assign:
		b.visitAssign(v)
	default:
		// Generic descent. Phase 1b will replace many of these with
		// scope-aware cases (FnDef, Lambda, ForLoop, Switch, Defer).
		for _, child := range n.Children() {
			b.visit(child)
		}
	}
}

// visitAssign handles `x = expr`, `x, y = ...`, and compound-assign /
// incr-decr (which the converter desugars and marks via UpdateEnclosing).
//
// Order matters: we visit the RHS first so an expression like
// `x = x + 1` resolves the read-of-x to the pre-assign binding before
// the LHS creates or rebinds it. For compound assigns the LHS already
// exists, so this distinction doesn't matter, but doing it uniformly
// keeps the logic simple.
func (b *binder) visitAssign(a *rl.Assign) {
	for _, val := range a.Values {
		b.visit(val)
	}
	for _, target := range a.Targets {
		b.declareTarget(target, a)
	}
	if a.Catch != nil {
		b.visitCatch(a.Catch)
	}
}

// declareTarget introduces a binding for an assignment target. Plain
// identifiers create a SymLocal in the current scope on first sight.
// VarPath targets (a.b, xs[i]) don't introduce a new binding - they
// mutate an existing one - so we just visit them as expressions.
//
// Invalid LHS shapes (assigning to a literal, call, etc.) are caught
// by addInvalidAssignmentLHSErrorsAST and don't need to be re-diagnosed
// here.
func (b *binder) declareTarget(target rl.Node, _ *rl.Assign) {
	switch t := target.(type) {
	case *rl.Identifier:
		if sym := b.current.Lookup(t.Name); sym != nil {
			// Existing binding - this is a rebinding, not a new
			// declaration. Record the use so goto-def points at the
			// original decl.
			b.resolved.Uses[t] = sym
			return
		}
		b.declare(t.Name, SymLocal, t.Span(), t)
		// Also record this declaring identifier as its own use so the
		// LSP hover at the decl site finds the symbol.
		b.resolved.Uses[t] = b.current.Symbols[t.Name]
	case *rl.VarPath:
		// Mutation of an existing path. Visit it as an expression so
		// the root identifier (if any) resolves to its binding.
		b.visit(t)
	default:
		// Invalid LHS - leave the diagnostic to
		// addInvalidAssignmentLHSErrorsAST.
	}
}

// visitCatch walks the body of an error-catch block. The block does not
// introduce a binding in Phase 1a - the catch-variable narrowing rule
// lands with the rest of catch handling in Phase 4.
func (b *binder) visitCatch(c *rl.CatchBlock) {
	for _, stmt := range c.Stmts {
		b.visit(stmt)
	}
}
