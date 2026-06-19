package docir

import (
	"strings"
	"testing"
)

func TestRewriteInlineLinks_ProseRewrittenCodeUntouched(t *testing.T) {
	src := lines(
		"See [docs](./x.md) here.",
		"",
		"```rad",
		"// [not a link](./y.md)",
		"```",
		"",
		"And [more](./z.md).",
	)
	got := RewriteInlineLinks(src, func(text, href string) string {
		return text + " <" + href + ">"
	})
	if !strings.Contains(got, "docs <./x.md>") || !strings.Contains(got, "more <./z.md>") {
		t.Fatalf("prose links not rewritten:\n%s", got)
	}
	if !strings.Contains(got, "[not a link](./y.md)") {
		t.Fatalf("link inside code block should be untouched:\n%s", got)
	}
}

func TestRewriteInlineLinks_MultiLineLinkText(t *testing.T) {
	// The opening bracket and the rest of the link wrap across a line
	// break - a real pattern in our docs.
	src := lines(
		"intro [",
		"`load()`](./fn.md) outro",
	)
	got := RewriteInlineLinks(src, func(text, href string) string {
		return strings.TrimSpace(text) + " (" + href + ")"
	})
	if strings.Contains(got, "](") {
		t.Fatalf("multi-line link not rewritten:\n%s", got)
	}
	if !strings.Contains(got, "`load()` (./fn.md)") {
		t.Fatalf("expected rewritten link text:\n%s", got)
	}
}
