---
red: 2
title: Why Rad
status: Implemented
kind: Language
created: 2026-06-05
decided: 2026-06-05
released: v0.1.0
supersedes:
superseded-by:
related:
---

# RED-2: Why Rad

## Summary

Rad exists to kill one specific, repetitive pain: writing Bash scripts that take some
input, hit a JSON API, pull out a few fields, and print them as a table. It was born in
August 2024, and the name says what it was for - **R**ad = **R**equest **A**nd **D**isplay.

The project took the form of a CLI tool *and* its own purpose-built scripting language
(originally **RSL**, the Rad Scripting Language). Building a whole new language - rather than
reaching for a library in Bash, Python, or Go - is the largest decision here: only a
purpose-built language could give that workflow the ergonomics and domain-tailored syntax we
wanted, and a greenfield language had the highest ceiling.

At birth, Rad was deliberately narrow: a JSON request-and-display tool, not a general
scripting language. The broad "replace Bash for most scripting" identity it carries today was
*not* the original vision - it grew later, as the walls of the narrow scope kept pushing us
back to Bash.

## Context / Motivation

The pain was concrete and recurring. Writing Bash scripts at work, the same shape kept coming
up: take some user input, resolve an endpoint from it, `curl` the JSON response, pull a few
fields out with `jq`, and print them as a table with `column`. The rad block - the construct
built to deliver exactly this workflow - is the feature the project is named after.

Every such script had the same frustrations:

- **Bash is arcane.** It's a product of its time; syntax has moved on a lot in the decades
  since. Simple things are deceptively laborious.
- **Argument parsing got unwieldy fast.** Fine for trivial cases, painful past them - and for
  the scripts being targeted, good arg handling matters a lot.
- **The toolchain is rough.** `jq`'s syntax is tricky, pipe handling is fiddly, and even
  working scripts came out with rough edges.

These scripts were imperative where they wanted to be declarative. The job - *what to query,
what to extract, how to display it* - is small and well-shaped, but expressing it in Bash was
out of proportion to the task. That gap is the reason to build anything at all.

## Decision

We built **Rad**: a CLI tool that runs scripts written in a **purpose-built scripting
language** designed for the request-and-display workflow. `rad <script>` parses and validates
the script's declared arguments, runs it, and handles query execution and table formatting.

Here is the genesis-era syntax, taken from the earliest README (shipped in v0.1.0):

```
args:
    repo string    # The repo to query. Format: user/project
    limit int = 20 # The max commits to return.

url = "https://api.github.com/repos/{repo}/commits?per_page={limit}"

Time = json[].commit.author.date
Author = json[].commit.author.name
SHA = json[].sha

rad url:
    Time, Author, SHA
    sort Time desc, Author, SHA
```

```
> rad commits spf13/cobra 3

Time                   Author                 SHA
2024-07-28T16:18:07Z   Gabe Cook              756ba6d...
2024-07-16T23:36:29Z   Sebastiaan van Stijn   371ae25...
2024-06-01T10:31:11Z   Ville Skyttä           e94f6d0...
```

Every line earns its place: the `args:` block declares and documents inputs (and generates
the usage string), interpolation builds the URL, the `json[]...` lines declare the fields to
extract, and the `rad` block runs the request and renders the table. The whole language exists
to make these few lines express the entire workflow.

Several sub-decisions sit underneath "build Rad," and the rest of this RED is mostly about
*why* each went the way it did:

- **It's a language, not a library.** Rad is its own grammar and interpreter, not a package
  you import into Bash, Python, or Go.
- **Born narrow.** v0.1.0 Rad was a JSON request-and-display tool. General-purpose ambitions
  were explicitly out of scope - so much so that **user-defined functions were not going to be
  a thing**, and even `if` statements weren't a given early on.
- **A tree-walk interpreter, written in Go.**
- **Grammar-first.** The EBNF grammar was written on day one, before any lexer or parser code.
- **Flat variable scoping.** Variables are visible everywhere after they're assigned, with no
  block-local scope.
- **Pre-1.0, breaking changes allowed between minor versions.**

## Rationale

### Why build anything

The pain was real, recurring, and badly served by existing tools (see Context). The task
itself is small and regular - request, extract, display - which is exactly the kind of thing
worth investing tooling in: the work saved per script is high, and it recurs constantly.

### Why a whole language, not a library

This is the central call. A library lives inside its host language and inherits that host's
syntax and constraints. Stay in Python and you're still writing Python; the result will never
be as tight as something designed for the job. Going greenfield removed that ceiling:

- **Domain-tailoring.** Owning the syntax meant the sky was the limit - the `args` block and
  the `rad` block could be exactly as ergonomic as the domain allowed, instead of as ergonomic
  as a host language permitted.
- **Ergonomics was the point.** The overriding goal was making these CLI scripts genuinely
  easy and fast to write. A general-purpose language optimizes for generality, not for *this*.
- **Maximum potential and full control.** A greenfield language has the highest ceiling, and
  full control means you can do anything. (That same instinct shows up later in Rad's habit of
  forking and vendoring its dependencies - dropping Cobra, shipping its own table writer - to
  keep that control end to end.)

The cost was understood and accepted: building a language is a lot of work. The bet was that
the ceiling justified it.

### Guiding principles

A handful of principles have guided Rad from the start. They aren't ranked, and they aren't a
closed list - but they were present at the genesis and have held throughout:

