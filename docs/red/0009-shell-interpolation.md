---
red: 9
title: Shell commands quote their interpolations
status: Implemented
kind: Language
created: 2026-08-10
decided: 2026-08-10
released: v0.12.0
supersedes:
superseded-by:
related: 2, 6, 10
---

# RED-9: Shell commands quote their interpolations

## Summary

A value interpolated into a shell command is now quoted, so it reaches the program as
exactly one argument no matter what characters it contains. Interpolating a list expands
it to one argument per element. A command that is a list rather than a string skips the
shell entirely and is exec'd directly. A command that is a string a script assembled
itself still runs verbatim, which is the escape hatch for anything the other two forms
can't express.

The rule underneath all three: **the literal text you write is shell, and interpolations
are data.** This is a breaking change, shipped with a static diagnostic (RAD40023) that
finds the affected lines.

## Context / Motivation

Rad pasted interpolated values into the command string as raw text and handed the result
to `<shell> -c`. The shell then parsed that string as a program, with no way to tell which
characters came from the script's source and which came from a variable. Every value was
potential syntax.

The trigger was a security report: a value carrying `; rm -rf ~` into a command would run
it, and Rad's own kan hooks interpolate user-typed card titles straight into `kan edit`.
But injection is not the reason we changed the default, and treating it as one would have
produced the wrong fix. Injection needs a hostile value. **Word splitting needs a space**,
and that fires constantly on ordinary data. In our own script corpus, one script forwarded
its arguments with `get_args()[1:].join(" ")` and destroyed any argument containing a
space; another called `message.replace("'", "")` to *delete apostrophes* from user text
rather than fight the quoting. Neither author believed they were handling untrusted input.
They were handling filenames and reminder text.

That distinction decides the shape of the fix. A helper you must remember to call is a
plausible answer to a security problem and a useless one to a correctness problem, because
the people who need it don't know they do.

The deeper reason is that **no quoting a script writes in advance can be correct**. The
correct quoting depends on the value, and the script doesn't have the value yet:

| Written           | Value            | Old result                        |
|-------------------|------------------|-----------------------------------|
| `-m "{msg}"`      | `it's`           | works                             |
| `-m "{msg}"`      | `100% of $USERS` | silently becomes `100% of `       |
| `-m '{msg}'`      | `a b`            | works                             |
| `-m '{msg}'`      | `it's o'clock`   | silently becomes `its oclock`     |
| `-m {msg}`        | `a b`            | silently becomes two arguments    |

Double quotes still expand `$` and command substitution; single quotes cannot contain an
apostrophe at all. Both spellings work until they don't, and when they fail they mostly
fail silently.

There was no prior decision here to overturn. `docs/thinking/shell_cmds.md` (2024-10)
considered whether Rad should sandbox scripts that run arbitrary commands and concluded it
shouldn't - a question about trusting script *authors*. Escaping interpolated *data* is
not discussed anywhere in that document. The design space was open, not settled.

Finally, this is a claim Rad already makes. [RED-2](0002-why-rad.md) positions Rad against
Bash, and `VISION.md` names "a quoting minefield" and "generated bash is still bash
(quoting bugs, no validation)" as the thing a Rad script is supposed to beat. The shell
guide opened by calling shell commands "safe". Leaving the default unsafe made that
marketing.

## Decision

Three forms, distinguished by what the script wrote, because that is what says who owns
the shell syntax.

**A command literal: the text is shell, interpolations are data.**

```rad
message = "it's fine"
files = ["My Notes.txt", "b.txt"]

$`git commit -m {message}`
// runs: git commit -m 'it'\''s fine'

$`cat {files} | grep needle > out.txt`
// runs: cat 'My Notes.txt' b.txt | grep needle > out.txt
```

Pipes, redirects, `&&`, globs and `$VAR` in the literal text keep working. Only
interpolations changed. A scalar becomes one argument; a list becomes one argument per
element, and an empty list contributes none. Because the shell concatenates adjacent
quoted fragments, `--flag={x}` and `{a}-{b}` remain single arguments without the script
quoting anything.

Values made only of `A-Z a-z 0-9 @ % + = : , . / _ -` are emitted bare, so most real
commands produce byte-identical output to before.

**A list command: an argument vector, no shell at all.**

```rad
cmd = ["ffmpeg", "-i", path]
if start:
    cmd += ["-ss", start]
$cmd
```

