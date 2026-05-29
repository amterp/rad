// gen-funcs-page regenerates the public-facing functions reference
// at `docs-web/docs/reference/functions.md` from the canonical
// per-function markdown files under `docs/funcs/*.md`.
//
// The pipeline:
//
//	docs/funcs/*.md   ── source of truth (per-function)
//	      │
//	      ▼ gen-funcs-page (this binary)
//	docs-web/docs/reference/functions.md   ── derived aggregate
//
// The hand-written "How to Read This Document" preamble lives in
// `tools/gen-funcs-page/preamble.md` and is prepended verbatim - it
// describes the project's general signature syntax, which isn't
// per-function content. Edit it there to update the rendered page.
//
// Layout:
//   - `## <Category>` section per distinct category (alphabetical).
//   - Within a category, `### <name>` sections in alphabetical order.
//   - Each function: description, signature in a ```rad block,
//     first example block (if present), optional ## Notes paragraph,
//     optional See also footer.
//
// Usage:
//
//	go run ./tools/gen-funcs-page
//
// Defaults assume invocation from the repo root. Pass `-source`
// and `-out` to override.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amterp/rad/rts"
)

//go:embed preamble.md
var preamble string

func main() {
	var (
		source = flag.String("source", "docs/funcs", "path to the source docs/funcs/ directory")
		out    = flag.String("out", "docs-web/docs/reference/functions.md", "path to write the aggregated functions reference")
		dryRun = flag.Bool("dry-run", false, "print the generated output to stdout instead of writing")
		stdout = flag.Bool("stdout", false, "alias for -dry-run; emit to stdout")
	)
	flag.Parse()

	if err := run(*source, *out, *dryRun || *stdout); err != nil {
		fmt.Fprintf(os.Stderr, "gen-funcs-page: %v\n", err)
		os.Exit(1)
	}
}

func run(source, out string, dryRun bool) error {
	docs, err := loadFuncDocs(source)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("no func docs found under %s", source)
	}

	body := render(docs)
	combined := preamble + body

	if dryRun {
		fmt.Print(combined)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(out), err)
	}
	if err := os.WriteFile(out, []byte(combined), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	fmt.Printf("gen-funcs-page: wrote %d functions across %d categories to %s\n",
		len(docs), countCategories(docs), out)
	return nil
}

func loadFuncDocs(dir string) ([]*rts.FuncDoc, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	var docs []*rts.FuncDoc
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		stem := strings.TrimSuffix(name, ".md")
		if !rts.IsValidFuncDocStem(stem) {
			continue // README.md, scratch notes, etc.
		}
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		doc, err := rts.ParseFuncDoc(stem, string(body))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func render(docs []*rts.FuncDoc) string {
	byCategory := map[string][]*rts.FuncDoc{}
	for _, d := range docs {
		byCategory[d.Category] = append(byCategory[d.Category], d)
	}
	for _, fns := range byCategory {
		sort.Slice(fns, func(i, j int) bool { return fns[i].Name < fns[j].Name })
	}
	cats := make([]string, 0, len(byCategory))
	for c := range byCategory {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	var b strings.Builder
	for _, cat := range cats {
		b.WriteString("\n## ")
		b.WriteString(displayCategory(cat))
		b.WriteString("\n")
		for _, fn := range byCategory[cat] {
			b.WriteString(renderFunc(fn))
		}
	}
	return b.String()
}

func renderFunc(fn *rts.FuncDoc) string {
	var b strings.Builder
	b.WriteString("\n### ")
	b.WriteString(fn.Name)
	b.WriteString("\n\n")
	if fn.Description != "" {
		b.WriteString(fn.Description)
		b.WriteString("\n\n")
	}
	b.WriteString("```rad\n")
	b.WriteString(fn.Signature)
	b.WriteString("\n```\n")
	for _, ex := range fn.Examples {
		b.WriteString("\n```rad\n")
		b.WriteString(strings.TrimRight(ex, "\n"))
		b.WriteString("\n```\n")
	}
	if strings.TrimSpace(fn.Notes) != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(fn.Notes))
		b.WriteString("\n")
	}
	if len(fn.SeeAlso) > 0 {
		b.WriteString("\nSee also: ")
		for i, name := range fn.SeeAlso {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("`")
			b.WriteString(name)
			b.WriteString("`")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// displayCategory turns the lowercase one-word category into a
// title-cased section header. e.g. "io" -> "IO", "strings" ->
// "Strings". Two-letter all-caps (io, http) are kept upper.
func displayCategory(c string) string {
	switch c {
	case "io":
		return "IO"
	case "http":
		return "HTTP"
	}
	if c == "" {
		return "Misc"
	}
	return strings.ToUpper(c[:1]) + c[1:]
}

func countCategories(docs []*rts.FuncDoc) int {
	seen := map[string]struct{}{}
	for _, d := range docs {
		seen[d.Category] = struct{}{}
	}
	return len(seen)
}
