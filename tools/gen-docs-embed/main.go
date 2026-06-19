// gen-docs-embed mirrors the canonical documentation pages listed
// in `docs-web/mkdocs.yml`'s `nav` into an embedded tree at
// `core/embedded_docs/` so the runtime's `//go:embed embedded_docs/*`
// directive picks them up. That embedded tree is what `rad docs`
// reads, which means `rad docs` always serves docs matching the
// installed binary's version - no network, no drift.
//
// This is the third destination compiled from the same source set:
//   - the website (mkdocs reads docs-web/docs/ directly)
//   - llms.txt / llms-full.txt (docs-web/hooks/generate_llms_txt.py)
//   - rad docs (this generator)
//
// The Python hook resolves titles/H2s/nav the same way for its table
// of contents, so those rules here (`_strip_front_matter`,
// `_resolve_title`, `_extract_h2s`, `_parse_nav`) intentionally mirror
// it. For the full page bodies, the hook now reads this generator's
// cleaned output (core/embedded_docs/) rather than re-cleaning the raw
// sources, so llms-full.txt and `rad docs all` are byte-for-byte the
// same content. The drift gate test in core/testing guards parity
// between the nav and the embedded tree.
//
// Behaviour:
//   - Parses `nav`, keeping only the sections whose pages we embed
//     (Guide, Reference, Examples) - the same set the Python hook
//     inlines into llms-full.txt.
//   - For each page reads docs-web/docs/<path> (following the
//     reference/syntax.md symlink to root SYNTAX.md, and reading the
//     already-generated functions.md / errors.md), strips YAML front
//     matter, resolves the title, and extracts H2s for the TOC.
//   - Normalizes the body through tools/docir: mkdocs-only markup
//     (admonitions, content tabs, result <div>s, ragged GFM tables)
//     and HTML-comment authoring notes don't survive a terminal
//     renderer, so docir converts them to clean markdown that reads
//     well both rendered in a TTY and piped raw into LLM context.
//   - Writes the normalized body to core/embedded_docs/<slug>.md and a
//     manifest.json capturing ordered {slug, section, title, h2s}.
//   - Idempotent (no mtime bump when unchanged) and prunes stale
//     slugs, like tools/gen-funcs-go.
//
// Usage (defaults assume invocation from the repo root):
//
//	go run ./tools/gen-docs-embed
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/tools/docir"
	"gopkg.in/yaml.v3"
)

// embeddedSections are the nav sections whose pages we embed. Kept
// in sync with FULL_CONTENT_SECTIONS in
// docs-web/hooks/generate_llms_txt.py so `rad docs all` and
// llms-full.txt cover the same corpus.
var embeddedSections = map[string]bool{
	"Guide":     true,
	"Reference": true,
	"Examples":  true,
}

// skipPaths mirrors SKIP_PAGES in the Python hook: the bare home
// page is navigational, not content. (Section-prefixed index pages
// such as examples/index.md are kept.)
var skipPaths = map[string]bool{"index.md": true}

// skipH2s mirrors SKIP_H2S - navigational headings that are noise in
// a table of contents.
var skipH2s = map[string]bool{"Summary": true, "Next": true}

var (
	h2Pattern      = regexp.MustCompile(`(?m)^## (.+)$`)
	h1Pattern      = regexp.MustCompile(`(?m)^# (.+)$`)
	fmTitlePattern = regexp.MustCompile(`(?m)^title:\s*(.+)$`)
)

func main() {
	var (
		navFile  = flag.String("nav", "docs-web/mkdocs.yml", "path to mkdocs.yml (the nav manifest)")
		docsDir  = flag.String("docs", "docs-web/docs", "path to the docs-web/docs/ directory")
		funcsDir = flag.String("funcs", "docs/funcs", "path to the docs/funcs/ directory (per-function source of truth)")
		target   = flag.String("out", "core/embedded_docs", "path to the embedded_docs output directory")
		dryRun   = flag.Bool("dry-run", false, "print planned actions without writing")
	)
	flag.Parse()

	if err := run(*navFile, *docsDir, *funcsDir, *target, *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "gen-docs-embed: %v\n", err)
		os.Exit(1)
	}
}

type navPage struct {
	section string
	path    string // cleaned, relative to docsDir (e.g. "guide/basics.md")
	title   string // explicit nav title, or "" to resolve from content
}

