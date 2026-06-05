---
red: 1
title: The RED process
status: Accepted
kind: Process
created: 2026-06-04
decided: 2026-06-04
released:
supersedes:
superseded-by:
related:
---

# RED-1: The RED process

## Summary

A **RED (Rad Evolution Document)** is a numbered, status-tracked markdown document that
records a significant decision or proposal for Rad - what we decided, why, and what we
rejected. A single RED spans its whole life: it starts as a proposal and matures, in place,
into the permanent record of the decision. This RED defines the format, lifecycle, and
conventions for all REDs that follow.

## Context / Motivation

Rad accumulates decisions constantly - syntax, semantics, interpreter architecture, tooling,
process. Most of the *reasoning* behind them lives only in my head or scattered across
commits, and it evaporates. Six months later the question "why does it work this way, and did
we consider X?" has no answer, and we risk re-litigating settled ground or silently undoing a
deliberate choice.

Our reference docs (`docs-web/`) describe *current behavior* - the *what*, for users. But
nothing captures the crystallized *why*: "we decided X, because Y, having rejected Z, accepting
trade-off W." That's the gap REDs fill. Reference docs say what Rad does today; REDs say why.

As Rad grows into a serious project that others may contribute to, this record becomes
infrastructure - for onboarding and for consistency.

## Decision

We adopt REDs as described below.

### When to write a RED

Write a RED for a decision that is **hard to reverse** or whose **rationale is worth
preserving**. Rules of thumb - write one when:

- it shapes the language, architecture, tooling, or process in a way users or future maintainers will
  question;
- the choice is hard to undo (breaking changes, established syntax, muscle memory);
- the interesting part is *why this and not the obvious alternative*.

Don't write one for routine bugfixes, mechanical refactors, or easily-reversible choices. When
in doubt: would you want to know the *why* in a year? If yes, write it. Don't write a RED for
every small thing - the bar is high.

### Lifecycle and statuses

A RED moves through a linear lifecycle. The status is the single signal of whether a RED is
still a proposal or now a record of reality.

| Status        | Meaning                                                                   |
|---------------|---------------------------------------------------------------------------|
| `Draft`       | Being figured out / actively proposed. Freely editable. Forward-looking.  |
| `Accepted`    | Decided - we're committing - but not yet built (or N/A for process).      |
| `Implemented` | Built and shipped. The document now describes how Rad actually works.     |
| `Rejected`    | Considered and declined. **Kept**, with a `## Rejection` section saying why. |
| `Superseded`  | Replaced by a later RED (see `superseded-by`).                            |

Backfilled records of long-standing decisions are simply *born* at `Implemented` - they skip
the early stages.

### Frontmatter

Every RED begins with YAML frontmatter:

| Field           | Required      | Meaning                                                                                  |
|-----------------|---------------|------------------------------------------------------------------------------------------|
| `red`           | yes           | The number, as an integer (e.g. `1`). References and filenames zero-pad to four digits.  |
| `title`         | yes           | Short descriptive title.                                                                 |
| `status`        | yes           | One of the lifecycle values above.                                                       |
| `kind`          | yes           | `Language`, `Architecture`, `Tooling`, or `Process`.                                     |
| `created`       | yes           | Date the RED was started (`YYYY-MM-DD`).                                                 |
| `decided`       | once decided  | Date it reached `Accepted`/`Rejected`. Blank while `Draft`.                              |
| `released`      | code REDs     | Rad version the change shipped in, e.g. `v0.11.0`. Omit for `Process`/non-shipping REDs. |
| `supersedes`    | if applicable | RED number(s) this one replaces.                                                         |
| `superseded-by` | if applicable | RED number that replaced this one. Canonical - the `Superseded` status derives from it.  |
| `related`       | optional      | Related RED numbers.                                                                     |

### The freeze rule

While a RED is in `Draft`, edit it freely - it's a work in progress. Once it's decided
(`Accepted`/`Implemented`/`Rejected`), **the substance is frozen**: if what we decided or why
would change, you don't rewrite it - you write a new RED that supersedes it. This keeps every
RED an honest snapshot of what we believed and why *at the time*, which is the entire reason
decision records are worth keeping.

What's frozen is the *meaning*, not the exact characters. Minor, inconsequential edits - typos,
spelling, grammar, formatting, wording that reads better without changing the decision - are
always fine. So is appending to the **History footer**, which logs each status transition
(including a future supersession). Git is the real backstop here - the full edit history is
always recoverable - so the freeze is a norm, not something we enforce in tooling.

### Body

Every RED follows the same skeleton. The **core sections are always present**; the rest are
included when they have something to say. Sections shift tense and fill-level as a RED matures
(e.g. *Consequences* is anticipated while `Accepted`, observed once `Implemented`), but the
skeleton stays fixed whether the RED is a live proposal or a settled record.

Core:

- **Summary** - a few sentences; the whole decision graspable from this alone.
- **Context / Motivation** - the problem and the forces around it.
- **Decision** - the actual choice/design. **Show concrete syntax and expected behavior with
  examples** - a RED should make the change tangible, not just describe it abstractly.
- **Rationale** - why this choice; the reasoning and trade-offs that led here.
- **Alternatives Considered** - the roads not taken, and *why not*. For a language this is the
  highest-value section: it stops us re-litigating settled ground.
- **Compatibility & Migration** - does this break existing scripts? What's the migration
  path? Placed first because Rad's promise of script stability makes breakage the
  highest-stakes consequence; state "no impact" explicitly when there is none.
