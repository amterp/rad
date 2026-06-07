---
red: 6
title: Embedding Rad in Bash
status: Implemented
kind: Language
created: 2026-06-07
decided: 2026-06-07
released: v0.1.0
supersedes:
superseded-by:
related: 2, 4
---

# RED-6: Embedding Rad in Bash

## Summary

Within weeks of Rad's birth we added a small cluster of features so a Rad script could live
*inside* a Bash script and the two could cooperate. The pieces: `---` file headers and
single-quote strings, so Rad source embeds in a Bash string without backslash-escaping; a
`--shell` mode that prints the script's variables as shell assignments for Bash to `eval`; and
disciplined stdout/stderr routing so that `eval` stays clean. The point was the **hybrid
script** - even when Rad couldn't replace a Bash script outright, you could still borrow its
best parts (above all declarative arg parsing) from within Bash. Today it is semi-legacy.

## Context / Motivation

Rad exists to replace Bash (see [RED-2](0002-why-rad.md)). That is the long-term ambition, but
it was never going to happen overnight, and early Rad was deliberately narrow - a JSON
request-and-display tool. There would always be scripts Rad couldn't fully express: many at the
start, a rare few even now. Rather than make that an all-or-nothing wall - rewrite the whole
thing in Rad or get nothing from Rad - we wanted the two to coexist in a single script.

A hybrid script is genuinely powerful. You keep the Bash you need, and you reach for Rad
exactly where Rad shines. The standout case is argument parsing: Bash arg handling turns
unwieldy the moment you have flags or optional arguments, and fixing that was one of Rad's
founding motivations (see [RED-4](0004-declarative-args.md)). If a Bash author could lean on
Rad's declarative `args:` block - typed parsing, positional-or-flag flexibility, and a
generated usage string - without abandoning Bash, that was already a win.

But embedding ran into friction immediately. The natural way to carry a Rad script inside Bash
is a string variable, and the natural Bash string is double-quoted. Two pieces of Rad syntax
fought that: the original `"""` file-header delimiter and double-quoted Rad strings both
collided with the surrounding Bash quotes, forcing a thicket of backslashes. And once a script
*was* embedded, Rad's own output - errors, help text - landed on stdout and corrupted the very
output Bash was trying to `eval`.

## Decision

We built out embedding as one coherent capability, in four parts.

### `---` file headers instead of `"""`

The file header was originally delimited by triple double-quotes, inspired by Python
docstrings. Inside a double-quoted Bash variable that meant escaping every quote:

```bash
rad_script="
\"\"\"
Greet someone.
\"\"\"
args:
    name str
"
```

Switching the delimiter to `---` removed the collision entirely:

```bash
rad_script="
---
Greet someone.
---
args:
    name str
"
```

`---` is simple to type, easy to pick out visually, and echoes how Markdown and similar text
formats mark a divider. It reads cleanly whether or not the script is embedded.

### Single-quote strings

Rad originally accepted only double-quoted strings, which hit the same wall - a string literal
inside a double-quoted Bash variable needed escaping:

```bash
rad_script="
name = \"alice\"
print(name)
"
```

Allowing single quotes sidesteps it:

```bash
rad_script="
name = 'alice'
print(name)
"
```

Single and double quotes are otherwise interchangeable - both interpolate, both behave
identically. Single quotes are purely an ergonomic alias, and embedding is what motivated
adding them.

### `--shell` export mode

The `--shell` flag tells Rad to print, after the script runs, every variable in its environment
as a shell assignment. Bash `eval`s that output to pull the values into its own scope. Combined
with reading the script from stdin (`rad -`), a Bash script can hand its arguments to Rad and
get back parsed, typed values:

```bash
#!/usr/bin/env bash
rad_script="
---
Greet someone.
---
args:
    name str         # Name of the person
    age int = 30     # Age of the person
"

eval "$(rad - --shell "$@" <<< "$rad_script")"

echo "Hi $name, age $age"
```

```
> ./greet.sh alice --age 40
Hi alice, age 40
```

Bash forwards its arguments (`"$@"`) to Rad; Rad parses them against the `args:` block - so
`alice` binds positionally and `--age 40` by flag, exactly as a normal Rad invocation would -
and emits `name="alice"` and `age=40` for the `eval` to absorb. The Bash author gets Rad's full
arg-parsing and usage generation for free, having written no parsing loop.

The flag began life as `--BASH`, alongside `--STDIN`, both uppercased to stand apart from a
script's own (lowercase) flags. It was soon renamed `--SHELL` - the mechanism isn't Bash-specific
- and later lowercased to `--shell` when all global flags were. Reading a script from stdin,
originally the `--STDIN` flag, became the Unix-conventional `rad -`.

### Output-stream discipline

For `eval` to work, stdout must carry *only* the shell assignments. So Rad routes its
human-facing output - errors, help - to stderr, leaving stdout clean. And when Rad exits early
in shell mode (an error, or `--help`), it prints an `exit` line to stdout so the wrapping `eval`
propagates the outcome rather than letting the Bash script blunder on:

