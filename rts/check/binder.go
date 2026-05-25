package check

import (
	"strconv"

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
			Builtin:      builtinScope,
			Uses:         map[rl.Node]*Symbol{},
			Decls:        map[rl.Node]*Symbol{},
			ForLoopVars:  map[*rl.ForLoop][]*Symbol{},
			ParamSymbols: map[rl.Node][]*Symbol{},
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
// On miss, emits a static Error-severity diagnostic with a did-you-
// mean suggestion drawn from the same Levenshtein-threshold logic
// the runtime uses (findSimilarNames mirrors core/env.go's
// FindSimilarVars). This gates broken scripts at check time and
// gives LSP an actionable signal for quick-fix code actions.
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
	// Internal `_rad_*` builtins are registered in FnSignaturesByName
	// but deliberately excluded from FunctionSet (and from completion
	// / hover). They're still legal calls - the runtime needs them
	// for the embedded `check`, `explain`, etc. scripts - so accept
	// them silently here. User scripts that hit a `_rad_*` typo will
	// still get the diagnostic via the lookup below.
	if _, ok := rts.FnSignaturesByName[ident.Name]; ok {
		sym := b.ensureBuiltin(ident.Name)
		b.resolved.Uses[ident] = sym
		return sym
	}
	b.emitUndefinedIdentifier(ident)
	return nil
}

// emitUndefinedIdentifier records the diagnostic and attaches a
// "did you mean: X" suggestion when one of the visible names is
// close enough to be worth showing. The threshold matches the
// runtime's FindSimilarVars so the static and runtime suggestion
// sets line up.
func (b *binder) emitUndefinedIdentifier(ident *rl.Identifier) {
	var builtinNames map[string]bool
	if b.builtins != nil {
		builtinNames = b.builtins.Names()
	}
	similar := findSimilarNames(b.current, builtinNames, ident.Name, 3)
	suggestion := formatDidYouMean(similar)
	b.resolved.Issues = append(b.resolved.Issues, BindIssue{
		Span:       ident.Span(),
		Severity:   IssueError,
		Code:       rl.ErrUndefinedVariable,
		Message:    "Undefined identifier '" + ident.Name + "'",
		Suggestion: suggestion,
	})
}

