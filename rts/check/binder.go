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
// it resolves to, or nil if the name is unknown.
//
// Unresolved identifier uses are NOT surfaced as diagnostics here. The
// runtime emits a rich undefined-variable error with "did you mean"
// suggestions; emitting a static-time error of the same kind would
// short-circuit the runtime path that drives several test scenarios
// (defer behavior, suggestion strings, scoping snapshots). A future
// commit can broaden the static check to cover all uses, with the
// runtime-side test migration that goes with it.
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

// addIssue appends a structural binder finding to the resolved view.
// The checker layer converts these to user-facing Diagnostics.
func (b *binder) addIssue(span rl.Span, code rl.Error, msg string) {
	b.resolved.Issues = append(b.resolved.Issues, BindIssue{
		Span:    span,
		Code:    code,
		Message: msg,
	})
}

// bindFile is the entry point. It opens the file scope, hoists named
// functions and arg-block declarations (both visible across the entire
// file body), then walks every top-level statement.
func (b *binder) bindFile(file *rl.SourceFile) {
	fileScope := b.pushScope(ScopeFile, nil)
	b.resolved.File = fileScope
	defer b.popScope()

	// Pre-pass: declare everything that's visible from anywhere in the
	// file, so later visits (function bodies, cmd callbacks) resolve
	// against the complete set.
	//
	//   - Hoist top-level functions so calls earlier in the file can
	//     refer to definitions later.
	//   - Args-block declarations become ambient: the runtime populates
	//     them from CLI flags before the body runs.
	//   - Cmd-block args become ambient too: the runtime populates the
	//     invoked command's args into the file env before its callback
	//     runs, so a named-function callback can reference them.
	//     Multiple cmds with the same arg name share one symbol -
	//     they're mutually exclusive at runtime.
	for _, stmt := range file.Stmts {
		if fn, ok := stmt.(*rl.FnDef); ok {
			b.declare(fn.Name, SymHoistedFn, fn.DefSpan, fn)
		}
	}
	if file.Args != nil {
		for i := range file.Args.Decls {
			decl := &file.Args.Decls[i]
			b.declare(decl.Name, SymArg, decl.Span(), decl)
		}
	}
	for _, cmd := range file.Cmds {
		for i := range cmd.Decls {
			decl := &cmd.Decls[i]
			b.declare(decl.Name, SymCmdArg, decl.Span(), decl)
		}
	}

	// Visit arg-block default expressions now that every ambient name
	// is declared - this surfaces any undefined references inside
	// defaults without false positives from forward references.
	if file.Args != nil {
		for i := range file.Args.Decls {
			if file.Args.Decls[i].Default != nil {
				b.visit(file.Args.Decls[i].Default)
			}
		}
	}

	// Walk statements and cmd blocks.
	for _, stmt := range file.Stmts {
		b.visit(stmt)
	}
	for _, cmd := range file.Cmds {
		b.visit(cmd)
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
	case *rl.FnDef:
		b.visitFnDef(v)
	case *rl.Lambda:
		b.visitLambda(v)
	case *rl.ForLoop:
		b.visitForLoop(v)
	case *rl.WhileLoop:
		b.visitWhileLoop(v)
	case *rl.ListComp:
		b.visitListComp(v)
	case *rl.Switch:
		b.visitSwitch(v)
	case *rl.Defer:
		b.visitDefer(v)
	case *rl.CmdBlock:
		b.visitCmdBlock(v)
	default:
		// Generic descent for unhandled node kinds. Subsequent commits
		// in this phase replace more of these with scope-aware cases
		// (ForLoop, Switch, Defer, CmdBlock, RadBlock, ListComp).
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
		b.declareTarget(target, a.UpdateEnclosing)
	}
	if a.Catch != nil {
		b.visitCatch(a.Catch)
	}
}

