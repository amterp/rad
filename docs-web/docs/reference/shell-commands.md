---
title: Shell Commands
---

## Basic Shell Commands

```rad
$`ls -l`
fail:
    print("Command failed! Exiting script!")
```

```rad
$`ls -l`
recover:
    print("Command failed! Continuing script...")
```

## Critical Shell Commands

```rad
$!`ls -l`
```

## Unsafe Shell Commands

```rad
unsafe $`ls -l`
```

## Output Capture

```rad
err_code = $!`ls -l`
err_code, stdout = $!`ls -l`
err_code, stdout, stderr = $!`ls -l`
```

## Suppress Announcements

By default, Rad will 'announce' (i.e. print) commands as they're executed. Example:

```rad title="Without quiet"
$!`ls`
$!`echo hello`
```

```title="Without quiet output"
⚡️ Running: ls
pick.rad  simple.rad  sorting.rad
⚡️ Running: echo hello
hello
```

These announcements can be suppressed with the `quiet` keyword. It does not impact stdout/stderr output for the command.

```rad title="With quiet"
quiet $!`ls`
quiet $!`echo hello`
```

```title="With quiet output"
pick.rad  simple.rad  sorting.rad
hello
```
