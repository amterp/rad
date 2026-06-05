---
red: 3
title: The rad block
status: Implemented
kind: Language
created: 2026-06-04
decided: 2026-06-04
released: v0.1.0
supersedes:
superseded-by:
related: A, B
---

# RED-3: The rad block

## Summary

The **rad block** is Rad's namesake construct - "Rad" is short for *Request And
Display*. It is a declarative block that fetches JSON, extracts named fields by path, and
prints the result as a table. It collapses the everyday `curl | jq | column` pipeline into
a few lines of readable code. This RED records why request-and-display was built as a
first-class *language block* rather than a library or a set of functions.

## Context / Motivation

Rad grew out of a specific, repetitive pain: writing Bash scripts at work that all did the
same thing. Take some user input, resolve an endpoint from it, `curl` the JSON response,
pull out a few fields with `jq`, and print them as a table with `column`. Every such script
was imperative, never declarative; Bash's syntax is arcane, `jq`'s is tricky, pipe handling
is fiddly, and the results - even when they worked - always had rough edges.

So Rad began as a humble DSL with one purpose: make *this* workflow - request-and-display
for JSON APIs - effortless and readable. The rad block is the construct that delivers it,
and the feature the whole language is named after.

## Decision

A rad block names a source (a URL), declares the fields to extract as **json path
definitions**, and renders a table. This was the canonical example shipped in v0.1.0:

```rad
args:
    repo string    # The repo to query. Format: user/project
    limit int = 20 # The max commits to return.

url = "https://api.github.com/repos/{repo}/commits?per_page={limit}"

Time = json[].commit.author.date
Author = json[].commit.author.name
SHA = json[].sha

rad url:
    fields Time, Author, SHA
```

Running it queries GitHub, extracts the three fields from the JSON response, and prints them
as a table - sorted, formatted, no extra code.

The block is more than a function call dressed up: its indented body hosts **statements**.
From the earliest design it was meant to carry `fields`, `sort`, per-field modifiers
(`truncate`, `color`, later `map`/`filter`), and - importantly - **conditional logic**, so
that which fields display can depend on the script's arguments:

```rad
rad url:
    if hideLocation:
        fields Temp
    else:
        fields Location, Temp
    sort Temp desc
```

`json` is a magic root identifier representing the response blob; a path like `json[].sha`
says "treat the response as a list, and for each item read `sha`." The block prints by
default.

## Rationale

**Why a block and not a function.** The construct carries sub-syntax - `fields`, `sort`,
per-field modifier sub-blocks, and conditional `if` statements. Expressed as nested function
calls or a config object, this would be verbose and ugly; the sorting and mapping
intricacies alone would make for unpleasant syntax. Conditional logic is the clincher: a
function call cannot host `if`/`else` over its own arguments cleanly. A block reads
top-to-bottom like a declarative spec of *what you want*, which fits Rad's readability-first
principle.

**Minimal, readable syntax.** The guiding question was "what is the fewest words needed to
express request-and-display while staying readable and not obtuse?" Name-based field binding
(`fields Time, Author`) and the bare `json` root both fall out of that: refer to data by the
names you already gave it, and to the response by an obvious keyword.

**Print by default.** The dominant use case is exactly the work-script pattern that motivated
Rad, where you *do* want to see the table. Printing is assumed so the common case needs no
ceremony.

**`json` as the root.** Every API in scope returned JSON, and the tool's entire purpose was
JSON request-and-display. A keyword to name "the JSON blob" and dig into it was the natural,
obvious fit for that scope.

## Alternatives Considered

- **A function / library API** - something like `display_table(fetch_json(url), [...],
  sort=...)`. Rejected: the modifier and conditional-logic sub-syntax don't fit a function
  signature; it would be verbose and hard to read, defeating the entire point.
- **Staying in Bash** (`curl | jq | column`) - the status quo that motivated Rad. Rejected
  as arcane, imperative, and perpetually rough.
- **A general-purpose language plus a library** (e.g. a Python package). This is really the
  "why build a language at all?" question, which belongs to the language-creation RED
  (forthcoming RED-2), not here. Granting the language, a declarative block was clearly the
  right shape for the feature.

## Compatibility & Migration

No impact. The rad block was foundational - present in Rad's first release (v0.1.0). Nothing
pre-existed for it to break.

## Other Consequences & Trade-offs

- **Unfamiliar syntax.** No other language has this construct, so users must learn it. Accepted
  as the cost of a purpose-built ergonomic that pays for itself immediately.
- **Print-by-default needs an opt-out** for the cases where you want extraction without a
  table. That tension surfaced later and shaped the block's evolution (see RED-A, RED-B).
- **JSON-specific by design.** The `json` root deliberately scopes the block to JSON, matching
  the tool's original remit rather than a general data model.
- **The construct became the center of the language** - conceptually and literally, in its
  name.

## References

- RED-A - the later split into `rad`/`request`/`display` blocks.
- RED-B - the still-later re-unification into a single `rad` keyword.
- Forthcoming RED-2 - the creation of the language as a whole.

---

## History

- 2026-06-04 Backfilled as Implemented. The rad block was built over Aug-Sept 2024 and
  shipped in v0.1.0 (2024-09-08), Rad's first release.