// declareTarget introduces a binding for an assignment target.
//
//   - Plain '=' on an identifier introduces a fresh local in the current
//     scope. If a same-named binding exists in an enclosing scope, the
//     new local shadows it - this matches Python and is the Rad runtime
//     behavior.
//   - Compound assigns ('+=', '++', etc.) and unpacking-with-rebind set
//     UpdateEnclosing on the AST node. In that mode we resolve up the
//     scope chain and treat the target as a *use* of the existing
//     binding, not a new declaration. Without an existing binding the
//     compound op would have nothing to operate on.
//   - VarPath targets (a.b, xs[i]) mutate an existing path's contents
//     and don't introduce a new binding. We just visit them as
//     expressions so the root identifier resolves.
//
// Invalid LHS shapes (assigning to a literal, call, etc.) are caught
// by addInvalidAssignmentLHSErrorsAST and don't need to be re-diagnosed
// here.
func (b *binder) declareTarget(target rl.Node, updateEnclosing bool) {
	switch t := target.(type) {
	case *rl.Identifier:
		if updateEnclosing {
			if sym := b.current.Lookup(t.Name); sym != nil {
				b.resolved.Uses[t] = sym
				return
			}
			// Compound-assign needs an existing binding to operate on,
			// but we defer the static diagnostic to the same future
			// commit that broadens the undefined-variable check. The
			// runtime will emit a clear error when it tries to execute.
			return
		}
		// Plain '=' shadows any enclosing-scope binding with the same name.
		sym := b.declare(t.Name, SymLocal, t.Span(), t)
		// Record this declaring identifier as its own use so an LSP
		// hover at the decl site finds the symbol.
		b.resolved.Uses[t] = sym
	case *rl.VarPath:
		// Mutation of an existing path. Visit it as an expression so
		// the root identifier (if any) resolves to its binding.
		b.visit(t)
	default:
		// Invalid LHS - leave the diagnostic to
		// addInvalidAssignmentLHSErrorsAST.
	}
}

// visitFnDef binds a named function definition. The function name was
// already declared at file scope by the hoisting pass (or, for nested
// functions, by the enclosing visit before this point), so the body
// can reference itself for recursion via normal lookup.
//
// Parameter default expressions evaluate in the *enclosing* scope, not
// the function scope. This matches what callers intuitively expect
// (defaults reference state visible at definition time) and avoids the
// surprise of one parameter's default seeing a later parameter's name.
func (b *binder) visitFnDef(fn *rl.FnDef) {
	// Nested function definitions are not hoisted; declare them at
	// the current scope at point of visit. Top-level FnDefs are
	// already in the file scope from bindFile's pre-pass and
	// declare() returns the existing symbol unchanged in that case.
	if fn.Name != "" {
		b.declare(fn.Name, SymHoistedFn, fn.DefSpan, fn)
	}
	b.bindFnLike(fn.Typing, fn.Body, ScopeFunction, fn)
}

// visitLambda binds an anonymous function. Same shape as FnDef minus
// the name-introduction step.
func (b *binder) visitLambda(l *rl.Lambda) {
	b.bindFnLike(l.Typing, l.Body, ScopeLambda, l)
}

// bindFnLike opens a function-like scope, declares parameters in it,
// and walks the body. Param defaults are visited *before* the scope
// opens so they bind to names in the enclosing scope.
func (b *binder) bindFnLike(typing *rl.TypingFnT, body []rl.Node, kind ScopeKind, owner rl.Node) {
	// Step 1: visit param defaults in the enclosing scope.
	if typing != nil {
		for i := range typing.Params {
			p := &typing.Params[i]
			if p.DefaultAST != nil && p.DefaultAST.Node != nil {
				b.visit(p.DefaultAST.Node)
			}
		}
	}

	// Step 2: open the function scope and declare params.
	b.pushScope(kind, owner)
	defer b.popScope()

	if typing != nil {
		for i := range typing.Params {
			p := &typing.Params[i]
			if p.Name == "" {
				continue
			}
			// Same-scope collisions are the actual error case here -
			// shadowing an outer-scope binding via a parameter is a
			// legitimate, common pattern. Only flag when two params
			// in the *same* parameter list share a name.
			if _, dup := b.current.Symbols[p.Name]; dup {
				b.addIssue(owner.Span(), rl.ErrDuplicateParameter,
					"Duplicate parameter '"+p.Name+"'")
				continue
			}
			b.declare(p.Name, SymParam, owner.Span(), nil)
		}
	}

	// Step 3: walk the body in the new scope.
	for _, stmt := range body {
		b.visit(stmt)
	}
}

