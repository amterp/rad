package testing

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/amterp/rad/rts/check"
)

// snippetRoot describes one corpus of markdown files to scan.
// IncludeFile narrows which files within the root we extract from -
// useful for skipping meta files like core/error_docs/AGENTS.md that
// share the directory with real error docs but aren't error docs.
type snippetRoot struct {
	// Path is relative to the repo root.
	Path        string
	IncludeFile func(filename string) bool
}

// startsWithDigit matches the error-doc naming convention used in core/error_docs
// (e.g. 10001.md, 30007.md), mirroring isErrorDocFile in error_docs_test.go.
func startsWithDigit(filename string) bool {
	if !strings.HasSuffix(filename, ".md") || filename == "" {
		return false
	}
	c := filename[0]
	return c >= '0' && c <= '9'
}

func anyMarkdown(filename string) bool {
	return strings.HasSuffix(filename, ".md")
}

var docSnippetRoots = []snippetRoot{
	{Path: "docs-web/docs", IncludeFile: anyMarkdown},
	{Path: "core/error_docs", IncludeFile: startsWithDigit},
}

// excludedDocPaths skips an entire file from validation. Prefer per-snippet
// tolerance entries; use this only when a file is mid-rework or is structurally
// not-runnable (e.g. a reference of function signatures).
//
// Keys are paths relative to the repo root.
var excludedDocPaths = map[string]bool{
	// reference/functions.md contains function signatures (e.g.
	// `print(*_items: any, *, sep: str = " ") -> void`), not runnable
	// Rad. Pending restructuring of that file - revisit afterwards.
	"docs-web/docs/reference/functions.md": true,
}

// Path from core/testing/ up to the repo root.
const repoRootFromTestDir = "../.."

// Matches the opening fence of a rad code block. We accept:
//
//	```rad
//	```rad linenums="1" hl_lines="0"
//	```rad title="example.rad"
//
// but not ```radish, ```rad-foo, etc.
var openFenceRe = regexp.MustCompile("^\\s*```rad(\\s.*)?$")

// Matches a bare closing fence: ``` with optional whitespace.
var closeFenceRe = regexp.MustCompile("^\\s*```\\s*$")

type docSnippet struct {
	// ID is "<relPath>#<8-hex-content-hash>", e.g.
	// "docs-web/docs/guide/error-handling.md#a1b2c3d4".
	ID string
	// RelPath is the snippet's source file relative to the repo root.
	RelPath string
	// Line is the 1-based line number of the opening ```rad fence.
	Line int
	// Index is the 1-based index of this snippet within its file
	// (purely informational, used for failure reports - not part of the ID).
	Index int
	// Content is the raw snippet body, lines joined by '\n', no trailing newline.
	Content string
	// Hash is the 8-hex prefix of sha256(Content).
	Hash string
}

// TestDocSnippets runs `rad check` over every ```rad block under
// docSnippetRoots. Default expectation per snippet: zero diagnostics.
// Exceptions live in docSnippetTolerances.
//
// Also fails if any tolerance entry didn't match a snippet (stale entries).
func TestDocSnippets(t *testing.T) {
	snippets, err := collectAllSnippets()
	if err != nil {
		t.Fatalf("failed to collect snippets: %v", err)
	}
	if len(snippets) == 0 {
		t.Fatal("collected zero snippets - extractor is broken or roots are wrong")
	}

	// Sanity: detect intra-file ID collisions where the CONTENT
	// differs - that's a real hash collision and would break tolerance
	// lookup. Two snippets with the same content (e.g. a tutorial's
	// Preview at top + the final-form snippet at the bottom) share the
	// same ID by design; they're benign because any tolerance applies
	// uniformly to both.
	contentByID := make(map[string]string)
	for _, s := range snippets {
		if prev, seen := contentByID[s.ID]; seen && prev != s.Content {
			t.Errorf("hash collision on snippet ID %q (two snippets in the same file with different content but identical 8-hex prefix). Edit one to disambiguate.", s.ID)
		} else {
			contentByID[s.ID] = s.Content
		}
	}

	seenIDs := make(map[string]bool, len(snippets))

	for _, snip := range snippets {
		snip := snip
		seenIDs[snip.ID] = true
		t.Run(snip.ID, func(t *testing.T) {
			tol, hasTol := docSnippetTolerances[snip.ID]
			if hasTol && strings.TrimSpace(tol.Reason) == "" {
				t.Errorf("tolerance entry for %s has empty Reason - every entry must explain why it exists", snip.ID)
			}
			if hasTol && tol.Skip {
				return
			}

			chk, err := check.NewChecker()
			if err != nil {
				t.Fatalf("NewChecker: %v", err)
			}
			chk.UpdateSrc(snip.Content)
			result, err := chk.Check()
			if err != nil {
				t.Fatalf("Check: %v", err)
			}

			failures := evaluateDiagnostics(result.Diagnostics, tol)
			if len(failures) > 0 {
				t.Error(formatFailure(snip, failures, tol))
			}
		})
	}

	t.Run("StaleTolerances", func(t *testing.T) {
		var stale []string
		for id := range docSnippetTolerances {
			if !seenIDs[id] {
				stale = append(stale, id)
			}
		}
		sort.Strings(stale)
		if len(stale) > 0 {
			var b strings.Builder
			fmt.Fprintf(&b, "%d tolerance entries no longer match any snippet (file removed, or snippet content edited):\n", len(stale))
			for _, id := range stale {
				fmt.Fprintf(&b, "  - %s\n", id)
			}
			b.WriteString("Remove these entries from docSnippetTolerances, or re-add them under the new content hash.")
			t.Error(b.String())
		}
	})
}

