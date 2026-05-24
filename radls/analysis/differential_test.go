package analysis

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Differential testing rationale:
//
// Even though radls today does full reparse + re-analysis on every
// didChange (no incremental analysis yet), the LSP snapshot contract
// promises that the observable end state of a document depends only
// on its current text - not on the path of edits that led there.
//
// These tests pin that contract down. They build the SAME final text
// two ways:
//   (1) Incrementally: AddDoc(t1) + UpdateDoc(t2) + UpdateDoc(t3)
//   (2) From scratch: AddDoc(t3) on a fresh State
// and assert the observable state is identical. If we ever move to
// incremental parsing or caching, these tests are the canary: any
// drift between the two paths shows up as a divergence here before
// it shows up as a flaky LSP feature.
//
// What counts as "observable state": text, diagnostics, AST shape.
// What's intentionally excluded: version counter (different by
// construction), FileID (different by construction), pointer
// identity (different by construction).

func TestDifferentialFinalStateMatchesFromScratch(t *testing.T) {
	cases := []struct {
		name  string
		edits []string
	}{
		{
			name:  "valid-progression",
			edits: []string{"x = 1", "x = 1\ny = 2", "x = 1\ny = 2\nz = 3"},
		},
		{
			name:  "introduce-then-fix-error",
			edits: []string{"x = 1", "x =", "x = 5"},
		},
		{
			name:  "ascii-then-multibyte",
			edits: []string{"x = 1", "x = \"中\"", "x = \"中\"\ny = 2"},
		},
		{
			name: "empty-then-content",
			edits: []string{
				"",
				"\n",
				"x = 1\n",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			final := tc.edits[len(tc.edits)-1]

			incremental := freshState()
			const uri = "file:///diff.rad"
			incremental.AddDoc(uri, tc.edits[0])
			for _, edit := range tc.edits[1:] {
				incremental.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{
					{Text: edit},
				})
			}
			incSnap := incremental.Snapshot(uri)
			defer incSnap.Release()

			scratch := freshState()
			scratch.AddDoc(uri, final)
			scrSnap := scratch.Snapshot(uri)
			defer scrSnap.Release()

			assertEquivalent(t, incSnap, scrSnap)
		})
	}
}

// TestDifferentialEditCommutativityForIndependentDocs verifies that
// editing doc A doesn't perturb doc B's analysis state. With a shared
// parser this is a real risk - if the parser carried per-call state,
// interleaved edits could leak between documents.
func TestDifferentialEditCommutativityForIndependentDocs(t *testing.T) {
	const (
		uriA = "file:///a.rad"
		uriB = "file:///b.rad"
	)
	textA := "x = 1\ny = 2"
	textB := "p = \"hi\"\nq = 3"

	// Path 1: open A first, then B; edit each.
	s1 := freshState()
	s1.AddDoc(uriA, "x = 0")
	s1.AddDoc(uriB, "p =")
	s1.UpdateDoc(uriA, []lsp.TextDocumentContentChangeEvent{{Text: textA}})
	s1.UpdateDoc(uriB, []lsp.TextDocumentContentChangeEvent{{Text: textB}})

	// Path 2: open B first, then A; edit interleaved.
	s2 := freshState()
	s2.AddDoc(uriB, "p =")
	s2.AddDoc(uriA, "x = 0")
	s2.UpdateDoc(uriB, []lsp.TextDocumentContentChangeEvent{{Text: textB}})
	s2.UpdateDoc(uriA, []lsp.TextDocumentContentChangeEvent{{Text: textA}})

	s1A := s1.Snapshot(uriA)
	s2A := s2.Snapshot(uriA)
	s1B := s1.Snapshot(uriB)
	s2B := s2.Snapshot(uriB)
	defer s1A.Release()
	defer s2A.Release()
	defer s1B.Release()
	defer s2B.Release()
	assertEquivalent(t, s1A, s2A)
	assertEquivalent(t, s1B, s2B)
}

func freshState() *State {
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	return s
}

// assertEquivalent compares observable state of two snapshots. Skips
// fields that are different by construction (version, FileID).
func assertEquivalent(t *testing.T, a, b *DocumentVersion) {
	t.Helper()
	if a == nil || b == nil {
		t.Fatalf("got nil snapshot(s): a=%v b=%v", a, b)
	}
	if a.Text() != b.Text() {
		t.Errorf("text mismatch:\n  a: %q\n  b: %q", a.Text(), b.Text())
	}
	if !sameDiagnostics(a.Diagnostics(), b.Diagnostics()) {
		t.Errorf("diagnostics differ:\n  a: %s\n  b: %s",
			diagSummary(a.Diagnostics()), diagSummary(b.Diagnostics()))
	}
	if astShape(a) != astShape(b) {
		t.Errorf("AST shape differs (s-exp):\n  a: %s\n  b: %s",
			astShape(a), astShape(b))
	}
	if a.LineIndex().LineCount() != b.LineIndex().LineCount() {
		t.Errorf("line count: a=%d b=%d",
			a.LineIndex().LineCount(), b.LineIndex().LineCount())
	}
}

// sameDiagnostics compares two diagnostic slices ignoring order. The
// checker doesn't promise a specific emit order, so a sort-then-equal
// comparison is the right way to test "same set of findings."
func sameDiagnostics(a, b []lsp.Diagnostic) bool {
	if len(a) != len(b) {
		return false
	}
	keyOf := func(d lsp.Diagnostic) string {
		var sb strings.Builder
		sb.WriteString(d.Message)
		sb.WriteByte('|')
		writeRange(&sb, d.Range)
		return sb.String()
	}
	as := make([]string, len(a))
	bs := make([]string, len(b))
	for i, d := range a {
		as[i] = keyOf(d)
	}
	for i, d := range b {
		bs[i] = keyOf(d)
	}
	sort.Strings(as)
	sort.Strings(bs)
	return reflect.DeepEqual(as, bs)
}

func diagSummary(ds []lsp.Diagnostic) string {
	var parts []string
	for _, d := range ds {
		parts = append(parts, d.Message)
	}
	sort.Strings(parts)
	return strings.Join(parts, " ; ")
}

// astShape returns the tree-sitter s-expression of the snapshot's
// parse tree. Two snapshots with the same source should produce
// identical s-expressions - that's the invariant we care about.
func astShape(v *DocumentVersion) string {
	if v == nil || v.Tree() == nil {
		return "<nil>"
	}
	return v.Tree().Sexp()
}

// writeRange is a tiny range-stringifier kept local so the test
// file doesn't pull strconv/fmt for one print site. (The helper
// used to be named `fmt`, but that shadowed the stdlib package
// name and broke once any other file in analysis/ imported "fmt".)
func writeRange(sb *strings.Builder, r lsp.Range) {
	itoaInto(sb, r.Start.Line)
	sb.WriteByte(':')
	itoaInto(sb, r.Start.Character)
	sb.WriteByte('-')
	itoaInto(sb, r.End.Line)
	sb.WriteByte(':')
	itoaInto(sb, r.End.Character)
}

func itoaInto(sb *strings.Builder, n int) {
	sb.WriteString(itoa(n))
}
