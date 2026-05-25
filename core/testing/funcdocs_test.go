package testing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amterp/rad/rts"
)

// TestFuncDocsValid verifies every docs/funcs/*.md file in the
// embedded set parses cleanly through ParseFuncDoc. Catches
// authors landing a malformed doc that the hover layer would
// silently drop.
func TestFuncDocsValid(t *testing.T) {
	names := rts.FuncDocNames()
	if len(names) == 0 {
		t.Skip("no embedded func docs yet")
	}
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			doc := rts.GetFuncDoc(name)
			if doc == nil {
				t.Fatalf("FuncDocNames listed %q but GetFuncDoc returned nil", name)
			}
			if doc.Name != name {
				t.Errorf("name mismatch: stem=%q, doc.Name=%q", name, doc.Name)
			}
			if doc.Signature == "" {
				t.Errorf("%s: empty signature", name)
			}
			if !strings.Contains(doc.Signature, name+"(") {
				t.Errorf("%s: signature %q doesn't start with the function name",
					name, doc.Signature)
			}
			if len(doc.Examples) == 0 {
				t.Errorf("%s: no example code blocks", name)
			}
			if doc.Category == "" {
				t.Errorf("%s: empty category", name)
			}
		})
	}
}

// TestFuncDocsMatchRegisteredBuiltins verifies that every embedded
// doc names a function the runtime actually registers. Catches the
// reverse-drift case: a doc author renames the file to `say.md`
// while the runtime still exposes `print`.
//
// Note: the opposite assertion - every registered builtin has a
// doc - is intentionally NOT in this test yet. The doc migration
// is incremental; gating CI on 100% coverage would block landing
// any incremental work. The completeness assertion will land once
// the migration finishes.
func TestFuncDocsMatchRegisteredBuiltins(t *testing.T) {
	for _, name := range rts.FuncDocNames() {
		if _, ok := rts.FnSignaturesByName[name]; !ok {
			t.Errorf("doc exists for %q but no such builtin is registered", name)
		}
	}
}

// TestFuncDocsSignatureMatchesRegistered verifies the signature
// line in each doc matches the registered builtin's signature
// byte-for-byte. Catches the case where a doc author updates the
// signature in source without updating the doc (or vice versa).
func TestFuncDocsSignatureMatchesRegistered(t *testing.T) {
	for _, name := range rts.FuncDocNames() {
		doc := rts.GetFuncDoc(name)
		sig, ok := rts.FnSignaturesByName[name]
		if !ok {
			continue // covered by TestFuncDocsMatchRegisteredBuiltins
		}
		if doc.Signature != sig.Signature {
			t.Errorf("%s: doc signature %q != registered %q",
				name, doc.Signature, sig.Signature)
		}
	}
}

// TestFuncDocsSourceMatchesEmbedded verifies the source docs at
// docs/funcs/ are in sync with the embedded copy in
// rts/embedded_funcs/. The embedded files are the artefact the
// runtime actually reads; the docs/funcs/ tree is the canonical
// editable source.
//
// TODO(codegen): replace the manual mirror with a build step that
// reads docs/funcs/ and writes rts/embedded_funcs/. Until then,
// this test is the drift gate - if you edit one tree, edit the
// other or the test fails. Grep "TODO(codegen)" to find related
// sites.
func TestFuncDocsSourceMatchesEmbedded(t *testing.T) {
	sourceDir := "../../docs/funcs"
	embeddedDir := "../../rts/embedded_funcs"
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Skipf("source dir %s missing", sourceDir)
	}
	if _, err := os.Stat(embeddedDir); os.IsNotExist(err) {
		t.Skipf("embedded dir %s missing", embeddedDir)
	}

	sources, err := collectDocSet(sourceDir)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	embedded, err := collectDocSet(embeddedDir)
	if err != nil {
		t.Fatalf("read embedded: %v", err)
	}

	for name, content := range sources {
		if name == "README.md" {
			continue
		}
		emb, ok := embedded[name]
		if !ok {
			t.Errorf("%s exists in docs/funcs/ but not in rts/embedded_funcs/", name)
			continue
		}
		if string(content) != string(emb) {
			t.Errorf("%s differs between docs/funcs/ and rts/embedded_funcs/", name)
		}
	}
	for name := range embedded {
		if _, ok := sources[name]; !ok {
			t.Errorf("%s exists in rts/embedded_funcs/ but not in docs/funcs/", name)
		}
	}
}

// collectDocSet walks dir and returns md files keyed by their
// path relative to dir. Keying by relative path (not filename
// alone) matters once docs/funcs/ grows subdirectories - the
// README documents an `internal/` subfolder for _rad_* builtin
// docs, and a `print.md` in both top-level and `internal/`
// would silently collide under filepath.Base keying.
func collectDocSet(dir string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		// Normalize to forward slashes so the map keys round-trip
		// across platforms.
		rel = filepath.ToSlash(rel)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[rel] = content
		return nil
	})
	return out, err
}
