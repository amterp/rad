---
title: Global Flags
---

Rad offers a range of global flags that are available with every Rad script. We'll explore some of them in this section.

## `help`

The most basic global flag is `--help` or `-h`. *All* Rad scripts automatically generate a usage string that can be displayed by invoking this flag.

`--help` also prints available *global* flags:

```
Global options:
  -h, --help                Print usage string.
  -i, --interactive         Interactively prompt for script args not already provided, then run.
  -d, --debug               Enables debug output. Intended for Rad script developers.
      --rad-debug           Enables Rad debug output. Intended for Rad developers.
      --color mode          Control output colorization. Valid values: [auto, always, never] (default auto)
  -q, --quiet               Suppresses some output.
      --shell               Outputs shell/bash exports of variables, so they can be eval'd
  -v, --version             Print rad version information.
      --confirm-shell       Confirm all shell commands before running them.
      --tls-insecure        Skip TLS certificate verification for all HTTP requests.
      --src                 Instead of running the target script, just print it out.
      --cst-tree            Instead of running the target script, print out its CST (concrete syntax tree).
      --ast-tree            Instead of running the target script, print out its AST (abstract syntax tree).
      --rad-args-dump       Instead of running the target script, print out an args dump for debugging argument parsing.
      --mock-response str   (optional) Add mock response for json requests (pattern:filePath)
      --reply line:value    Answer a prompt when there's no terminal. Repeatable; repeat a line to answer it again.
      --reply-na line       Assert a prompt won't be reached on this run; rad fails cleanly if it is.
```

