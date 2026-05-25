package rts

import (
	"embed"
	"io/fs"
	"strings"
	"sync"
)

//go:embed embedded_funcs/*.md
var embeddedFuncDocs embed.FS

// FuncDocs is the lazily-loaded map from function name to its parsed
// FuncDoc. Populated on first access from the embedded
// `embedded_funcs/` directory.
//
// The embedded_funcs/ tree is the byte-for-byte mirror of
// docs/funcs/ that the runtime embeds at compile time.
// `tools/gen-funcs-go` keeps the two trees in sync; run
// `go generate ./rts` after editing under docs/funcs/.
// TestFuncDocsSourceMatchesEmbedded is the drift gate that catches
// contributors editing only one side.
var (
	funcDocs     map[string]*FuncDoc
	funcDocsOnce sync.Once
)

// GetFuncDoc returns the parsed FuncDoc for a built-in name, or
// nil if the function has no doc file. The LSP hover layer uses
// this to render a function's description alongside its signature.
func GetFuncDoc(name string) *FuncDoc {
	funcDocsOnce.Do(loadFuncDocs)
	if doc, ok := funcDocs[name]; ok {
		return doc
	}
	return nil
}

// FuncDocNames returns the sorted list of function names that
// have a doc entry. Used by completeness tests to compare against
// the registered builtin set.
func FuncDocNames() []string {
	funcDocsOnce.Do(loadFuncDocs)
	names := make([]string, 0, len(funcDocs))
	for n := range funcDocs {
		names = append(names, n)
	}
	return names
}

func loadFuncDocs() {
	funcDocs = make(map[string]*FuncDoc)
	entries, err := fs.ReadDir(embeddedFuncDocs, "embedded_funcs")
	if err != nil {
		// Embedded dir is empty or missing - that's fine; the
		// runtime still works, just without rich hover.
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		stem := strings.TrimSuffix(name, ".md")
		if !IsValidFuncDocStem(stem) {
			continue
		}
		src, err := fs.ReadFile(embeddedFuncDocs, "embedded_funcs/"+name)
		if err != nil {
			continue
		}
		doc, err := ParseFuncDoc(stem, string(src))
		if err != nil {
			// Malformed doc - skip so a bad file doesn't break
			// the runtime. The codegen test catches these at
			// build time; this branch is defense-in-depth.
			continue
		}
		funcDocs[stem] = doc
	}
}

// IsValidFuncDocStem matches the README's rule: file stems must be
// valid Rad identifiers ([a-z_][a-z0-9_]*). Filters out README.md
// and contributor notes that happen to land in this directory.
// Exported so the three codegen tools under tools/gen-funcs-* share
// the same rule without re-implementing it - if the rule ever
// changes (e.g. allowing uppercase), there's exactly one place to
// update.
func IsValidFuncDocStem(s string) bool {
	if s == "" {
		return false
	}
	first := s[0]
	if !(first == '_' || (first >= 'a' && first <= 'z')) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
