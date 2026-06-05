package radfmt

import (
	"strings"

	"github.com/amterp/rad/rts"
)

// normalizeLineEndings converts CRLF and bare CR to LF so formatting is
// line-ending agnostic and output is canonical. (radfmt is below core in the
// import graph, so it can't reuse core's helper - and this handles bare CR,
// which core's CRLF-only version does not.)
// [F2] normalize line endings to LF
func normalizeLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

// Format canonically re-formats a Rad script.
//
// It returns the formatted source, whether it differs from the input, and
// whether formatting succeeded. Formatting is deliberately conservative about
// failure: if the source has syntax errors, or the formatter panics on some
// construct, Format returns the original source unchanged with ok=false. This
// guarantees `rad fmt` can never emit corrupted output - the worst case is a
// no-op.
func Format(src string) (out string, changed bool, ok bool) {
	// Compare against the original input (not the normalized form) so a
	// line-ending-only fix (CRLF -> LF) still reports changed=true and is
	// actually written. On any failure we return this original untouched.
	original := src

	// Any panic in the format pass degrades to a safe no-op rather than
	// producing (or persisting) corrupted output.
	defer func() {
		if r := recover(); r != nil {
			out, changed, ok = original, false, false
		}
	}()

	src = normalizeLineEndings(src)

	formatted, wantSig, wantComments, ok := formatRaw(src)
	if !ok {
		return original, false, false
	}

	// Last line of defense: never emit output that parses to a different code
	// structure, or that dropped/duplicated a comment. If the formatter got it
	// wrong, degrade to a safe no-op rather than corrupting the user's script.
	if !structurallyEquivalent(formatted, wantSig, wantComments) {
		return original, false, false
	}

	return formatted, formatted != original, true
}

// formatRaw parses src and renders the formatted output WITHOUT the
// structural-equivalence guard. It returns the formatted text plus the input's
// structural signature and comment count so the caller (Format) can verify
// equivalence. ok is false when src has parse errors. Tests use formatRaw to
// inspect raw formatter output (and assert structure preservation with a
// readable diff) rather than getting a silent no-op from the guard.
func formatRaw(src string) (out string, wantSig string, wantComments int, ok bool) {
	parser, err := rts.NewRadParser()
	if err != nil {
		return src, "", 0, false
	}
	defer parser.Close()

	tree := parser.Parse(src)
	if tree.HasInvalidNodes() {
		return src, "", 0, false
	}

	root := tree.Root()
	if root == nil {
		return src, "", 0, false
	}

	wantSig, wantComments = structuralSig(root)

	p := &printer{src: src}
	out = PrintDocToString(p.format(root), MaxWidth)
	return out, wantSig, wantComments, true
}
