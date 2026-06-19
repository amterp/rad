package radfmt

import (
	"strings"
	"unicode/utf8"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// argAlignOutlier: within a declaration alignment group, a declaration whose
// canonical code width exceeds the next-widest kept declaration by more than
// this many columns is dropped from the column alignment (rendered with a plain
// two-space comment gap) so a single outsized declaration can't balloon the
// whole group's columns. It also removes any incentive to hand-massage layout -
// the formatter excludes the outlier itself rather than the author reaching for
// a blank line or a reorder (which would change positional-arg semantics).
// Calibrated against snapshots.
const argAlignOutlier = 8

// formatArgBlock formats the `args:` block: a tight `args:` header, then an
// indented body of declarations followed by constraints. Declarations in a
// contiguous group are column-aligned (type / default / comment); the
// declaration group and the constraint group are separated by exactly one blank
// line.
//
// [F37] args block: tight `args:` header, indented body
func (p *printer) formatArgBlock(n *ts.Node) Doc {
	headerComment, items := blockBody(n)
	header := concat(text("args"), text(tColon))
	if headerComment != nil {
		// [F11] a comment trailing the `args:` header stays on the header line.
		header = concat(header, lineSuffix(concat(text(" "), text(p.nodeText(headerComment)))))
	}
	return concat(header, p.argBody(items))
}

type argRowKind int

const (
	argRowDecl argRowKind = iota
	argRowConstraint
	argRowComment
)

// argRow is one logical body line: a declaration, a constraint, or a standalone
// comment. Declarations carry their cells separately so a group can be aligned;
// constraints and comments carry a pre-rendered string.
type argRow struct {
	kind       argRowKind
	blankAbove bool

	// declaration cells (argRowDecl) - joined with single spaces within a row,
	// padded per-column when aligned. The rename stays glued to the name (`name`
	// holds `*name "rename"`); the shorthand is its own column.
	name, short, typ, def string

	// trailing comment with its marker ("# ..." or "// ..."), or "".
	comment string

	// final rendered line. For declarations it's filled by alignDeclGroups;
	// constraints/comments fill it directly.
	text string
}

// argBody lays out the indented block body: it builds rows, enforces the
// declaration/constraint blank line, aligns declaration groups, and joins
// everything with hardlines (a second hardline wherever a blank line belongs).
func (p *printer) argBody(items []*ts.Node) Doc {
	rows := p.buildArgRows(items)
	if len(rows) == 0 {
		return text("")
	}

	// [F41] exactly one blank line between the declaration group and the first
	// constraint that follows it.
	for i := 1; i < len(rows); i++ {
		if rows[i].kind == argRowConstraint && rows[i-1].kind == argRowDecl {
			rows[i].blankAbove = true
		}
	}

	p.alignDeclGroups(rows)

	var parts []Doc
	for i := range rows {
		line := p.renderArgRow(&rows[i])
		if i > 0 {
			parts = append(parts, hardLine()) // [F8] one hardline between rows
			if rows[i].blankAbove {
				parts = append(parts, hardLine()) // [F6][F8] at most one blank line
			}
		}
		parts = append(parts, text(line))
	}
	return indent(concat(hardLine(), concat(parts...)))
}

// buildArgRows classifies body items into rows, folding a `//` comment that
// trails a row on the same line into that row rather than starting a new line.
func (p *printer) buildArgRows(items []*ts.Node) []argRow {
	var rows []argRow
	var prev *ts.Node
	for _, it := range items {
		if isComment(it) && prev != nil && len(rows) > 0 && sameRow(prev, it) {
			// [F10] a trailing same-line comment stays on the row's line.
			rows[len(rows)-1].comment = p.nodeText(it)
			prev = it
			continue
		}

		blank := prev != nil && blankBetween(prev, it)
		switch {
		case isComment(it):
			rows = append(rows, argRow{kind: argRowComment, blankAbove: blank, text: p.nodeText(it)})
		case it.Kind() == rl.K_ARG_DECLARATION:
			name, short, typ, def, comment := p.argDeclCells(it)
			rows = append(rows, argRow{
				kind: argRowDecl, blankAbove: blank,
				name: name, short: short, typ: typ, def: def, comment: comment,
			})
		default:
			rows = append(rows, argRow{kind: argRowConstraint, blankAbove: blank, text: p.argConstraintStr(it)})
		}
		prev = it
	}
	return rows
}

// argDeclCells builds a declaration's aligned cells plus its trailing
// description comment, sourced by field:
//
//	[*]name [rename]   short   type[?]   = default   # comment
//	└───── name ────┘  └─────┘ └ typ ┘   └─ def ─┘
//
// The rename stays glued to the name (it documents that name and reads as a
// unit); the single-character shorthand gets its own column so shorthands line
// up across a group.
//
// [F38] declaration spacing: single spaces; `*`/`?` bind tight
// [F39] `#` description comment: a single space after the marker
func (p *printer) argDeclCells(n *ts.Node) (name, short, typ, def, comment string) {
	var b strings.Builder
	if childByField(n, rl.F_VARIADIC_MARKER) != nil {
		b.WriteString("*")
	}
	if nm := childByField(n, rl.F_ARG_NAME); nm != nil {
		b.WriteString(p.nodeText(nm))
	}
	if r := childByField(n, rl.F_RENAME); r != nil {
		b.WriteString(" ")
		b.WriteString(p.nodeText(r))
	}
	name = b.String()

	if s := childByField(n, rl.F_SHORTHAND); s != nil {
		short = p.nodeText(s)
	}

	if t := childByField(n, rl.F_TYPE); t != nil {
		typ = p.nodeText(t)
	}
	if childByField(n, rl.F_OPTIONAL) != nil {
		typ += "?"
	}

	if d := childByField(n, rl.F_DEFAULT); d != nil {
		// Default value is emitted verbatim for now; canonicalizing list-default
		// spacing is a noted follow-up.
		def = "= " + p.nodeText(d)
	}

	if c := childByField(n, rl.F_COMMENT); c != nil {
		comment = normalizeArgComment(p.nodeText(c))
	}
	return
}

// normalizeArgComment renders an arg description comment canonically: the `#`
// marker, one space, then the content with its surrounding whitespace stripped.
// An empty comment is just `#`. The content is otherwise preserved exactly, so
// this content-rewriting rule can't alter the help text Rad derives from it.
//
// [F39] `#` comment marker spacing
func normalizeArgComment(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "#"
	}
	return "# " + content
}