[//]: # (todo script something to keep the above blob in check)

Note that when `--shell` is enabled, help output goes to stderr instead of stdout, and stdout gets an `exit 0` statement.
This keeps stdout safe to `eval`: wrappers using `eval "$(rad - --shell "$@" <<< "$script")"` won't execute usage text as shell code, and stop cleanly after showing help.

## `interactive`

Use `--interactive` or `-i` to be walked through a script's args interactively instead of typing them all out. Rad prompts for each arg you haven't already provided on the command line, using a prompt suited to the arg's type:

- enum-constrained strings offer a picker of the valid values
- bools are gathered into a single toggleable multi-pick (defaults pre-checked); a lone bool just asks y/n
- ints, floats, and regex-constrained strings validate as you submit, re-prompting inline on invalid values
- list and variadic args collect one value per line until an empty line

Each answered prompt collapses to a compact transcript line (`--env staging`, `--replicas (skip - default: 3)`), so you can see at a glance what was chosen. Optional args can be skipped (Enter on an empty field, or a dedicated "skip" row in pickers), in which case their defaults apply. If the script defines [commands](./script-commands.md) and you didn't specify one, you first pick the command, then get prompted for its args.

Relational constraints react as you go: an arg excluded by something you've already set is skipped with a note (`Skipping --url (excluded by --file)`), and an arg required by one becomes mandatory - its prompt blocks an empty Enter with `required by --<arg>`. Press Ctrl+C at any prompt to abort without running the script.

`-i` composes with partially-supplied args - anything already on the command line isn't re-prompted:

```
rad deploy.rad -i prod
```

Before running, Rad prints the equivalent non-interactive invocation to stderr, so you can copy it to rerun or script the same call directly:

```
Equivalent: rad deploy.rad --env prod --replicas 3 --force
```

This makes `-i` double as an invocation builder for unfamiliar scripts: answer the prompts once, keep the one-liner.

## `debug`

[`debug`](../reference/functions.md#debug) is an built-in function which behaves exactly like `print`, except that it only prints if the global flag `--debug` is enabled. You can use them in your script for debugging as desired.

For example, given this example:

```rad title="debug.rad"
print("1")
debug("2")
print("3")
```

the following invocations will give the respective outputs:

```
rad debug.rad
```

<div class="result">
```
1
3
```
</div>

```
rad debug.rad -d
```

<div class="result">
```
1
DEBUG: 2
3
```
</div>

## `quiet`

Use `--quiet` or `-q` to suppress *some* outputs, including print statements and errors. Some outputs still get printed e.g. shell command outputs.

## `color`

```
--color mode
    Control output colorization.
    Valid values: [auto, always, never].
    (default auto)
```

A lot of Rad's outputs have colors e.g. [`pick`](../reference/functions.md#pick) interaction or [`pprint`](../reference/functions.md#pprint) JSON formatted output.
By default (`auto`), Rad checks your terminal to detect if it's appropriate to enable colors or not. Things like piping or redirecting output will disable coloring.

However, you can override the automatic detection by explicitly setting `--color=always` or `--color=never` to force having colors, or force *not* having colors, respectively. 

## `src`

Use `--src` to print the source code of a script instead of running it. This is handy when you want to quickly inspect a script without opening it in an editor - for example, checking what a script does before running it.

```
rad my-script.rad --src
```

## `cst-tree`

Use `--cst-tree` to print the concrete syntax tree (CST) of a script instead of running it. The CST is the raw parse tree that directly mirrors the grammar - every token, whitespace, and syntactic element is represented.

This flag bypasses argument validation, so you can inspect the tree even if the script expects arguments you haven't provided.

```
rad my-script.rad --cst-tree
```

This is primarily useful for debugging the parser or understanding how Rad tokenizes your script.

## `ast-tree`

Use `--ast-tree` to print the abstract syntax tree (AST) instead of running the script. The AST is the simplified, semantic tree that Rad actually interprets - syntactic sugar has been desugared, and irrelevant tokens are stripped away.

Like `--cst-tree`, this bypasses argument validation.

```
rad my-script.rad --ast-tree
```

Comparing `--cst-tree` and `--ast-tree` output can help you understand how Rad transforms your code before execution - for instance, how compound assignments like `x += 1` get desugared.

## `tls-insecure`

Use `--tls-insecure` to skip TLS certificate verification for all HTTPS requests made by the script. This is useful when developing against servers with self-signed certificates.

```
rad my-api-script.rad --tls-insecure
```

!!! warning "Development only"

    Don't use this flag in production. It disables certificate verification for *all* requests in the script, making them vulnerable to man-in-the-middle attacks.

## `mock-response`

You might be writing a script which hits a JSON API and uses its output e.g. formatting it into a table using a [`rad` block](./rad-blocks.md).

In writing said script, you may wish to test it against certain responses that the live API isn't giving you at the moment, perhaps because the server is down. To accomplish this, you can use the `mock-response` flag.

`mock-response` takes an argument in a `<url regex>:<file path>` format.
In other words, you can mock responses based on a regex match of the queried URL, and make them return the contents of a specified file.

For example, if you wanted to mock a response from GitHub's API, you could define an example response in a file:

```json title="commits.json"
[
  {
    "sha": "306f3a4ddb3b09747d61a5eab264c3d72fbbc36e",
    "commit": {
      "author": {
        "name": "Alice Smith",
        "date": "2025-01-11T04:15:06Z"
      }
    }
  },
  {
    "sha": "2b642c482b32e4d87924839b2f5a9592670dee69",
    "commit": {
      "author": {
        "name": "Charlie Johnson",
        "date": "2025-01-10T12:21:03Z"
      }
    }
  }
]
```

And then define it as the mock response with the following example invocation:

```shell
rad commits.rl --mock-response "api.github.*:commits.json"
```

Before executing the HTTP request, Rad checks for defined mock responses and if there's a regex match against the URL, it will short circuit,
avoiding the HTTP request, and simply returning the contents of the mocked response.

!!! tip "Match all URLs with .*"

    It's common for scripts to perform just one API query, in which case the regex filter doesn't need to be specific.
    Instead, you can just write `.*` e.g. `.*:commits.json`.

[//]: # (todo can be set several times?)

## `reply`

`input`, `confirm`, `pick`, `pick_kv`, `pick_from_resource`, `multipick`, and `confirm`-gated shell commands all need a person at a terminal. Rad looks for one on stdin and, failing that, on `/dev/tty`, which still works when stdin carries data rather than keystrokes. In CI, cron, and AI agent tool calls, neither is there.

Rather than run until it hits a prompt and dies, rad stops first and prints what the script would ask:

```shell
rad deploy.rad prod
```

```title="stderr"
error[RAD20046]: this script needs to ask you something, but there's no terminal

  deploy.rad prompts in 3 places. rad found no terminal - stdin isn't one and
  /dev/tty isn't available - so it can't ask. Supply the answers up front,
  keyed by line number.

  deploy.rad:5  confirm "Deploy to prod?"
  deploy.rad:6  input   "Release notes"
  deploy.rad:7  pick    options: web-1, web-2, db-1

    rad deploy.rad prod --reply '5:<yes|no>' --reply '6:<value>' --reply '7:<option>'
```

Nothing in the script has run yet, so there is no half-finished work to reason about before you retry. The last line is your own command with a `--reply` per prompt: keys in place, answers blank. Rad won't guess those - whether you mean yes, or which server you want, isn't in the script. Fill in each `<...>` and run it; rad takes whatever you put there literally.

Rad exits `7` when it stops like this, so a wrapper can tell "this run needs input" apart from "this run failed" without reading the message.

Each listed prompt may also carry a note about how it behaves:

| Note | Meaning |
|---|---|
| `filtered - may not prompt, and then ignores its answer` | A `pick` whose filter rad can't work through ahead of time. If it narrows to one option it asks nothing, and any answer you gave is consumed and dropped. |
| `settles on "web-1" - won't ask` | A `pick` whose options and filter are both written literally, leaving one survivor. It takes that option without asking, so rad answers it with `--reply-na` for you. |
| `used as a value - ...` | An interactive function passed around rather than called. Takes `--reply-na` only. |
| `may run more than once - repeat --reply per run` | It sits in a loop, or in a function called from several places. One `--reply` answers one execution. |
| `secret - cannot be answered on the command line` | See [secret inputs](#reply-na) below. |
| `options computed at runtime` | Rad can't list the choices without running the script. They're checked when the prompt is reached. |

### Prompts rad can't answer

One shape takes no `--reply`: passing an interactive function around as a value rather than calling it.

```rad
ask = pick
answer = ask(["dev", "prod"])
```

Answers are addressed by the position of the call, and rad can't see where such a value ends up or how many times it runs - so there is nothing for a value to attach to. It's still listed and still keyed, so `--reply-na` can assert the value is never invoked:

```shell
rad dispatch.rad --reply-na 1
```

That matters for a script that merely *mentions* one on a path this run won't take - a handler map where the interactive entry isn't selected. If the assertion turns out to be wrong, the script stops at the prompt with [RAD20047](../reference/errors.md) rather than acting on a guess.

Answers are matched by the kind of prompt on that line - which is what the blank names - and only `multipick` needs escaping:

| Prompt | Blank | Answer |
|---|---|---|
| `confirm`, `confirm $\`...\`` | `<yes\|no>` | `yes` or `no` (also `y`, `n`, `true`, `false`) |
| `input` | `<value>` | everything after the colon, verbatim; empty takes the arg's `default` |
| `pick`, `pick_kv`, `pick_from_resource` | `<option>` | one option, matched exactly |
| `multipick` | `<option,...>` | comma-separated; `\,` is a literal comma, `\\` a literal backslash |

Only the first colon separates the key from the value, so `--reply 6:https://example.com` works without quoting. For `pick_kv`, answer with the key the user would see, not the value it returns - the listing names it `pick_kv` rather than `pick` for exactly this reason.

Matching is exact. A near miss fails rather than running the wrong thing.

A `multipick` answer must name each option at most once, and must satisfy the prompt's own `min` and `max`. Where those are written literally in the script, a bad count fails at pre-flight rather than partway through the run. Answer with an empty value (`--reply 3:`) to select nothing, which a `min=0` prompt accepts.

### Prompts that run more than once

A prompt inside a loop asks once per pass. Repeat its key to answer each one, in order:

```shell
rad cleanup.rad --reply 12:yes --reply 12:no
```

Answers queue per line rather than globally, so adding a prompt earlier in the script never silently re-targets a later answer. If the loop runs out of answers, rad stops. It does not reuse the last one - a single `yes` approving five hundred deletions is the accident this avoids. When there *is* a terminal, running out just means rad asks you for the rest, so you can pre-answer the first few passes and type the others.

One answer is consumed per execution of the line, not per prompt you would have seen. A filtered `pick` that narrows to one option asks nothing, but still takes its answer - otherwise a pass that quietly skipped would hand every later answer to the wrong one. That answer is then discarded: you answer prompts, and no prompt happened, so the script's own filter decides. The same applies when a `pick`'s list simply happens to hold one entry on this run.

## `reply-na`

Some prompts sit on branches a given run won't take. Rather than invent a value you expect to go unused, assert that the prompt won't be reached:

```shell
rad deploy.rad --dry-run --reply-na 20
```

If the script reaches it anyway, rad fails with [RAD20047](../reference/errors.md) instead of acting on a guess.

This is also a good answer for a `pick` given a filter. Such a call narrows its options first and never asks when one survives, so it often needs no answer at all. Where the options and the filter are both written literally, rad works that out for you and puts `--reply-na` in the suggested command itself. Where either is computed while the script runs, it can't, so the prompt is listed like any other - reach for `--reply-na` on the runs you expect the filter to settle.

Secret inputs are the one case with no `--reply` at all. Command-line arguments are visible to other processes, so a secret `input` can only be answered at a terminal. Use `--reply-na` if this run won't reach it. This holds whether `secret` is written as a literal or computed while the script runs - in the second case rad can't warn you up front, but it still refuses the answer.

!!! note "Scripts that disable global options"

    A script setting `@enable_global_options = false` has no `--reply` flag to give it. Rad says so rather than suggesting a command the script would reject; such a script can only be answered at a terminal.

## Summary

- Rad provides several global flags that can be used across all Rad scripts.
- Use `--reply` and `--reply-na` to run scripts that prompt in CI, cron, or an AI agent.
- Use `--src`, `--cst-tree`, and `--ast-tree` to inspect scripts without running them.
- Use `--tls-insecure` for development against self-signed certs.
- Use `--mock-response` to test your scripts against canned API responses.
- To browse this documentation from the terminal, see [`rad docs`](./built-in-commands.md#rad-docs).

!!! info "Script args can shadow global flags"

    If a script defines an arg such as `debug`, conflicting with an existing global flag, then the script arg will **shadow** the global flag.

    This means that the global flag's functionality is effectively disabled for the script. It gets removed from the script's usage string, and
    the script itself defines the behavior of the flag.

## Next

Sometimes you may wish to run commands before your script ends, either normally or via an error, such as cleanups.
Rad provides a way to do this that we will explore in the next section: [Defer & Errdefer](./defer-errdefer.md).
