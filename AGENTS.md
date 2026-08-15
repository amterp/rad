# AGENTS.md

This file provides guidance for AI agents when working with code in this
repository.

## What This Is

Rad is a CLI-first scripting language (Python-like syntax) with declarative arg
parsing and built-in JSON, HTTP, and interactive-prompt support. This repo is a
single Go module containing the `rad` interpreter, the `radls` language server,
and supporting tooling.

Key orientation docs:

- `SYNTAX.md` (repo root) - comprehensive Rad language reference, written for
  dev sessions. Read it before writing or modifying any Rad code (`.rad`
  files, embedded scripts, snapshot test inputs), and update it when changing
  language syntax or semantics.
- `NAVIGATE.md` - maps this repo and sibling repos (tree-sitter-rad grammar,
  go-tbl, homebrew-rad).
- `docs/red/` - Rad Evolution Documents (REDs): decision records for
  significant design choices. Check them before revisiting a settled design;
  significant new decisions should get one.
- Directory-specific agent guides take precedence in their subtrees:
  `core/error_docs/AGENTS.md`, `rts/radfmt/AGENTS.md`,
  `docs-web/docs/guide/AGENTS.md`, `docs-web/docs/reference/AGENTS.md`.

## Commands

```sh
make all              # generate + format + build + test (the full local loop)
make build            # build dev binary to bin/radd
make test             # go test ./core/testing/... ./rts/... ./radls/lstesting/...
make format           # gofmt + goimports (run before committing)
make generate         # run all codegen (see Generated Code below)
./dev --validate      # go mod tidy + make all, via the dev Rad script
```

The dev binary is `bin/radd`, not `rad` - a bare `rad` invokes the installed
release. Test local changes with `./bin/radd script.rad`. Useful debugging
flags: `--src`, `--cst-tree`, `--ast-tree`, `--rad-debug`,
`--mock-response pattern:file.json`.

Run a single test the standard Go way:

```sh
go test ./core/testing/ -run TestName
```

### Snapshot tests

Snapshot suites are the dominant test style - prefer adding a snapshot case
over bespoke Go assertions when either would do. Five suites:
`core/testing/snapshots/` (end-to-end script runs), `rts/test/st_snapshots/`
(syntax trees), `rts/check/snapshots/` (checker diagnostics),
`rts/radfmt/snapshots/` (formatter), `radls/lstesting/snapshots/` (LSP).

The engine is go-snap (`github.com/amterp/go-snap`); each suite declares its own
sections next to its runner, so the section names differ by suite. `.snap` files
hold multiple cases delimited by `### TITLE ###` and `### INPUT ###` (Rad
source), then:

| Suite | Sections |
|---|---|
| `core/testing` | `ARGS` / `RAW_ARGS` / `STDIN` / `TERM_WIDTH` / `KEYS` in; `STDOUT` / `STDERR` / `FRAMES` / `EXIT` out |
| `rts` (syntax trees) | `CST` (tree-sitter dump) and `AST` out |
| `rts/check` | `ARGS` in; `CHECK` (binder + type checker dump) out |
| `rts/radfmt` | `FORMATTED` out |
| `radls/lstesting` | repeatable action headers (`### HOVER 1:6 ###`, ...); `STDOUT` out |

`INPUT` is the script a case runs; `STDIN` is what that run reads. A `rad repl`
case is all `STDIN` and no `INPUT`.

A header naming a section its suite never declared is a parse error, not
content. An absent output section asserts that channel is empty, so accepting an
update deletes any section it can reproduce. Mark a section `[raw]` (e.g.
`### INPUT [raw] ###`) to store it as a Go-quoted string compared byte for byte,
which is how the formatter's line-ending and trailing-whitespace cases work.

When you intentionally change behavior, update snapshots scoped to what you
changed:

```sh
# Rewrite only snapshot files whose path contains the substring(s):
go test ./core/testing/ -run TestSnapshots -update=types/str_lexing
go test ./core/testing/ -run TestSnapshots -update=errors/validation,control_flow

# Same, but reaching every package in one run:
SNAP_UPDATE=types/str_lexing go test ./...

# Blanket rewrite (use sparingly): ./dev --update, or per-package -update-all
```

Under `-update=...`, mismatches in non-targeted files still fail, so unrelated
regressions can't be silently baked in. Write the value with `=`: a bare
`-update -run X` swallows `-run` as the update value (go-snap refuses a target
starting with `-` for exactly this reason). Always review the `.snap` diff
afterward.

Prefer `SNAP_UPDATE` for anything wider than one package: `-update` is only
registered by packages that link go-snap, so `go test ./... -update=x` fails on
the ones that don't, while the environment variable reaches all of them.

Two things to know about accepting an update. It is refused when `CI` is set,
unless `SNAP_UPDATE_CI=1`. And it rewrites a file whenever go-snap's canonical
rendering differs from what is on disk, not only when an expectation changed -
so `-update-all` doubles as a formatter, and a format change can be migrated
across every existing snapshot in one pass.

### CI

