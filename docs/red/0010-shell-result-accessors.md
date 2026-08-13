---
red: 10
title: Shell invocations are expressions when you name a result
status: Implemented
kind: Language
created: 2026-08-13
decided: 2026-08-13
released: v0.12.0
supersedes:
superseded-by:
related: 9
---

# RED-10: Shell invocations are expressions when you name a result

## Summary

A shell invocation is an expression when it is immediately followed by one of
four accessors - `.stdout`, `.stderr`, `.code`, `.ok`. There is no result object
type; the accessor is part of the invocation syntax.

Making that parse required narrowing what `$` binds. It used to take a whole
expression, so everything written after a command joined the *command*. That
binding was quietly wrong in three ways, two of which killed scripts silently,
and fixing it is the larger half of this change.

## Context / Motivation

Reading one value out of a command took an assignment and a throwaway variable:

```rad
stdout = $`git branch --show-current` catch:
    print_err("Failed to get branch")
    exit(1)
branch = stdout.trim()
```

Query-style commands - `git status --porcelain`, `git branch --show-current`,
`which x` - are among the most common things Rad scripts run, and every one cost
two lines and a name nobody wanted. Invocations were statements, so a command's
output could not feed an `if`, a function argument, or an interpolation.

The trigger was different from the reason. The idea had been parked since the
2025-10-13 "one last tangent thought" in `docs/thinking/shell_cmds.md`; what
moved it was a review of the author's own script corpus, where the same
four-line dance appeared repeatedly for a single value.

What made it urgent was what we found while checking whether it could parse.
`shell_cmd` bound `field("command", $.expr)`, the top of the whole precedence
cascade, so `$` swallowed everything after it:

```rad
x = $`echo hi`.upper()      // ran ECHO HI
x = $`cmd` catch "fallback" // caught the command string; handler never ran
x = $`cmd` ?? "fallback"    // the same
```

`rad check` reported nothing for any of them. The last two are the serious ones:
they read exactly like working error handling, and instead the failure went
unhandled and killed the script. Nobody had reported them, which we read as
evidence that people wrote the statement form and never tried the shape that
breaks.

## Decision

Four accessors, and an invocation is legal in expression position only with one:

```rad
branch = $`git branch --show-current`.stdout.trim() catch "unknown"

if $`git status --porcelain`.stdout.trim():
    print_err("Uncommitted changes!")

if not $`which docker`.ok:
    exit(1)
```

**Capture follows the accessor**, exactly as it follows target names in a named
assignment. `.stdout` captures stdout and lets stderr through to the terminal;
`.stderr` the reverse; `.code` and `.ok` capture nothing. So ``x = $`cmd`.stdout``
and ``x = $`cmd` `` behave identically.

**The output accessors raise; the status accessors never do.** `.stdout` and
`.stderr` propagate on a non-zero exit and compose with `??`, the `catch`
operator and `catch:` blocks. `.ok` (bool) and `.code` (int) return whatever
happened. This is deliberately asymmetric with statement-form `code = $cmd`,
which still raises.

**No truthiness.** ``if $`cmd`:`` is a check-time error (RAD40025) naming the
four accessors.

**One value inline.** The expression form never yields both streams from one
run; multi-capture stays with statement assignment.

**The ⚡️ echo stays on.** Under `and`/`or` short-circuit it shows which commands
actually ran. `quiet` and `confirm` remain prefix modifiers and work in
expressions.

`$` now takes only a primary: a string, a list, a parenthesized expression, or
an identifier. A computed command moves inside parentheses - `$(cmds[1])` -
which already parsed and ran, so no new form was invented.

## Rationale

**The accessor is syntax because a result object would be a type we'd have to
own.** The 2025-10-13 sketch returned a map of `code`/`stdout`/`stderr`, which
reuses map access and needs no new syntax. It also commits Rad to a value that
outlives the command, can be stored, passed and returned, and needs a name, a
type spelling, and an answer for what `.stdout` means on a command that captured
neither stream. Four accessor names cost nothing and close all of that.

**Refusing truthiness is the whole point.** In shell, `if cmd` tests the exit
code and `if $(cmd)` tests whether it printed something - two different questions
that look nearly identical, and a reliable source of bugs. A default here would
have inherited the ambiguity. Making the author write `.ok` or `.stdout` means
the reader never guesses.

**The status accessors cannot raise, or they are useless.** `.ok` exists to be
asked about a command that might fail. Raising there would make
``if not $`which docker`.ok:`` unwritable, which is the single most-wanted use.
The general rule underneath: asking about an outcome *is* handling it, so grep's
exit 1 is data. Asking for output is not, because output a failed command never
produced is not an answer.