// argConstraintStr renders a constraint line canonically, reconstructed from
// fields so source spacing is normalized.
//
// [F42] constraint spacing (enum / regex / range / requires / excludes)
func (p *printer) argConstraintStr(n *ts.Node) string {
	name := p.fieldText(n, rl.F_ARG_NAME)
	switch n.Kind() {
	case rl.K_ARG_ENUM_CONSTRAINT:
		return name + " enum " + p.flatStringList(childByField(n, rl.F_VALUES))
	case rl.K_ARG_REGEX_CONSTRAINT:
		return name + " regex " + p.fieldText(n, rl.F_REGEX)
	case rl.K_ARG_RANGE_CONSTRAINT:
		return name + " range " + p.argRangeStr(n)
	case rl.K_ARG_REQUIRES_CONSTRAINT:
		return p.argRelationStr(name, n, "requires", rl.F_REQUIRED)
	case rl.K_ARG_EXCLUDES_CONSTRAINT:
		return p.argRelationStr(name, n, "excludes", rl.F_EXCLUDED)
	default:
		return strings.TrimRight(p.nodeText(n), "\n")
	}
}

// argRangeStr renders a range constraint payload: `[min, max]`, with a missing
// bound omitted and no padding space against the comma (`[0,]`, `[, 10]`),
// matching idiom. The opener/closer brackets are preserved (inclusive `[]` vs
// exclusive `()`).
func (p *printer) argRangeStr(n *ts.Node) string {
	var b strings.Builder
	b.WriteString(p.fieldText(n, rl.F_OPENER))
	if minN := childByField(n, rl.F_MIN); minN != nil {
		b.WriteString(p.nodeText(minN))
	}
	b.WriteString(",")
	if maxN := childByField(n, rl.F_MAX); maxN != nil {
		b.WriteString(" ")
		b.WriteString(p.nodeText(maxN))
	}
	b.WriteString(p.fieldText(n, rl.F_CLOSER))
	return b.String()
}

// argRelationStr renders a requires/excludes relation: `<arg> [mutually]
// <keyword> a, b`.
func (p *printer) argRelationStr(name string, n *ts.Node, keyword, idField string) string {
	var b strings.Builder
	b.WriteString(name)
	if childByField(n, rl.F_MUTUALLY) != nil {
		b.WriteString(" mutually")
	}
	b.WriteString(" ")
	b.WriteString(keyword)
	b.WriteString(" ")
	var ids []string
	for i, c := range childPtrs(n) {
		if n.FieldNameForChild(uint32(i)) == idField {
			ids = append(ids, p.nodeText(c))
		}
	}
	b.WriteString(strings.Join(ids, ", "))
	return b.String()
}

