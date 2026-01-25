---
title: Welcome to Rad
---

**Rad is a scripting language designed to make writing CLI tools delightful.**

Familiar, Python-like syntax with CLI superpowers built-in: declarative argument parsing, JSON processing, shell integration, and interactive prompts.

## A Quick Taste

```rad
#!/usr/bin/env rad
args:
    name str        # Name of the person to greet.
    times int = 1   # How many times to greet them.

for i in range(times):
    print("Hello, {name}!")
```

```
> ./greet --help
Usage:
  greet <name> [times]

Script args:
      --name str    Name of the person to greet.
      --times int   How many times to greet them. (default 1)
```

No argparse. No boilerplate. Just readable code that does what you want.

## Explore

- [**Getting Started**](./guide/getting-started.md) – Installation & your first script
- [**Args**](./guide/args.md) – Declarative argument parsing
- [**Rad Blocks**](./guide/rad-blocks.md) – JSON processing & API queries
- [**Shell Commands**](./guide/shell-commands.md) – Running shell commands
- [**Examples**](./examples/index.md) – Real-world scripts

**→ [Install Rad](./guide/getting-started.md#installation)**

!!! note "Early Development"
    Rad is in early development but useful today. Core features work well, though expect breaking changes between versions. [Feedback welcome!](https://github.com/amterp/rad/discussions)
