# radfmt - contributor guide

`radfmt` implements `rad fmt`, a gofmt-style canonical re-printer for Rad
scripts. It is **rule-based**: every formatting decision is a numbered rule in
[`RULES.md`](./RULES.md), which is the source of truth. The code enforces the
rules, the snapshots demonstrate them, and `rules_test.go` keeps all three in
lockstep.

## Vision

The north star: **any syntactically valid Rad script, however it was originally
written, formats to a single canonical output.** `rad fmt` is maximally
opinionated - it does not nudge your whitespace, it discards your layout and
regenerates it. Two different layouts of the same code converge to the same bytes.

We are not there yet, and the gap is **coverage, not architecture**. A construct
with no dedicated rule falls through to `verbatim()`, which re-emits its exact
source span. That is the one thing that breaks convergence today (the block
constructs - `args:`, `rad`, `fn`, ... - are still verbatim). `verbatim()` is
load-bearing *scaffolding*: it keeps the formatter safe and lets rules grow
incrementally, and its footprint should shrink toward zero as rules graduate. The
prioritized path from here is the **Roadmap** in [`RULES.md`](./RULES.md).

What stays verbatim *forever* is only the deliberate, semantic-preserving
passthroughs - the "within reason" boundary where canonicalizing would change
meaning:
- Literal value text - `3.140` is never rewritten to `3.14` (F33).
- Map key form - a bareword key and a quoted key are semantically distinct in Rad,
  so we never convert between them (F28).
- Shebang and `--- ... ---` header - free text, emitted as-is (F34/F35).

When weighing a new rule, the default is **more opinionated**: if a difference is
non-semantic, canonicalize it. Reach for passthrough only when the change would
alter meaning (then it's a `passthrough` rule, locked by a test) or rewrites token
*content* rather than structure (then it needs its own value-preservation guard -
see the safety-model note below on why F30 is deferred).

## The golden rule

**Behavior and `RULES.md` move together.** If you change what the formatter
emits, you are changing a rule - update `RULES.md` in the same change. The
coverage test fails the build if you don't.

## To add, change, or remove a rule

1. **`RULES.md`** - add or edit the rule. New rules take the next free `Fn`
   (IDs are append-only - never renumber, never reuse). A removed rule becomes a
   tombstone in the Changelog; its number is retired. Keep the heading shape
   exact: `### Fn - Title \`status\`` (the status is the only backticked word -
   nothing may follow it, or the coverage test rejects the heading). Place the
   rule in the topical section it belongs to, not in numeric order, so IDs read
   out of sequence in the file (F36 sits among the comment rules, F31 among the
   statements). That's expected: the section gives context, the ID is just
   append-order.
2. **Code** - implement it and tag the enforcing site with `// [Fn]`. One tag
   per rule is enough; put it where the decision lives.
3. **Snapshot** - add or update a case under `snapshots/` whose title contains
   `[Fn]` (e.g. `[F27] List spacing`). A case may demonstrate several rules:
   `[F12][F13] ...`. Author the messy input and the canonical output. The
   coverage test only checks the `[Fn]` tag is present in some title - it can't
   tell whether the case actually exercises the rule. Write one that genuinely
   does: for an `implemented` rule, the input must differ from the output, or the
   snapshot proves nothing.
4. **Run** `go test ./rts/radfmt/`. Regenerate snapshot outputs after an
   intentional behavior change with `-run TestFmtSnapshots -update`, then read
   the diff before committing.

### Status values (drive enforcement - see the legend in `RULES.md`)

| status        | needs code tag | needs snapshot | meaning                                  |
|---------------|:---:|:---:|------------------------------------------|
| `implemented` | yes | yes | active canonicalizing rule               |
| `passthrough` | yes | yes | deliberate non-action, locked by a test  |
| `limitation`  | optional | no | known gap to close later               |
| `deferred` / `roadmap` | no | no | decided/planned, not built          |

### Byte-level rules and `[raw]` snapshots

Rules about characters the line-based snapshot text can't hold - CR bytes
(the scanner strips them), trailing whitespace (editors strip it), the exact
trailing newline - still get real snapshots. Mark the case title `[raw]` and
write its `### INPUT ###` and `### STDOUT ###` as a single Go-quoted string:

```
### TITLE ###
[F2][raw] CRLF and bare CR normalized to LF
### INPUT ###
"x = 1\r\ny = 2\r\n"
### STDOUT ###
"x = 1\ny = 2\n"
```

The harness decodes both with `strconv.Unquote` and compares byte-for-byte (no
trailing-newline trim). Every rule is a snapshot; there is no unenforced status.

## Safety model (do not weaken)

Three layers guarantee `rad fmt` can never corrupt code:

1. **Parse-error no-op** - a tree with invalid nodes is returned unchanged
   (`ok=false`).
2. **Panic recovery** - a panic during formatting degrades to a safe no-op.
3. **Structural-equivalence guard** - the output is re-parsed and its
   named-node + comment structure compared to the input; a mismatch discards the
   formatting and returns the original.

A construct with no dedicated formatter falls through to `verbatim`, which
re-emits its exact source span - always safe, just not canonicalized. This is
what lets us grow rules incrementally. **The structural guard compares node
kinds, not text**, so a rule that rewrites token *content* (e.g. string quote
normalization, F30) needs its own value-preservation check and tests - that's
why F30 is deferred.

## Layout

| file | role |
|------|------|
| `doc.go`, `render.go` | the Doc IR + Wadler render machine (the engine) |
| `printer.go` | `format(node)` dispatch + `formatSeq` + `verbatim` |
| `print_stmt.go`, `print_expr.go`, `print_lit.go` | construct formatters |
| `cst.go`, `safety.go` | CST helpers and the structural-equivalence guard |
| `RULES.md` | the rule catalog (source of truth) |
| `rules_test.go` | coverage test linking spec ↔ code ↔ snapshots |
| `snapshots/*.snap` | `### INPUT ###` / `### STDOUT ###` cases, titled by rule |
| `DESIGN.md` | design-time research (Doc IR + comment attachment theory) |