This is exec'd directly. Nothing in it can be reinterpreted, because nothing parses it. It
gives up what the shell provides - pipes, redirects, globs, `$VAR`, and builtins like `cd`
- and in exchange it is safe by construction rather than safe by quoting.

**A string command: verbatim.**

```rad
cmd = `ls {dir}`   // interpolation happens here, in an ordinary string
$cmd               // by now it is just text
```

Unchanged. The interpolation already happened when the string was built, so there is
nothing left for `$` to protect.

Two shapes that previously produced a mangled command now fail under RAD20045: a value
with no single-argument form (a map, or a null that would have become the word `null`),
and a list that isn't standing alone as its own argument, such as `--file={files}`.

## Rationale

**Correctness has to be the default; security features can be opt-in.** This is the whole
argument. Framed as a security tool, a quoting helper is something a careful author reaches
for when handling untrusted input. Framed as what it is - a correctness bug that fires on
any filename with a space - it has to be automatic, because the failure lands hardest on
people who don't know they're at risk.

**Only the interpreter knows enough to quote correctly.** Rad resolves the shell at
runtime from `$SHELL`, which can be `sh`, `bash`, `zsh`, `pwsh` or `cmd.exe`. A quoting
function evaluated during expression evaluation has no idea which of those will consume
its output; the execution seam does. Any design that puts escaping in a builtin is
structurally unable to get this right, and would have had to be replaced later.

**The lexical rule is simpler than what it replaces.** "The text is shell, the
interpolations are data" is one sentence, and it matches what someone reading
`` $`git commit -m {message}` `` already assumes. The behavior it replaced - hand-placed
quotes that work for some values - was never stateable that briefly.

**A list command gives us the argv form for free.** The grammar already accepted any
expression after `$`; a list simply errored at runtime. Making it meaningful cost no
syntax and no tree-sitter release, and it answers the "N filenames as N arguments" problem
that scripts previously solved by writing a temp file and using `--files-from`.

**Splatting lists implicitly follows the same logic as quoting scalars.** We considered
requiring an explicit spread. A list interpolated into a command was a hard error before,
so there was no ambiguity to resolve and no existing behavior to preserve - and demanding
a marker would reintroduce the thing we removed, a piece of ceremony the script must
remember or silently get wrong.

## Alternatives Considered

- **A `shell_escape()` builtin.** The obvious fix, and the one first proposed. Rejected on
  three counts. It leaves the default unsafe, so it only helps authors who already know
  they have a problem. It cannot be correct in principle, because a pure function doesn't
  know which shell will consume its output. And it recreates the footgun it removes: it is
  only correct *outside* quotes, so `` $`git commit -m "{shell_escape(m)}"` `` is broken,
  looks careful, and is what a careful person writes.
- **A `{x:q}` or `{x!q}` format specifier.** Terser, and precedented in Jinja and Ansible.
  Rejected: still opt-in, so it inherits the first objection above. It also wants a
  tree-sitter release, and it puts an escaping concern in a slot that is otherwise purely
  about display - alignment, padding, precision. Rad's "no cryptic symbols" principle
  argues against the `!q` spelling specifically.
- **An argv-only form, with no change to command literals.** Safe by construction and
  clean, but it abandons the commands that need a real shell - roughly a fifth of the
  corpus uses pipes, redirects, `&&`, `cd` or background jobs. Those would have kept the
  old hazard permanently. We took this as the *second* form rather than the only one.
- **A lint with no behavior change.** Rejected: `rad check` already demoted one
  correct-but-noisy diagnostic to `--strict` after it produced roughly three-quarters of
  all hints on a real-script sweep. A warning on every interpolated command would land in
  the same bucket, and warning about a hazard we could simply remove is the wrong trade.
- **Absorbing quotes written around an interpolation**, so `"{x}"` and `{x}` mean the same
  thing. This would have made most of the migration a no-op. Rejected: it does nothing for
  `"literal text {x}"`, which needs restructuring either way, and it makes the source lie
  about what runs. A diagnostic that says "delete these quotes" is more honest.

## Compatibility & Migration

**This breaks scripts, and one of the two broken spellings fails silently.** A command
with hand-written quotes around an interpolation regresses: `-m "{msg}"` with a value
containing an apostrophe produces visible garbage, while `-m '{msg}'` with a value
containing a space silently splits into two arguments. Clean values are unaffected,
because they are emitted bare.

RAD40023 flags every interpolation sitting inside a quote the script wrote. It tracks quote
state across the command literal rather than only checking the adjacent characters, so it
also catches `"Bump to {version}"`, where the value is quoted along with a literal prefix.
A sweep of the author's 43-script `bin` directory found 45 sites across 15 files.

