package testing

import (
	"testing"

	"github.com/amterp/rad/core"
	"github.com/stretchr/testify/assert"
)

func TestSymbolTable_DefineAndLookup(t *testing.T) {
	st := core.NewSymbolTable()

	// Define a symbol
	sym := st.Define("foo", nil)
	assert.Equal(t, "foo", sym.Name)
	assert.Equal(t, 0, sym.UsageCount)

	// Lookup should find it
	found := st.Lookup("foo")
	assert.NotNil(t, found)
	assert.Equal(t, "foo", found.Name)

	// Lookup unknown should return nil
	notFound := st.Lookup("bar")
	assert.Nil(t, notFound)
}

func TestSymbolTable_ScopeNesting(t *testing.T) {
	st := core.NewSymbolTable()

	// Define in global scope
	st.Define("global_var", nil)

	// Enter child scope
	st.EnterScope("function")
	st.Define("local_var", nil)

	// Both should be visible
	assert.NotNil(t, st.Lookup("global_var"))
	assert.NotNil(t, st.Lookup("local_var"))

	// Exit scope
	st.ExitScope()

	// Global still visible, local not
	assert.NotNil(t, st.Lookup("global_var"))
	assert.Nil(t, st.Lookup("local_var"))
}

func TestSymbolTable_MarkUsed(t *testing.T) {
	st := core.NewSymbolTable()

	st.Define("foo", nil)
	assert.Equal(t, 0, st.Lookup("foo").UsageCount)

	st.MarkUsed("foo")
	assert.Equal(t, 1, st.Lookup("foo").UsageCount)

	st.MarkUsed("foo")
	assert.Equal(t, 2, st.Lookup("foo").UsageCount)

	// Marking unknown symbol is a no-op
	st.MarkUsed("unknown")
}

func TestSymbolTable_UnusedSymbols(t *testing.T) {
	st := core.NewSymbolTable()

	st.Define("used", nil)
	st.Define("unused", nil)
	st.Define("_ignored", nil) // underscore prefix means intentionally unused
	st.DefineBuiltIn("print")

	st.MarkUsed("used")

	unused := st.Current.UnusedSymbols()
	assert.Len(t, unused, 1)
	assert.Equal(t, "unused", unused[0].Name)
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"kitten", "sitting", 3},
		{"foobar", "foobaz", 1},
	}

	for _, tc := range tests {
		result := core.Levenshtein(tc.a, tc.b)
		assert.Equal(t, tc.expected, result, "Levenshtein(%q, %q)", tc.a, tc.b)
	}
}

func TestSymbolTable_FindSimilar(t *testing.T) {
	st := core.NewSymbolTable()

	st.Define("username", nil)
	st.Define("user_name", nil)
	st.Define("password", nil)
	st.Define("email", nil)

	// Should find similar names
	similar := st.FindSimilar("usernme", 3)
	assert.Contains(t, similar, "username")

	// Should not find exact match
	similar = st.FindSimilar("password", 3)
	assert.NotContains(t, similar, "password")

	// Should return empty for very different names
	similar = st.FindSimilar("xyz123", 3)
	assert.Empty(t, similar)
}

func TestSymbolTable_LookupLocal(t *testing.T) {
	st := core.NewSymbolTable()

	st.Define("global", nil)

	st.EnterScope("child")
	st.Define("local", nil)

	// LookupLocal only finds in current scope
	assert.NotNil(t, st.LookupLocal("local"))
	assert.Nil(t, st.LookupLocal("global"))

	// Regular Lookup finds both
	assert.NotNil(t, st.Lookup("local"))
	assert.NotNil(t, st.Lookup("global"))
}