type docPageMeta struct {
	Slug    string   `json:"slug"`
	Section string   `json:"section"`
	Title   string   `json:"title"`
	H2s     []string `json:"h2s"`
}

type manifest struct {
	Pages []docPageMeta `json:"pages"`
	Funcs []string      `json:"funcs"`
}

func run(navFile, docsDir, funcsDir, target string, dryRun bool) error {
	navBytes, err := os.ReadFile(navFile)
	if err != nil {
		return fmt.Errorf("read nav %s: %w", navFile, err)
	}
	var cfg struct {
		Nav []any `yaml:"nav"`
	}
	if err := yaml.Unmarshal(navBytes, &cfg); err != nil {
		return fmt.Errorf("parse nav %s: %w", navFile, err)
	}

	var pages []navPage
	parseNav(cfg.Nav, "", &pages)

	// Keep only the sections we embed, and drop the navigational
	// home page. Order is preserved from the nav list.
	var kept []navPage
	for _, p := range pages {
		if !embeddedSections[p.section] || skipPaths[p.path] {
			continue
		}
		kept = append(kept, p)
	}
	if len(kept) == 0 {
		return fmt.Errorf("no embeddable pages found in %s nav", navFile)
	}

	// Per-function pages: docs/funcs/<name>.md is the source of truth.
	// `rad docs <name>` resolves to these (routed like an error code).
	// Loaded first so the link rewriter knows which names are addressable.
	funcDocs, err := loadFuncDocs(funcsDir)
	if err != nil {
		return err
	}
	sort.Slice(funcDocs, func(i, j int) bool { return funcDocs[i].Name < funcDocs[j].Name })

	funcNames := make([]string, 0, len(funcDocs))
	funcSet := make(map[string]bool, len(funcDocs))
	for _, fn := range funcDocs {
		if fn.Name == "all" {
			return fmt.Errorf("a function named %q would shadow `rad docs all`", fn.Name)
		}
		funcNames = append(funcNames, fn.Name)
		funcSet[fn.Name] = true
	}

	// Slug set drives link resolution: a relative .md link resolving to
	// a known page becomes `rad docs <slug>`.
	slugSet := make(map[string]bool, len(kept))
	for _, p := range kept {
		slugSet[strings.TrimSuffix(p.path, ".md")] = true
	}

	// normalize is the shared cleanup: rewrite web links first (so table
	// alignment reflects the final cell text), then parse + emit through
	// docir to strip mkdocs-only markup and align tables. baseSlug gives
	// the link rewriter the page's location for resolving relative paths.
	normalize := func(src, baseSlug string) []byte {
		src = docir.RewriteInlineLinks(src, func(linkText, href string) string {
			return resolveDocLink(linkText, href, baseSlug, slugSet, funcSet)
		})
		return []byte(docir.EmitTerminal(docir.Parse(src)))
	}

	type stagedFunc struct {
		name string
		body []byte
	}
	stagedFuncs := make([]stagedFunc, 0, len(funcDocs))
	for _, fn := range funcDocs {
		stagedFuncs = append(stagedFuncs, stagedFunc{
			name: fn.Name,
			body: normalize(renderFuncPage(fn), "reference/functions"),
		})
	}

	type staged struct {
		meta docPageMeta
		body []byte
	}
	stagedPages := make([]staged, 0, len(kept))
	for _, p := range kept {
		source := filepath.Join(docsDir, filepath.FromSlash(p.path))
		raw, err := os.ReadFile(source) // follows the syntax.md symlink
		if err != nil {
			return fmt.Errorf("read page %s: %w", source, err)
		}
		text := string(raw)
		slug := strings.TrimSuffix(p.path, ".md")
		stagedPages = append(stagedPages, staged{
			meta: docPageMeta{
				Slug:    slug,
				Section: p.section,
				Title:   resolveTitle(text, p.title, p.path),
				H2s:     extractH2s(text),
			},
			body: normalize(stripFrontMatter(text), slug),
		})
	}

	if !dryRun {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", target, err)
		}
	}

	// Every emitted .md self-announces as generated (stripped at serve
	// time by core's GetDocPage/GetFuncDoc and the llms.txt hook).
	banner := []byte(rts.GeneratedBanner("tools/gen-docs-embed", "docs-web/docs/ + docs/funcs/") + "\n")

	wrote := 0
	want := make(map[string]struct{}, len(stagedPages)+len(stagedFuncs)+1)
	writeFile := func(rel string, body []byte) error {
		body = append(append([]byte{}, banner...), body...)
		want[rel] = struct{}{}
		dst := filepath.Join(target, rel)
		if dryRun {
			fmt.Printf("would write %s\n", dst)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
		}
		if existing, err := os.ReadFile(dst); err == nil && string(existing) == string(body) {
			return nil // already in sync, don't bump mtime
		}
		if err := os.WriteFile(dst, body, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dst, err)
		}
		wrote++
		return nil
	}

	for _, sp := range stagedPages {
		if err := writeFile(filepath.FromSlash(sp.meta.Slug+".md"), sp.body); err != nil {
			return err
		}
	}
	for _, sf := range stagedFuncs {
		if err := writeFile(filepath.FromSlash("funcs/"+sf.name+".md"), sf.body); err != nil {
			return err
		}
	}

	// Write the manifest (deterministic: ordered by nav, indented).
	m := manifest{Pages: make([]docPageMeta, 0, len(stagedPages)), Funcs: funcNames}
	for _, sp := range stagedPages {
		m.Pages = append(m.Pages, sp.meta)
	}
	manifestBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')
	want["manifest.json"] = struct{}{}
	manifestPath := filepath.Join(target, "manifest.json")
	if dryRun {
		fmt.Printf("would write %s\n", manifestPath)
	} else if existing, err := os.ReadFile(manifestPath); err != nil || string(existing) != string(manifestBytes) {
		if err := os.WriteFile(manifestPath, manifestBytes, 0o644); err != nil {
			return fmt.Errorf("write manifest: %w", err)
		}
		wrote++
	}

	removed, err := pruneTarget(target, want, dryRun)
	if err != nil {
		return err
	}

	fmt.Printf("gen-docs-embed: %d pages; %d functions; %d updated; %d pruned\n",
		len(stagedPages), len(stagedFuncs), wrote, removed)
	return nil
}

