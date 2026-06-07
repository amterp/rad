---
red: 5
title: Removing null from the language
status: Superseded
kind: Language
created: 2026-06-07
decided: 2026-06-07
released: v0.1.0
supersedes:
superseded-by:        # the null re-addition (prov. AEA, v0.5.32) - not yet backfilled
related: 2, 4
---

# RED-5: Removing null from the language

## Summary

Very early in development - days before the first public release - we removed `null` from
the language. It had briefly existed as an ordinary value you could write (`x = null`). The
immediate trigger was implementation pain (to allow a "no value" state, *every* runtime value
was a Go pointer so it could be `nil`, which was making the interpreter unwieldy), but the
real motivation was a standing aversion to null and a wish to try a null-free scripting
language while the cost of experimenting was near zero.

Null stayed gone for Rad's entire early public life - across v0.1.0 through v0.5.31, roughly
eight months - before being re-added in v0.5.32 for JSON type parity. **This RED records the
removal; the later re-addition supersedes it.**

## Context / Motivation

This was the genesis era: a handwritten lexer, parser, and tree-walk interpreter, all being
built at once (see [RED-2](0002-why-rad.md)). At that point `null` was a first-class value in
the language - there was a `null` literal in the grammar (`literal -> STRING | NUMBER | BOOL |
NULL`) and a corresponding node in the syntax tree.

Supporting it carried a pervasive cost. To let any value be absent, every runtime value was
stored as a pointer that could be `nil`, with `nil` standing in for null. That decision
rippled through the whole interpreter: every place that read or combined values had to thread
pointers around and account for the possibility of nothing being there.

The weight of that became hard to ignore while building out the binary operators - the `+`,
`==`, `<` machinery that has to inspect and combine operands. Doing that against
pointers-that-might-be-`nil` at every turn had made the implementation tangled and unwieldy,
out of proportion to what null was actually buying the language.

Implementation convenience has never been a guiding principle for Rad, though - our bias is to
spare the *user* effort even when that costs us, the implementers, plenty. The pointer tangle
was what forced the question; it isn't why we answered it the way we did.

The real reason was a genuine wariness of null. The author had been bitten by null-pointer
bugs many times over the years - enough to be skeptical of baking null into a language from
the outset, it being a famously hard thing to get right. Removing it was a chance to find out
whether Rad could avoid that whole class of problem.

And the timing made the experiment cheap. The language was pre-release with no users, so
trying something bold - "what if we simply don't have null?" - risked nothing. That fits how
Rad has evolved generally: we make a call that isn't fully worked out, ship it, and let real
use deliver the verdict. Removing null was one of those experiments.

## Decision

We removed `null` from the language outright. The `null` literal and its token were deleted,
so `null` was no longer something a script could write:

```rad
// Valid before the removal, a syntax error after it:
x = null
```

The grammar's `literal` rule lost its `NULL` alternative:

```
// before
literal -> STRING | NUMBER | BOOL | NULL
// after
literal -> STRING | NUMBER | BOOL
```

Underneath, this unwound the pointers: runtime values stopped being nullable pointers and
became plain values again, which stripped a great deal of pointer-handling out of the
interpreter and simplified it considerably.

That left one open question: what about optional arguments the user doesn't supply? Null had
been the obvious answer - an unset optional arg would simply *be* null. We judged this to be
null's *only* genuinely compelling use, and planned to cover it another way: a "set" flag on
such args, reading them while unset disallowed, and an `isSet(arg)` function for scripts that
needed to check. In practice, optionality leaned on **defaults** instead - the first release
only documented optional args *with* a default (`limit int = 20`).

The null-free language shipped in the first public release, **v0.1.0** (2024-09-08).

## Rationale

**We were wary of null to begin with.** Null-pointer bugs are a recurring, hard-to-eliminate
class of error - the author had been caught by them enough times to be reluctant to build null
into a language from the start. Sidestepping that entire category was worth a real attempt.

