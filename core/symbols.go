package core

import (
	"sort"
	"strings"
)

// Symbol represents a variable definition in the symbol table.
type Symbol struct {
	Name       string
	Span       *Span // Where the symbol was defined
	UsageCount int   // How many times it's been read
	IsArg      bool  // True if this is a script argument
	IsBuiltIn  bool  // True if this is a built-in (print, len, etc.)
}

// Scope represents a lexical scope in the program.
type Scope struct {
	Parent  *Scope
	Symbols map[string]*Symbol
	Name    string // For debugging: "global", "function:foo", "for-loop", etc.
}

// SymbolTable tracks variable definitions and usages across scopes.
type SymbolTable struct {
	Global  *Scope
	Current *Scope
}

// NewSymbolTable creates a new symbol table with a global scope.
func NewSymbolTable() *SymbolTable {
	global := &Scope{
		Parent:  nil,
		Symbols: make(map[string]*Symbol),
		Name:    "global",
	}
	return &SymbolTable{
		Global:  global,
		Current: global,
	}
}

// EnterScope creates a new child scope and makes it current.
func (st *SymbolTable) EnterScope(name string) {
	child := &Scope{
		Parent:  st.Current,
		Symbols: make(map[string]*Symbol),
		Name:    name,
	}
	st.Current = child
}

// ExitScope returns to the parent scope.
// Returns the scope that was exited (for checking unused variables).
func (st *SymbolTable) ExitScope() *Scope {
	exited := st.Current
	if st.Current.Parent != nil {
		st.Current = st.Current.Parent
	}
	return exited
}

// Define adds a symbol to the current scope.
func (st *SymbolTable) Define(name string, span *Span) *Symbol {
	sym := &Symbol{
		Name:       name,
		Span:       span,
		UsageCount: 0,
		IsArg:      false,
		IsBuiltIn:  false,
	}
	st.Current.Symbols[name] = sym
	return sym
}

// DefineArg adds an argument symbol to the current scope.
func (st *SymbolTable) DefineArg(name string, span *Span) *Symbol {
	sym := st.Define(name, span)
	sym.IsArg = true
	return sym
}

// DefineBuiltIn adds a built-in symbol to the current scope.
func (st *SymbolTable) DefineBuiltIn(name string) *Symbol {
	sym := &Symbol{
		Name:       name,
		Span:       nil, // Built-ins have no source location
		UsageCount: 0,
		IsArg:      false,
		IsBuiltIn:  true,
	}
	st.Current.Symbols[name] = sym
	return sym
}

// Lookup finds a symbol by name, searching from current scope up to global.
// Returns nil if not found.
func (st *SymbolTable) Lookup(name string) *Symbol {
	return st.Current.Lookup(name)
}

// Lookup finds a symbol in this scope or any enclosing scope.
func (s *Scope) Lookup(name string) *Symbol {
	if sym, ok := s.Symbols[name]; ok {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

// LookupLocal finds a symbol only in the current scope (not parents).
func (st *SymbolTable) LookupLocal(name string) *Symbol {
	if sym, ok := st.Current.Symbols[name]; ok {
		return sym
	}
	return nil
}

// MarkUsed increments the usage count for a symbol.
func (st *SymbolTable) MarkUsed(name string) {
	if sym := st.Lookup(name); sym != nil {
		sym.UsageCount++
	}
}

// UnusedSymbols returns all symbols in the given scope that were never used.
// Excludes symbols starting with '_' (intentionally unused) and built-ins.
func (s *Scope) UnusedSymbols() []*Symbol {
	var unused []*Symbol
	for _, sym := range s.Symbols {
		if sym.UsageCount == 0 && !sym.IsBuiltIn && !strings.HasPrefix(sym.Name, "_") {
			unused = append(unused, sym)
		}
	}
	// Sort for deterministic output
	sort.Slice(unused, func(i, j int) bool {
		return unused[i].Name < unused[j].Name
	})
	return unused
}

// AllSymbolNames returns all symbol names visible from the current scope.
// Used for "did you mean?" suggestions.
func (st *SymbolTable) AllSymbolNames() []string {
	seen := make(map[string]bool)
	var names []string

	for scope := st.Current; scope != nil; scope = scope.Parent {
		for name := range scope.Symbols {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}

	sort.Strings(names)
	return names
}

// Levenshtein calculates the Levenshtein distance between two strings.
// Used for "did you mean?" suggestions.
func Levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// FindSimilar finds symbols with names similar to the given name.
// Returns at most maxResults symbols, sorted by similarity.
func (st *SymbolTable) FindSimilar(name string, maxResults int) []string {
	type candidate struct {
		name     string
		distance int
	}

	var candidates []candidate
	allNames := st.AllSymbolNames()

	// Only suggest names within a reasonable edit distance
	maxDistance := len(name)/2 + 1
	if maxDistance < 2 {
		maxDistance = 2
	}

	for _, n := range allNames {
		dist := Levenshtein(name, n)
		if dist <= maxDistance && dist > 0 {
			candidates = append(candidates, candidate{n, dist})
		}
	}

	// Sort by distance
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].name < candidates[j].name
	})

	// Return top results
	result := make([]string, 0, maxResults)
	for i := 0; i < len(candidates) && i < maxResults; i++ {
		result = append(result, candidates[i].name)
	}
	return result
}
