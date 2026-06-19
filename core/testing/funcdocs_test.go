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
			if marker := danglingNotesMarker(doc.Notes); marker != "" {
				t.Errorf("%s: Notes ends with a dangling bold label %q with no "+
					"content after it - drop the marker or fill it in", name, marker)
			}
		})
	}
}

// danglingNotesMarker returns the last line of a Notes section when
// that line is a lone bold label (e.g. "**Examples:**") with nothing
// following it, otherwise "". The docs/funcs source-of-truth
// migration left such empty markers behind, which the functions.md
// generator then rendered as empty headers (issue 128). This guard
// keeps the smell from recurring for any label, not just Examples.
func danglingNotesMarker(notes string) string {
	lines := strings.Split(strings.TrimRight(notes, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		last := strings.TrimSpace(lines[i])
		if last == "" {
			continue
		}
		if strings.HasPrefix(last, "**") && strings.HasSuffix(last, ":**") {
			return last
		}
		return ""
	}
	return ""
}

// TestParseFuncDocNormalizesCRLF guards the cross-platform bug where
// Windows git checks the embedded .md files out with CRLF, go:embed
// bakes the \r in, and the carriage returns leak into LSP hover output
// - mismatching snapshots generated on LF platforms.
func TestParseFuncDocNormalizesCRLF(t *testing.T) {
	src := strings.Join([]string{
		"# foo",
		"",
		"First line of the description.",
		"Second line.",
		"",
		"## Signature",
		"",
		"`foo() -> void`",
		"",
		"## Examples",
		"",
		"```rad",
		"foo()",
		"```",
		"",
		"## Category",
		"",
		"Test",
	}, "\n")
	crlf := strings.ReplaceAll(src, "\n", "\r\n")

	doc, err := rts.ParseFuncDoc("foo", crlf)
	if err != nil {
		t.Fatalf("CRLF doc failed to parse: %v", err)
	}

	fields := append([]string{doc.Description, doc.Signature, doc.Category, doc.Notes}, doc.Examples...)
	for _, f := range fields {
		if strings.Contains(f, "\r") {
			t.Errorf("carriage return survived parsing in field %q", f)
		}
	}
}

// TestFuncDocsMatchRegisteredBuiltins verifies that every embedded
// doc names a function the runtime actually registers. Catches the
// reverse-drift case: a doc author renames the file to `say.md`
// while the runtime still exposes `print`.
//
// We also assert the inverse: every public registered builtin has
// a docs/funcs/<name>.md. The migration that landed this gate
// covered all 113 then-existing builtins; new builtins added since
// must come with a doc. Internal _rad_* signatures are excluded -
// they live in docs/funcs/internal/ if at all.
func TestFuncDocsMatchRegisteredBuiltins(t *testing.T) {
	docs := make(map[string]struct{}, 0)
	for _, name := range rts.FuncDocNames() {
		docs[name] = struct{}{}
		if _, ok := rts.FnSignaturesByName[name]; !ok {
			t.Errorf("doc exists for %q but no such builtin is registered", name)
		}
	}
	for name, sig := range rts.FnSignaturesByName {
		if sig.IsInternal {
			continue
		}
		if _, ok := docs[name]; !ok {
			t.Errorf("registered builtin %q has no doc in docs/funcs/%s.md", name, name)
		}
	}
}

// TestFuncDocsSignatureMatchesRegistered verifies the signature
// line in each doc matches the registered builtin's signature
// byte-for-byte. Under the post-codegen pipeline,
// rts/signatures_gen.go is generated from docs/funcs/*.md - so a
// mismatch here means someone edited a .md without running
// `go generate ./rts` (or vice versa: edited signatures_gen.go
// directly, which they shouldn't). The test is the drift gate
// that catches stale codegen.
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
// The mirror is produced by `tools/gen-funcs-go` (run via
// `go generate ./rts`). This test is the drift gate that fires
// when someone edits one tree without regenerating the other.
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
		// The embedded copy carries a leading generated banner; strip it
		// to confirm the rest is a byte-for-byte mirror of source.
		if string(content) != rts.StripGeneratedBanner(string(emb)) {
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