```bash
# Rad fails to parse → stderr shows the error, stdout carries `exit 1`,
# the eval runs it, and the Bash script stops instead of using empty vars.
eval "$(rad - --shell "$@" <<< "$rad_script")"
```

This stdout/stderr split is the right default well beyond embedding, but embedding is the
concrete problem that forced it.

## Rationale

**Coexistence is the pragmatic path to a Bash-replacement.** Replacing Bash wholesale is a long
road. Insisting on all-or-nothing would have left real scripts with no way to adopt Rad at all.
Letting Rad and Bash share a script means you use as much Rad as you can today, even while you
can't yet use it for everything - and you get the benefit incrementally instead of waiting for
Rad to grow complete.

**The escape hatch keeps Rad useful at its own edges.** Some problems Rad couldn't solve early
on, and a few it still can't. Embedding means those cases don't push you entirely back to Bash:
you can still pull in Rad and benefit from its strongest features, for the parts of the job it fits.

**The syntax changes were about removing escaping pain.** `---` and single quotes both exist so
that embedding a script in a Bash string doesn't degenerate into backslash soup. That `---` also
just looks cleaner was a welcome bonus, not the driver.

**stderr routing was bash-triggered but independently correct.** The motivating bug was concrete
- errors on stdout were breaking `eval`s - but sending diagnostics to stderr and reserving
stdout for real output is simply the right Unix behavior. Embedding is what made us do it; it
would have been right regardless.

**Keeping the export data in-process beat writing a file.** The values could have been handed
back through a sourceable temp file, but passing them as `eval`-able text keeps everything in the
data flow between the two processes, with nothing to create or clean up on disk. That felt more
robust, and it was the obvious shape.

## Alternatives Considered

- **Keep `"""` headers and escape them in Bash.** Rejected: `\"\"\"` is ugly and error-prone, and
  `---` is a cleaner delimiter on its own merits.
- **Escape double-quoted Rad strings in Bash** instead of adding single quotes. Rejected for the
  same reason - escaping defeats the goal of making embedding pleasant.
- **A sourceable temp file** for the exported variables, rather than `eval`-able stdout. Rejected:
  writing to the filesystem is less robust and less clean than keeping the hand-off as in-process
  data.
- **Single quotes with different semantics from double** (for instance, non-interpolating, as in
  Python or many shells). It crossed our minds, and making the language *feel* familiar was a real
  consideration, but it wasn't what drove adding single quotes, and we kept them identical to
  double quotes. Rad's quote forms - single, double, and later backticks - have varied in behavior
  at points in Rad's history but today all behave identically, a deliberate simplification.

## Compatibility & Migration

No impact at the time. All of this landed in the genesis era, before Rad had users, so there was
nothing to break.

One later refinement was technically breaking: lowercasing the global flags (v0.5.11) renamed
`--SHELL` to `--shell` and `--STDIN` to its lowercase form. By then Rad had a handful of users,
but this is a niche feature, so the blast radius was minimal. The migration is trivial - use the
lowercase flag.

## Other Consequences & Trade-offs

- **Hybrid scripts became possible**, and with them a partial-adoption path: you can bring Rad
  into a Bash codebase one capability at a time instead of committing to a full rewrite.
- **stdout is reserved for real output.** The discipline embedding forced - diagnostics to stderr,
  data to stdout - is good Unix citizenship that benefits every Rad invocation, not just embedded
  ones.
- **Single quotes opened the door to Rad's multiple quote forms.** That flexibility is convenient,
  and the decision to keep all forms behaving identically is what keeps it from becoming a source
  of confusion.
- **Doing Bash interop *well* is an ongoing cost.** Embedding touches arg parsing, output routing,
  exit handling, and value serialization, and each has to keep working as those areas evolve.
  That maintenance burden is the main reason the feature's future is uncertain.
- **It is semi-legacy today.** Rad has grown complete enough that almost any script can be written
  in pure Rad, so the case for embedding has shrunk to the rare script that must stay Bash - where
  borrowing Rad's arg parsing is still appealing. Whether the long-term upkeep is worth that
  narrow benefit is an open call (see Future Directions).

## Future Directions

The future of Bash embedding is undecided. It still earns its keep for the occasional
Bash-only script, but the value is narrow and doing it well indefinitely is a real maintenance
commitment. Whether to keep investing, freeze it as-is, or eventually retire it is a judgment we
have deferred rather than made.

## References

- [RED-2](0002-why-rad.md) - Rad's reason for existing is to replace Bash; embedding is the
  pragmatic counterpoint to that mission while the replacement is still in progress.
- [RED-4](0004-declarative-args.md) - declarative argument parsing, the marquee Rad feature that
  embedding lets a Bash script borrow.

---

## History

- 2026-06-07 Implemented (backfilled). The embedding features were built across the genesis era:
  the stdin/export mode shipped in v0.1.0 (2024-09-08), the `---` header and single-quote strings
  in v0.2.0 (2024-09-09), and the output-stream discipline through v0.2.2 (2024-09-11). Later
  refinements followed - lowercasing the global flags (v0.5.11, 2025-03) and replacing the
  `--STDIN` flag with the `rad -` convention. This record reconstructs the decision after the fact
  from git history and the author's recollection.
</content>
</invoke>
