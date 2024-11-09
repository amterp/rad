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
