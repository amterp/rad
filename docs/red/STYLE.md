# RED Style Guide

[RED-1](0001-red-process.md) defines the process: lifecycle, frontmatter, the section
skeleton, the freeze rule. This guide covers what RED-1 doesn't - the voice, approach, and
prose that make a RED read like the others. Before writing one, read a couple of existing
REDs as models: [RED-5](0005-removing-null.md) (a backfill, including a superseded
decision) and [RED-C](C-radish.md) (a live, forward-looking draft). The models are where
the examples live - emulate their moves at the same altitude, in fresh words; don't reuse
their phrasing.

Everything here applies to both live REDs and backfills (retroactive records of decisions
already shipped); advice specific to backfills is marked as such.

## Voice

- **Write as "we."** A RED is the people who built Rad explaining a decision they made.
  Reserve "the author" (or "I") for the genuinely personal: a taste, a lived experience
  that shaped the call. Default to "we".
- **Confident and plain.** Direct declarative prose, whether arguing a proposal or
  recording a decision. No hedging, no apologizing for the choice.
- **Honest about costs and failures.** A RED is a record, not a pitch. Name what turned
  out awkward or wrong in plain words, and mark accepted downsides explicitly as accepted
  rather than burying them.
- **Stance follows status.** A Draft is a proposal and should read like one - argued with
  conviction, but not written as a done deal. Once decided, the document speaks as
  commitment and then record: Implemented and backfilled REDs use past/observed tense.
  Promotion is the moment to sweep the prose from proposal to record.

## Approach

- **Truth over tidiness.** Never fabricate rationale to round out a story. When
  backfilling, if a reason or alternative is unrecoverable, record that honestly rather
  than inventing one. A plausible-but-wrong *why* poisons the record permanently.
- **Distinguish the trigger from the reason.** What forced the question at that moment is
  often not why it was answered that way. Name both and keep them apart - RED-5 turns on
  exactly this distinction.
- **Why, not what.** Reference docs (`docs-web/`) describe current behavior; a RED
  explains the decision. Show enough behavior to make the decision concrete, then stop -
  don't drift into user documentation.
- **Make it tangible.** Show real syntax, real output, real version numbers and dates.
  One concrete before/after or example block beats paragraphs of abstract description.
- **Low-level implementation details are out of scope.** The high-level implementation
  approach can *be* the decision - RED-C's pure-model/swappable-I/O split is the whole
  point of that RED - so record architecture at that altitude. What doesn't belong is the
  code-level inventory: files, structs, function names. For those, the commits are the
  record.
- **Weave the web.** Link related REDs and say *how* they relate - which one acts on,
  realizes, or supersedes which - in prose, not just in frontmatter. Tie the decision to
  Rad's recurring patterns and ethos where it genuinely fits (e.g. the
  own-your-foundational-deps precedent, the ship-and-let-usage-decide habit).

## Prose

- **Every sentence earns its place.** Cut throat-clearing, concessive strawmen ("It would
  be easy to call this X, but..."), and metadiscourse that rates or announces its own
  content ("this is a minor note", "the rest of this RED is about why"). The test: if
  deleting a sentence loses no information, delete it.
- **Lead with a bolded thesis.** In Rationale and Trade-offs, open each paragraph or
  bullet with a short bolded claim, then spend the rest of the paragraph unpacking it.
  This makes the document scannable at the section level.
- **Bold the load-bearing phrases** sparingly within prose, so a skimmer catches the spine
  of the argument. Italics for word-level emphasis.
- **No method narration** (backfills). The archaeology is how the story was recovered, not
  part of it: no "the commit says", "git history shows", "as best reconstructed", no
  quoted commit messages as rationale. Mine commits for the why, then state that why in
  clean design prose. The History footer is the only place the reconstruction is
  acknowledged.
- **Mechanics:** hyphen surrounded by spaces ( - ), never em dashes; American spelling;
  hard-wrap prose around 95-100 columns (tables may run longer); refer to REDs as `RED-5`
  in prose (zero-padding is for filenames only).

## Section notes

Cues from the strongest existing REDs, beyond what the template comments say:

- **Summary** - the whole decision graspable alone: what, the driving why, and any status
  caveat the reader must not miss (a supersession, for example).
- **Context / Motivation** - the forces, including project setting and precedent. This is
  where the trigger-vs-reason distinction usually lives.
- **Decision** - the choice made concrete: syntax, examples, expected behavior, what
  shipped and what's deferred. Tangible enough that a reader could recognize the feature.
- **Rationale** - bold-thesis paragraphs, one per reason. Order by importance.
- **Alternatives Considered** - one bullet per road not taken, each following the shape
  "**The alternative.** Rejected: why." Only alternatives genuinely on the table - never
  manufacture roads-not-taken to fill the section. In a backfill, an honest "not
  considered at the time" is a valid entry.
- **Compatibility & Migration** - "No impact" is a valid answer; state it explicitly and
  say why.
- **Other Consequences & Trade-offs** - gains *and* costs, including the ones that aged
  badly. The least flattering bullet is often the most valuable one.
- **History** - append-only; for backfills, the single home for "reconstructed from git
  history and the author's recollection" plus the real-world timeline.
