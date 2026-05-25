package check

import (
	"sort"
)

// levenshtein is the standard edit-distance metric. Kept local to
// rts/check rather than imported from core/fuzzy because rts is
// below core in the dependency graph - the runtime imports the
// checker, not the other way around. The implementation is
// deliberately the same shape as core/fuzzy.go's; if one diverges
// from the other, that's a bug to fix not a feature.
func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	matrix := make([][]int, len(ra)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(rb)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(rb); j++ {
		matrix[0][j] = j
	}
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			matrix[i][j] = min3(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}
	return matrix[len(ra)][len(rb)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// findSimilarNames returns up to `limit` names visible from `scope`
// (walking the parent chain) plus the builtin set, sorted by edit
// distance to `target`. Returns nil when nothing is close enough
// to be worth showing - "did you mean glorp?" when the name is
// nothing like anything is worse than no suggestion at all.
//
// Threshold: max(len/2 + 1, 2). Matches what the runtime
// Env.FindSimilarVars uses, so the static and runtime suggestion
// sets line up.
func findSimilarNames(scope *Scope, builtins map[string]bool, target string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	maxDist := len(target)/2 + 1
	if maxDist < 2 {
		maxDist = 2
	}

	type cand struct {
		name string
		dist int
	}
	seen := map[string]bool{target: true}
	var cands []cand

	collect := func(name string) {
		if seen[name] {
			return
		}
		seen[name] = true
		d := levenshtein(target, name)
		if d > 0 && d <= maxDist {
			cands = append(cands, cand{name: name, dist: d})
		}
	}

	for cur := scope; cur != nil; cur = cur.Parent {
		for name := range cur.Symbols {
			collect(name)
		}
	}
	for name := range builtins {
		collect(name)
	}

	sort.Slice(cands, func(i, j int) bool {
		if cands[i].dist != cands[j].dist {
			return cands[i].dist < cands[j].dist
		}
		return cands[i].name < cands[j].name
	})

	if len(cands) > limit {
		cands = cands[:limit]
	}
	out := make([]string, len(cands))
	for i, c := range cands {
		out[i] = c.name
	}
	return out
}

// formatDidYouMean renders a list of suggestion candidates into the
// `did you mean ...?` line shown after a static diagnostic. Three
// cases:
//
//   - 0 candidates: empty string (the caller suppresses the line).
//   - 1 candidate:  "did you mean 'X'?"
//   - 2 candidates: "did you mean 'X' or 'Y'?"
//   - 3+ candidates: Oxford-or, "did you mean 'X', 'Y', or 'Z'?"
//
// The Oxford-or pattern matches what Rust's rustc emits and reads
// more naturally than the pre-formatter "one of 'X', 'Y'" shape it
// replaces. Empty entries are skipped defensively so callers don't
// have to filter.
func formatDidYouMean(candidates []string) string {
	// Filter empties (defensive; findSimilarNames doesn't produce them).
	filtered := candidates[:0:0]
	for _, c := range candidates {
		if c != "" {
			filtered = append(filtered, c)
		}
	}
	switch len(filtered) {
	case 0:
		return ""
	case 1:
		return "did you mean '" + filtered[0] + "'?"
	case 2:
		return "did you mean '" + filtered[0] + "' or '" + filtered[1] + "'?"
	default:
		// Oxford-or: 'A', 'B', or 'C'.
		var b string
		for i, n := range filtered {
			switch {
			case i == 0:
				b = "'" + n + "'"
			case i == len(filtered)-1:
				b += ", or '" + n + "'"
			default:
				b += ", '" + n + "'"
			}
		}
		return "did you mean " + b + "?"
	}
}
