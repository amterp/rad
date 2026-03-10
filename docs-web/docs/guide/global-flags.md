---
title: Global Flags
---

Rad offers a range of global flags that are available with every Rad script. We'll explore some of them in this section.

## `help`

The most basic global flag is `--help` or `-h`. *All* Rad scripts automatically generate a usage string that can be displayed by invoking this flag.

`--help` also prints available *global* flags:

```
Global flags:
  -h, --help                   Print usage string.
  -d, --debug                  Enables debug output. Intended for Rad script developers.
      --rad-debug              Enables Rad debug output. Intended for Rad developers.
      --color mode             Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet                  Suppresses some output.
      --shell                  Outputs shell/bash exports of variables, so they can be eval'd
  -v, --version                Print rad version information.
      --stdin script-name      Enables reading Rad from stdin, and takes a string arg to be treated as the 'script name'.
      --confirm-shell          Confirm all shell commands before running them.
      --src                    Instead of running the target script, just print it out.
      --cst-tree               Instead of running the target script, print out its concrete syntax tree.
      --ast-tree               Instead of running the target script, print out its abstract syntax tree.
      --tls-insecure           Skip TLS certificate verification for HTTPS requests.
      --mock-response string   Add mock response for json requests (pattern:filePath)
```

[//]: # (todo script something to keep the above blob in check)

Note that, outside of `help`, all the global flags are ALL CAPS.

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

## Summary

- Rad provides several global flags that can be used across all Rad scripts.
- Use `--src`, `--cst-tree`, and `--ast-tree` to inspect scripts without running them.
- Use `--tls-insecure` for development against self-signed certs.
- Use `--mock-response` to test your scripts against canned API responses.

!!! info "Script args can shadow global flags"

    If a script defines an arg such as `debug`, conflicting with an existing global flag, then the script arg will **shadow** the global flag.

    This means that the global flag's functionality is effectively disabled for the script. It gets removed from the script's usage string, and
    the script itself defines the behavior of the flag.

## Next

Sometimes you may wish to run commands before your script ends, either normally or via an error, such as cleanups.
Rad provides a way to do this that we will explore in the next section: [Defer & Errdefer](./defer-errdefer.md).
