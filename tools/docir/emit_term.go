package docir

import "strings"

// EmitTerminal renders blocks back to clean markdown for the embedded
// `rad docs` corpus. The output is still markdown (not ANSI): the
// runtime renderer applies TTY styling, and piping it raw yields clean
// text for LLM context. Constructs the runtime renderer can't handle
// (admonitions, tabs, ragged tables, HTML, comments) are gone by here.
func EmitTerminal(blocks []Block) string {
	out := strings.Join(emitBlocks(blocks), "\n\n")
	return strings.TrimRight(out, "\n") + "\n"
}

func emitBlocks(blocks []Block) []string {
	var parts []string
	for _, b := range blocks {
		if s := emitBlock(b); strings.TrimSpace(s) != "" {
			parts = append(parts, s)
		}
	}
	return parts
}

func emitBlock(b Block) string {
	switch v := b.(type) {
	case Text:
		return strings.Join(v.Lines, "\n")
	case Code:
		return emitCode(v)
	case Callout:
		return emitCallout(v)
	case Tabs:
		return emitTabs(v)
	case Table:
		return emitTable(v)
	default:
		return ""
	}
}

func emitCode(c Code) string {
	var b strings.Builder
	b.WriteString("```")
	b.WriteString(c.Lang)
	b.WriteString("\n")
	if c.Body != "" {
		b.WriteString(c.Body)
		b.WriteString("\n")
	}
	b.WriteString("```")
	return b.String()
}

// emitCallout turns an admonition into a bold header line plus an
// indented body. The bold survives the runtime renderer; the indent
// nests the body visually, and an indented fence inside still renders
// as code (the runtime fence matcher allows leading whitespace).
func emitCallout(c Callout) string {
	header := "**" + titleizeKind(c.Kind)
	if c.Title != "" {
		header += ": " + c.Title
	}
	header += "**"
	body := indent(strings.Join(emitBlocks(c.Body), "\n\n"), "    ")
	if body == "" {
		return header
	}
	return header + "\n\n" + body
}

// emitTabs flattens a content-tab group into labeled, indented
// sections - there's no tabbing in a terminal.
func emitTabs(t Tabs) string {
	var parts []string
	for _, tab := range t.Tabs {
		header := "**" + tab.Label + "**"
		body := indent(strings.Join(emitBlocks(tab.Body), "\n\n"), "    ")
		if body == "" {
			parts = append(parts, header)
		} else {
			parts = append(parts, header+"\n\n"+body)
		}
	}
	return strings.Join(parts, "\n\n")
}

func titleizeKind(kind string) string {
	if kind == "" {
		return "Note"
	}
	return strings.ToUpper(kind[:1]) + kind[1:]
}

// indent prefixes every non-blank line; blank lines stay empty.
func indent(s, prefix string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n")
}