The migration is mechanical in two of three shapes: delete the quotes. The third -
literal text plus a value in one argument - needs the string hoisted into a variable
first, which is better code anyway.

**What the diagnostic cannot see** is a command assembled into a string elsewhere and run
with `$cmd`. By then the interpolation has happened and there is nothing left to flag. That
form is unchanged, but it is also where the old quoting bugs live, so scripts that build
command strings need reading by hand. Converting them to list commands is the fix.

## Other Consequences & Trade-offs

**List commands work on Windows, where shell commands don't.** Shell command support on
Windows is currently broken outright. The argv form resolves no shell, so it sidesteps the
problem entirely and is the first form usable there. That was a consequence, not a goal.

**The `⚡️` echo now shows the command after quoting**, which is what actually ran. It
diverges from the source text for values that needed quoting. We accept that: the echo is
a debugging aid, and showing text that differs from what was executed would be worse.

**Quoting is POSIX-only.** We emit the `'\''` single-quote idiom, verified correct on
`sh`, `bash`, `zsh`, `dash` and `ksh`. It is wrong on at least four shells, and they are
not all on Windows:

- `cmd.exe` gives `'` no quoting meaning, so a value still word-splits and the quotes
  arrive as literal characters.
- PowerShell escapes an embedded quote as `''`, so `'\''` ends the string early.
- `csh`/`tcsh` expand `!` for history even inside single quotes, so `hello!world`
  interpolates to `Event not found`.
- `xonsh` rejects the `'\''` idiom outright, so any apostrophe or newline in a value is a
  syntax error.

`fish` is untested. `resolveShell` reads `$SHELL` verbatim and falls back to a
`pwsh`/`powershell`/`cmd` chain on Windows, so it can hand the quoter any of these. The
earlier draft of this section called that a Windows-only debt on the grounds that shell
commands don't work there; that was wrong twice over. Windows resolution is real and
tested, and `csh`/`tcsh`/`xonsh` are not Windows at all. On those shells a value that
used to paste through unquoted now fails outright, so this is a regression for them, not
a no-op.

The list form has no such assumption - it never reaches a shell - which is the answer for
anyone on an affected shell, and a further reason to prefer it generally. The real fix is
to pick the quoting from the resolved shell, or to refuse the string form where we cannot
quote correctly.

**Interpolation now behaves differently inside a `$` command than in an ordinary string.**
[RED-6](0006-bash-embedding.md) deliberately made Rad's quote forms interchangeable, and
this is the first place where context changes what `{x}` means. We think the distinction is
justified - a command is not a string, it is a program - but it is a genuine new thing to
learn, and the `cmd = \`ls {dir}\`` / `$cmd` pair now behaves differently from the same
text written inline. That asymmetry is the least comfortable part of this design.

**Rejecting maps and nulls makes some previously-running scripts fail.** A null used to
become the word `null` in a command. That was always a bug; it now stops the script instead
of producing a wrong command. We consider this a gain and note it as a break.

## Future Directions

The `--shell` export mode has the same defect independently: it emits `NAME="value"` with
no escaping, and that output is `eval`'d by the calling Bash script - the embedding pattern
[RED-6](0006-bash-embedding.md) documents. It is tracked separately and unfixed.

Whether the string command form (`$cmd` on assembled text) should eventually warn, or be
retired in favor of list commands, is open. It is the only remaining place where a script
can build a command unsafely, but it is also the escape hatch, and we would want evidence
that it isn't carrying real weight before touching it.

## References

- [RED-2](0002-why-rad.md) - Rad exists to replace Bash; the quoting minefield is one of
  the specific things it claims to beat, and this RED is what makes that claim true for
  shell commands.
- [RED-6](0006-bash-embedding.md) - established that Rad's quote forms behave identically;
  this decision introduces the first context where an interpolation does not.
- zx (`google.github.io/zx/quotes`) - the closest prior art, being the same
  `` $`...` `` template-literal shape. It escapes interpolations automatically, splats
  arrays as one quoted element each, and ships no opt-out at all, directing users to
  assemble arrays instead. We landed in the same place independently and kept a raw form
  where zx has none.
- Nushell - external commands are argv-native with an explicit `...list` spread. We chose
  implicit splatting; see the Rationale.

---

## History

- 2026-08-10 Implemented. Proposed, decided and shipped in a single session, replacing an
  earlier plan to add a `shell_escape()` builtin.
