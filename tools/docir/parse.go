package docir

import (
	"regexp"
	"strings"
)

var (
	// commentRe matches the `[//]: # (...)` authoring-comment hack:
	// a link-reference definition that renders to nothing on the web
	// but leaks verbatim through a plain terminal renderer.
	commentRe = regexp.MustCompile(`^\s*\[//\]:`)

	// htmlCommentOpenRe matches the start of a block-level HTML comment
	// (`<!-- ... -->`). These are web-only authoring notes and the
	// "GENERATED ... DO NOT EDIT" banners the doc generators prepend to
	// their source pages - invisible on the web, literal noise in a
	// terminal. Dropped like commentRe.
	htmlCommentOpenRe = regexp.MustCompile(`^\s*<!--`)

	// fenceRe matches an opening (or closing) code fence: 3+ backticks
	// or tildes, optionally indented, with an optional info string.
	fenceRe = regexp.MustCompile("^(\\s*)(`{3,}|~{3,})(.*)$")

	// admonitionRe matches `!!! kind "Title"` and the collapsible
	// `???`/`???+` variants. Title is optional.
	admonitionRe = regexp.MustCompile(`^(\s*)(?:!!!|\?\?\?\+?)\s+([\w-]+)(?:\s+"([^"]*)")?\s*$`)

	// tabRe matches a content-tab marker `=== "Label"`.
	tabRe = regexp.MustCompile(`^(\s*)===\s+"([^"]*)"\s*$`)

	// divOpenRe / divCloseRe match block-level <div ...> wrappers used
	// purely for web styling (e.g. <div class="result">).
	divOpenRe  = regexp.MustCompile(`^\s*<div\b([^>]*)>\s*$`)
	divCloseRe = regexp.MustCompile(`^\s*</div>\s*$`)

	// tableSepRe matches a GFM table separator row (the dashes-and-
	// colons line under the header).
	tableSepRe = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?\s*$`)
)

// Parse turns a markdown document into a flat list of Blocks. Front
// matter is expected to already be stripped by the caller.
func Parse(src string) []Block {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	return parseLines(strings.Split(src, "\n"))
}

func parseLines(lines []string) []Block {
	var blocks []Block
	var text []string

	flush := func() {
		if cleaned := cleanTextLines(text); len(cleaned) > 0 {
			blocks = append(blocks, Text{Lines: cleaned})
		}
		text = nil
	}

	i := 0
	for i < len(lines) {
		line := lines[i]

		switch {
		case commentRe.MatchString(line):
			// Drop authoring comments entirely (along with the line).
			i++

		case htmlCommentOpenRe.MatchString(line):
			// Drop HTML comments (single- or multi-line). A malformed
			// comment with no closing `-->` falls back to plain text so
			// it can't swallow the rest of the doc.
			if next, ok := skipHTMLComment(lines, i); ok {
				i = next
			} else {
				text = append(text, line)
				i++
			}

		case fenceRe.MatchString(line):
			flush()
			code, next := parseFence(lines, i)
			blocks = append(blocks, code)
			i = next

		case admonitionRe.MatchString(line):
			flush()
			callout, next := parseAdmonition(lines, i)
			blocks = append(blocks, callout)
			i = next

		case tabRe.MatchString(line):
			flush()
			tabs, next := parseTabs(lines, i)
			blocks = append(blocks, tabs)
			i = next

		case divOpenRe.MatchString(line):
			if inner, next, ok := collectDiv(lines, i); ok {
				flush()
				m := divOpenRe.FindStringSubmatch(line)
				innerBlocks := parseLines(dedent(inner))
				if strings.Contains(m[1], "result") {
					markResult(innerBlocks)
				}
				blocks = append(blocks, innerBlocks...)
				i = next
			} else {
				text = append(text, line)
				i++
			}

		case isTableStart(lines, i):
			flush()
			tbl, next := parseTable(lines, i)
			blocks = append(blocks, tbl)
			i = next

		default:
			text = append(text, line)
			i++
		}
	}
	flush()
	return blocks
}

// parseFence consumes a fenced code block starting at start, returning
// the Code block and the index just past the closing fence.
func parseFence(lines []string, start int) (Code, int) {
	m := fenceRe.FindStringSubmatch(lines[start])
	indent, marker, info := m[1], m[2], m[3]
	lang := firstToken(info)

	var body []string
	i := start + 1
	for i < len(lines) {
		if isClosingFence(lines[i], marker) {
			i++
			break
		}
		body = append(body, stripIndent(lines[i], len(indent)))
		i++
	}
	return Code{Lang: lang, Body: strings.Join(body, "\n")}, i
}

func parseAdmonition(lines []string, start int) (Callout, int) {
	m := admonitionRe.FindStringSubmatch(lines[start])
	markerIndent := len(m[1])
	kind, title := m[2], m[3]
	body, next := collectIndented(lines, start+1, markerIndent)
	return Callout{Kind: kind, Title: title, Body: parseLines(dedent(body))}, next
}

// parseTabs coalesces a run of consecutive `=== "Label"` markers at
// the same indentation into one Tabs block.
func parseTabs(lines []string, start int) (Tabs, int) {
	markerIndent := len(tabRe.FindStringSubmatch(lines[start])[1])
	var tabs []Tab
	i := start
	for i < len(lines) {
		m := tabRe.FindStringSubmatch(lines[i])
		if m == nil || len(m[1]) != markerIndent {
			break
		}
		body, next := collectIndented(lines, i+1, markerIndent)
		tabs = append(tabs, Tab{Label: m[2], Body: parseLines(dedent(body))})
		i = next
	}
	return Tabs{Tabs: tabs}, i
}

func parseTable(lines []string, start int) (Table, int) {
	header := splitTableRow(lines[start])
	align := parseAlign(splitTableRow(lines[start+1]))
	var rows [][]string
	i := start + 2
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" || !strings.Contains(lines[i], "|") {
			break
		}
		rows = append(rows, splitTableRow(lines[i]))
		i++
	}
	return Table{Header: header, Align: align, Rows: rows}, i
}

// isTableStart reports whether the line at i begins a GFM table: a
// header row containing a pipe immediately followed by a separator
// row that also contains a pipe (the pipe requirement keeps a bare
// `---` thematic break from being read as a table).
func isTableStart(lines []string, i int) bool {
	if strings.TrimSpace(lines[i]) == "" || !strings.Contains(lines[i], "|") {
		return false
	}
	if i+1 >= len(lines) {
		return false
	}
	next := lines[i+1]
	return strings.Contains(next, "|") && tableSepRe.MatchString(next)
}

// skipHTMLComment returns the index just past a block-level HTML
// comment starting at start, and ok=true. The comment may close on the
// same line or several lines down. If no closing `-->` is found, returns
// ok=false so the caller can treat the line as plain text rather than
// dropping the rest of the document.
func skipHTMLComment(lines []string, start int) (next int, ok bool) {
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], "-->") {
			return i + 1, true
		}
	}
	return start, false
}

// collectDiv gathers the lines between a block-level <div ...> and its
// matching </div>, handling nesting. Returns ok=false when there's no
// closing tag, so the caller can fall back to treating the line as
// plain text.
func collectDiv(lines []string, start int) (inner []string, next int, ok bool) {
	depth := 1
	for i := start + 1; i < len(lines); i++ {
		if divOpenRe.MatchString(lines[i]) {
			depth++
			inner = append(inner, lines[i])
			continue
		}
		if divCloseRe.MatchString(lines[i]) {
			depth--
			if depth == 0 {
				return inner, i + 1, true
			}
			inner = append(inner, lines[i])
			continue
		}
		inner = append(inner, lines[i])
	}
	return nil, start, false
}

// collectIndented gathers the body of an admonition or tab: subsequent
// lines that are blank or indented past markerIndent. Returns the body
// (with leading/trailing blank lines trimmed) and the index of the
// first line that ends the body.
func collectIndented(lines []string, start, markerIndent int) (body []string, next int) {
	i := start
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" {
			body = append(body, lines[i])
			i++
			continue
		}
		if leadingSpaces(lines[i]) > markerIndent {
			body = append(body, lines[i])
			i++
			continue
		}
		break
	}
	return trimBlankEdges(body), i
}

// markResult flags top-level Code blocks as result/output. Called for
// the contents of a <div class="result">.
func markResult(blocks []Block) {
	for i, b := range blocks {
		if c, ok := b.(Code); ok {
			c.IsResult = true
			blocks[i] = c
		}
	}
}

// splitTableRow splits a table row on unescaped pipes, treating
// backtick-delimited inline code as atomic so a literal pipe inside a
// type like `int | float` doesn't get read as a column boundary.
func splitTableRow(line string) []string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "|")
	s = strings.TrimSuffix(s, "|")

	var cells []string
	var cur strings.Builder
	inCode := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && s[i+1] == '|':
			cur.WriteByte('|')
			i++
		case c == '`':
			inCode = !inCode
			cur.WriteByte(c)
		case c == '|' && !inCode:
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	cells = append(cells, strings.TrimSpace(cur.String()))
	return cells
}