// resolveDocLink turns a markdown link into terminal-friendly text.
// Web links keep the bare URL (the runtime renderer auto-links those);
// relative .md links that resolve to a known page or function become
// "text (rad docs <topic>)"; everything else (in-page anchors, images,
// unknown targets) collapses to just the link text. baseSlug is the
// current page's slug, used to resolve relative paths.
func resolveDocLink(text, href, baseSlug string, slugs, funcs map[string]bool) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return text + " (" + href + ")"
	}

	rel, anchor := href, ""
	if i := strings.Index(href, "#"); i >= 0 {
		rel, anchor = href[:i], href[i+1:]
	}
	// Pure in-page anchor, or a non-page target (mailto, image): nothing
	// to point at from a terminal, so keep just the text.
	if rel == "" || !strings.HasSuffix(rel, ".md") {
		return text
	}

	baseDir := ""
	if i := strings.LastIndex(baseSlug, "/"); i >= 0 {
		baseDir = baseSlug[:i]
	}
	target := path.Join(baseDir, strings.TrimSuffix(rel, ".md"))

	// A deep link into the function reference (#pick) is best served by
	// that function's own page.
	if target == "reference/functions" && funcs[anchor] {
		return text + " (rad docs " + anchor + ")"
	}
	if slugs[target] {
		return text + " (rad docs " + target + ")"
	}
	return text
}

