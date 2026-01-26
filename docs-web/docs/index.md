---
title: Welcome to Rad
hide:
  - navigation
  - toc
---

<div class="rad-hero" markdown>
<div class="rad-hero-inner" markdown>

<div class="rad-hero-content" markdown>

<h1 class="rad-hero-title">Rad</h1>

<p class="rad-hero-tagline">
A scripting language designed to make writing <strong>CLI tools delightful</strong>. Familiar, Python-like syntax with CLI superpowers built-in.
</p>

<div class="rad-hero-buttons" markdown>

[Get Started](./guide/getting-started.md){ .md-button .md-button--primary }
[View Examples](./examples/index.md){ .md-button }

</div>

</div>

<div class="rad-hero-code" markdown>
<div class="rad-hero-code-header">
<span class="rad-hero-code-dot"></span>
<span class="rad-hero-code-dot"></span>
<span class="rad-hero-code-dot"></span>
<span class="rad-hero-code-title">greet.rad</span>
</div>

```rad
#!/usr/bin/env rad
args:
    name str        # Name to greet.
    times int = 1   # How many times.

for i in range(times):
    print("Hello, {name}!")
```

</div>

</div>
</div>

## Why Rad?

**No boilerplate.** Declarative argument parsing, built-in JSON processing, and shell integration—all in a clean, readable syntax.

**Familiar.** If you know Python, you already know most of Rad. The learning curve is gentle; the productivity gains are immediate.

**Delightful.** Rad is designed to be fun to use. Small touches like helpful error messages and intuitive syntax make scripting enjoyable again.

## Quick Example

```
> ./greet --help
Usage:
  greet <name> [times]

Script args:
      --name str    Name to greet.
      --times int   How many times. (default 1)
```

```
> ./greet Alice --times 2
Hello, Alice!
Hello, Alice!
```

## Explore

- [**Getting Started**](./guide/getting-started.md) – Installation & your first script
- [**Args**](./guide/args.md) – Declarative argument parsing
- [**Rad Blocks**](./guide/rad-blocks.md) – JSON processing & API queries
- [**Shell Commands**](./guide/shell-commands.md) – Running shell commands
- [**Examples**](./examples/index.md) – Real-world scripts

!!! note "Early Development"
    Rad is in early development but useful today. Core features work well, though expect breaking changes between versions. [Feedback welcome!](https://github.com/amterp/rad/discussions)
