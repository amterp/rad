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
// To keep all three in lockstep, the front-matter-stripping and
// title-resolution rules here intentionally mirror the Python hook
// (`_strip_front_matter`, `_resolve_title`, `_extract_h2s`,
// `_parse_nav`). The drift gate test in core/testing guards parity
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
//   - Writes the stripped body to core/embedded_docs/<slug>.md and a
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
	"path/filepath"
	"regexp"
	"strings"

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
		navFile = flag.String("nav", "docs-web/mkdocs.yml", "path to mkdocs.yml (the nav manifest)")
		docsDir = flag.String("docs", "docs-web/docs", "path to the docs-web/docs/ directory")
		target  = flag.String("out", "core/embedded_docs", "path to the embedded_docs output directory")
		dryRun  = flag.Bool("dry-run", false, "print planned actions without writing")
	)
	flag.Parse()

	if err := run(*navFile, *docsDir, *target, *dryRun); err != nil {
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
}

func run(navFile, docsDir, target string, dryRun bool) error {
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
			body: []byte(stripFrontMatter(text)),
		})
	}

	if !dryRun {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", target, err)
		}
	}

	wrote := 0
	want := make(map[string]struct{}, len(stagedPages)+1)
	for _, sp := range stagedPages {
		rel := filepath.FromSlash(sp.meta.Slug + ".md")
		want[rel] = struct{}{}
		dst := filepath.Join(target, rel)
		if dryRun {
			fmt.Printf("would write %s\n", dst)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
		}
		if existing, err := os.ReadFile(dst); err == nil && string(existing) == string(sp.body) {
			continue // already in sync, don't bump mtime
		}
		if err := os.WriteFile(dst, sp.body, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dst, err)
		}
		wrote++
	}

	// Write the manifest (deterministic: ordered by nav, indented).
	m := manifest{Pages: make([]docPageMeta, 0, len(stagedPages))}
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

	fmt.Printf("gen-docs-embed: %d pages; %d updated; %d pruned\n", len(stagedPages), wrote, removed)
	return nil
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
