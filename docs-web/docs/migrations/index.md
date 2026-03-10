---
title: How to Migrate
---

# How to Migrate

Rad is pre-1.0, so breaking changes can occur between minor versions. When they do, the version gets its own migration guide (listed below this page in the sidebar) detailing exactly what changed and how to update your scripts.

This page covers the general process for upgrading smoothly.

## Upgrade Workflow

### 1. Check your current version

```shell
rad --version
```

### 2. Read the migration guide

Before upgrading, skim the migration guide for your target version. You'll find them in the sidebar under **Migrations**. Each guide lists every breaking change with before/after examples and migration steps.

If you're jumping multiple versions, read each guide in order - changes can build on each other.

### 3. Upgrade Rad

Upgrade using whichever method you originally installed with - your package manager, `go install`, or by downloading the latest binary from the [releases page](https://github.com/amterp/rad/releases). See [Installation](../guide/getting-started.md#installation) for details.

### 4. Check your scripts

Run `rad check --from-logs` to bulk-check scripts you've recently used:

```shell
rad check --from-logs all
```

This reads Rad's invocation log to find scripts you've actually run, then checks each one for compatibility issues. Here's what the output looks like:

```
                                                        Errors  Warns   Info  Hints
  ✓ /Users/alice/scripts/deploy                              0      0      0      0
  ✗ /Users/alice/scripts/brewu                               1      0      0      0
  ✓ /Users/alice/scripts/epoch                               0      1      0      0
  ✓ /Users/alice/src/myproject/dev                           0      0      0      1
  ...

───────────────────────────────────────────────────────────────────────────────────────
Checked 38 scripts: 37 passed, 1 failed (1 warn, 1 hint, 5 skipped).
```

It's the fastest way to find what needs updating.

The `--from-logs` flag takes a duration value that controls how far back to look. Use `all` to check everything, or narrow it down if the full log catches too many old or irrelevant scripts:

```shell
rad check --from-logs 30d
```

Duration values support units like `d` (days), `w` (weeks), `h` (hours), and combinations like `2w3d`.

You can also check individual scripts directly:

```shell
rad check ./my-script.rad
```

!!! tip "Keep invocation logging enabled"
    `rad check --from-logs` relies on Rad's invocation logging, which is enabled by default since v0.9. If you've disabled it, consider re-enabling it so it's ready when you need it. See the [Configuration guide](../guide/config.md) for details.

### 5. Fix issues

When you hit a breaking change - either through `rad check` or by running a script - Rad tells you what changed and how to fix it. For example, running a script that uses a removed keyword produces:

```
error[RAD40008]: 'request' blocks have been removed. Use 'rad' instead.
  --> script.rad:1:1
   |
 1 | request "https://api.example.com/users":
   | ^^
   |
   = help: See migration guide: https://amterp.github.io/rad/migrations/v0.9/
   = info: rad explain RAD40008
```

The inline error gives you the gist, but if you want more detail, run `rad explain` with the error code:

```shell
rad explain RAD40008
```

```
RAD40008: Deprecated Block Keyword

The request and display block keywords have been removed in v0.9.
All rad block variants now use the unified rad keyword.

Migration

Replace request and display with rad:

    # Before (no longer works)
    request "https://api.example.com/data":
        fields Name, Age

    # After
    rad "https://api.example.com/data":
        noprint
        fields Name, Age
```

Between the inline hint and `rad explain`, you should have enough context to update your script and move on to the next one.

### 6. Test

Run your scripts and verify they behave as expected. Static checks (`rad check`) catch syntax and naming issues, but some changes are behavioral - like an operator producing a different result for the same input. The migration guides call these out, so you'll know what to watch for.
