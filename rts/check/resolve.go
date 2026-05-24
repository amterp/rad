package check

import "github.com/amterp/rad/rts/rl"

// SymbolKind classifies what introduced a symbol.
type SymbolKind int

const (
	// SymBuiltin is an ambient name supplied by the runtime (e.g. `print`).
	// These symbols are synthesized on first reference; they have no decl
	// span in the user's source.
	SymBuiltin SymbolKind = iota + 1
	// SymHoistedFn is a top-level named function. Visible across the file
	// regardless of textual order; this is how callers reference a function
	// defined further down.
	SymHoistedFn
	// SymArg is declared in the script-level `args:` block. These act as
	// ambient locals in the file scope: the runtime populates them from
	// CLI flags before the body executes.
	SymArg
	// SymCmdArg is declared inside a `cmd_block` args section. The
	// binding lives in the enclosing (file) scope because the runtime
	// populates it as a global before the command's callback runs;
	// the kind distinguishes it from SymLocal so LSP hover and
	// goto-def can point users at the cmd block's decl.
	SymCmdArg
	// SymParam is a function/lambda parameter.
	SymParam
	// SymLocal is anything else assigned in normal statement flow.
	SymLocal
	// SymLoopVar is the binding introduced by `for x in ...`.
	SymLoopVar
	// SymWith is the `with` context binding on a `for` loop.
	SymWith
)

// ScopeKind tracks why a scope exists.
//
// Rad opens new scopes only at function-like boundaries - named
// functions, lambdas, and the implicit top-level file scope. Loops,
// switch cases, defer bodies, list comprehensions, and cmd blocks do
// NOT open a scope; they share the enclosing env (matching the
// runtime's runBlock behavior, where loop variables and body-locals
// persist after the construct ends).
type ScopeKind int

const (
	ScopeBuiltin  ScopeKind = iota // ambient runtime names
	ScopeFile                      // script body
	ScopeFunction                  // named function body
	ScopeLambda                    // anonymous function body
)

// Symbol is the declaration record for a name in some scope.
//
// Each *use* of a name resolves to exactly one Symbol via Resolved.Uses.
// The Symbol is shared across all uses so later passes (type checker,
// goto-def, find-refs) can route through one identity per binding.
//
// Note: parameter and loop-variable spans currently fall back to the
// owner's span (the FnDef/Lambda or ForLoop) because TypingFnParam
// and ForLoop.Vars don't carry per-name spans on the AST. LSP
// goto-def on a param will point at the function keyword until the
// AST gains per-name span info.
type Symbol struct {
	Name     string
	Kind     SymbolKind
	DeclSpan rl.Span // location of the declaration in source; zero for builtins
	DefNode  rl.Node // the AST node that declared the symbol; nil for builtins
	Scope    *Scope  // scope this symbol lives in; the builtin scope for SymBuiltin
}

// Scope is a lexical name -> Symbol table chained to its parent. Lookup
// walks the parent chain; declaration is local.
type Scope struct {
	Parent  *Scope
	Kind    ScopeKind
	Owner   rl.Node // node that introduced the scope; nil for file/builtin
	Symbols map[string]*Symbol
}

// Lookup walks this scope and its parents for a symbol named `name`.
// Returns nil if the name is not in scope.
func (s *Scope) Lookup(name string) *Symbol {
	for cur := s; cur != nil; cur = cur.Parent {
		if sym, ok := cur.Symbols[name]; ok {
			return sym
		}
	}
	return nil
}

// Resolved is the output of name resolution: a scope tree plus indexes
// from AST nodes to the symbols they refer to or declare.
//
// All maps key on AST node pointer identity, so a Resolved is safe to
// pass to readers that hold the same AST.
type Resolved struct {
	// Builtin is the ambient scope holding lazily-synthesized symbols
	// for runtime-provided names.
	Builtin *Scope
	// File is the top-level script scope.
	File *Scope
	// Uses maps an identifier-reference node to the Symbol it resolves to.
	// Identifiers that fail to resolve are absent from this map.
	Uses map[rl.Node]*Symbol
	// Decls maps a declaring node to the Symbol it introduced. Useful for
	// goto-def (jump to decl span) and for hover (show declared type).
	Decls map[rl.Node]*Symbol
	// Issues are problems the binder detected during resolution
	// (undefined references, duplicate parameters, etc.). Callers
	// convert these to whatever diagnostic shape they need; the binder
	// stays src-free.
	Issues []BindIssue
}

// IssueSeverity is the binder's severity classification. Kept here
// rather than on Diagnostic so resolve.go stays src-free.
type IssueSeverity int

const (
	IssueError IssueSeverity = iota
	IssueWarning
	IssueHint
)

// BindIssue is a problem detected during name resolution. It carries
// just the bare information - span, severity, message, error code - so
// that the binder doesn't depend on the source text or the wider
// Diagnostic type. Callers turn each issue into their preferred
// diagnostic shape (e.g. check.Diagnostic with src-derived range info).
//
// Suggestion is an optional "= help: ..." line. When non-empty, the
// rendered diagnostic shows it after the source-context block, the
// same way the runtime renders its `emitErrorWithHint` calls. Used
// to give the static check parity with runtime diagnostics that
// already provide actionable follow-up (the v0.9 `+` migration hint
// is the canonical example).
type BindIssue struct {
	Span       rl.Span
	Severity   IssueSeverity
	Code       rl.Error
	Message    string
	Suggestion string
}