// flatStringList rebuilds a string list with canonical `", "` spacing,
// preserving each element's exact text (quote style included).
func (p *printer) flatStringList(n *ts.Node) string {
	if n == nil {
		return "[]"
	}
	var parts []string
	for _, c := range childPtrs(n) {
		if c.IsNamed() {
			parts = append(parts, p.nodeText(c))
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// alignDeclGroups renders each declaration row, aligning the type / default /
// comment columns within each contiguous group of declarations (a group breaks
// at a blank line, a constraint, or a standalone comment).
//
// [F40] declaration column alignment (type / default / comment)
func (p *printer) alignDeclGroups(rows []argRow) {
	i := 0
	for i < len(rows) {
		if rows[i].kind != argRowDecl {
			i++
			continue
		}
		j := i + 1
		for j < len(rows) && rows[j].kind == argRowDecl && !rows[j].blankAbove {
			j++
		}
		p.renderDeclGroup(rows[i:j])
		i = j
	}
}

// renderDeclGroup computes the per-column widths over the group's non-outlier
// declarations and fills each row's rendered text.
func (p *printer) renderDeclGroup(group []argRow) {
	width := func(r *argRow) int {
		w := runeLen(r.name)
		if r.short != "" {
			w += 1 + runeLen(r.short)
		}
		w += 1 + runeLen(r.typ)
		if r.def != "" {
			w += 1 + runeLen(r.def)
		}
		return w
	}

	kept := make([]bool, len(group))
	for i := range kept {
		kept[i] = true
	}
	// Exclude outliers: peel off the single widest while it exceeds the
	// next-widest kept row by more than the threshold.
	for {
		widestIdx, widest, second := -1, -1, -1
		for i := range group {
			if !kept[i] {
				continue
			}
			if w := width(&group[i]); w > widest {
				second, widest, widestIdx = widest, w, i
			} else if w > second {
				second = w
			}
		}
		if widestIdx < 0 || second < 0 || widest-second <= argAlignOutlier {
			break
		}
		kept[widestIdx] = false
	}

	wName, wShort, wType, wDef := 0, 0, 0, 0
	for i := range group {
		if !kept[i] {
			continue
		}
		wName = maxInt(wName, runeLen(group[i].name))
		wShort = maxInt(wShort, runeLen(group[i].short))
		wType = maxInt(wType, runeLen(group[i].typ))
		wDef = maxInt(wDef, runeLen(group[i].def))
	}

	for i := range group {
		if kept[i] {
			group[i].text = renderAlignedDecl(&group[i], wName, wShort, wType, wDef)
		} else {
			group[i].text = renderPlainDecl(&group[i])
		}
	}
}

// renderAlignedDecl pads a declaration's cells to the group's column widths,
// joining present columns with single spaces. A column no row in the group uses
// (shorthand, default) is omitted entirely; an empty cell in a column the group
// does use is padded so later columns and the comment stay aligned.
func renderAlignedDecl(r *argRow, wName, wShort, wType, wDef int) string {
	cols := []string{pad(r.name, wName)}
	if wShort > 0 {
		cols = append(cols, pad(r.short, wShort))
	}
	cols = append(cols, pad(r.typ, wType))
	if wDef > 0 {
		cols = append(cols, pad(r.def, wDef))
	}
	s := strings.Join(cols, " ")
	if r.comment != "" {
		s += "  " + r.comment // [F40] two-space gap before the comment column
	}
	return strings.TrimRight(s, " ")
}

// renderPlainDecl renders a declaration with single-space cells and a two-space
// comment gap, without column alignment (used for outliers and lone rows).
func renderPlainDecl(r *argRow) string {
	parts := []string{r.name}
	if r.short != "" {
		parts = append(parts, r.short)
	}
	parts = append(parts, r.typ)
	if r.def != "" {
		parts = append(parts, r.def)
	}
	s := strings.Join(parts, " ")
	if r.comment != "" {
		s += "  " + r.comment
	}
	return strings.TrimRight(s, " ")
}

// renderArgRow returns a row's final line. Declarations are already rendered by
// alignDeclGroups; constraints and comments render here.
func (p *printer) renderArgRow(r *argRow) string {
	switch r.kind {
	case argRowDecl:
		return r.text
	case argRowConstraint:
		s := r.text
		if r.comment != "" {
			s += "  " + r.comment
		}
		return strings.TrimRight(s, " ")
	default:
		return strings.TrimRight(r.text, " ")
	}
}

func (p *printer) fieldText(n *ts.Node, field string) string {
	if c := childByField(n, field); c != nil {
		return p.nodeText(c)
	}
	return ""
}

func runeLen(s string) int { return utf8.RuneCountInString(s) }

func pad(s string, w int) string {
	if n := w - runeLen(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
