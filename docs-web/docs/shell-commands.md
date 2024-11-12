---
title: Shell Commands
---

## Basic Shell Commands

```rsl
$`ls -l`
fail:
    print("Command failed! Exiting script!")
```

```rsl
$`ls -l`
recover:
    print("Command failed! Continuing script...")
```

## Critical Shell Commands

```rsl
$!`ls -l`
```

## Unsafe Shell Commands

```rsl
unsafe $`ls -l`
```

## Output Capture

```rsl
err_code = $!`ls -l`
err_code, stdout = $!`ls -l`
err_code, stdout, stderr = $!`ls -l`
```

## Suppress Announcements

By default, Rad will 'announce' (i.e. print) commands as they're executed. Example:

```rsl title="Without quiet"
$!`ls`
$!`echo hello`
```

```title="Without quiet output"
⚡️ Running: ls
pick.rsl  simple.rsl  sorting.rsl
⚡️ Running: echo hello
hello
```

These announcements can be suppressed with the `quiet` keyword. It does not impact stdout/stderr output for the command.

```rsl title="With quiet"
quiet $!`ls`
quiet $!`echo hello`
```

```title="With quiet output"
pick.rsl  simple.rsl  sorting.rsl
hello
```
