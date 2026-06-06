---
red: 4
title: Declarative argument parsing
status: Implemented
kind: Language
created: 2026-06-05
decided: 2026-06-05
released: v0.1.0
supersedes:
superseded-by:
related: 2, 3
---

# RED-4: Declarative argument parsing

## Summary

Rad scripts declare the arguments they accept in an `args:` block, and Rad does the rest:
it parses user input, validates it, coerces it to the declared types, and **generates the
script's help/usage string automatically**. You describe *what* arguments exist; you never
write the parsing loop or the help text by hand.

Declarative arg parsing is not Rad's namesake - that's the rad block (see
[RED-3](0003-rad-block.md)) - but it was part of the language's design from the very start,
and it is one of the load-bearing pillars that make a Rad script dramatically better than the
Bash equivalent.

## Context / Motivation

A big share of the pain that motivated Rad (see [RED-2](0002-why-rad.md)) was argument
parsing in Bash. The problem was that Bash makes it *imperative*: you loop over the
arguments yourself, pick apart the flags by hand, and write separate logic depending on
whether you want order-independent flags, positional parameters, or both. Every script
re-implements the same fiddly machinery, and gets it subtly wrong.

The second half of the same pain is the help string. Writing a good `--help`/usage message
is tedious busywork, so people skip it - and when they do write one, it drifts out of sync
with what the script actually accepts. The usage text and the real argument handling are two
copies of the same information that have to be kept consistent by hand, and in practice they
don't stay consistent.

And even a perfectly maintained Bash usage string is just plain text the author formatted by
hand. Because Rad takes the argument information into the language *semantically*, it can
render that usage *well* - typed, aligned, sectioned, colored in a real terminal - for free.
That points at a broader idea behind the feature: Rad doesn't merely do these chores for you,
it does them well, so you end up with a genuinely good script even if you never set out to
craft one.

These frustrations are all the *imperative-where-it-wants-to-be-declarative* gap that Rad
exists to close. The insight that shaped the feature: all of it - parsing, ordering, type
coercion, validation, *and* a genuinely useful help string - can collapse into a single
declarative construct that derives everything from one description of the arguments.

## Decision

A script declares its arguments in an `args:` block near the top. From that single
declaration, Rad parses, validates, type-coerces, and generates help. The shape has been
stable since v0.1.0 - back then the type keyword `str` was spelled `string`:

```rad
args:
    repo str        # The repo to query. Format: user/project
    limit int = 20  # The max commits to return.
```

Each declaration packs several pieces into one readable line - the modern anatomy is:

```
<name> [rename] [shorthand flag] <type> [= default] [# arg comment]
```

From this, Rad gives the script author three things for free:

**Parsing with a dual nature.** Every argument is *both* a positional parameter and a named
flag - automatically, with no extra work and no choice forced on the author. Users can be
terse, explicit, or mix the two in one invocation. All of these were equivalent in v0.1.0:

```shell
rad commits spf13/cobra 3
rad commits --repo spf13/cobra --limit 3
rad commits spf13/cobra --limit 3
```

**Type coercion and validation.** Declaring `limit int` means the script receives an actual
integer; Rad rejects `--limit abc` with a clear error before the script runs. The author
never accepts a string and parses it themselves.

**An automatic help string.** The `#` comments and the declarations together produce the
usage output - no separately maintained help text, and formatted (aligned, sectioned, and
colored in a real terminal) so a human can scan it at a glance:

```
Usage:
  commits <repo> [limit] [OPTIONS]

Script args:
      --repo str    The repo to query. Format: user/project
      --limit int   The max commits to return. (default 20)
```

The block sits **before any executable code**, near the top of the script - so the
arguments read as the script's interface, and the variables they bind are declared before
they're used, like in any other language.

Two adjacent concerns are deliberately *out of scope* for this RED, each warranting its own:
the **constraint sub-language** (`enum`, `range`, `regex`, relational `requires`/`excludes`)
that validates argument *values*, and the **parsing-engine history** (Cobra → pflag → Rad's
own `Ra` library) underneath the declarative model. This RED is about the language-level
declarative model itself.

## Rationale

### Why declarative

The whole point is to invert Bash's imperative model. Instead of writing the parsing loop,
you write a description, and Rad does the parsing. This is the same declarative instinct that
runs through Rad's other signature features (the rad block, JSON path extraction): *declare
what you want, and let the tool get you as close to it as possible.*

What makes the args block in particular so powerful is how *rich* that description is. In one
declaration the author hands Rad the name, an optional short flag, the type, whether it's
optional, a default, and a human-readable description. That's a lot of meaningful structure
about each argument - enough that Rad genuinely understands what the script is asking for,
and can do a great deal on the author's behalf from it. The dual positional/flag nature, type
coercion and validation, and a well-formatted help string all fall out of that one rich
declaration.

### Why automatic help is co-equal, not a bonus

Generating a genuinely useful help string with no effort is not a side benefit of the args
block - it's half the reason the block exists, on the same level as parsing. People avoid
writing usage strings, and hand-written ones drift out of sync with the real arguments. By
making the declaration the single source of truth, the help string *cannot* go stale: it's
derived from the same thing that does the parsing. Unifying the two into one cohesive
structure is the core value. And the documentation cuts both ways: the declarations - names,
types, defaults, and `#` descriptions - also document the script for the *next developer* who
opens it, not only for the user reading `--help`.

### Ergonomics so that good scripts are the default