- **Other Consequences & Trade-offs** - everything else we accept, good and bad: what gets
  easier, what gets harder. "Other" meaning beyond the compatibility covered just above.

Optional:

- **Future Directions** - deferred-but-plausible follow-ons (distinct from undecided points).
- **Open Questions** - genuinely unresolved points (heavy in `Draft`, empty by `Implemented`).
- **References** - links to related REDs and any external prior art.

And the **History** footer (append-only).

**Rejection** - *rejected REDs only.* When a RED is declined, add a `## Rejection` section
immediately after the Summary, capturing why it was declined (and what, if anything, would
change the answer). Writing it is part of the act of rejecting - the same decision moment that
freezes any other RED - so it's not a freeze exception. Keeping the full proposal beneath it is
the point: the record shows the case that was made *and* why it didn't win.

### Numbering, filenames, and references

- Numbers are assigned sequentially, next available.
- Files live in `docs/red/`, named `NNNN-short-slug.md` with a zero-padded four-digit number:
  `0001-red-process.md`. The padding is for filenames only - it keeps the directory sorted
  and aligned.
- Refer to a RED in prose without padding: `RED-1`, not `RED-0001`. The zero-padded form
  belongs to filenames alone.

### Supersession

Decisions reverse. When one does, don't edit the old RED - write a new one and link them: set
`supersedes:` on the new RED and `superseded-by:` on the old, and flip the old RED's status to
`Superseded`. The chain is the history; the latest un-superseded RED on a topic is current
truth. Keep the two links bidirectionally consistent (a good candidate for a future lint).

### Index

`docs/red/README.md` is an index of all REDs grouped by status. For now it's maintained by
hand; the frontmatter is structured precisely so it can later be generated automatically - a
natural thing to dogfood Rad itself for.

### REDs vs reference docs

Reference docs (`docs-web/`) are the *current behavior* layer - the *what*, for users. REDs are
not user documentation and shouldn't drift into it: they answer *why*, not *what Rad does
today*.

## Rationale

The design is a deliberate hybrid, taking the best of several prior systems while staying light
enough for a solo project to actually sustain:

- **One document for both proposal and record** (from Oxide's RFDs): a proposal that's accepted
  doesn't get replaced by a separate record - it *becomes* one. The `Accepted` → `Implemented`
  split is what lets a single doc signal "still a pitch" vs. "now describes reality".
- **Freeze-and-supersede over edit-in-place** (from ADRs): the value of a decision record is
  the *at-the-time* snapshot. Editing decided records destroys exactly that.
- **A fixed skeleton** (over MADR's mostly-optional sections): consistency makes REDs scannable
  and lowers the activation energy to write one.
- **A high trigger bar and lightweight process** (rejecting the heavyweight end of PEPs/RFCs):
  the dominant failure mode of decision-record practices is over-process - they get abandoned
  within weeks, so the default has to stay light. Today that also means no governance machinery
  - Rad is effectively solo-driven, so author, reviewer, and decider are one person and a
  `decided` date replaces a review apparatus. If that changes, a later RED can formalize
  process; lightweight stays the baseline.

## Alternatives Considered

- **Plain ADRs** (bare Context/Decision/Consequences; current state inferred from the
  supersession chain). Not really a rejection - REDs are a variation of ADRs (numbered,
  immutable, supersede-don't-edit). We diverge on three things: a richer fixed skeleton suited
  to language decisions, the single proposal→record lifecycle, and an index so finding current
  state doesn't mean tracing chains by hand. That last point we only *soften*, not solve - with
  enough supersessions "what's the rule now?" gets harder again, and the index (later, tooling)
  is our bet against it.
- **Separate "proposal" and "decision record" document types.** Rejected: the forward/backward
  distinction is already encoded by *status*; a second axis would just be one more thing to keep
  in sync. A proposal and its eventual record are the same artifact at different times.
- **Just keep informal notes.** Rejected: scattered exploratory notes meander, aren't
  commitments, and don't always get written. They capture the journey, not the decision - a RED
  pins the decision down.
- **Full community process** (RFC review periods, FCP bots, sponsors, steering councils).
  Rejected: pure overhead for a solo project. Adoptable later if Rad grows a contributor base.
- **An explicit `Ideation` status before `Draft`** (as Oxide RFDs have). Rejected: `Draft`
  already covers "still being figured out"; a separate pre-Draft status is the kind of sprawl
  that kills the practice.

## Compatibility & Migration

This RED introduces a new, additive process - nothing existing breaks, so there's nothing to
migrate. Going forward, significant decisions get a RED. **Backfilling** the rationale for
decisions already made (drawing on git history and memory) is valuable but
explicitly *future* work; those will be born at `Implemented`.

## Other Consequences & Trade-offs

- We gain a durable, discoverable record of *why* Rad is the way it is, stable references
  (`RED-1`) to point to from commits and discussions, and a forcing function to think through
  alternatives and compatibility before committing.
- It costs writing time, and the index needs maintaining until it's automated.
- Freezing decided REDs means reversals create supersession chains rather than edits. That's
  the intended trade: more documents, but an honest history.
- The high trigger bar means some decisions won't get a RED. Accepted - over-documenting is the
  bigger risk.

## Future Directions

- Backfilling historical decisions, likely via an agent interviewing me to fill the gaps.

## References

- Prior art this process borrows from: Python PEP 1 & PEP 12, Rust RFCs, Swift Evolution,
  Michael Nygard's ADRs and the MADR template, and Oxide Computer's RFD 1.

---

## History

- 2026-06-04 Accepted
