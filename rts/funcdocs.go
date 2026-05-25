package rts

import (
	"fmt"
	"strings"
)

// FuncDoc carries the documentation for a built-in function, parsed
// from `docs/funcs/<name>.md`. The signature is stored as the raw
// source line; downstream consumers (rts/signatures.go's parser,
// the LSP hover renderer) re-parse on demand rather than encoding
// the parsed form here. The intent is to keep this struct small
// and serializable; a future commit may grow it (e.g. with a
// pre-parsed param list) once we have a clearer view of what
// callers need.
type FuncDoc struct {
	Name        string
	Description string   // body of the H1 section, before any "##" subsection
	Signature   string   // single line inside the `## Signature` section
	Parameters  []FuncDocParam
	Examples    []string // each rad code block inside `## Examples`
	Category    string
	Notes       string
	SeeAlso     []string // function names listed in `## See also`
}

// FuncDocParam is one entry in the `## Parameters` list. The Type
// column carries whatever the doc author wrote in backticks - it
// isn't validated against the signature here; that check belongs
// in the codegen test.
type FuncDocParam struct {
	Name        string
	Type        string
	Description string
}

// ParseFuncDoc parses a docs/funcs/*.md file into a FuncDoc. The
// expected shape is documented in docs/funcs/README.md and locked
// by the codegen test - if you're updating the parser, update both.
//
// Lenient about extra blank lines and trailing whitespace; strict
// about required sections (returns an error when any of `Signature`,
// `Examples`, `Category` is missing). Optional sections (`Notes`,
// `See also`, `Parameters`) just leave the corresponding fields
// empty when absent.
func ParseFuncDoc(name, src string) (*FuncDoc, error) {
	doc := &FuncDoc{Name: name}

	// go:embed bakes in file bytes verbatim, and Windows git checks
	// these .md files out with CRLF. Normalize so parsed doc content
	// (and the hover output built from it) is byte-identical on every
	// platform rather than depending on the checkout's line endings.
	src = strings.ReplaceAll(src, "\r\n", "\n")

	lines := strings.Split(src, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty doc for %s", name)
	}

	// First non-blank line must be the H1 title and the name must
	// match the file stem - prevents silent renames that drift the
	// doc and the registered builtin out of sync.
	titleIdx := -1
	for i, l := range lines {
		stripped := strings.TrimSpace(l)
		if stripped == "" {
			continue
		}
		titleIdx = i
		break
	}
	if titleIdx < 0 {
		return nil, fmt.Errorf("%s: missing H1 title", name)
	}
	titleLine := strings.TrimSpace(lines[titleIdx])
	expected := "# " + name
	if titleLine != expected {
		return nil, fmt.Errorf("%s: H1 is %q, expected %q", name, titleLine, expected)
	}

	// Walk the rest of the file collecting sections. A section
	// starts at a `## <heading>` line and ends at the next `## `
	// or EOF.
	sections := map[string][]string{}
	var current string
	var pre []string // lines after the H1 before the first ## section
	beforeFirst := true
	for i := titleIdx + 1; i < len(lines); i++ {
		line := lines[i]
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "## ") {
			current = strings.TrimSpace(strings.TrimPrefix(stripped, "## "))
			sections[current] = []string{}
			beforeFirst = false
			continue
		}
		if beforeFirst {
			pre = append(pre, line)
			continue
		}
		sections[current] = append(sections[current], line)
	}

	doc.Description = strings.TrimSpace(strings.Join(pre, "\n"))

	signatureLines, ok := sections["Signature"]
	if !ok {
		return nil, fmt.Errorf("%s: missing '## Signature' section", name)
	}
	doc.Signature = extractInlineCode(signatureLines)
	if doc.Signature == "" {
		return nil, fmt.Errorf("%s: '## Signature' must contain a single backtick-quoted line", name)
	}

	if paramLines, ok := sections["Parameters"]; ok {
		doc.Parameters = parseParamsSection(paramLines)
	}

	exampleLines, ok := sections["Examples"]
	if !ok {
		return nil, fmt.Errorf("%s: missing '## Examples' section", name)
	}
	doc.Examples = extractCodeBlocks(exampleLines)
	if len(doc.Examples) == 0 {
		return nil, fmt.Errorf("%s: '## Examples' has no rad code blocks", name)
	}

	categoryLines, ok := sections["Category"]
	if !ok {
		return nil, fmt.Errorf("%s: missing '## Category' section", name)
	}
	doc.Category = strings.TrimSpace(strings.Join(categoryLines, " "))
	if doc.Category == "" {
		return nil, fmt.Errorf("%s: '## Category' is empty", name)
	}

	if notesLines, ok := sections["Notes"]; ok {
		doc.Notes = strings.TrimSpace(strings.Join(notesLines, "\n"))
	}
	if seeAlsoLines, ok := sections["See also"]; ok {
		doc.SeeAlso = parseSeeAlso(seeAlsoLines)
	}

	return doc, nil
}

