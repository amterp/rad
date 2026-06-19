// gen-funcs-go mirrors `docs/funcs/*.md` (the source-of-truth
// builtin function docs) into `rts/embedded_funcs/*.md` so the
// runtime's `//go:embed embedded_funcs/*.md` directive picks them
// up. The embedded copy is what the LSP hover layer reads via
// `rts.GetFuncDoc`.
//
// Before this tool existed the two trees were kept in sync by
// hand and the drift gate test in `core/testing/funcdocs_test.go`
// failed when authors updated only one side. Now contributors
// edit `docs/funcs/` and run this tool (via `go generate ./rts`).
//
// Behaviour:
//   - Reads `docs/funcs/*.md` (skipping README.md, the `internal/`
//     subdirectory, and any stem that isn't a valid Rad identifier
//     per the loader's `isValidFuncDocStem` rule).
//   - Parses each file through `rts.ParseFuncDoc` so a malformed
//     doc fails loudly here, before it can land in the embedded
//     tree and break runtime hover.
//   - Writes the verified files byte-for-byte into
//     `rts/embedded_funcs/`. Files no longer present in the
//     source are removed from the embedded tree so the two stay
//     in lockstep.
//
// Usage:
//
//	go run ./tools/gen-funcs-go
//
// Defaults assume invocation from the repo root. Pass `-source`
// and `-target` to override.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amterp/rad/rts"
)

func main() {
	var (
		source = flag.String("source", "docs/funcs", "path to the source docs/funcs/ directory")
		target = flag.String("target", "rts/embedded_funcs", "path to the target rts/embedded_funcs/ directory")
		dryRun = flag.Bool("dry-run", false, "print planned actions without writing")
	)
	flag.Parse()

	if err := run(*source, *target, *dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "gen-funcs-go: %v\n", err)
		os.Exit(1)
	}
}

func run(source, target string, dryRun bool) error {
	sourceFiles, err := collectFuncDocs(source)
	if err != nil {
		return fmt.Errorf("scanning source: %w", err)
	}
	if len(sourceFiles) == 0 {
		return fmt.Errorf("no func docs found under %s", source)
	}

	// Parse-validate each before staging - we won't ship a broken
	// doc into the embedded tree where the loader would silently
	// skip it (or worse, panic).
	for _, src := range sourceFiles {
		body, err := os.ReadFile(src.path)
		if err != nil {
			return fmt.Errorf("read %s: %w", src.path, err)
		}
		if _, err := rts.ParseFuncDoc(src.stem, string(body)); err != nil {
			return fmt.Errorf("validate %s: %w", src.path, err)
		}
		src.body = body
	}

	if !dryRun {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", target, err)
		}
	}

	// The embedded copy self-announces as generated so opening it
	// locally points the reader back at docs/funcs/. The runtime loader
	// (rts.StripGeneratedBanner) and the drift gate strip it, so the
	// tree stays a byte-for-byte mirror of source modulo this line.
	banner := []byte(rts.GeneratedBanner("tools/gen-funcs-go", "docs/funcs/") + "\n")

	wrote := 0
	for _, src := range sourceFiles {
		dst := filepath.Join(target, src.stem+".md")
		body := append(append([]byte{}, banner...), src.body...)
		if dryRun {
			fmt.Printf("would write %s\n", dst)
			continue
		}
		if existing, err := os.ReadFile(dst); err == nil && string(existing) == string(body) {
			continue // already in sync, don't bump mtime
		}
		if err := os.WriteFile(dst, body, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dst, err)
		}
		wrote++
	}

	// Remove embedded files that no longer exist in source.
	removed, err := pruneTarget(target, sourceFiles, dryRun)
	if err != nil {
		return err
	}

	fmt.Printf("gen-funcs-go: %d source files; %d updated; %d pruned\n",
		len(sourceFiles), wrote, removed)
	return nil
}

type funcDocFile struct {
	stem string
	path string
	body []byte
}

func collectFuncDocs(dir string) ([]*funcDocFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]*funcDocFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue // internal/ etc. skipped
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		stem := strings.TrimSuffix(name, ".md")
		if !rts.IsValidFuncDocStem(stem) {
			// README.md, scratch.txt, etc. excluded by the same
			// rule the runtime loader uses.
			continue
		}
		out = append(out, &funcDocFile{
			stem: stem,
			path: filepath.Join(dir, name),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].stem < out[j].stem })
	return out, nil
}

func pruneTarget(target string, sourceFiles []*funcDocFile, dryRun bool) (int, error) {
	entries, err := os.ReadDir(target)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	want := make(map[string]struct{}, len(sourceFiles))
	for _, sf := range sourceFiles {
		want[sf.stem+".md"] = struct{}{}
	}
	removed := 0
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
			continue // contributor notes in the embedded tree - leave alone
		}
		if _, keep := want[name]; keep {
			continue
		}
		path := filepath.Join(target, name)
		if dryRun {
			fmt.Printf("would remove %s\n", path)
			continue
		}
		if err := os.Remove(path); err != nil {
			return removed, fmt.Errorf("prune %s: %w", path, err)
		}
		removed++
	}
	return removed, nil
}
