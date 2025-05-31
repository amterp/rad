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
      --src-tree               Instead of running the target script, print out its syntax tree.
      --mock-response string   Add mock response for json requests (pattern:filePath)
```

[//]: # (todo script something to keep the above blob in check)

Note that, outside of `help`, all the global flags are ALL CAPS.

## `debug`

[`debug`](../reference/functions.md#debug) is an built-in function which behaves exactly like `print`, except that it only prints if the global flag `--debug` is enabled. You can use them in your script for debugging as desired.

For example, given this example:

```rad title="debug.rsl"
print("1")
debug("2")
print("3")
```

the following invocations will give the respective outputs:

```
rad debug.rsl
```

<div class="result">
```
1
3
```
</div>

```
rad debug.rsl -d
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

## Additional Commands

There are more global flags - see the [reference](../reference/global-flags.md) for a complete coverage of what's available.

## Summary

- Rad provides several global flags that can be used across all Rad scripts.
- Use `mock-response` to test your scripts.

!!! info "Script args can shadow global flags"

    If a script defines an arg such as `debug`, conflicting with an existing global flag, then the script arg will **shadow** the global flag.

    This means that the global flag's functionality is effectively disabled for the script. It gets removed from the script's usage string, and
    the script itself defines the behavior of the flag.

## Next

Sometimes you may wish to run commands before your script ends, either normally or via an error, such as cleanups.
Rad provides a way to do this that we will explore in the next section: [Defer & Errdefer](./defer-errdefer.md).
