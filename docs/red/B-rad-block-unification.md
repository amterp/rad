---
red: B          # provisional - to be renumbered to its chronological slot
title: Unify the rad block keywords
status: Implemented
kind: Language
created: 2026-06-04
decided: 2026-06-04
released: v0.9.0
supersedes: A
superseded-by:
related: 3
---

# RED-B: Unify the rad block keywords

## Summary

v0.9.0 collapsed the three rad block keywords (`rad`/`request`/`display`) back into a single
`rad` keyword that **dispatches on its source type at runtime** - a URL fetches, a list/map
extracts in-memory, and no source operates on existing variables. Whether the block prints
became an orthogonal `noprint` option. This reversed the three-keyword split (RED-A), which
had grown confusing.

## Context / Motivation

The three keywords conflated two genuinely independent concerns:

- **Where the data comes from** - a URL, in-memory data, or existing variables.
- **Whether the block prints.**

Mapped onto keywords, that was `request` = URL + never print, `display` = in-memory + always
print, `rad` = URL + always print. The combinations didn't cover the space - there was no
clean way to say "in-memory data, but don't print."

Worse, the keywords were **inconsistent about mutation**. `request` and `rad` modified the
input fields - a `sort` reordered the underlying lists - while `display` did not, even though
you sometimes wanted it to. Users couldn't predict whether a block changed their data. Trying
to fix the three keywords' semantics in place turned into a can of worms: it couldn't be made
consistent.

## Decision

One `rad` keyword that dispatches on its source, plus a `noprint` option for the printing axis:

- `rad <url>:` - fetch JSON from the URL (replaces `request`).
- `rad <list/map>:` - extract from in-memory data (replaces `display`).
- `rad:` - operate on existing variables (replaces sourceless `display`).
- `noprint` - suppress the table when you only want extraction.

Before and after:

```rad
# Old (no longer works)
request "https://api.example.com/users":
    fields Name, Age

display data:
    fields Name, Age

# New
rad "https://api.example.com/users":
    noprint
    fields Name, Age

rad data:
    fields Name, Age
```

Because `request` never printed but the unified `rad` prints by default, migrating a
`request` block means adding `noprint`; `display` already printed, so `display` → `rad` is a
clean rename.

To ease migration, the grammar still *accepts* `request`/`display` solely to emit a clear
error (`RAD40008`) pointing at the v0.9 migration guide, rather than a cryptic parse failure
- enforced by both `rad check` and the runtime. `rad check` also gained warnings (`RAD40007`)
for options with no effect in context, e.g. `noprint` on a block with no source.

## Rationale

Splitting the two orthogonal axes apart - **source** via runtime dispatch, **printing** via
`noprint` - is the whole cleanup. There's one keyword to learn, behavior is predictable, and
the mutation/print inconsistencies dissolve because there's a single, consistent model
instead of three subtly different ones. The readability argument that had favored distinct
keywords (RED-A) lost once the three-keyword model's own confusion outweighed it.

## Alternatives Considered

- **Keep the three keywords and fix their semantics.** Attempted, and abandoned as a can of
  worms - the mutation and print inconsistencies could not be made consistent while three
  keywords each hard-coded a source-and-print combination.

## Compatibility & Migration

A breaking change (`feat!`). `request` and `display` were removed as functional keywords.
Migration:

1. Replace `request` with `rad`, adding `noprint` to preserve the no-table behavior.
2. Replace `display` with `rad` - no other change needed.
3. Run `rad check` to catch remaining uses.

This is backed by a migration diagnostic (`RAD40008`), an `rad explain` error doc, and the
v0.9 migration guide.

## Other Consequences & Trade-offs

- **One construct instead of three**, with predictable, consistent mutation and print behavior.
- **`rad` now prints by default even for URL sources**, where `request` did not - a behavior
  change that migration must account for via `noprint`.
- **`rad` once again means "request and display" by default**, realigning the keyword with the
  language's name.
- **The readability concern RED-A tried to avoid returns** - readers must again understand that
  `rad` does both request and display. Accepted as worth it for the far simpler model.

---

## History

- 2026-06-04 Backfilled as Implemented. The unification shipped in v0.9.0 (~Mar 2026),
  superseding RED-A. Number `B` is provisional, pending chronological renumbering.
