package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

// DumpForSnapshot produces a deterministic textual summary of the
// checker's output suitable for snapshot testing. It records, in
// source order:
//
//   - Each Identifier reference, its source position, and the type
//     the checker synthesized for it at that point. This is the
//     load-bearing signal for narrowing: the same identifier name
//     can have different types at different positions, and the
//     dump captures that explicitly so a regression in narrowing
//     surfaces immediately.
//
//   - Each declared Symbol with its final SymbolTypes entry. This
//     covers symbols that may not have a use site (cleanly-declared
//     locals, params).
//
//   - Each diagnostic (severity, code, message). Sorted by source
//     position so the order is stable across runs.
//
// The format is plain text - no JSON - because the goal is for a
// human to diff snapshot changes and confirm they reflect intentional
// behavior changes. Each section is prefixed with a header so a
// future reader can read top-to-bottom without context.
func DumpForSnapshot(file *rl.SourceFile, info *TypeInfo, resolved *Resolved, diagnostics []Diagnostic) string {
	var sb strings.Builder

	// --- Identifier types -------------------------------------------------
	sb.WriteString("# Identifier types\n")
	idents := collectIdentifiers(file)
	sortByPos(idents)
	if len(idents) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, ident := range idents {
			t := "<no-type>"
			if v, ok := info.ExprTypes[ident]; ok && v != nil {
				t = v.Name()
			}
			fmt.Fprintf(&sb, "  %s @ %d:%d -> %s\n",
				ident.Name, ident.Span().StartLine(), ident.Span().StartColumn(), t)
		}
	}

	// --- Symbol types ----------------------------------------------------
	sb.WriteString("\n# Symbol types\n")
	type symRow struct {
		name string
		kind string
		t    string
	}
	var syms []symRow
	seen := map[*Symbol]bool{}
	for _, sym := range resolved.Uses {
		if sym == nil || seen[sym] || sym.Kind == SymBuiltin {
			continue
		}
		seen[sym] = true
		syms = append(syms, symRowFor(sym, info))
	}
	for _, sym := range resolved.Decls {
		if sym == nil || seen[sym] || sym.Kind == SymBuiltin {
			continue
		}
		seen[sym] = true
		syms = append(syms, symRowFor(sym, info))
	}
	sort.Slice(syms, func(i, j int) bool {
		if syms[i].name != syms[j].name {
			return syms[i].name < syms[j].name
		}
		return syms[i].kind < syms[j].kind
	})
	if len(syms) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, s := range syms {
			fmt.Fprintf(&sb, "  %s (%s): %s\n", s.name, s.kind, s.t)
		}
	}

	// --- Diagnostics -----------------------------------------------------
	//
	// Diagnostics come from the full Check() pipeline, not just the
	// type-checker's BindIssue list. This is what catches AST-level
	// problems (return-outside-fn, invalid LHS, etc.) that the binder
	// + type-checker alone don't see. Without this, snapshots could
	// claim "clean" while `rad check` emits real errors.
	sb.WriteString("\n# Diagnostics\n")
	diags := append([]Diagnostic(nil), diagnostics...)
	sort.SliceStable(diags, func(i, j int) bool {
		a, b := diags[i].Range, diags[j].Range
		if a.Start.Line != b.Start.Line {
			return a.Start.Line < b.Start.Line
		}
		if a.Start.Character != b.Start.Character {
			return a.Start.Character < b.Start.Character
		}
		return codeString(diags[i].Code) < codeString(diags[j].Code)
	})
	if len(diags) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, d := range diags {
			fmt.Fprintf(&sb, "  [%s] %s @ %d:%d - %s\n",
				severityShortDiag(d.Severity), codeString(d.Code),
				d.Range.Start.Line+1, d.Range.Start.Character+1,
				d.Message)
			if d.Suggestion != nil && *d.Suggestion != "" {
				fmt.Fprintf(&sb, "    help: %s\n", *d.Suggestion)
			}
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func codeString(c *rl.Error) string {
	if c == nil {
		return ""
	}
	return c.String()
}

func severityShortDiag(s Severity) string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warn"
	case Info:
		return "info"
	case Hint:
		return "hint"
	}
	return "?"
}

func symRowFor(sym *Symbol, info *TypeInfo) struct {
	name string
	kind string
	t    string
} {
	t := "<no-type>"
	if v, ok := info.SymbolTypes[sym]; ok && v != nil {
		t = v.Name()
	}
	return struct {
		name string
		kind string
		t    string
	}{
		name: sym.Name,
		kind: kindShort(sym.Kind),
		t:    t,
	}
}

func kindShort(k SymbolKind) string {
	switch k {
	case SymBuiltin:
		return "builtin"
	case SymHoistedFn:
		return "fn"
	case SymArg:
		return "arg"
	case SymCmdArg:
		return "cmdarg"
	case SymParam:
		return "param"
	case SymLocal:
		return "local"
	case SymLoopVar:
		return "loop"
	case SymWith:
		return "with"
	}
	return "?"
}

func severityShort(s IssueSeverity) string {
	switch s {
	case IssueError:
		return "error"
	case IssueWarning:
		return "warn"
	case IssueHint:
		return "hint"
	}
	return "?"
}

// collectIdentifiers walks the AST recursively and gathers every
// *rl.Identifier reference. Used by the snapshot dump so the per-use
// types are visible.
func collectIdentifiers(n rl.Node) []*rl.Identifier {
	if n == nil {
		return nil
	}
	var out []*rl.Identifier
	var walk func(rl.Node)
	walk = func(n rl.Node) {
		if n == nil {
			return
		}
		if id, ok := n.(*rl.Identifier); ok {
			out = append(out, id)
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return out
}

func sortByPos(idents []*rl.Identifier) {
	sort.SliceStable(idents, func(i, j int) bool {
		a, b := idents[i].Span(), idents[j].Span()
		if a.StartRow != b.StartRow {
			return a.StartRow < b.StartRow
		}
		if a.StartCol != b.StartCol {
			return a.StartCol < b.StartCol
		}
		return a.StartByte < b.StartByte
	})
}
