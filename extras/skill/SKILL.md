---
name: rad
description: >
  Write and debug Rad scripts. Use when asked to: create new Rad scripts,
  modify existing Rad code, debug Rad syntax errors, write argument parsing
  with args blocks, use shell commands with dollar-backtick syntax, create
  display/rad blocks for JSON processing, or follow Rad style conventions.
---

# Rad Scripting

Rad is a scripting language for writing CLI tools - Python-like syntax with CLI
essentials (declarative args, shell integration, JSON/HTTP, tables) built in.
Reach for it instead of Bash whenever a script needs arguments, structured
data, or would otherwise grow into unreadable Bash.

## Read the docs from the binary

The installed `rad` ships its own documentation, always matching its version.
**Before writing or modifying Rad code, pull what you need with `rad docs`:**

```sh
rad docs                      # list every topic
rad docs all                  # the entire corpus (guides + reference + examples)
rad docs reference/functions  # every built-in function and signature
rad docs reference/syntax     # the full language reference
rad docs guide/<topic>        # e.g. guide/args, guide/shell-commands, guide/rad-blocks
rad docs RAD20001             # explain an error code
```

Do NOT guess function names or signatures - `rad docs reference/functions` is
authoritative for the installed version. When in doubt about syntax, `rad docs
all` loads everything in one shot (pipe it - output is raw markdown when not a
TTY).

If `rad docs` is unavailable (an older Rad without it), fall back to
https://amterp.dev/rad/ and its function reference.

## Minimal script shape

```rad
#!/usr/bin/env rad
---
Script description for --help.
---
args:
    name str        # Required argument
    count int = 5   # Optional with default
    verbose v bool  # Flag with short form

print("Hello {name}!")
```

`rad docs guide/args` and `rad docs guide/getting-started` cover the rest.

## Gotchas (where Rad differs from Python/Bash intuition)

- **Logic/negation** are keywords: `not`, `and`, `or` - never `!`, `&&`, `||`.
- **Comments** are `//` in the body (`#` is only for arg help text).
- **Shell commands** use `$` + backticks (`` $`cmd` ``) and are *critical by
  default*: a non-zero exit propagates as an error unless you add a `catch:`.
  Capture forms: `code = $`cmd``, `code, stdout = $`cmd``, plus stderr.
- **Errors**: `x = risky() ?? fallback` for a default, or a `catch:` block for
  richer handling.
- **Arg names** use underscores (`dry_run`); Rad exposes them hyphenated
  (`--dry-run`) automatically.
- **UFCS**: prefer `text.trim().upper()` over `upper(trim(text))`.

## Style

- snake_case for variables and arg names.
- Prefer dollar-backtick shell syntax over wrapping `bash -c`.
- Align arg `#` comments two spaces past the longest arg.
- Lean on UFCS chains for readability.

## CLI Design Patterns

Two ways to structure a script's interface (judgment, not syntax):

### Command pattern
Mutually exclusive verbs, each with its own args - like `git`, `docker`.

```rad
command build:
    ---Build the project.---
    calls do_build

command test:
    ---Run tests.---
    calls do_test
```
Usage: `./script build` or `./script test` (pick one).

### Composable flags
Additive boolean flags that combine freely; each enables an independent step.

```rad
args:
    build b bool     # Build the project
    test t bool      # Run tests
    frontend f bool  # Build frontend

if build:
    do_build()
if frontend:
    do_frontend()
if test:
    do_test()
```
Usage: `./script -bf` (build + frontend), `./script -bft` (all three).

### When to use which
- **Command pattern**: actions are conceptually distinct or mutually exclusive.
- **Composable flags**: actions are steps users may want to combine. Prefer
  this for dev-automation scripts (build/test/run/deploy).
