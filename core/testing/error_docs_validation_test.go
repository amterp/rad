package testing

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/amterp/rad/rts/check"
)

// TestErrorDocsRadBlocksValid is the regression lock against doc drift.
// It extracts every ```rad block from core/error_docs/*.md, runs the
// static checker, and asserts:
//
//  1. No block emits "Unexpected '#'" - the comment-syntax bug that
//     kept docs from teaching the very error code they document.
//  2. Blocks tagged exclusively as `// Correct` / `// Fix` / `// Works`
//     produce no error-severity diagnostics. (Catches the case where
//     a "fix" example is itself broken.)
//
// Mixed blocks (Wrong + Correct in the same block, common for
// before/after demos) skip the second check - they can't be cleanly
// classified, and the diagnostics they emit are intentional.
// Untagged blocks are illustrative and only get the "# regression"
// assertion. We don't strictly require Wrong blocks to fire because
// many error codes are runtime-only and have no static detector.

type radBlock struct {
	file    string
	lineNum int // 1-based line where ```rad started
	content string
	tag     blockTag
}

type blockTag int

const (
	tagUntagged blockTag = iota
	tagWrong    // contains Wrong marker only
	tagCorrect  // contains Correct marker only
	tagMixed    // contains both - skip the strict check
)

var (
	hashCommentRE = regexp.MustCompile(`Unexpected '#'?`)

	wrongMarkerRE   = regexp.MustCompile(`(?i)//\s*(wrong|bad|error[: ])`)
	correctMarkerRE = regexp.MustCompile(`(?i)//\s*(correct|fix|works?|right|good)`)
)

func extractRadBlocks(path, src string) []radBlock {
	var blocks []radBlock
	lines := strings.Split(src, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		if strings.HasPrefix(strings.TrimSpace(line), "```rad") {
			start := i + 1
			i++
			var contentLines []string
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				contentLines = append(contentLines, lines[i])
				i++
			}
			content := strings.Join(contentLines, "\n")
			blocks = append(blocks, radBlock{
				file:    path,
				lineNum: start,
				content: content,
				tag:     classifyBlock(content),
			})
		}
		i++
	}
	return blocks
}

func classifyBlock(content string) blockTag {
	// Markers can appear either as their own comment line or inline
	// after code (e.g. `x = "hi"  // Wrong`). If both markers appear,
	// the block is a before/after demo and we can't apply the strict
	// "must compile" rule.
	hasWrong := wrongMarkerRE.MatchString(content)
	hasCorrect := correctMarkerRE.MatchString(content)
	switch {
	case hasWrong && hasCorrect:
		return tagMixed
	case hasWrong:
		return tagWrong
	case hasCorrect:
		return tagCorrect
	}
	return tagUntagged
}

func runCheckOnBlock(t *testing.T, src string) check.Result {
	t.Helper()
	checker, err := check.NewChecker()
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	checker.UpdateSrc(src)
	result, err := checker.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	return result
}

func TestErrorDocsRadBlocksValid(t *testing.T) {
	errorDocsDir := "../error_docs"
	if _, err := os.Stat(errorDocsDir); os.IsNotExist(err) {
		t.Skipf("error_docs directory not found: %s", errorDocsDir)
	}

	allBlocks := collectBlocks(t, errorDocsDir, func(name string) bool {
		return isErrorDocFile(name)
	})
	if len(allBlocks) == 0 {
		t.Fatal("no rad blocks discovered - extraction is broken")
	}

	for _, blk := range allBlocks {
		blk := blk
		name := strings.TrimPrefix(blk.file, errorDocsDir+"/") + ":" + itoa(blk.lineNum)
		t.Run(name, func(t *testing.T) {
			result := runCheckOnBlock(t, blk.content)

			// Assertion 1: no spurious # comment parse errors.
			for _, d := range result.Diagnostics {
				if hashCommentRE.MatchString(d.Message) {
					t.Errorf("block emits Unexpected '#' - did `#` slip back into a rad code block?\n  file: %s:%d\n  diag: %s\n  block:\n%s",
						blk.file, blk.lineNum, d.Message, indent(blk.content))
				}
			}

			// Assertion 2/3: wrong vs correct semantics.
			hasError := false
			for _, d := range result.Diagnostics {
				if d.Severity == check.Error {
					hasError = true
					break
				}
			}
			// Only the Correct-only case carries a strict assertion.
			// Wrong-only blocks may document runtime-only errors that
			// the static checker can't reach. Mixed blocks combine
			// both halves and intentionally emit diagnostics.
			if blk.tag == tagCorrect && hasError {
				var msgs []string
				for _, d := range result.Diagnostics {
					if d.Severity == check.Error {
						msgs = append(msgs, codeStr(d.Code)+": "+d.Message)
					}
				}
				t.Errorf("block tagged // Correct but produced error diagnostics\n  file: %s:%d\n  errors: %s\n  block:\n%s",
					blk.file, blk.lineNum, strings.Join(msgs, "; "), indent(blk.content))
			}
		})
	}
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = "    " + ln
	}
	return strings.Join(lines, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// collectBlocks walks a docs directory, extracting rad blocks from
// every .md file that matches `accept(name)`.
func collectBlocks(t *testing.T, dir string, accept func(name string) bool) []radBlock {
	t.Helper()
	var out []radBlock
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if info.IsDir() || !strings.HasSuffix(base, ".md") || !accept(base) {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out = append(out, extractRadBlocks(path, string(src))...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return out
}

// TestDocsWebRadBlocksValid mirrors the error_docs test for the
// public-facing docs under docs-web/. Only the "# comment" regression
// assertion is applied - docs-web blocks are heavily illustrative and
// classifying them as Wrong / Correct would be too noisy.
func TestDocsWebRadBlocksValid(t *testing.T) {
	docsDir := "../../docs-web/docs"
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		t.Skipf("docs-web/docs directory not found: %s", docsDir)
	}

	blocks := collectBlocks(t, docsDir, func(name string) bool {
		return true
	})
	if len(blocks) == 0 {
		t.Fatal("no rad blocks discovered under docs-web - extraction is broken")
	}

	for _, blk := range blocks {
		blk := blk
		rel := strings.TrimPrefix(blk.file, docsDir+"/")
		t.Run(rel+":"+itoa(blk.lineNum), func(t *testing.T) {
			result := runCheckOnBlock(t, blk.content)
			for _, d := range result.Diagnostics {
				if hashCommentRE.MatchString(d.Message) {
					t.Errorf("block emits Unexpected '#' - did `#` slip back into a rad code block?\n  file: %s:%d\n  diag: %s\n  block:\n%s",
						blk.file, blk.lineNum, d.Message, indent(blk.content))
				}
			}
		})
	}
}

func codeStr(c interface{}) string {
	type stringer interface{ String() string }
	if s, ok := c.(stringer); ok {
		return s.String()
	}
	return "?"
}
