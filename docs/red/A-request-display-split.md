---
red: A          # provisional - to be renumbered to its chronological slot
title: Separate request and display blocks
status: Superseded
kind: Language
created: 2026-06-04
decided: 2026-06-04
released: v0.3.1
supersedes:
superseded-by: B
related: 3
---

# RED-A: Separate request and display blocks

## Summary

For a stretch of Rad's early life, the rad block (RED-3) was split into **three** keywords:
`rad` (request + display), `request` (fetch and extract, no print), and `display` (print
data already in memory). The split unlocked a workflow the original single block couldn't
express - fetch data, process it further with general Rad code, *then* display it. This was
later reversed; see RED-B.

## Context / Motivation

The original rad block could only hit a URL and had to display whatever came back,
immediately. But a common need was to fetch and extract data, manipulate it with ordinary
Rad code, and only then display - or to display data you'd built up in memory with no
request at all. The single, all-in-one block had no way to say "fetch but don't print yet"
or "just print this in-memory data."

## Decision

Three block keywords, each a thin variation on the same underlying rad block:

- `rad <url>:` - request, extract, and display (the original, unchanged).
- `request <url>:` - request and extract, **no print**. For processing before display.
- `display <data>:` - display in-memory data, **no request**.

For example, fetch each person's `ids`, compute how *many* ids each has, then display that
derived value alongside their name - impossible before, because the old `rad` block would
have printed the raw `ids` immediately:

```rad
url = "some-url.com"

Name = json[].name
ids  = json[].ids

request url:
    fields Name, ids

NumIds = [len(x) for x in ids]

display:
    fields Name, NumIds
```

Under the hood all three were the same rad block with different fields set and different
validation applied - the simplest, least-duplicative way to cover the overlapping set of
statements each variant allowed.

## Rationale

Separating request from display unlocked intermediate processing - the core motivation.
The choice of **distinct keywords** over an option flag on a single block came down to
readability: writing `rad` and then immediately instructing it to *disable* its display read
awkwardly, whereas a `display` keyword stated plainly what the block did. The keywords were
meant to read like what they do.

## Alternatives Considered

- **An option flag on a single `rad` keyword** (e.g. a toggle to suppress display) instead of
  separate keywords. This was thinkable at the time but rejected on readability grounds - a
  `rad` block that then turns off its own display felt odd; clear, distinct keywords read
  better. Notably, this is the very approach Rad later adopted (RED-B): the readability
  calculus flipped once the three-keyword model's *own* costs became clear.

## Compatibility & Migration

Additive at the time. `request` and `display` were new keywords alongside the existing `rad`
block, so no existing scripts broke. (Their eventual removal is covered in RED-B.)

## Other Consequences & Trade-offs

- **Two orthogonal concerns got conflated into one knob.** Each keyword fixed *both* the data
  source (URL vs. in-memory) *and* whether the block printed. That coupling later proved
  confusing - developed fully in RED-B.
- **Inconsistent mutation semantics.** `request` and `rad` modified the input fields (a `sort`
  reordered the underlying lists), while `display` did not - an inconsistency that became a
  real pain point and a driver of the reversal.

---

## History

- 2026-06-04 Backfilled. The split shipped in v0.3.1 (~Oct 2024). Recorded here at status
  Superseded, because the v0.9.0 unification (RED-B, ~Mar 2026) reversed it. Number `A` is
  provisional, pending chronological renumbering.
