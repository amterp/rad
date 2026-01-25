---
title: Project Status
---

# Project Status

Rad is in **early development** but actively maintained and useful for real scripts today.

**What this means:**

- Core features work well
- Actively maintained and improving
- Breaking changes may occur between minor versions
- Some rough edges and missing features remain

## Platform Support

| Platform    | Status          | Notes                                      |
|-------------|-----------------|--------------------------------------------|
| **macOS**   | Fully Supported | Primary development platform               |
| **Linux**   | Fully Supported | Works well due to Unix similarity          |
| **Windows** | Experimental    | Limited support, some features unavailable |

### macOS

Primary development and testing platform. All features fully supported on both Apple Silicon (arm64) and Intel (amd64).

### Linux

Works reliably due to Unix similarity. Statically linked binaries provide broad compatibility across distributions. Both
amd64 and arm64 architectures supported.

### Windows

**Experimental support** - use with caution.

- Shell command integration (`` $`...` ``) does not currently work
- Some edge cases may have bugs
- Only x86-64 (Intel/AMD) architecture currently available (no ARM)
- Community bug reports welcome to help improve support

## Near-term Focus

We're balancing quick wins (new functions, bug fixes) with building out major features so we can learn how they fit together as the language evolves.

Currently we're working on:

- Continuing to build out the [script command syntax](./guide/script-commands.md)
- Iterating and adding to the json path syntax
- Making errors more informative and user-friendly
- Improving IDE integrations & support
- Ongoing: bug fixes, new functions, documentation improvements

## Long-term Vision

Rad aims to be *the best* way to write CLI scripts. That's a high bar, and we're building toward it deliberately.

What this means:

- Rad scripts will need to cover as many use cases as possible
  - Writing idiomatic Rad should be natural and easy, and the result should be better than equivalent scripts in other languages
- Writing and reading Rad should be a joy
  - User-friendly errors, clear syntax for newcomers, high-quality documentation, etc
- Top-class tooling and supporting infrastructure
  - Language servers, static analysis, editor integrations, easy installation, etc

This is a long-term ambition, and we're building it piece by piece.

## Get Involved

Your feedback directly shapes the language!

- [GitHub Discussions](https://github.com/amterp/rad/discussions) - Questions, ideas, and general discussion
- [GitHub Issues](https://github.com/amterp/rad/issues) - Bug reports and feature requests