// visitForLoop binds a `for vars in iter [with ctx]:` loop.
//
// In Rad, loop bodies do NOT open a new environment - the interpreter
// writes loop variables and any body-locals via SetVar on the
// enclosing env. As a consequence, `for i in range(3): pass; print(i)`
// is valid Rad: i is 2 after the loop. The binder mirrors that by
// declaring loop vars (and the optional 'with' context) in the
// current scope, not a synthetic loop scope.
//
// The iterable expression is visited first - it computes the value to
// iterate, which by definition cannot reference the loop var.
func (b *binder) visitForLoop(f *rl.ForLoop) {
	b.visit(f.Iter)
	for _, name := range f.Vars {
		if name == "" {
			continue
		}
		b.declare(name, SymLoopVar, f.Span(), f)
	}
	if f.Context != nil && *f.Context != "" {
		b.declare(*f.Context, SymWith, f.Span(), f)
	}
	for _, stmt := range f.Body {
		b.visit(stmt)
	}
}

// visitWhileLoop binds a `while [cond]:` loop. The header declares no
// names; the body shares the enclosing env (matches the interpreter's
// runBlock behavior). The condition is visited as a normal expression
// in that same scope.
func (b *binder) visitWhileLoop(w *rl.WhileLoop) {
	if w.Condition != nil {
		b.visit(w.Condition)
	}
	for _, stmt := range w.Body {
		b.visit(stmt)
	}
}

// visitListComp binds a `[expr for vars in iter if cond]` comprehension.
//
// Same scoping story as ForLoop - the runtime evaluates a list-comp
// via executeForLoop, which writes the iteration variable into the
// enclosing env. The vars and 'with' context become locals in the
// current scope; the iterable is visited first.
func (b *binder) visitListComp(c *rl.ListComp) {
	b.visit(c.Iter)
	for _, name := range c.Vars {
		if name == "" {
			continue
		}
		b.declare(name, SymLoopVar, c.Span(), c)
	}
	if c.Context != nil && *c.Context != "" {
		b.declare(*c.Context, SymWith, c.Span(), c)
	}
	if c.Condition != nil {
		b.visit(c.Condition)
	}
	b.visit(c.Expr)
}

// visitSwitch binds a switch statement. Discriminant, match keys, and
// case bodies all evaluate in the same scope - Rad has no per-case
// environment, so a local declared in one case body is visible after
// the switch and to subsequent cases (control flow permitting).
func (b *binder) visitSwitch(s *rl.Switch) {
	b.visit(s.Discriminant)
	for _, c := range s.Cases {
		for _, k := range c.Keys {
			b.visit(k)
		}
		b.visitSwitchAlt(c.Alt)
	}
	if s.Default != nil {
		b.visitSwitchAlt(s.Default.Alt)
	}
}

// visitSwitchAlt visits one case's right-hand side. The case-block
// form just walks its statements in the enclosing scope.
func (b *binder) visitSwitchAlt(alt rl.Node) {
	switch a := alt.(type) {
	case *rl.SwitchCaseBlock:
		for _, stmt := range a.Stmts {
			b.visit(stmt)
		}
	case *rl.SwitchCaseExpr:
		for _, v := range a.Values {
			b.visit(v)
		}
	default:
		b.visit(alt)
	}
}

// visitDefer walks a `defer:` / `errdefer:` block's body. The deferred
// code runs later in the enclosing function's env (the interpreter
// uses runBlock, not a child env), so anything declared inside is
// visible to the rest of the function after the defer is registered -
// matching how the runtime sees it.
func (b *binder) visitDefer(d *rl.Defer) {
	for _, stmt := range d.Body {
		b.visit(stmt)
	}
}

// visitCmdBlock walks a cmd_block's defaults and inline-lambda
// callback. The cmd's args were already declared at file scope by
// bindFile's pre-pass (the runtime populates them as globals before
// the callback runs), so this routine doesn't declare them again -
// it just visits the default expressions (which may reference other
// arg-block / cmd-arg names) and the lambda callback.
//
// Identifier-style callbacks (`calls handler`) are not resolved here
// because CmdCallback stores the name as a plain string with no AST
// identifier node. addUnknownCommandCallbackWarnings handles the
// "is the target visible at file scope?" question.
func (b *binder) visitCmdBlock(c *rl.CmdBlock) {
	for i := range c.Decls {
		if c.Decls[i].Default != nil {
			b.visit(c.Decls[i].Default)
		}
	}
	cb := c.Callback
	if cb.IsLambda && cb.Lambda != nil {
		b.visitLambda(cb.Lambda)
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
