package docir

import (
	"regexp"
	"strings"
)

// inlineLinkRe matches a markdown inline link `[text](href)`. Link text
// may itself contain inline code with backticks (e.g. [`str`](...)).
var inlineLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

// RewriteInlineLinks rewrites every markdown inline link outside fenced
// code via resolve(text, href), which returns the replacement text. The
// runtime terminal renderer only auto-links bare URLs - it has no notion
// of `[text](href)` syntax - so web-authored links would otherwise show
// as literal noise (relative .md paths) or dangle. resolve turns them
// into something useful in a terminal (e.g. "text (rad docs <topic>)").
//
// Links inside code blocks are left untouched (a URL in example code
// must survive verbatim). docir owns the mechanics; the caller supplies
// resolve since only it knows how hrefs map to `rad docs` topics.
func RewriteInlineLinks(md string, resolve func(text, href string) string) string {
	lines := strings.Split(md, "\n")
	var out, buf []string

	// Rewrite a contiguous non-code region as one string, so a link
	// whose text wraps across a line break still matches.
	flush := func() {
		if len(buf) == 0 {
			return
		}
		rewritten := inlineLinkRe.ReplaceAllStringFunc(strings.Join(buf, "\n"), func(m string) string {
			sub := inlineLinkRe.FindStringSubmatch(m)
			return resolve(sub[1], sub[2])
		})
		out = append(out, strings.Split(rewritten, "\n")...)
		buf = buf[:0]
	}

	inCode := false
	for _, line := range lines {
		if fenceRe.MatchString(line) {
			flush()
			out = append(out, line)
			inCode = !inCode
			continue
		}
		if inCode {
			out = append(out, line)
			continue
		}
		buf = append(buf, line)
	}
	flush()
	return strings.Join(out, "\n")
}
