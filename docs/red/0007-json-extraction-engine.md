---
red: 7
title: The JSON extraction engine
status: Implemented
kind: Architecture
created: 2026-06-07
decided: 2026-06-07
released: v0.1.0
supersedes:
superseded-by:
related: 3
---

# RED-7: The JSON extraction engine

## Summary

The rad block lets you name fields by JSON path (`Author = json[].commit.author.name`)
and prints them as a table. The engine that turns those path definitions into aligned
table columns compiles them into a **trie**, walks the response **once** to capture data as
it goes, then **merges** those captures into the equal-length columns a table needs. This RED
records why extraction was built as a single coordinated walk with a merge step - rather than
evaluating each field independently - and why every captured field is always a list. This is
the computational core beneath RED-3's block syntax.

## Context / Motivation

RED-3 covers *why* the rad block exists and what it looks like. This RED is about the
machinery underneath: given a set of field definitions and a JSON blob, how do you produce
the columns of a table?

```rad
Time   = json[].commit.author.date
Author = json[].commit.author.name
SHA    = json[].sha

rad url:
    fields Time, Author, SHA
```

The output is a table, and a table has a hard constraint: **every column must be the same
length**. So the real job isn't just "pull these values out" - it's "pull these values out
*and* line them up into equal-length columns," even when the paths dig to different depths,
fan out through arrays, or hit single scalar values. That alignment problem is what makes
this more than a loop over `jq`-style lookups, and it's what the engine is designed around.

## Decision

The engine has three stages: build a trie from the field definitions, traverse the JSON
once against it, and merge the captures into aligned columns.

### A trie of the defined fields

We start with a handful of field definitions - paths into the response:

```rad
Time   = json[].commit.author.date
Author = json[].commit.author.name
SHA    = json[].sha
```

Notice how much these paths overlap. All three begin at `json[]`; two of them continue
through `commit.author`. The natural structure for a set of paths that share prefixes is a
**trie** (a prefix tree): where paths share a beginning they share a node, and they branch
only where they genuinely diverge.

```
json
 └── []                        (array wildcard - "for each element")
      ├── commit
      │    └── author
      │         ├── date        → Time
      │         └── name        → Author
      └── sha                   → SHA
```

The property the whole engine leans on is that **every node corresponds to one scope in the
JSON**. Read the tree top to bottom: `json` is the entire response; `[]` says "the response
is an array - step into each element"; `commit`, `author`, and `date` each follow one key
deeper. A node also remembers which fields *end* there - `date` terminates `Time`, `sha`
terminates `SHA` - since those are the points where we'll actually record a value.

### One walk, capturing as it goes

With the trie built, we walk the JSON once, guided by it. At each node we move into the
matching scope of the data: follow a key, index into an array, iterate every element of an
array (`[]`), or iterate every key of a map (`*`). When the walk reaches a node that
terminates a field, it writes down the value sitting there.

The result of walking a node is what we'll call a **capture**: a set of named columns, each
column a list of values - one entry per row. A capture is a fragment of the eventual table.
Walking the `[]` node over a two-commit response, for example, yields a capture three columns
wide and two rows tall:

```
Time            Author          SHA
2024-09-02...   Linus Torvalds  a1b2c3
2024-09-01...   Linus Torvalds  d4e5f6
```

Because the trie holds every field at once, this single walk feeds all of them - the blob is
never traversed again per field.

### Merging captures into a table

This is the step that earns the algorithm its name. A node rarely captures in isolation: it
has children, each handing back its own capture, and the node must combine them into one. The
combining is the crux, because **children can hand back differently shaped captures**, while
the end result has to be a clean rectangle - a table, where every column is the same height.

Where do the different shapes come from? Fan-out. A node that loops - an array `[]` or a key
wildcard `*` - produces one row per element or key, so its capture can be many rows tall. A
plain key or a fixed index produces a single value: one row. So when a parent gathers its
children's captures, some arrive tall and some arrive short, and it has to reconcile them.

It does this by comparing the two captures' columns. Four cases cover it, checked in order:

1. **Same columns.** The two captures name the exact same fields, so they're two batches of
   one table - stack them and append the rows. (Two array elements, each producing a
   `{Name, Age}` row, combine into a two-row `{Name, Age}` capture.)
2. **Different columns, same height.** The captures describe different fields but have the
   same number of rows, so they're columns of one table standing side by side - append the
   columns.