// addIssue appends a structural binder finding to the resolved view.
// The checker layer converts these to user-facing Diagnostics, using
// the recorded severity rather than imposing one of its own.
func (b *binder) addIssue(span rl.Span, severity IssueSeverity, code rl.Error, msg string) {
	b.resolved.Issues = append(b.resolved.Issues, BindIssue{
		Span:     span,
		Severity: severity,
		Code:     code,
		Message:  msg,
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
	hoisted := make(map[string]rl.Span)
	for _, stmt := range file.Stmts {
		if fn, ok := stmt.(*rl.FnDef); ok {
			b.declare(fn.Name, SymHoistedFn, fn.NameSpan, fn)
			hoisted[fn.Name] = fn.NameSpan
		}
	}
	if file.Args != nil {
		for i := range file.Args.Decls {
			decl := &file.Args.Decls[i]
			// A hoisted function and an args-block decl claiming the
			// same name is a contradiction: the function definition
			// wins the symbol slot, but the user intended both. Emit
			// a focused issue pointing at the fn's def span (matches
			// the long-standing diagnostic message and location).
			if fnSpan, clash := hoisted[decl.Name]; clash {
				b.addIssue(fnSpan, IssueError, rl.ErrHoistedFunctionShadowsArgument,
					"Hoisted function '"+decl.Name+"' shadows an argument with the same name")
			}
			sym := b.declare(decl.Name, SymArg, decl.NameSpan, decl)
			if decl.Typing != nil {
				sym.Declared = decl.Typing
			}
		}
	}
	for _, cmd := range file.Cmds {
		for i := range cmd.Decls {
			decl := &cmd.Decls[i]
			sym := b.declare(decl.Name, SymCmdArg, decl.NameSpan, decl)
			// Multiple cmds may declare the same arg name; declare()
			// returns the existing symbol on collision. Only plant the
			// type when fresh so the first cmd's annotation wins; the
			// runtime enforces these are mutually exclusive at invoke
			// time.
			if decl.Typing != nil && sym.Declared == nil {
				sym.Declared = decl.Typing
			}
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
	case *rl.Shell:
		b.visitShell(v)
	case *rl.RadBlock:
		b.visitRadBlock(v)
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
	// Lambda-aware ordering: when value[i] is a Lambda, declare
	// target[i] FIRST so the lambda body can reference the new
	// binding (let-rec semantics for `f = fn(...) ... f(...)`).
	// For non-lambda RHSes we keep the old order - visit RHS
	// first so `x = x + 1` resolves the read-of-x to the
	// pre-assign binding, matching Python and Rad's runtime.
	//
	// The ordering is per pair, not all-or-nothing. In an unpack
	// like `f, g = fn() ..., fn() ...` (rare) each lambda gets
	// its target declared early; in `x, f = 5, fn() ... x ...`
	// only f's target moves up.
	for i, val := range a.Values {
		if _, isLambda := val.(*rl.Lambda); !isLambda {
			b.visit(val)
			continue
		}
		if i < len(a.Targets) {
			b.declareTarget(a.Targets[i], a.UpdateEnclosing)
		}
		b.visit(val)
	}
	for i, target := range a.Targets {
		// Skip targets already declared above (their value was a
		// Lambda and we did the early declaration). For unpack
		// shapes where len(Targets) > len(Values), every extra
		// target goes through the normal declare path.
		if i < len(a.Values) {
			if _, isLambda := a.Values[i].(*rl.Lambda); isLambda {
				continue
			}
		}
		b.declareTarget(target, a.UpdateEnclosing)
	}
	// Typed local: `x: int = 5`. The converter only attaches
	// DeclaredType on single-target assigns today, so we plant the
	// annotation on the first target's symbol if it's a fresh local
	// the binder just declared.
	//
	// Re-declaring an already-typed binding (`x: int = 5; x: str = ...`)
	// is rejected: the declared type is part of the binding's contract
	// and must not be retroactively rewritten - doing so would poison
	// every preceding assignment. Same-type redeclarations are also
	// flagged so the diagnostic catches "I forgot I declared this" bugs
	// uniformly, not just type-changing ones. The runtime currently
	// allows free reassignment; we may revisit redecl semantics if we
	// adopt full Python-style rebinding.
	if a.DeclaredType != nil && len(a.Targets) > 0 {
		if ident, ok := a.Targets[0].(*rl.Identifier); ok {
			if sym, ok := b.resolved.Uses[ident]; ok && sym != nil {
				if sym.Declared != nil {
					b.addIssue(ident.Span(), IssueError, rl.ErrDuplicateTypedDeclaration,
						"Cannot re-declare '"+ident.Name+"' with a type annotation (originally declared on line "+
							strconv.Itoa(sym.DeclSpan.StartLine())+"); drop ': "+(*a.DeclaredType).Name()+"' to reassign")
				} else {
					sym.Declared = *a.DeclaredType
				}
			}
		}
	}
	if a.Catch != nil {
		b.visitCatch(a.Catch)
	}
}

// visitRadBlock walks a rad block, treating field-name identifiers
// as declarations rather than references. Field names come from the
// data source (e.g. CSV columns, JSON keys) and aren't variables in
// the script's lexical scope - they're injected into the rad-block
// body by the runtime. Without this special case, every `fields
// Name age` would emit RAD20028 for each field name.
//
// Field names declared in `fields`, `sort`, and field-modifier
// statements all flow through the same path. Other rad-block stmts
// (RadIf, RadOption, etc.) descend normally so any user code inside
// the block (e.g. expressions in an option) gets the usual check.
func (b *binder) visitRadBlock(rb *rl.RadBlock) {
	if rb.Source != nil {
		b.visit(rb.Source)
	}
	for _, stmt := range rb.Stmts {
		switch s := stmt.(type) {
		case *rl.RadField:
			for _, ident := range s.Identifiers {
				if id, ok := ident.(*rl.Identifier); ok {
					b.declareLocal(id, false)
				}
			}
		case *rl.RadFieldMod:
			// Fields[] are target field-name identifiers; treat as
			// uses but resolve through declared field names rather
			// than emitting RAD20028 on a miss. For commit 17 the
			// simplest is to skip resolution on these.
			for _, arg := range s.Args {
				b.visit(arg)
			}
		default:
			b.visit(stmt)
		}
	}
}

// declareLocal introduces a fresh local for an identifier, dual-
// registering it as both decl and use so hover/find-refs work. Used
// for synthesized declarations (rad-block field names, etc.) where
// the runtime injects the binding implicitly.
func (b *binder) declareLocal(ident *rl.Identifier, updateEnclosing bool) {
	sym := b.declare(ident.Name, SymLocal, ident.Span(), ident)
	b.resolved.Uses[ident] = sym
}

// visitShell handles shell statements like `code = $\"echo hi\"`.
// The targets on the LHS are declarations (an exit-code variable,
// stdout/stderr captures), not references - same shape as Assign.
// Without explicit handling the generic descent would walk the
// target identifiers and resolveIdentifier would flag them as
// undefined. Visit the command expression and any catch block as
// normal.
func (b *binder) visitShell(s *rl.Shell) {
	if s.Cmd != nil {
		b.visit(s.Cmd)
	}
	for _, target := range s.Targets {
		b.declareTarget(target, false)
	}
	if s.Catch != nil {
		b.visitCatch(s.Catch)
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
		sym := b.declare(fn.Name, SymHoistedFn, fn.NameSpan, fn)
		// Plant the function's structural signature on the symbol so
		// references-by-name (`process(my_callback)`) synth to a
		// TypingFnT instead of Dynamic. With Declared set, the type
		// checker's seed pass copies it into SymbolTypes, and
		// IsAssignableFrom on the receiving param drives a real
		// contravariant-params + covariant-return shape comparison.
		// Without this plant, the structural check would never bite
		// because synthIdentifier would just see Dynamic.
		if sym != nil && fn.Typing != nil {
			sym.Declared = fn.Typing
		}
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
		paramSyms := make([]*Symbol, 0, len(typing.Params))
		for i := range typing.Params {
			p := &typing.Params[i]
			if p.Name == "" {
				continue
			}
			// Prefer the param name span when the typing resolver
			// captured it; fall back to the owning fn span for
			// synthesised params with no source token. The name span
			// is what lets LSP rename / find-refs / goto-def land on
			// the parameter name rather than the entire function.
			declSpan := p.NameSpan
			if declSpan == (rl.Span{}) {
				declSpan = owner.Span()
			}
			// Same-scope collisions are the actual error case here -
			// shadowing an outer-scope binding via a parameter is a
			// legitimate, common pattern. Only flag when two params
			// in the *same* parameter list share a name.
			if _, dup := b.current.Symbols[p.Name]; dup {
				b.addIssue(declSpan, IssueError, rl.ErrDuplicateParameter,
					"Duplicate parameter '"+p.Name+"'")
				continue
			}
			sym := b.declare(p.Name, SymParam, declSpan, nil)
			// Param type annotation acts like a typed local'\''s
			// declared type: it'\''s an immutable contract subsequent
			// reads / assigns must respect. Unannotated params stay
			// at Declared == nil (effectively Dynamic).
			if p.Type != nil {
				sym.Declared = *p.Type
			}
			paramSyms = append(paramSyms, sym)
		}
		if len(paramSyms) > 0 {
			b.resolved.ParamSymbols[owner] = paramSyms
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
	// Record the per-var symbol list. We can't use Decls for this -
	// every var keys on the same ForLoop node, so the map would only
	// retain one. Type checking the `for k, v in xs:` shape needs
	// both symbols, so we keep them in source order here.
	vars := make([]*Symbol, 0, len(f.Vars))
	for i, name := range f.Vars {
		if name == "" {
			continue
		}
		// VarSpans is parallel to Vars; populated by the converter.
		// Defensive fallback to f.Span() when a caller built a ForLoop
		// without per-var spans - we keep working but lose rename
		// precision. Current callers always populate VarSpans.
		declSpan := f.Span()
		if i < len(f.VarSpans) && f.VarSpans[i] != (rl.Span{}) {
			declSpan = f.VarSpans[i]
		}
		vars = append(vars, b.declare(name, SymLoopVar, declSpan, f))
	}
	if len(vars) > 0 {
		b.resolved.ForLoopVars[f] = vars
	}
	if f.Context != nil && *f.Context != "" {
		ctxSpan := f.Span()
		if f.ContextSpan != (rl.Span{}) {
			ctxSpan = f.ContextSpan
		}
		b.declare(*f.Context, SymWith, ctxSpan, f)
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
