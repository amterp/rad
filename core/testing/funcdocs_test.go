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
// editable source. Manual sync today; codegen later.
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

func collectDocSet(dir string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if info.IsDir() || !strings.HasSuffix(base, ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[base] = content
		return nil
	})
	return out, err
}