PR checks run `make verify-generated` (fails on stale codegen output) and the
test suite on Linux, macOS, and Windows. For faster platform-specific
feedback: `gh workflow run "Quick Platform Tests" --ref <branch>
[-f platforms=windows]`.

## Architecture

Subsystems within the one Go module:

- `core/` - interpreter and CLI. One big package for historical reasons; new
  code should be packaged appropriately where possible.
- `rts/` - "Rad Tree Sitter": wraps the tree-sitter-rad grammar. Owns the AST
  (`rts/rl/`), the static checker (`rts/check/`), and the formatter
  (`rts/radfmt/`). Consumed by both `core` and `radls`; **rts must not import
  core** (it sits upstream).
- `radls/` - LSP server; `radls/analysis/` is the feature layer, built on rts.
- `docs-web/` - MkDocs site; `tools/` - codegen generators; `vsc-extension/` -
  VS Code extension; `ci/` - Rad scripts run by GitHub Actions.

The `.rad` scripts in `ci/` are executed by the **released** `rad`, not by the
build under test - so they must stay compatible with the last release, not with
HEAD. A script written against unreleased semantics passes `make test` locally
and fails CI. Verify changes to them with the installed `rad`, not `./bin/radd`.

`benchmark/` is different: nothing in CI runs it. It is a macOS-only harness you
run by hand, and it measures `./bin/radd`, so it tracks HEAD.

### Execution pipeline

`main.go` → `core.RadRunner` (`core/runner.go`): reads the script, parses via
rts to a tree-sitter CST, converts it to the AST (`rts.ConvertCST` →
`rl.SourceFile`), runs the static checker (`rts/check`) and exits on errors,
registers script args and command blocks as CLI flags via the `ra` library,
then hands off to `core.Interpreter` (`core/interpreter.go`), a tree-walking
evaluator. `eval(node)` returns `EvalResult{RadValue, Ctrl}`, carrying both
value and control flow (break/continue/return).

`rad repl` (`core/repl*.go`) is the other host: it drives the same interpreter a
turn at a time. Ending execution is a policy rather than a hardcoded exit -
`RadExitHandler.SetUnwinding` makes a fatal diagnostic raise `*RadAbort` instead
of exiting, which the REPL catches at the turn boundary. Anything else that wants
to run Rad without owning the process uses that seam.

Runtime values are `RadValue` (`core/type_rad_value.go`), a tagged union over
Go types (`int64`, `float64`, `RadString`, `*RadList`, `*RadMap`, `RadFn`,
`*RadError`, ...). Per-type dispatch uses the visitor in
`core/type_visitor.go`. Adding a runtime type means touching `Type()`,
`newRadValue`, the visitor, equality/hash, and the `rl.RadType` enum - the
`panic("Bug! ...")` defaults surface missed spots loudly.

### Testability seams (core/global.go)

All IO and environment access goes through injectable globals (`RIo`,
`RClock`, `RShell`, `RReq` for HTTP, `RSleep`, `RSignal`, `RInteractive`,
`RTerminal`, `RExit`, ...), wired from `RunnerInput`. The tests in
`core/testing/` inject fakes for all of them and run the real runner end to
end, asserting on captured stdout/stderr/exit code (`setupAndRunCode`,
`assertOnlyOutput`, `assertError`, `NewTestParams(...).Keys(...)` for
interactive prompts, `NewTestParams(...).NoTerminal()` for the no-terminal
paths, ...). Don't bypass these seams with direct `os.*` / `time.*` / `exec.*`
calls in interpreter code paths - it breaks both testability and determinism.

`RTerminal` is the single authority on whether rad can prompt at all: it
probes stdin, then the controlling terminal (`/dev/tty`, `CONIN$` on Windows).
Both the prompt driver and the pre-flight guard ask it, so they can never
disagree. Note `com.IsTty` is NOT this - it inspects stdout and exists only for
cosmetic rendering choices.

### Built-in functions: docs are the source of truth

Function signatures are generated from Markdown, not written in Go. To add a
built-in:

1. Write `docs/funcs/<name>.md` following the format contract in
   `docs/funcs/README.md`. The `## Signature` line IS the type-checker's
   signature - there is no parallel definition in Go.
2. Run `make generate` to regenerate signatures and doc mirrors.
3. Add the `FUNC_X` const and `BuiltInFunc{Name, Execute}` registration in
   `core/funcs.go` (larger ones get their own `core/func_<name>.go`). The
   `Execute` body pulls pre-bound args via `f.GetStr(...)` etc. and returns
   via `f.Return(...)` / `f.ReturnErrf(...)`.
4. Add snapshot cases under `core/testing/snapshots/functions/`.

`function-metadata/extract.go` exists because rts can't import core: it dumps
the runtime function registry to `rts/embedded/functions.txt` for the checker's
did-you-mean suggestions. `make generate` handles it.

Code snippets in the user docs are run through `rad check` by tests
(`core/testing/docs_snippets_test.go`), so docs edits can fail the suite. They
are checked, never executed - a snippet may shell out or hit the network
without CI doing either. A snippet that is *meant* to be invalid (error docs
are full of them) needs an entry in `core/testing/docs_snippet_tolerances.go`
naming the codes it may emit.