- **Tailor to the use case.** Rad has a specific domain; lean into it rather than chasing
  generality.
- **Ergonomics and productivity.** Scripts should be quick to write - many are throwaway -
  so favor the pragmatic, productive path over the strict one.
- **Readability and familiarity.** Python-like and self-explanatory; none of Bash's or `jq`'s
  arcane symbols. Stick with familiar norms unless there's clear value in diverging. The
  guiding test - not strictly enforced, but a north star - is that someone who has never heard
  of Rad should be able to open a Rad script and understand what it does.
- **Portability and shareability.** A single self-contained binary with no dependencies: if a
  Rad script runs on your machine, it runs on others' too.

### Why Go

Go was the natural pick. It's pragmatic, productive, and portable - it compiles to a single
static binary, which directly serves the portability-and-shareability principle. Performance
isn't critical for Rad's domain, so a garbage-collected language was a fine fit rather than a
compromise.

### Why a tree-walk interpreter

Performance was never the constraint, so an interpreted language was completely fine, and a
tree-walk interpreter was the quickest way to get runnable scripts. The interpreter was built
alongside the parser from early on specifically to get end-to-end execution working sooner
rather than finishing the parser in isolation first.

### Why grammar-first

Formalizing the EBNF grammar before writing implementation forced a clear picture of the
syntax up front, and that grammar then guided how the lexer and parser were built. It was a
deliberately methodical way to start: decide what the language *is* before building the
machine that runs it.

### Why flat scoping

Flat scoping - no block-local variables - was a deliberate, pragmatic choice in service of
productivity. For a scripting language where scripts are often short and throwaway, being
strict about block scope adds ceremony without much payoff. The practical path won, and it has
not been a source of regret. (Functions, added much later, are the one exception - a function
body runs in its own scope and closes over where it was defined - but that's beyond the
genesis.)

### Why pre-1.0 with breaking changes between minor versions

A 1.0 release remains the long-term plan, but Rad stays pre-1.0 deliberately - much as Zig has
for years. Being pre-1.0 signals that the language is still in active development and isn't
bound to strict semantic-versioning guarantees yet. Allowing breaking changes between minor
versions is how that's expressed: languages are a lot of work, and locking down compatibility
prematurely would slow the very evolution Rad still needs.

## Alternatives Considered

- **Keep using Bash (build nothing).** The status quo, and the thing Rad replaced. Rejected:
  arcane syntax, arg parsing that doesn't scale, and a generally rough authoring experience
  for the targeted workflow.
- **A library in a general-purpose language** (Python, Go, etc.). Rejected: general-purpose
  languages aren't designed for the CLI-tooling domain, and a library is constrained by its
  host's syntax. Whatever you built would be less ergonomic than a purpose-built language, and
  general-purpose languages also carry their own friction around environment management and
  sharing scripts reliably.
- **Build on an existing tool** (a `jq` wrapper, extending HTTPie, a Bash framework). Rejected
  in favor of greenfield: wrapping an existing tool inherits its limits, whereas a new language
  had the highest ceiling and full control over the result.
- **A compiler or bytecode VM** instead of a tree-walk interpreter. Not seriously considered:
  performance isn't critical for the domain, so the added complexity bought nothing.
- **Other implementation languages.** These came down to the author's own preferences, not
  objective rankings. *Rust* - Go simply felt more productive to write, and Rust (in the
  author's experience) carried a steeper learning curve and was less enjoyable; Go had neither
  downside. *Java* - barely entered the running despite the author's Java background; it was
  dismissed almost out of hand, the JVM being a poor fit for a single-binary CLI tool. *C++* -
  never a real contender: unfamiliar to the author, and with the freedom to pick any stack (a
  luxury of solo development), there was little reason to take on its reputation for sharp
  edges when more modern, ergonomic languages were on offer. Go was the clear, pragmatic fit.

## Compatibility & Migration

No impact. This is the genesis of the project - there was nothing before it to break or
migrate.

## Other Consequences & Trade-offs

- **A language is a lot of work.** The whole premise trades a large, ongoing implementation and
  maintenance burden for a much higher ceiling. Two years on and counting, that's the deal
  we're still living with - and still think was right.
- **Full control means owning the stack.** The greenfield bet led to forking and vendoring
  dependencies over time (Cobra dropped, a custom table writer, and more) to keep control end
  to end. That's leverage and maintenance burden in the same breath.
- **Born narrow, grown broad.** The original scope was JSON request-and-display, and several
  early shapes reflect that. The clearest artifact: because user-defined functions were never
  meant to exist, *all* functions could just live in one global namespace - which is why
  built-in functions are still globally scoped today, even though custom functions now exist
  and see heavy use. The broadening began early, though: even the v0.1.0 README's promise to
  handle "95% of your scripts" was already reaching past the narrow workflow - by the time it
  was written, the vision had started to generalize beyond request-and-display. Each time the
  narrow scope later forced a return to Bash, the ambition grew further, until Rad became the
  general scripting language it is now.
- **Flat scoping is a watched trade-off.** Pragmatic and so far unregretted, but a known place
  the language chose convenience over strictness; whether that ever needs revisiting remains
  open.

---

## History

- 2026-06-05 Implemented (backfilled). The decision itself dates to the project's creation in
  August 2024 and shipped in v0.1.0; this record reconstructs it after the fact from git
  history and the author's recollection.
