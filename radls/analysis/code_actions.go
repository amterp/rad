package analysis

import (
	"strings"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// addDiagnosticQuickFixes emits a code action per snapshot
// diagnostic whose Range overlaps the request range, for the
// subset of diagnostics that have a machine-applicable fix.
//
// We only emit STRUCTURED quick fixes - actions that come with
// a real WorkspaceEdit the client can apply. An earlier
// iteration also surfaced "info-only" actions (the diagnostic's
// Suggestion string with no edit), but a user clicking those
// gets a no-op: the LSP client invokes apply, there's nothing
// to apply, and the menu closes silently. That trains users to
// distrust the lightbulb. Suggestion text already renders as
// part of the diagnostic itself ("help: ..." line under the
// error message); the lightbulb's job is to perform fixes, not
// repeat help text.
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
		}
	}
}

// structuredFixFor returns a machine-applicable quick fix for a
// diagnostic when we recognize the pattern, or (zero, false)
// when we don't.
//
// Dispatch is keyed on the diagnostic's Code (a stable
// identifier the binder/checker owns) rather than on suggestion
// text. The earlier shape - strings.Contains(*d.Suggestion,
// "T?") - silently coupled fix recognition to human-readable
// prose; editing the suggestion to say "T? syntax" or similar
// would have broken the fix with no compiler help. Each fix
// pattern we recognize now has its own case here, paired with
// a builder that knows how to construct the replacement from
// the diagnostic's RangedSrc.
func structuredFixFor(snap *DocumentVersion, d check.Diagnostic) (lsp.CodeAction, bool) {
	if d.Code == nil {
		return lsp.CodeAction{}, false
	}
	switch *d.Code {
	case rl.ErrUnexpectedToken:
		// The null-in-union heuristic surfaces this code with a
		// RangedSrc like "|null". buildNullUnionFix returns false
		// for anything else under this code, so other
		// ErrUnexpectedToken sites pass through harmlessly.
		replacement, ok := buildNullUnionFix(d.RangedSrc)
		if !ok {
			return lsp.CodeAction{}, false
		}
		target := fromByteRange(checkRangeToLSP(d.Range), snap)
		return lsp.NewQuickFix("Replace '|null' with '?'", snap.uri, target, replacement), true
	case rl.ErrUndefinedVariable:
		// The binder's emitUndefinedIdentifier puts the suggestion
		// in the format "did you mean 'X'?" or "did you mean one
		// of 'X', 'Y', 'Z'?". Surface ONE quick fix per candidate
		// rather than packing them all into one action - users
		// pick the one they meant from the menu.
		if d.Suggestion == nil {
			return lsp.CodeAction{}, false
		}
		names := extractDidYouMeanNames(*d.Suggestion)
		if len(names) == 0 {
			return lsp.CodeAction{}, false
		}
		// Code actions are returned one-per-call; return the top
		// pick. The caller's loop visits all diagnostics, so we
		// could grow this to fan out into multiple actions later -
		// for now the single most-likely rename matches what
		// users typically need.
		target := fromByteRange(checkRangeToLSP(d.Range), snap)
		return lsp.NewQuickFix("Rename to '"+names[0]+"'", snap.uri, target, names[0]), true
	}
	return lsp.CodeAction{}, false
}

// extractDidYouMeanNames parses the suggestion text produced by
// emitUndefinedIdentifier back into the candidate names. Going
// through structured fields on the diagnostic would be cleaner;
// the suggestion string is the only carrier today and adding a
// new field cascades through every BindIssue producer. Cheap
// enough to do here at action-build time.
func extractDidYouMeanNames(suggestion string) []string {
	// Patterns we emit:
	//   "did you mean 'X'?"
	//   "did you mean one of 'X', 'Y', 'Z'?"
	const prefix1 = "did you mean '"
	const prefixN = "did you mean one of '"
	if strings.HasPrefix(suggestion, prefix1) && !strings.HasPrefix(suggestion, prefixN) {
		s := strings.TrimPrefix(suggestion, prefix1)
		s = strings.TrimSuffix(s, "'?")
		if s == "" {
			return nil
		}
		return []string{s}
	}
	if strings.HasPrefix(suggestion, prefixN) {
		s := strings.TrimPrefix(suggestion, prefixN)
		s = strings.TrimSuffix(s, "'?")
		parts := strings.Split(s, "', '")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return nil
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