func collectAllSnippets() ([]docSnippet, error) {
	var out []docSnippet
	for _, root := range docSnippetRoots {
		abs := filepath.Join(repoRootFromTestDir, root.Path)
		err := filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !root.IncludeFile(info.Name()) {
				return nil
			}
			rel, relErr := filepath.Rel(filepath.Join(repoRootFromTestDir, root.Path), path)
			if relErr != nil {
				return relErr
			}
			repoRel := filepath.ToSlash(filepath.Join(root.Path, rel))
			if excludedDocPaths[repoRel] {
				return nil
			}
			snips, perr := extractSnippetsFromFile(path, root.Path)
			if perr != nil {
				return fmt.Errorf("%s: %w", path, perr)
			}
			out = append(out, snips...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// extractSnippetsFromFile reads a markdown file and returns one docSnippet
// per ```rad ... ``` block. relRoot is the snippet-root prefix (e.g.
// "docs-web/docs") so we can build a path relative to the repo root.
func extractSnippetsFromFile(absPath, relRoot string) ([]docSnippet, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Relative path from repo root for ID purposes.
	rootAbs := filepath.Join(repoRootFromTestDir, relRoot)
	rel, err := filepath.Rel(rootAbs, absPath)
	if err != nil {
		return nil, err
	}
	relFromRepoRoot := filepath.ToSlash(filepath.Join(relRoot, rel))

	var (
		snippets []docSnippet
		inRad    bool
		buf      []string
		openLine int
		idx      int
	)
	scanner := bufio.NewScanner(f)
	// Allow long lines (some snippets have wide tables / long strings).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if !inRad {
			if openFenceRe.MatchString(line) {
				inRad = true
				buf = buf[:0]
				openLine = lineNo
			}
			continue
		}
		// Inside a rad block.
		if closeFenceRe.MatchString(line) {
			idx++
			content := strings.Join(buf, "\n")
			h := sha256.Sum256([]byte(content))
			hash := hex.EncodeToString(h[:])[:8]
			snippets = append(snippets, docSnippet{
				ID:      fmt.Sprintf("%s#%s", relFromRepoRoot, hash),
				RelPath: relFromRepoRoot,
				Line:    openLine,
				Index:   idx,
				Content: content,
				Hash:    hash,
			})
			inRad = false
			buf = buf[:0]
			continue
		}
		buf = append(buf, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inRad {
		return nil, fmt.Errorf("unclosed ```rad fence opened at line %d", openLine)
	}
	return snippets, nil
}

// evaluateDiagnostics returns a list of human-readable failure strings.
// An empty result means the snippet's diagnostic set satisfies the tolerance.
func evaluateDiagnostics(diags []check.Diagnostic, tol Tolerance) []string {
	if tol.Skip {
		return nil
	}

	maxSev, hasMax := parseSeverity(tol.MaxSeverity)

	expected := make(map[string]bool, len(tol.ExpectedCodes))
	seenExpected := make(map[string]bool, len(tol.ExpectedCodes))
	for _, c := range tol.ExpectedCodes {
		expected[c] = true
	}

	var fails []string
	for _, d := range diags {
		code := diagCode(d)
		// Severity tolerance: diagnostics at or below MaxSeverity are accepted.
		// (Note: the Severity iota goes Hint < Warning < Info < Error.)
		if hasMax && d.Severity <= maxSev {
			if code != "" && expected[code] {
				seenExpected[code] = true
			}
			continue
		}
		// Code tolerance: codes in ExpectedCodes are accepted.
		if code != "" && expected[code] {
			seenExpected[code] = true
			continue
		}
		fails = append(fails, formatDiag(d))
	}

	for code := range expected {
		if !seenExpected[code] {
			fails = append(fails, fmt.Sprintf("expected diagnostic %s was not produced", code))
		}
	}
	return fails
}

func parseSeverity(s string) (check.Severity, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hint":
		return check.Hint, true
	case "warning", "warn":
		return check.Warning, true
	case "info":
		return check.Info, true
	case "error":
		return check.Error, true
	default:
		return 0, false
	}
}

func diagCode(d check.Diagnostic) string {
	if d.Code == nil {
		return ""
	}
	return d.Code.String()
}

func formatDiag(d check.Diagnostic) string {
	code := diagCode(d)
	if code == "" {
		code = "(no code)"
	}
	return fmt.Sprintf("[%s %s] line %d: %s", d.Severity.String(), code, d.Range.Start.Line+1, d.Message)
}

func formatFailure(s docSnippet, failures []string, tol Tolerance) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (snippet #%d at line %d, id %s)\n", s.RelPath, s.Index, s.Line, s.Hash)
	if len(failures) == 1 {
		fmt.Fprintf(&b, "Got 1 unexpected result:\n")
	} else {
		fmt.Fprintf(&b, "Got %d unexpected results:\n", len(failures))
	}
	for _, f := range failures {
		fmt.Fprintf(&b, "  %s\n", f)
	}
	fmt.Fprintf(&b, "\nSnippet:\n")
	for _, line := range strings.Split(s.Content, "\n") {
		fmt.Fprintf(&b, "    %s\n", line)
	}
	fmt.Fprintf(&b, "\nTo accept, add to docSnippetTolerances:\n")
	fmt.Fprintf(&b, "    %q: {\n", s.ID)
	if tol.Skip {
		fmt.Fprintf(&b, "        Skip:   true,\n")
	} else {
		fmt.Fprintf(&b, "        // ExpectedCodes: []string{\"RADxxxxx\"},\n")
		fmt.Fprintf(&b, "        // MaxSeverity:   \"warning\",\n")
		fmt.Fprintf(&b, "        // Skip:          true,\n")
	}
	fmt.Fprintf(&b, "        Reason: \"TODO: explain why\",\n")
	fmt.Fprintf(&b, "    },\n")
	return b.String()
}
