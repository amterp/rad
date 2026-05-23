package check

import "github.com/amterp/rad/rts/rl"

// SymbolKind classifies what introduced a symbol.
type SymbolKind int

const (
	SymUnknown SymbolKind = iota
	// SymBuiltin is an ambient name supplied by the runtime (e.g. `print`).
	// These symbols are synthesized on first reference; they have no decl
	// span in the user's source.
	SymBuiltin
	// SymHoistedFn is a top-level named function. Visible across the file
	// regardless of textual order; this is how callers reference a function
	// defined further down.
	SymHoistedFn
	// SymArg is declared in the script-level `args:` block. These act as
	// ambient locals in the file scope: the runtime populates them from
	// CLI flags before the body executes.
	SymArg
	// SymCmdArg is declared inside a `cmd_block` args section. Visible only
	// inside that command's callback scope.
	SymCmdArg
	// SymParam is a function/lambda parameter.
	SymParam
	// SymLocal is anything else assigned in normal statement flow.
	SymLocal
	// SymLoopVar is the binding introduced by `for x in ...`.
	SymLoopVar
	// SymWith is the `with` context binding on a `for` loop.
	SymWith
	// SymRadField is a field name introduced inside a rad block.
	SymRadField
)

// ScopeKind tracks why a scope exists. Useful for diagnostics ("break
// outside loop") and later for narrowing rules that key off scope shape
// (e.g. loop-entry widening, lambda capture preservation).
type ScopeKind int

const (
	ScopeBuiltin  ScopeKind = iota // ambient runtime names
	ScopeFile                      // script body
	ScopeFunction                  // named function body
	ScopeLambda                    // anonymous function body
	ScopeLoop                      // for/while body
	ScopeBlock                     // switch case body, defer body, etc.
	ScopeListComp                  // list comprehension
	ScopeRadBlock                  // rad block body
	ScopeCmdBlock                  // cmd block body
)

// Symbol is the declaration record for a name in some scope.
//
// Each *use* of a name resolves to exactly one Symbol via Resolved.Uses.
// The Symbol is shared across all uses so later passes (type checker,
// goto-def, find-refs) can route through one identity per binding.
//
// Declared and Inferred type slots are populated by the type checker
// (Phase 2). They are nil for now.
type Symbol struct {
	Name     string
	Kind     SymbolKind
	DeclSpan rl.Span // location of the declaration in source; zero for builtins
	DefNode  rl.Node // the AST node that declared the symbol; nil for builtins
	Scope    *Scope  // scope this symbol lives in; nil for builtins

	// Declared is the static type from an explicit annotation, if any.
	// Stays nil for unannotated locals; the checker treats nil as
	// "annotation-free, infer from RHS / fall back to Dynamic".
	Declared rl.TypingT
	// Inferred is the type the checker computed (from RHS, narrowing,
	// etc.). Phase 2 will populate this; for now it stays nil.
	Inferred rl.TypingT
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
}
