package testing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amterp/rad/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var embeddedDocSections = map[string]bool{"Guide": true, "Reference": true, "Examples": true}

// TestDocsManifestConsistency verifies the embedded docs manifest and
// the embedded .md tree are in lockstep: every manifest page has a
// readable file, every file is in the manifest, and sections are the
// expected set. The nav->embedded drift (a new doc page that wasn't
// regenerated) is gated separately by `make verify-generated`, which
// re-runs gen-docs-embed and diffs.
func TestDocsManifestConsistency(t *testing.T) {
	manifest := core.GetDocsManifest()
	require.NotEmpty(t, manifest, "embedded docs manifest is empty - did gen-docs-embed run?")

	manifestSlugs := map[string]bool{}
	for _, p := range manifest {
		assert.Truef(t, embeddedDocSections[p.Section], "unexpected section %q for %q", p.Section, p.Slug)
		assert.NotEmptyf(t, p.Title, "page %q has no title", p.Slug)
		content, ok := core.GetDocPage(p.Slug)
		assert.Truef(t, ok, "manifest page %q has no embedded file", p.Slug)
		assert.NotEmptyf(t, content, "manifest page %q is empty", p.Slug)
		manifestSlugs[p.Slug] = true
	}

	// Every embedded .md on disk must be in the manifest (no orphans
	// left behind by a stale generation).
	root := "../embedded_docs"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		slug := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
		assert.Truef(t, manifestSlugs[slug], "embedded file %q is not in the manifest", slug)
		return nil
	})
	require.NoError(t, err)

	// Sanity: the core reference + entry pages are always present.
	for _, slug := range []string{"reference/syntax", "reference/functions", "reference/errors", "guide/basics"} {
		_, ok := core.GetDocPage(slug)
		assert.Truef(t, ok, "expected core page %q to be embedded", slug)
	}
}

// TestDocsTopicResolution verifies `rad docs <topic>` routing: error
// codes (with or without the RAD prefix) resolve to error docs, page
// slugs resolve to pages, and unknown topics miss. This is what makes
// `rad docs` the single entry point, error codes included.
func TestDocsTopicResolution(t *testing.T) {
	// Error codes route to error docs (the old `rad explain` behavior).
	doc, ok := core.GetDocTopic("RAD10001")
	require.True(t, ok, "RAD10001 should resolve")
	assert.Contains(t, doc, "RAD10001")

	docNoPrefix, ok := core.GetDocTopic("10001")
	require.True(t, ok, "10001 (no prefix) should resolve")
	assert.Equal(t, doc, docNoPrefix, "prefixed and unprefixed codes resolve identically")

	// Page slugs route to pages.
	_, ok = core.GetDocTopic("reference/functions")
	assert.True(t, ok, "reference/functions should resolve")

	// Unknown topics miss.
	_, ok = core.GetDocTopic("nonsense")
	assert.False(t, ok, "unknown topic should not resolve")
	_, ok = core.GetDocTopic("RAD99999")
	assert.False(t, ok, "unknown error code should not resolve")
}

// TestDocsFullCoversWholeCorpus verifies `rad docs all` inlines every
// embedded page (matching llms-full.txt) and leads with the TOC.
func TestDocsFullCoversWholeCorpus(t *testing.T) {
	toc := core.BuildDocsTOC()
	full := core.BuildDocsFull()

	for _, section := range []string{"## Guide", "## Reference", "## Examples"} {
		assert.Contains(t, toc, section, "TOC should list every embedded section")
	}
	assert.Contains(t, toc, "rad docs guide/basics", "TOC should show the per-page command")

	assert.True(t, strings.HasPrefix(full, toc), "full corpus should lead with the TOC")
	// The corpus is the TOC plus every page's inlined body, so it dwarfs
	// the index alone - cheap proof that content (not just the TOC) shipped.
	assert.Greater(t, len(full), 2*len(toc), "full corpus should inline page bodies, not just the TOC")
}
