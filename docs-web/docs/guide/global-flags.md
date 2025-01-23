---
title: Global Flags
---

RSL offers a range of global flags that are available with every RSL script. We'll explore some of them in this section.

## `help`

The most basic global flag is `--help` or `-h`. *All* RSL scripts automatically generate a usage string that can be displayed by invoking this flag.

`--help` also prints available *global* flags:

```
Global flags:
  -h, --help                   Print usage string.
  -D, --DEBUG                  Enables debug output. Intended for RSL script developers.
      --RAD-DEBUG              Enables Rad debug output. Intended for Rad developers.
      --NO-COLOR               Disable colorized output.
  -Q, --QUIET                  Suppresses some output.
      --SHELL                  Outputs shell/bash exports of variables, so they can be eval'd
  -V, --VERSION                Print rad version information.
      --STDIN script-name      Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.
      --MOCK-RESPONSE string   Add mock response for json requests (pattern:filePath)
```

Note that, outside of `help`, all the global flags are ALL CAPS.

## `DEBUG`

[`debug`](../reference/functions.md#debug) is an built-in function which behaves exactly like `print`, except it only prints if the global flag `--DEBUG` is enabled. You can use them in your script for debugging as desired.

For example, given this example:

```rsl title="debug.rsl"
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
rad debug.rsl -D
```

<div class="result">
```
1
DEBUG: 2
3
```
</div>

## `QUIET`

Use `--QUIET` or `-Q` to suppress *some* outputs, including print statements and errors. Some outputs still get printed e.g. shell command outputs.

## `NO-COLOR`

A lot of rad's outputs have colors e.g. [`pick`](../reference/functions.md#pick) interaction or [`pprint`](../reference/functions.md#pprint) JSON formatted output.
Sometimes you just want monochrome output, and while rad aims to do this automatically when it detects e.g. you're redirecting output to file, you can force it by using the 
`--NO-COLOR` flag.

## `MOCK-RESPONSE`

You might be writing a script which hits a JSON API and uses its output e.g. formatting it into a table using a [`rad` block](./rad-blocks.md).

In writing said script, you may wish to test it against certain responses that the live API isn't giving you at the moment, perhaps because the server is down. To accomplish this, you can use the `MOCK-RESPONSE` flag.

`MOCK-RESPONSE` takes an argument in a `<url regex>:<file path>` format.
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
rad commits.rsl --MOCK-RESPONSE "api.github.*:commits.json"
```

Before executing the HTTP request, rad checks for defined mock responses and if there's a regex match against the URL, it will short circuit,
avoiding the HTTP request, and simply returning the contents of the mocked response.

!!! tip "Match all URLs with .*"

    It's common for scripts to perform just one API query, in which case the regex filter doesn't need to be specific.
    Instead, you can just write `.*` e.g. `.*:commits.json`.

[//]: # (todo can be set several times?)

## Additional Commands

There are more global flags - see the [reference](../reference/global-flags.md) for a complete coverage of what's available.

## Learnings Summary

- Rad provides several global flags that can be used across all RSL scripts.
- Generally, global flags are in ALL CAPS, such as `DEBUG` and `QUIET`.
- Use `MOCK-RESPONSE` to test your scripts.

## Next

Sometimes you may wish to run commands before your script ends, either normally or via an error, such as cleanups.
RSL provides a way to do this that we will explore in the next section: [Defer & Errdefer](./defer-errdefer.md).