// loadFuncDocs parses every per-function doc under dir into a FuncDoc.
// Mirrors gen-funcs-page's loader: non-function files (README, scratch
// notes) are skipped via the stem validity check.
func loadFuncDocs(dir string) ([]*rts.FuncDoc, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	var docs []*rts.FuncDoc
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), ".md")
		if !rts.IsValidFuncDocStem(stem) {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		doc, err := rts.ParseFuncDoc(stem, string(body))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// renderFuncPage reconstructs a standalone `rad docs <name>` page from
// a parsed FuncDoc. Unlike the compact aggregate (gen-funcs-page),
// this is the deep-dive view: an H1 name, all examples, and structured
// Parameters/Notes/See also sections. The `## Category` metadata is
// deliberately omitted - it's a docs-pipeline detail, not user content.
func renderFuncPage(fn *rts.FuncDoc) string {
	var b strings.Builder
	b.WriteString("# ")
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
	if len(fn.Parameters) > 0 {
		b.WriteString("\n## Parameters\n\n")
		for _, p := range fn.Parameters {
			b.WriteString("- `")
			b.WriteString(p.Name)
			b.WriteString("`")
			if p.Type != "" {
				b.WriteString(" (`")
				b.WriteString(p.Type)
				b.WriteString("`)")
			}
			if p.Description != "" {
				b.WriteString(": ")
				b.WriteString(p.Description)
			}
			b.WriteString("\n")
		}
	}
	if strings.TrimSpace(fn.Notes) != "" {
		b.WriteString("\n## Notes\n\n")
		b.WriteString(strings.TrimSpace(fn.Notes))
		b.WriteString("\n")
	}
	if len(fn.SeeAlso) > 0 {
		b.WriteString("\n## See also\n\n")
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

// parseNav walks the nav config and appends (section, path, title)
// for every page. Mirrors _parse_nav in the Python hook: a string
// entry is an untitled page; a one-key map is either a titled page
// (string value) or a section (list value).
func parseNav(nav []any, section string, out *[]navPage) {
	for _, entry := range nav {
		switch e := entry.(type) {
		case string:
			*out = append(*out, navPage{section: section, path: cleanPath(e)})
		case map[string]any:
			for key, value := range e {
				switch v := value.(type) {
				case []any:
					parseNav(v, key, out)
				case string:
					*out = append(*out, navPage{section: section, path: cleanPath(v), title: key})
				}
			}
		}
	}
}

// cleanPath strips the leading "./" mkdocs nav paths carry, matching
// the Python hook's path.lstrip("./").
func cleanPath(path string) string {
	return strings.TrimLeft(path, "./")
}

// resolveTitle mirrors _resolve_title: explicit nav title > front
// matter title > first H1 > titleized filename.
func resolveTitle(raw, explicit, path string) string {
	if explicit != "" {
		return explicit
	}
	if fm := extractFrontMatter(raw); fm != "" {
		if m := fmTitlePattern.FindStringSubmatch(fm); m != nil {
			return strings.Trim(strings.TrimSpace(m[1]), `"'`)
		}
	}
	if m := h1Pattern.FindStringSubmatch(raw); m != nil {
		return strings.TrimSpace(m[1])
	}
	return titleize(strings.TrimSuffix(filepath.Base(path), ".md"))
}

func extractH2s(raw string) []string {
	out := []string{}
	for _, m := range h2Pattern.FindAllStringSubmatch(raw, -1) {
		h := strings.TrimSpace(m[1])
		if skipH2s[h] {
			continue
		}
		out = append(out, h)
	}
	return out
}

// extractFrontMatter returns the YAML front matter block (without the
// --- fences), or "" if absent. Mirrors _extract_front_matter.
func extractFrontMatter(raw string) string {
	if !strings.HasPrefix(raw, "---") {
		return ""
	}
	end := strings.Index(raw[3:], "---")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(raw[3 : 3+end])
}

// stripFrontMatter removes a leading YAML front matter block.
// Mirrors _strip_front_matter.
func stripFrontMatter(raw string) string {
	if !strings.HasPrefix(raw, "---") {
		return raw
	}
	end := strings.Index(raw[3:], "---")
	if end == -1 {
		return raw
	}
	return strings.TrimLeft(raw[3+end+3:], "\n")
}

func titleize(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "-", " "))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// pruneTarget removes .md files and now-empty directories under
// target that aren't in the wanted set, keeping the embedded tree in
// lockstep with the nav.
func pruneTarget(target string, want map[string]struct{}, dryRun bool) (int, error) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return 0, nil
	}
	removed := 0
	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(target, path)
		if relErr != nil {
			return relErr
		}
		if _, keep := want[rel]; keep {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil // leave non-generated files (e.g. README) alone
		}
		if dryRun {
			fmt.Printf("would remove %s\n", path)
			return nil
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("prune %s: %w", path, err)
		}
		removed++
		return nil
	})
	if err != nil {
		return removed, err
	}
	if !dryRun {
		pruneEmptyDirs(target)
	}
	return removed, nil
}

// pruneEmptyDirs removes empty subdirectories left behind after
// pruning (e.g. a whole section was dropped from the nav).
func pruneEmptyDirs(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		pruneEmptyDirs(dir)
		if children, err := os.ReadDir(dir); err == nil && len(children) == 0 {
			os.Remove(dir)
		}
	}
}