// extractInlineCode pulls the first ``...`` inline-code span from
// the given lines. Returns "" when none is found. The signature
// section is expected to be a single line of inline code; this
// helper is forgiving about leading blank lines before that line.
func extractInlineCode(lines []string) string {
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}
		start := strings.Index(stripped, "`")
		if start < 0 {
			return ""
		}
		end := strings.LastIndex(stripped, "`")
		if end <= start {
			return ""
		}
		return strings.TrimSpace(stripped[start+1 : end])
	}
	return ""
}

// parseParamsSection turns a `## Parameters` bullet list into
// structured FuncDocParam entries. Each bullet is expected to
// match `- <name> (<type>): <description>`; bullets that don't
// match are skipped silently so prose between bullets stays
// non-fatal.
func parseParamsSection(lines []string) []FuncDocParam {
	var params []FuncDocParam
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if !strings.HasPrefix(stripped, "- ") {
			continue
		}
		body := strings.TrimPrefix(stripped, "- ")
		// Find the first `<name>` (backtick-quoted) followed by
		// optional `(<type>)` and a `:` description.
		nameStart := strings.Index(body, "`")
		if nameStart < 0 {
			continue
		}
		nameEnd := strings.Index(body[nameStart+1:], "`")
		if nameEnd < 0 {
			continue
		}
		nameEnd += nameStart + 1
		name := body[nameStart+1 : nameEnd]
		rest := body[nameEnd+1:]
		// Optional `(<type>)`.
		typeStr := ""
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, "(") {
			closeParen := strings.Index(rest, ")")
			if closeParen > 0 {
				typeStr = strings.TrimSpace(rest[1:closeParen])
				// Strip surrounding backticks if present.
				typeStr = strings.Trim(typeStr, "`")
				rest = strings.TrimSpace(rest[closeParen+1:])
			}
		}
		// Optional `:<description>`.
		desc := ""
		if strings.HasPrefix(rest, ":") {
			desc = strings.TrimSpace(strings.TrimPrefix(rest, ":"))
		}
		params = append(params, FuncDocParam{
			Name:        name,
			Type:        typeStr,
			Description: desc,
		})
	}
	return params
}

// extractCodeBlocks pulls every ```rad ... ``` block out of a
// section body. Blocks tagged with anything other than "rad" are
// skipped - the docs use shell blocks for invocation examples
// and we don't want to render those as Rad code.
func extractCodeBlocks(lines []string) []string {
	var blocks []string
	var current []string
	inBlock := false
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if !inBlock {
			if strings.HasPrefix(stripped, "```rad") {
				inBlock = true
				current = current[:0]
			}
			continue
		}
		if strings.HasPrefix(stripped, "```") {
			blocks = append(blocks, strings.Join(current, "\n"))
			inBlock = false
			current = current[:0]
			continue
		}
		current = append(current, line)
	}
	return blocks
}

// parseSeeAlso turns the `## See also` body into a flat list of
// function names. Accepts both bullet-list and comma-separated
// forms; backticks around names are stripped.
func parseSeeAlso(lines []string) []string {
	body := strings.Join(lines, "\n")
	// Split on commas + newlines + bullet markers; trim backticks
	// and whitespace on each token.
	body = strings.ReplaceAll(body, "\n", ",")
	body = strings.ReplaceAll(body, "- ", "")
	parts := strings.Split(body, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "`")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
