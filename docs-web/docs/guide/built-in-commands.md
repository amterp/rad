---
title: Built-in Commands
---

Rad is more than a script runner - the `rad` binary ships with a small toolbox of commands
for creating, checking, formatting, and managing your scripts. Run `rad` on its own to see them listed.

Every command documents itself: `rad <command> --help` is the complete, always-up-to-date
reference for its flags. This page is a tour, so you know what exists and when to reach for it.

## `rad new`

`rad new` scaffolds a new script, saving you the ritual of creating a file,
adding a shebang, and marking it executable:

```shell
rad new greet
```

<div class="result">
```
⚡️ touch greet
⚡️ chmod +x greet
greet is ready to go.
```
</div>

The result is a ready-to-run executable containing a small Hello World template to edit.
If you just want the shebang without the template, pass `--shebang`.
You can also jump straight into editing with `--open`:

```shell
rad new greet --shebang --open vim
```

## `rad check`

`rad check` validates and lints a script without running it:

```shell
rad check greet
```

It prints any diagnostics - errors, warnings, and hints - with the offending line and an
explanation, and exits non-zero if there are errors. Add `--strict` to also surface
advisory checks, such as unhandled fallible calls.

It has a flag `--from-logs`, which bulk-checks scripts you've actually used - handy
after upgrading Rad:

```shell
rad check --from-logs 30d   # scripts used in the last 30 days ('all' for everything)
```

This is powered by invocation logging - see [Configuration](./config.md#invocation-logging) for details.

## `rad fmt`

`rad fmt` rewrites scripts into Rad's canonical style, in place:

```shell
rad fmt greet
```

<div class="result">
```
Formatted greet
```
</div>

A few useful variations:

- `rad fmt --stdout greet` prints the formatted result instead of writing it.
- `rad fmt --check greet` writes nothing and exits non-zero if anything is unformatted - built for CI.
- With no paths, it formats stdin to stdout, making it easy to plug into editors and pipes.

## `rad docs`

`rad docs` is Rad's built-in documentation browser. It serves docs embedded directly in the binary,
so it always reflects the version you have installed - no browser or network required.

```
rad docs                     # list all topics grouped by section
rad docs guide/basics        # read the Basics guide page
rad docs len                 # look up a single built-in function
rad docs reference/functions # read the functions reference
rad docs migrations          # list the migration guides
rad docs migrations/v0.12    # read one of them
rad docs all                 # dump the full doc corpus
```

`rad docs all` covers the guide, reference, and examples - the same set as
[llms-full.txt](https://amterp.dev/rad/llms-full.txt). Migration guides are
addressable but stay out of it, since they demonstrate syntax that no longer
works.

Error codes are also accessible this way:

```
rad docs RAD20003           # explain a specific error code
rad docs 20003              # shorthand (no RAD prefix needed)
```

By default, `rad docs` renders markdown with color and formatting on a TTY,
and outputs raw markdown when piped. Use `--raw` to force raw output, or `--render`
to force rendered output regardless of destination. Pass `--web` / `-w` to open the
corresponding page on the documentation website instead.

## `rad stash`

Scripts can persist their own data in [stashes](./stashes.md). `rad stash` lets you
inspect and manage those stashes for scripts on your PATH:

```shell
rad stash myscript --state   # view the script's persisted state
rad stash myscript --id      # print the stash ID
rad stash myscript --delete  # delete the stash
```

## `rad gen-id`

`rad gen-id` generates a unique string ID:

```shell
rad gen-id
```

<div class="result">
```
VOMPLCB1B4GN7
```
</div>

By default it generates [FIDs](https://github.com/amterp/flexid) - short, time-ordered,
URL-safe identifiers. Pass `--uuid4` or `--uuid7` if you prefer UUIDs. Its classic use
case is minting [stash IDs](./stashes.md) for new scripts.

## `rad home`

`rad home` prints Rad's home directory - where config, stashes, and logs live:

```shell
rad home
```

<div class="result">
```
/home/alice/.rad
```
</div>

Useful when you want to poke around Rad's files, or reference them from other tools.

## `rad completion`

`rad completion` generates shell tab-completion - for the `rad` CLI itself and even for
your own scripts. It gets a dedicated page: [Shell Completion](./shell-completion.md).

## Summary

- The `rad` binary includes built-in commands beyond running scripts - run `rad` alone to list them.
- `rad <command> --help` is the full reference for any command's flags.
- `rad new` scaffolds scripts; `rad check` lints them; `rad fmt` formats them.
- `rad docs` browses this documentation offline, straight from your terminal.
- `rad stash`, `rad gen-id`, and `rad home` help manage Rad's persisted data.

## Next

One of those commands - `rad completion` - deserves a closer look: tab completion for
rad commands, script flags, and even enum values.

Continue to [Shell Completion](./shell-completion.md).