3. **Different columns, different heights, one side a single row.** One capture is tall, the
   other has just one row. We read that single row as a value that holds for the whole group,
   **repeat (broadcast)** it down to match the taller capture, and then set them side by side.
   (A country is the same for every city beneath it, so it's copied down each row.)
4. **Anything else.** Different columns, different heights, and neither side is a single row.
   There's no honest way to fold these into one rectangle, so it's an error.

Rule 3 is the subtle one - it's how a single value lines up against a list. Consider this
response and these fields:

```json
{
  "Australia": {
    "Sydney": [ { "name": "Alice" }, { "name": "Bob" } ]
  }
}
```

```rad
Country = json.*
City    = json.*.*
Name    = json.*.*[].name
```

A key wildcard (`*`) captures the *name* of the matched key, so `Country` and `City` resolve
to the single keys `Australia` and `Sydney`. The array wildcard (`[]`) in `Name` iterates the
list and yields two rows. The merge broadcasts the single values to match, so the captures
become:

```
Country = [Australia, Australia]
City    = [Sydney, Sydney]
Name    = [Alice, Bob]
```

which prints as a clean table:

```
COUNTRY    CITY    NAME
Australia  Sydney  Alice
Australia  Sydney  Bob
```

The repeated `Australia` looks surprising read as a raw list, but it's exactly what an
equal-length table demands - and it's the same broadcast rule (3) doing the work everywhere.

### Every field is a list

A captured field is **always** wrapped in a list, even when its path plainly yields a single
value (`Length = json.len`). The reason is the merge: a sibling field can force a "single
value" to broadcast into N rows. If we unwrapped, a field's array-ness would depend on the
response data *and* on whatever other fields you happened to query alongside it. We wrap
unconditionally so a field's type is knowable from the script alone, not from any particular
response.

## Rationale

**One walk, not one walk per field.** It struck us strongly from the outset that you should
be able to capture everything in a single pass over the blob, rather than re-walking it once
per field. The trie is what makes "all at once" work: shared path prefixes collapse into
shared nodes, so the walk does each piece of common work exactly once.

**One node, one move into the data.** Every trie node maps to exactly one step in the JSON -
follow a key, take an index, iterate an array, or iterate a map's keys - and nothing more.
That clean correspondence is what makes the walk mechanical: each node performs its single
move and recurses, and no node has to work out where it sits in the larger structure or juggle
several shapes at once. This was the insight that turned an earlier, ad-hoc version of the
algorithm into a sound one - before it, the traversal was a tangle of special cases.

**Merge because the output is a table.** Table output was front-of-mind the whole time. The
four merge rules aren't arbitrary - they're the minimal set needed to reconcile differently
shaped captures into the one shape a table accepts: equal-length columns. Doing extraction
and alignment together in one coordinated walk is far more natural than extracting freely and
trying to square the shapes afterward.

**Always-wrap for predictability.** We briefly did the opposite - left single-value captures
unwrapped - and it was confusing. A field's type shouldn't shift between list and scalar
depending on the data or on its neighbours. Wrapping everything trades a little ergonomic
convenience for a type you can reason about while writing the script.

## Alternatives Considered

- **Independent per-field evaluation** - walk the blob once *per field*, collect each
  field's values, then assemble the columns afterward. We didn't pursue it. A shared walk
  avoids re-traversing the blob for every field, and the column alignment falls out of a
  coordinated walk; independent evaluation would still have to reconcile the shapes into a
  table at the end, so it buys nothing and costs traversals. In honesty this wasn't deeply
  weighed - the single shared walk simply felt right and proved out.

- **Adopting an existing extraction model** - `jq`, JSONPath, or GraphQL. The author had
  used `jq` but hadn't studied how it (or JSONPath, or GraphQL, which weren't on the radar as
  related at the time) actually models this internally. The trie-and-merge design was arrived
  at cold, straight from the request-and-display need; it felt natural and we ran with it. A
  proper study of that prior art is overdue and likely to reshape a future revision (see
  Future Directions) - there are extractions the current path syntax simply can't express.

- **Leaving single-value fields unwrapped** - we actually shipped this for a stretch:
  `Length = json.len` captured a bare value rather than a one-element list. We reversed it.
  Once the merge can broadcast a single value across rows, a field's array-ness becomes a
  property of the response and its sibling fields rather than of the script, which made
  scripts hard to reason about.

## Compatibility & Migration

No meaningful impact. The engine was foundational - the trie extractor was in Rad's first
release (v0.1.0), so nothing pre-existed for it to break. Its later refinements did change
behavior (key wildcards in v0.2.6; the always-wrap reversal in the v0.4.7 rewrite), but those
landed during Rad's earliest development, well before any stability promise and with a
negligible user base. There was nothing to migrate.

## Other Consequences & Trade-offs

- **No field triggers its own full re-walk.** Shared path prefixes collapse into shared trie
  nodes and are visited once, so traversal cost tracks the trie's shape against the data rather
  than re-reading the whole blob once per field.
- **The merge rules can surprise.** Broadcasting and wildcard duplication produce repeated
  values that are correct for a table but puzzling if you read a captured list directly. The
  table is the intended lens; the lists underneath carry the alignment.
- **Always-wrap costs a little ergonomics.** Even an obviously-scalar `json.len` comes back
  as a one-element list. We accept that for a predictable type.
- **There's an expressiveness ceiling.** Some extractions can't be expressed today - flattening
  keys across levels was a known example, deferred at design time. The v0.4.7 design was built
  to be extendable, but that headroom has gone largely unused.
- **It has been stable and untouched since v0.4.7** - evidence the architecture is sound
  enough, but also that its known gaps remain open.

## Future Directions

- **Revisit with prior art in hand.** A deliberate study of `jq`, JSONPath, and GraphQL to
  absorb their lessons and close the expressiveness gaps the current syntax has. This is the
  most likely source of the engine's next iteration.
- **Cross-level key flattening** - extracting and flattening keys across nesting levels was
  floated and deferred when the v0.4.7 design landed; never built.
- **A single-value assertion** - a way to declare that a field should hold one value and be
  unwrapped, opting out of always-wrap where the script author knows the shape. Considered
  alongside the always-wrap decision; never built.

## References

- RED-3 - the rad block, the syntax this engine powers.
- The forthcoming data-model RED (mixed-type arrays and maps for JSON parity) covers the
  value model these captures populate; this RED is only about the extraction algorithm.
- Prior art not consulted at design time but relevant to any future revision: `jq`,
  JSONPath, GraphQL.

---

## History

- 2026-06-07 Backfilled as Implemented. The trie extraction engine shipped in v0.1.0
  (2024-09-08); key-wildcard support followed in v0.2.6; and the principled rewrite around
  trie-node-as-JSON-scope plus the four-case merge - the architecture described here, still in
  use today - landed in v0.4.7 (~Nov 2024).
