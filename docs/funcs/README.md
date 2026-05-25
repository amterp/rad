# `docs/funcs/`: source of truth for built-in function docs

These markdown files are the canonical documentation for Rad's
built-in functions. Each file describes exactly one function. Two
downstream consumers read this directory at build time:

1. `tools/gen-funcs-go/main.go` mirrors `docs/funcs/*.md` into
   `rts/embedded_funcs/`. The runtime's `//go:embed
   embedded_funcs/*.md` directive picks them up, and the LSP hover
   layer reads them via `rts.GetFuncDoc` to show a function's
   description alongside its signature. Run via
   `go generate ./rts` (or `go run ./tools/gen-funcs-go`).
2. `tools/gen-funcs-page/main.go` regenerates
   `docs-web/docs/reference/functions.md`, the public-facing
   functions reference. The aggregate page stays a derived artifact -
   editing it directly invites drift.

Treat these `.md` files as the single source. Editing the generated
artifacts directly is a one-way ticket to mismatched docs.

## File naming

One file per function, named `<fn>.md` where `<fn>` is the function's
name in Rad source (`print.md`, `range.md`, `parse_int.md`). The
codegen skips this `README.md` explicitly and any file whose stem
doesn't match the identifier rule `[a-z_][a-z0-9_]*`, so contributor
notes (`scratch.txt`, `2025-plan.md`) won't be picked up.

Internal `_rad_*` builtins do not belong here. Add their docs in the
`docs/funcs/internal/` subdirectory if they need any documentation
at all - the codegen ignores that path so the public surface stays
clean.

## Required sections

Every file must contain these sections in this order:

```markdown
# <fn>

Short one-paragraph description.

## Signature

`<fn>(...) -> <return_type>`

## Parameters

- `param_name` (`type`): description

## Examples

\`\`\`rad
example_code()
\`\`\`

## Category

<one word>
```

`# <fn>` is the H1 title. The function name has to match the file
stem - codegen rejects mismatches.

`## Signature` holds exactly one line of inline-code: the function's
signature in `signatures.go`'s syntax. The codegen parses this
through the same signature parser as the registered fns, so a typo
fails the doc test.

`## Parameters` lists each positional / keyword parameter. Order
matches the signature.

`## Examples` holds one or more rad code blocks. The first block is
what hover renders inline; later blocks appear in the public docs
page.

## What hover renders

The LSP hover for a built-in shows:

1. The signature line (always).
2. A horizontal rule.
3. The H1 description paragraph (only).
4. The first `## Examples` rad code block.

The `## Parameters`, `## Notes`, and `## See also` sections do NOT
appear in hover - they're for the public docs page. If a parameter
needs to be discoverable from hover (e.g. a confusing `_arg1`
overload), describe it in the H1 description instead of relying on
the Parameters list.

`## Category` is a single word the public docs use to group
functions ("io", "strings", "lists", "math", "time", "random",
"shell", "system").

## Optional sections

- `## Notes` - call out edge cases or related concepts.
- `## See also` - link to related fns by name.

## Tests

`core/testing/funcdocs_test.go` validates the doc set on every test run:

- Every `.md` parses cleanly into the structured shape above.
- Every signature line parses through the runtime's signature parser.
- The `docs/funcs/` and `rts/embedded_funcs/` trees are byte-for-byte
  identical (drift gate against editing only one side - the codegen
  in `tools/gen-funcs-go` keeps them in sync).
- Every registered builtin has a `.md`. New builtins without a doc
  file fail the build.

`make verify-generated` is the additional gate at the CI layer: it
snapshots every generated artifact, reruns `make generate`, and
fails if anything changed. That catches stale `signatures_gen.go`,
embedded docs, or the public reference page.