### Error codes

Codes live in `rts/rl/errors.go`, banded: 1xxxx syntax, 2xxxx runtime, 3xxxx
types, 4xxxx validation/lint. Numbers are never reused (tombstone rule:
retired codes keep their number with a `_retired` suffix). Each code gets
user-facing docs at `core/error_docs/<code>.md` (style guide in that dir's
AGENTS.md), surfaced by `rad docs RADxxxxx`. The interpreter has two failure
mechanisms: `RadPanic` for catchable, data-driven errors user code can
`catch`, and `emitError*` for hard exits on programming errors.

Both `rad check` and runtime errors render through `core/diagnostic_render.go`,
which fits output to the terminal - prose caps at 100 columns, the source
snippet uses the full width. Anything else that hand-formats error text must
wrap with `com.Wrap` (`core/common/text.go`) to `core.DiagnosticProseWidth()`
rather than picking its own width; `core/preflight.go` is the worked example.
Keep inline messages to one sentence and leave the explanation to the error
doc - the budget and its rationale are in `core/error_docs/AGENTS.md`. Hints
shared between the checker and the interpreter live in `rts/rl/hints.go` so the
two can't drift.

### Embedded commands are Rad scripts

`rad`'s own subcommands (`new`, `check`, `docs`, `fmt`, `stash`, `explain`,
...) are Rad scripts in `core/embedded/`, compiled into the binary and backed
by internal `_rad_*` functions (`core/funcs_internal.go`; their signatures are
hand-listed in `rts/signatures.go`). Dogfooding: language changes can affect
the CLI itself.

## Generated Code - Never Hand-Edit

`rts/signatures_gen.go`, `rts/embedded_funcs/`, `rts/embedded/functions.txt`,
`docs-web/docs/reference/functions.md`, `docs-web/docs/reference/errors.md`,
`core/embedded_docs/`. Edit the sources instead (`docs/funcs/*.md`,
`core/error_docs/*.md`, docs-web pages) and run `make generate`. CI's
`verify-generated` blocks stale output; when adding a generator, append its
output path to `GENERATED_PATHS` in the Makefile.

## Breaking Changes

Rad is pre-1.0 and breaks compatibility in minor versions, but every breaking
change ships with migration help on three layers:

1. A **migration diagnostic** that detects the old pattern and emits a hint
   linking to `https://amterp.dev/rad/migrations/v0.X/` (e.g. a renamed
   function's old name gets an `emitErrorWithHint` case). Prefer static
   detection in `rts/check/` where possible, so editors surface it too.
2. An **error doc** in `core/error_docs/<code>.md` with a before/after example
   and fix steps.
3. A **migration guide entry** in `docs-web/docs/migrations/v0.X.md` with the
   full context and rationale.

Use `feat!:` / `fix!:` commit prefixes for breaking changes.

## Conventions

- Conventional commit prefixes (`feat:`, `fix:`, `docs:`, `refactor:`,
  `test:`). Commit messages explain the why; `git blame` is treated as
  documentation here.
- Keep this file current: if a change alters the dev workflow, project
  structure, or invalidates anything stated here, update AGENTS.md in the
  same commit. Stale agent guidance is worse than none - it gets followed.
- Cross-platform: rad targets Linux, macOS, and Windows. Paths returned to
  user code are normalized to forward slashes via `NormalizePath`; all
  platform-specific behavior is centralized in `core/common/platform.go` - no
  scattered `runtime.GOOS` checks. `NormalizeLineEndings` applies only to Rad
  source, never to user data (round-trip safety).
- Dependencies under `github.com/amterp/` are first-party - `ra` (CLI arg
  parsing), `radish` (interactive prompts), `tree-sitter-rad` (grammar),
  `go-tbl` (tables), and smaller utils. Treat them as extensions of this
  codebase, not immutable third-party code: they exist largely to serve Rad,
  and a workaround in Rad is debt filed in the wrong repo. If a change is
  fighting one of them - wrapping an awkward API, duplicating logic the lib
  should own, patching around a bug traced into the lib - the right fix is
  usually upstream. Say so and make it properly in the lib (checkouts live at
  `~/src/<lib>`; iterate via a replace directive) instead of working around
  it in Rad. Tagging and releasing lib versions stays with the user.
- Never commit `replace` directives in go.mod. They're for local development
  only (e.g. a temporary `replace github.com/amterp/tree-sitter-rad => <local
  path>` while iterating on the grammar, or a gitignored go.work); remove them
  before committing. `./dev --release` also refuses to release with them
  present.
- Don't leave task-specific messages to the user as comments in the code -
  put them in your response instead.
- After user-facing changes (new/changed functions, syntax, errors, CLI
  behavior), invoke the Rad Docs Maintainer subagent to assess whether docs
  need updating.
- `AI_POLICY.md` governs AI-assisted external contributions (maintainers are
  exempt).