func parseAlign(sep []string) []Align {
	align := make([]Align, len(sep))
	for i, c := range sep {
		c = strings.TrimSpace(c)
		left := strings.HasPrefix(c, ":")
		right := strings.HasSuffix(c, ":")
		switch {
		case left && right:
			align[i] = AlignCenter
		case right:
			align[i] = AlignRight
		case left:
			align[i] = AlignLeft
		default:
			align[i] = AlignNone
		}
	}
	return align
}

// --- small line helpers ---

func isClosingFence(line, openMarker string) bool {
	t := strings.TrimSpace(line)
	if t == "" {
		return false
	}
	char := openMarker[0]
	if strings.Trim(t, string(char)) != "" {
		return false
	}
	return len(t) >= len(openMarker)
}

func firstToken(info string) string {
	return strings.TrimSpace(strings.SplitN(strings.TrimSpace(info), " ", 2)[0])
}

func leadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func stripIndent(line string, n int) string {
	for i := 0; i < n && len(line) > 0 && line[0] == ' '; i++ {
		line = line[1:]
	}
	return line
}

// dedent removes the common leading-space prefix from a block of
// lines (blank lines ignored when measuring).
func dedent(lines []string) []string {
	min := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if lead := leadingSpaces(l); min == -1 || lead < min {
			min = lead
		}
	}
	if min <= 0 {
		return lines
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			out[i] = ""
		} else {
			out[i] = l[min:]
		}
	}
	return out
}

func trimBlankEdges(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// cleanTextLines trims blank edges and collapses internal runs of
// blank lines to a single blank, so dropped comments don't leave gaps.
func cleanTextLines(lines []string) []string {
	lines = trimBlankEdges(lines)
	var out []string
	prevBlank := false
	for _, l := range lines {
		blank := strings.TrimSpace(l) == ""
		if blank && prevBlank {
			continue
		}
		out = append(out, l)
		prevBlank = blank
	}
	return out
}