**The implementation pain was the catalyst, not the reason.** The pointer tangle made the
removal urgent and attractive at that moment, but it isn't *why* we did it - Rad doesn't shed
language features to make the interpreter's life easier.

**A null-free language is cheap to *try* before 1.0 with no users.** The experiment cost
almost nothing and could teach us something real. That's a recurring shape in Rad's evolution:
commit to a direction, ship it, and let usage decide rather than deliberating in the abstract.

**The "proper" null-free design doesn't fit a scripting language.** The model we had in mind
for doing without null *well* is Rust's: an `Option` type plus pattern matching and
decomposition, which forces you to handle the absent case. That's a genuinely better way to be
null-free - but it's a lot of ceremony, and ceremony is exactly what a terse scripting language
can't afford. Rad scripts are often short and throwaway; routing every possibly-absent value
through match arms would have fought the language's whole ergonomic aim. That realization is
also what eventually ended the experiment: without the Rust-grade machinery, the pragmatic
choice for a scripting language is plain `null` - which is what v0.5.32 returned to.

## Alternatives Considered

- **An `Option`/maybe type with pattern matching (the Rust approach).** Genuinely considered,
  and the most serious alternative. Rejected: it's the right way to be null-free, but its
  ceremony - wrapping values and decomposing them through match arms - is a poor fit for a
  terse scripting language where many scripts are short and throwaway. This is the road that,
  once ruled out, made plain null the pragmatic end state.
- **Keeping null but implementing it without pervasive pointers** (e.g. a tagged value/union
  rather than a pointer per value). Not considered at the time. The removal was a fast call,
  not a survey of cleaner implementations.
- **A narrower "undefined"/sentinel concept** distinct from a full null value. Not considered.

## Compatibility & Migration

No impact. The removal landed days before the first public release, so no shipped script ever
contained `null` - there was nothing to break and nothing to migrate.

## Other Consequences & Trade-offs

- **The interpreter got dramatically simpler.** Dropping null removed pervasive
  pointer-handling and let binary-operator evaluation - the work that had exposed the problem -
  run on plain values. This was the immediate, intended payoff.
- **Optional-argument handling was left a little awkward.** With no value for "absent," an
  unset optional arg without a default fell back to its type's zero value rather than to a
  clean nothing, and the `isSet()`/"set flag" we'd sketched was never actually built. Something
  adjacent, `is_defined()`, only arrived much later (v0.5.14, for relational arg constraints) -
  not as the optional-arg mechanism we'd imagined. The fuller account of how absence regained a
  proper representation belongs to the RED doc on re-addition.
- **The experiment didn't hold.** Living without null turned out not to be viable, chiefly
  because Rad aims for type parity with JSON - and JSON has null. Deserializing API responses
  that contained null values had nowhere good to land without a null type, forcing ad-hoc
  workarounds. The re-addition in v0.5.32 resolves this and supersedes the present RED.

## References

- [RED-2](0002-why-rad.md) - the genesis of Rad and its pre-1.0, experiment-friendly ethos,
  the backdrop for trying a null-free language.
- [RED-4](0004-declarative-args.md) - declarative argument parsing; optional arguments were
  null's one conceded use case, and how they were handled changed because of this removal.
- The null **re-addition** (provisionally `AEA`, shipped v0.5.32) supersedes this decision. It
  has not been backfilled yet; when it is, the `superseded-by` link here will be completed.

---

## History

- 2026-06-07 Backfilled, recorded directly at `Superseded`. The removal itself happened on
  2024-09-02, days before the first public release; null was absent across v0.1.0–v0.5.31 and
  re-added in v0.5.32 (~April 2025), which supersedes this record. It is born `Superseded`
  rather than `Implemented` so it can't be read as Rad's current position on null - it isn't.
  The superseding RED (provisionally `AEA`) is not yet backfilled, so the `superseded-by`
  frontmatter link is left pending and will be filled in when that RED is written. Reconstructed
  from git history and the author's recollection.