**The names come from the capture vocabulary.** `stdout`, `stderr` and `code`
are already the reserved names for named assignment. Reusing them means one
thing to learn. `ok` is the one addition, and it is a predicate over the exit
code rather than a fourth stream - a distinction the implementation keeps
explicit, because the positional capture order is a three-element rule that a
fourth member would silently break.

**We let the old spellings parse so the checker could speak.** `$cmds[1]` is now
a command followed by an index, which is meaningless - but the grammar accepts
it, because a syntax error can neither say what the code used to mean nor offer
`$(cmds[1])` as a fix. The three shapes get three different diagnostics, since
only two of them ever did anything.

## Alternatives Considered

**A result map (the 2025-10-13 sketch).** Rejected as above. It also never
resolved its own open question: what the assigned variable holds inside a
`catch:` block. The document floated the error code, `null`, and the map itself,
and settled nothing. Accessors inherit the existing answer - the same rule as any
other expression with a catch - rather than needing a new one.

**Requiring the accessor in the grammar.** Cleaner to specify and strictly
worse. `$cmds[1]`, ``$`x`.upper()`` and ``if $`cmd`:`` all become syntax errors
carrying no explanation, no fix, and no editor action. Every migration
diagnostic this change ships depends on the grammar staying permissive.

**A distinct token for the inline form, say `$$cmd`.** Zero risk to the existing
grammar and non-breaking. Rejected because it leaves all three silent footguns
alive permanently - they live in the `$` binding, not in the new feature - and
adds a second spelling for the same operation.

**Keeping `shell_stmt` as its own rule alongside a new expression rule.**
Rejected on structure. Dissolving it into `assign` and `expr_stmt` means capture
targets, `catch:` blocks and operator precedence come from rules that already
define them, rather than being restated in a second place that can drift.

## Compatibility & Migration

**This breaks scripts, loudly.** Three shapes stop working, and all three are
check-time errors with an explanation:

| Old | New | Diagnostic |
|---|---|---|
| `$cmds[1]` | `$(cmds[1])` | RAD40026, with an editor quick fix |
| ``$`echo hi`.upper()`` | ``$`echo hi`.stdout.upper()`` | RAD40026 |
| ``$`cmd` catch "x"`` | ``$`cmd`.stdout catch "x"`` | RAD40025 |

**No silent reinterpretations.** Accessor names on a command string already
crashed at runtime before this change, so no script can quietly change meaning.
The two `catch`/`??` shapes were already killing the scripts that used them.

The statement forms are untouched. Of the seven syntax-tree snapshots that
moved, every AST is byte-identical - only the parse trees changed. The one
exception is a span that grew to cover parentheses that used to belong to lambda
body syntax.

## Other Consequences & Trade-offs

**A shell command that could not be started is now catchable.** It used to panic
as an internal Rad bug, with a Go stack trace and an invitation to file an
issue, while the same mistake written as a command string got a clean 127 from
the shell. The list form now reports 127 and 126 like a POSIX shell. This was a
prerequisite: `.ok` promises never to raise, and an uncatchable panic underneath
would have made that a lie.

**A non-zero exit finally has its own error code** (RAD20048). It reported as
the generic RAD20000 before, so the runtime error Rad users hit most often was
the one `rad docs` could say least about.

**`rad fmt` now formats shell statements**, because they are ordinary statements
rather than their own verbatim-emitted rule. That exposed a pre-existing bug:
statements with a trailing `catch:` block had their block dropped, which the
structural guard caught, so `rad fmt` declined the whole file. Fixed by emitting
those statements verbatim - a placeholder, not a decision that they should stay
unformatted.

**Accepted: `.ok == false` now means three things** - the command ran and
failed, you declined a `confirm`, or it could not be started. They are all "the
command did not succeed", so the semantics hold, but the error doc says it
plainly because someone will ask.

**Accepted: ``$`cmd`.ok`` works as a bare statement** - run it, stream it, never
fail. It falls out of the grammar and we allow it, but we do not document it as
the idiom over `catch: pass`. We expect a better spelling for that case later.

**Accepted: the parenthesized-command fix is uglier than what it replaces.**
`$(cmds[1])` reads worse than `$cmds[1]`. We take it: the old spelling's
readability came from a binding that also produced three silent failures.

## References

- [RED-9](0009-shell-interpolation.md) - the other v0.12 shell change. Both narrow
  what a command literal is allowed to mean, for the same reason: shell's
  defaults fail quietly, and Rad's job is to make them fail loudly instead.
- `docs/thinking/shell_cmds.md`, 2025-10-13 - where the idea and the rejected
  result-map design come from.
- [Migrating to v0.12](https://amterp.dev/rad/migrations/v0.12/)

---

## History
- 2026-08-13 Draft
- 2026-08-13 Implemented