A guiding aim is to make writing a *good* script the path of least resistance. Users hate
running a script with no usage - "how do I even use this thing?" - but in Bash you can't
really fault the author, because doing argument parsing and help generation properly is a
real chore. Rad removes the excuse: the low-effort, ergonomic path *is* the one that yields
typed parsing, validation, and a polished help string. Good behavior falls out almost for
free, and what comes out is easy to write, easy to maintain, and easy to read.

### Why the dual nature

Allowing every argument to be passed positionally *or* as a flag felt simply natural -
maximum flexibility for the user at no cost to the author. The common frustration with other
tools is being told "this must be a flag" or "this must be positional," or finding that once
you've passed a flag you can no longer pass positional values. Rad has none of that: anything
positional can always be given as a flag instead. And there's no reason to make the *author*
do extra work declaring which arguments are which - the automatic dual nature is strictly
more ergonomic for everyone.

### Why typed arguments

Types are the most natural form of validation, and they remove busywork. If a script wants an
integer, having to accept a string and parse it on every entry is exactly the manual labor
the declarative model is meant to eliminate. Declaring the type lets Rad do the coercion and
rejection, getting the script as close as possible to the values it actually wants.

## Alternatives Considered

- **Imperative parsing (the Bash status quo).** The thing being replaced - loop over args,
  pick apart flags by hand. Rejected outright; it is the pain that motivated the feature.
- **Author-written help strings.** Rejected: tedious enough that people skip it, and it
  drifts out of sync with the actual arguments. Deriving help from the declaration is the
  whole win.
- **Forcing the author to choose positional *or* flag** (the model nearly every CLI
  framework uses). Rejected: it's more author work for a strictly worse user experience. The
  automatic dual nature gives users more flexibility for free.
- **Untyped / string-only arguments** (Bash's `$1`). Rejected: pushes parsing and validation
  back onto every script, defeating the declarative goal.
- **Modeling on an existing arg library** (Cobra, Python's `click`, Rust's `clap`,
  `argparse`). Not done. These were known, and there may be unconscious influence, but the
  design was opinionated and built from first principles for Rad rather than templated on any
  of them - the dual nature in particular came from a sense of what felt natural, not from
  prior art. Cobra was adopted as the *engine* early on, but only as a means to an end; it
  did not shape the language model and was later dropped.

## Compatibility & Migration

No impact. The `args:` block was foundational - part of the original design, present in Rad's
first release (v0.1.0). Nothing pre-existed for it to break.

## Other Consequences & Trade-offs

- **The `#` comment carries semantic meaning.** Code comments use `//`; inside the args
  block, `#` is *not* a throwaway comment - it's part of the syntax tree and drives the help
  string. The choice was pragmatic and `#` was a familiar-enough character. The trade-off,
  acknowledged: newcomers can be briefly confused about which marker means what, and that
  `#` carries meaning at all. No better alternative has surfaced, and the confusion clears
  quickly once you understand how the block works.
- **A lot of syntax in one line.** Name, optional rename, short flag, type, default, and
  comment all share a single declaration line. Fitting that much in while keeping it readable
  took deliberate tweaking, and the shape was largely settled this way from the original
  grammar.
- **The flexible parsing has edge cases.** Because anything can be a flag, values that look
  like flags need care - negative numbers especially (`--count=-5`, or `--` to force
  positionals). This was accepted: negative numbers are awkward in any flag parser, it isn't
  really inherent to the dual nature, and the handling has held up well in practice.
- **Bools later became flag-only.** Originally bools followed the same dual nature as every
  other type and could be passed positionally too. That was eventually pared back to
  flag-only (v0.5.42), since a bare positional `true`/`false` is awkward and not something
  anyone actually writes - passing the flag (`-d`/`--debug`) is simpler. A small ergonomic
  refinement to the original model.
- **Opinionated parsing, not strict POSIX.** The model prioritizes modern ergonomics over
  strict conformance to convention. Most notably, POSIX keeps *options* and *operands*
  (positionals) as separate categories, whereas Rad's dual nature deliberately merges them -
  any argument can be given either way, and flags can be interspersed freely among
  positionals. Other conveniences (long `--flags`, `--flag=value`, repeat-to-count ints like
  `-vvv`) follow common modern GNU-style practice rather than the POSIX baseline. The bet is
  that predictable, ergonomic behavior serves Rad's users better than standards conformance;
  the cost is occasionally surprising someone who expects strict conventions.
- **The block must appear before executable code.** An early doubt - the commit introducing
  argument parsing openly wondered whether the placement rule would survive - that resolved
  into a kept, settled rule. The unease, as best reconstructed, was about the *ordering
  restriction itself*: it felt a little odd to impose a fixed order (file header, then args
  block, then the rest) and to worry that users wouldn't grasp they can't put code *before*
  the block. In the end there's no better alternative, and the rule reads right
  conceptually: the args are the script's interface, parsing them is the first logical step a
  script takes, and declaring them at the top before use mirrors normal top-to-bottom code
  flow. It'll almost certainly stay long-term.

## References

- [RED-2](0002-why-rad.md) - the creation of Rad; argument-parsing pain was a founding
  motivation.
- [RED-3](0003-rad-block.md) - the rad block, the sibling declarative pillar (the output
  side to args' input side).

---

## History

- 2026-06-05 Implemented (backfilled). Declarative argument parsing was part of Rad's original
  design - sketched in the project's first grammar in August 2024 - and shipped in v0.1.0
  (2024-09-08). This record reconstructs the decision after the fact from git history and the
  author's recollection.
