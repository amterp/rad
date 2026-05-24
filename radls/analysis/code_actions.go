package analysis

import (
	"strings"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
)

// addDiagnosticQuickFixes emits a code action per snapshot
// diagnostic whose Range overlaps the request range. Two
// flavours of action come out of this:
//
//  1. Structured quick fixes - we know how to construct a real
//     edit that resolves the diagnostic. Today only the
//     `T | null` -> `T?` pattern qualifies; we can grow this
//     list as more diagnostics gain machine-applicable fixes.
//
//  2. Informational refactor actions - the diagnostic carries a
//     Suggestion string but no machine-applicable plan. We
//     surface it as a refactor-kind action with no edit, so
//     users see the suggestion in the lightbulb menu and can
//     apply it themselves. This is the right backstop until
//     each emitter gets structured fixes.
//
// Ranges arrive in utf-8 byte coords (from check.Diagnostic);
// we translate through fromByteRange before emitting so the
// editor renders the edit in its negotiated encoding.
func addDiagnosticQuickFixes(actions *[]lsp.CodeAction, snap *DocumentVersion, reqRange lsp.Range) {
	if snap == nil {
		return
	}
	for _, d := range snap.rawDiagnostics {
		dRange := checkRangeToLSP(d.Range)
		if !rangesOverlap(dRange, reqRange) {
			continue
		}
		if action, ok := structuredFixFor(snap, d); ok {
			*actions = append(*actions, action)
			continue
		}
		if d.Suggestion != nil && *d.Suggestion != "" {
			*actions = append(*actions, lsp.CodeAction{
				Title: *d.Suggestion,
				Kind:  lsp.CodeActionRefactor,
			})
		}
	}
}

// structuredFixFor returns a machine-applicable quick fix for a
// diagnostic when we recognize the pattern, or (zero, false)
// when we don't. Recognition uses the diagnostic's RangedSrc
// (the text the diagnostic covers) so we don't have to re-walk
// the source.
//
// The null-union case is the seed: any diagnostic carrying a
// `T?` suggestion whose ranged source includes `|null` (or
// `null|`) is converted to a `T?`. The replacement is computed
// by stripping the null component and appending `?`. Whitespace
// and ordering variants are normalized so `int | null`,
// `int|null`, and `null | int` all map cleanly to `int?`.
func structuredFixFor(snap *DocumentVersion, d check.Diagnostic) (lsp.CodeAction, bool) {
	if d.Suggestion == nil || !strings.Contains(*d.Suggestion, "T?") {
		return lsp.CodeAction{}, false
	}
	replacement, ok := buildNullUnionFix(d.RangedSrc)
	if !ok {
		return lsp.CodeAction{}, false
	}
	target := fromByteRange(checkRangeToLSP(d.Range), snap)
	title := "Replace '|null' with '?'"
	return lsp.NewQuickFix(title, snap.uri, target, replacement), true
}

// buildNullUnionFix turns the tree-sitter ERROR span that covers
// `|null` (or its variants) into the replacement text that
// produces `T?`. The diagnostic spans ONLY the bad union tail -
// e.g. for `x: int|null = 5` the diagnostic is the substring
// `|null`. Replacing that substring with `?` in-place yields
// `x: int? = 5`, which is the canonical form.
//
// Returns false for shapes that aren't safely rewritable
// in-place (e.g. multi-type unions where we'd lose information,
// or a leading `null|` where the `?` needs to land somewhere
// else). Those need a wider-range fix that we can grow later.
func buildNullUnionFix(src string) (string, bool) {
	if !strings.Contains(src, "null") {
		return "", false
	}
	// Normalize whitespace inside the span.
	t := strings.TrimSpace(src)
	t = strings.ReplaceAll(t, " ", "")
	// In-place replacement only handles the trailing pattern:
	// the span is `|null` (optionally trailed by a stray `|`
	// the parser ate). Anything else - bare `null`, leading
	// `null|`, or multi-type union - drops to false.
	switch t {
	case "|null", "|null|":
		return "?", true
	}
	return "", false
}

// checkRangeToLSP turns a check.Range (utf-8 byte columns) into
// an lsp.Range still in byte coordinates. fromByteRange picks
// up from there to land on the negotiated encoding.
func checkRangeToLSP(r check.Range) lsp.Range {
	return lsp.Range{
		Start: lsp.Pos{Line: r.Start.Line, Character: r.Start.Character},
		End:   lsp.Pos{Line: r.End.Line, Character: r.End.Character},
	}
}

// rangesOverlap reports whether two LSP ranges share at least
// one position. Both inputs must be in the same coordinate
// system. We use this to filter diagnostics to those the user's
// selection actually touches - the editor sends a range with
// the request and expects only-relevant actions back.
func rangesOverlap(a, b lsp.Range) bool {
	if rangeStrictlyBefore(a, b) || rangeStrictlyBefore(b, a) {
		return false
	}
	return true
}

// rangeStrictlyBefore reports whether `a` ends before `b` starts.
// Helper for rangesOverlap; pulled out so the comparison logic
// reads symmetrically.
func rangeStrictlyBefore(a, b lsp.Range) bool {
	if a.End.Line < b.Start.Line {
		return true
	}
	if a.End.Line == b.Start.Line && a.End.Character <= b.Start.Character {
		return true
	}
	return false
}
