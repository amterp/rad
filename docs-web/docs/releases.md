---
title: Release Notes
---

# Release Notes

All Rad releases. Newest first.

---

## [v0.7.1](https://github.com/amterp/rad/releases/tag/v0.7.1) - 2026-01-26

- Release VSCode extension v0.3.0 ([9443d85](https://github.com/amterp/rad/commit/9443d85))
- ci: add automated docs deployment and release notes ([883a593](https://github.com/amterp/rad/commit/883a593))
- ci: add cross-platform testing for macOS and Windows ([e085011](https://github.com/amterp/rad/commit/e085011))
- ci: fix Windows build output filename in cross-platform tests ([4474ee3](https://github.com/amterp/rad/commit/4474ee3))
- ci: treat all platforms as equal first-class citizens ([668a509](https://github.com/amterp/rad/commit/668a509))
- feat: add platform abstraction for Windows compatibility ([5fdf8fd](https://github.com/amterp/rad/commit/5fdf8fd))
- feat: add underscore to v0.7 for-loop migration hint detection ([cf4af1a](https://github.com/amterp/rad/commit/cf4af1a))
- feat: complete platform normalization for Windows compatibility ([83c839e](https://github.com/amterp/rad/commit/83c839e))
- fix: improve gen_fid collision resistance ([0ee37ff](https://github.com/amterp/rad/commit/0ee37ff))

---

## [v0.7.0](https://github.com/amterp/rad/releases/tag/v0.7.0) - 2026-01-16

- docs(SYNTAX.md): add script commands ([9e4e595](https://github.com/amterp/rad/commit/9e4e595))
- feat!: redesign for-loop syntax with explicit context access ([e98f7fc](https://github.com/amterp/rad/commit/e98f7fc))
  * See [v0.7 migration guide](https://amterp.github.io/rad/migrations/v0.7/) for more information.
- feat: add context support to rad block map/filter lambdas ([21976ec](https://github.com/amterp/rad/commit/21976ec))


---

## [v0.6.27](https://github.com/amterp/rad/releases/tag/v0.6.27) - 2026-01-11

- Release VSCode extension v0.2.0 ([80c4102](https://github.com/amterp/rad/commit/80c4102))
- fix: multiline string interpolation followed by content ([5365366](https://github.com/amterp/rad/commit/5365366))

---

## [v0.6.26](https://github.com/amterp/rad/releases/tag/v0.6.26) - 2026-01-02

- ci: optimize ci benchmarks by interleaving & reducing runs ([3a4a7f4](https://github.com/amterp/rad/commit/3a4a7f4))
- feat(pick): implement 'prefer_exact' named arg ([a3092c7](https://github.com/amterp/rad/commit/a3092c7))
- feat: improve syntax error messages with specific diagnostics ([82d9472](https://github.com/amterp/rad/commit/82d9472))
- feat: unpackage radls from vs code extension, use PATH instead ([d79186f](https://github.com/amterp/rad/commit/d79186f))

---

## [v0.6.25](https://github.com/amterp/rad/releases/tag/v0.6.25) - 2025-12-16

- Release VSCode extension v0.1.13 ([c492a49](https://github.com/amterp/rad/commit/c492a49))
- docs(thinking): Rewrite imports.md and add some quick counter thoughts ([f50a960](https://github.com/amterp/rad/commit/f50a960))
- feat: add additional get_path fields e.g. modified_millis ([f536f12](https://github.com/amterp/rad/commit/f536f12))
- feat: allow parse_epoch to accept floats ([c15def5](https://github.com/amterp/rad/commit/c15def5))
- fix: avoid Go 'MISSING' formatting malformed print in shell cmds ([6a69141](https://github.com/amterp/rad/commit/6a69141))

---

## [v0.6.24](https://github.com/amterp/rad/releases/tag/v0.6.24) - 2025-12-14

- feat: add flat_map function ([0f72798](https://github.com/amterp/rad/commit/0f72798))

---

## [v0.6.23](https://github.com/amterp/rad/releases/tag/v0.6.23) - 2025-12-10

- feat: allow min/max to accept var args of numbers ([fb6582d](https://github.com/amterp/rad/commit/fb6582d))


---

## [v0.6.21](https://github.com/amterp/rad/releases/tag/v0.6.21) - 2025-12-03

- ci: use full brew hash ([4dfb9ed](https://github.com/amterp/rad/commit/4dfb9ed))

---

## [v0.6.20](https://github.com/amterp/rad/releases/tag/v0.6.20) - 2025-12-03

- ci: attempt brew CI fix by pulling latest commit ([775afad](https://github.com/amterp/rad/commit/775afad))

---

## [v0.6.19](https://github.com/amterp/rad/releases/tag/v0.6.19) - 2025-12-02

- ci: attempt to fix homebrew release ([da006f4](https://github.com/amterp/rad/commit/da006f4))

---

## [v0.6.18](https://github.com/amterp/rad/releases/tag/v0.6.18) - 2025-11-29

- feat: implement 'filter' rad block field modifier syntax ([022267b](https://github.com/amterp/rad/commit/022267b))
- fix: un-indent command descriptions ([5f1480e](https://github.com/amterp/rad/commit/5f1480e))

---

## [v0.6.17](https://github.com/amterp/rad/releases/tag/v0.6.17) - 2025-11-20

- Release VSCode extension v0.1.12 ([d70dbbe](https://github.com/amterp/rad/commit/d70dbbe))
- feat(checker): add warning against undefined cmd callback references ([868d6f8](https://github.com/amterp/rad/commit/868d6f8))
- feat: add function 'multipick' ([8c96666](https://github.com/amterp/rad/commit/8c96666))
- fix: delete pointless multipick tests ([6a0ebba](https://github.com/amterp/rad/commit/6a0ebba))

---

## [v0.6.16](https://github.com/amterp/rad/releases/tag/v0.6.16) - 2025-11-19

- Implement commands in Rad ([cef9661](https://github.com/amterp/rad/commit/cef9661))

---

## [v0.6.15](https://github.com/amterp/rad/releases/tag/v0.6.15) - 2025-11-09

- Release VSCode extension v0.1.11 ([1c6bb65](https://github.com/amterp/rad/commit/1c6bb65))
- feat: recognize hoisted functions for unknown functions check ([e535795](https://github.com/amterp/rad/commit/e535795))
- fix: allow --version --src and --src-tree on invalid scripts ([4220c18](https://github.com/amterp/rad/commit/4220c18))
- fix: avoid passing arg-less prints through printf ([7e16b1a](https://github.com/amterp/rad/commit/7e16b1a))
- fix: fix x/crypto dependency ([5bb2ea7](https://github.com/amterp/rad/commit/5bb2ea7))

---

## [v0.6.14](https://github.com/amterp/rad/releases/tag/v0.6.14) - 2025-11-04

- feat(check): add error check against hoisted functions shadowing args ([ff839e6](https://github.com/amterp/rad/commit/ff839e6))
- feat: allow non-str types in colorize function ([ba71c19](https://github.com/amterp/rad/commit/ba71c19))

---

## [v0.6.13](https://github.com/amterp/rad/releases/tag/v0.6.13) - 2025-11-03

- Release VSCode extension v0.1.10 ([e27b21c](https://github.com/amterp/rad/commit/e27b21c))
- ci: fix test-runner.rad syntax ([18c9bad](https://github.com/amterp/rad/commit/18c9bad))
- docs(guide): add complete example to getting started ([c4df48a](https://github.com/amterp/rad/commit/c4df48a))
- docs(guide): add error-handling.md ([4a7f9f5](https://github.com/amterp/rad/commit/4a7f9f5))
- docs(guide): add type-annotations.md ([20fe398](https://github.com/amterp/rad/commit/20fe398))
- docs(guide): shorten shell-commands.md header ([d704c25](https://github.com/amterp/rad/commit/d704c25))
- docs(guide): update args.md ([ab615c6](https://github.com/amterp/rad/commit/ab615c6))
- docs(guide): update basics.md ([0a38d12](https://github.com/amterp/rad/commit/0a38d12))
- docs(guide): update functions.md ([67b5fdc](https://github.com/amterp/rad/commit/67b5fdc))
- docs(guide): update getting started ([7974f4c](https://github.com/amterp/rad/commit/7974f4c))
- docs(guide): update rad-blocks.md ([5c58d83](https://github.com/amterp/rad/commit/5c58d83))
- docs(guide): update shell-commands.md ([863c3c1](https://github.com/amterp/rad/commit/863c3c1))
- docs(guide): update strings-advanced.md ([2776ef6](https://github.com/amterp/rad/commit/2776ef6))
- feat: revamp http url encoding behavior ([d91350b](https://github.com/amterp/rad/commit/d91350b))
- fix(fomatting): accept thousands_separator before alignment/padding ([c18c971](https://github.com/amterp/rad/commit/c18c971))
- fix(shell): fix shell modifier keywords ([e448f72](https://github.com/amterp/rad/commit/e448f72))
- fix: check built-in function signatures for parsing errors ([17b14fe](https://github.com/amterp/rad/commit/17b14fe))
- fix: use Print, not Printf for --src ([8579dac](https://github.com/amterp/rad/commit/8579dac))
- tests: fix broken tests from earlier commit ([0dba710](https://github.com/amterp/rad/commit/0dba710))

---

## [v0.6.12](https://github.com/amterp/rad/releases/tag/v0.6.12) - 2025-10-16

## Breaking!

This release changes shell and error handling syntax, expect breaks.

```rad
// OLD: Critical command
$!`make build`

// NEW: Same behavior (critical by default)
$`make build`

// OLD: Unsafe command
unsafe $`command_that_might_fail`

// NEW: Use catch block
$`command_that_might_fail` catch:
    pass

// OLD: Recover block
$`curl {url}`
recover:
    print_err("Failed but continuing")

// NEW: Use catch block
$`curl {url}` catch:
    print_err("Failed but continuing")

// OLD: Fail block
$`curl {url}`
fail:
    print_err("Failed, exiting")
    exit(1)

// NEW: Use catch block with exit
$`curl {url}` catch:
    print_err("Failed, exiting")
    exit(1)
```

- Release VSCode extension v0.1.9 ([46e3eed](https://github.com/amterp/rad/commit/46e3eed))
- feat: implement new error handling syntax (catch suffix) ([5c0253c](https://github.com/amterp/rad/commit/5c0253c))
- feat: implement new shell syntax + catch suffix syntax ([7ef3f46](https://github.com/amterp/rad/commit/7ef3f46))
- feat: update syntax dump tests for new shell & error handling syntax ([78a4857](https://github.com/amterp/rad/commit/78a4857))
- feat: update textmate grammar for new shell + error handling syntax ([981fdfb](https://github.com/amterp/rad/commit/981fdfb))


---

## [v0.6.11](https://github.com/amterp/rad/releases/tag/v0.6.11) - 2025-10-12

- feat: add --from-logs flag to check cmd ([468b455](https://github.com/amterp/rad/commit/468b455))
- feat: add global config loading ([a068c49](https://github.com/amterp/rad/commit/a068c49))
- feat: add invocation logging & rolling ([c5ee8e7](https://github.com/amterp/rad/commit/c5ee8e7))
- refactor: move Run up in runner.go ([2257aa7](https://github.com/amterp/rad/commit/2257aa7))

---

## [v0.6.10](https://github.com/amterp/rad/releases/tag/v0.6.10) - 2025-10-06

- Release VSCode extension v0.1.8 ([5b94a4e](https://github.com/amterp/rad/commit/5b94a4e))
- feat: add stdin support for Unix-style piping ([cd880f7](https://github.com/amterp/rad/commit/cd880f7))

---

## [v0.6.9](https://github.com/amterp/rad/releases/tag/v0.6.9) - 2025-10-05

- feat: allow scientific notation for int defaults ([8fd724a](https://github.com/amterp/rad/commit/8fd724a))
- feat: validate script immediately after parsing ([127f484](https://github.com/amterp/rad/commit/127f484))
- fix: correctly allow unknown flags to be absorbed in var args ([34ee9be](https://github.com/amterp/rad/commit/34ee9be))
- fix: resolve shell command pipe race condition ([319a956](https://github.com/amterp/rad/commit/319a956))

---

## [v0.6.8](https://github.com/amterp/rad/releases/tag/v0.6.8) - 2025-09-30

- Release VSCode extension v0.1.7 ([4b6613e](https://github.com/amterp/rad/commit/4b6613e))
- ci: implement 5 more PR benchmarks ([f8310f0](https://github.com/amterp/rad/commit/f8310f0))
- feat: allow int short count arg input style ([05b7ee2](https://github.com/amterp/rad/commit/05b7ee2))

---

## [v0.6.7](https://github.com/amterp/rad/releases/tag/v0.6.7) - 2025-09-28

- Bump version to v0.6.7 ([379d785](https://github.com/amterp/rad/commit/379d785))
- ci: add PR benchmarking ([046a3c2](https://github.com/amterp/rad/commit/046a3c2))
- ci: add PR check automation ([9efd9b4](https://github.com/amterp/rad/commit/9efd9b4))
- ci: add benchmark-scripts ([7f6349c](https://github.com/amterp/rad/commit/7f6349c))
- feat: support formatting numbers with commas separating thousands ([9e2bc45](https://github.com/amterp/rad/commit/9e2bc45))
- tests: fix local TZ-dependent time tests ([8e2dd59](https://github.com/amterp/rad/commit/8e2dd59))

---

## [v0.6.6](https://github.com/amterp/rad/releases/tag/v0.6.6) - 2025-09-27

- ci: iterate goreleaser ([f3a331a](https://github.com/amterp/rad/commit/f3a331a))

---

## [v0.6.5](https://github.com/amterp/rad/releases/tag/v0.6.5) - 2025-09-27

- ci: iterate goreleaser ([284797b](https://github.com/amterp/rad/commit/284797b))

---

## [v0.6.2](https://github.com/amterp/rad/releases/tag/v0.6.2) - 2025-09-21


---

## [v0.6.1](https://github.com/amterp/rad/releases/tag/v0.6.1) - 2025-09-21


---

## [v0.6.0](https://github.com/amterp/rad/releases/tag/v0.6.0) - 2025-09-21

* fix: returning from fn in for loop by @amterp in https://github.com/amterp/rad/pull/44
* feat(args): support var args by @amterp in https://github.com/amterp/rad/pull/45

---

## [v0.5.59](https://github.com/amterp/rad/releases/tag/v0.5.59) - 2025-09-14

* Replace pflag with Ra by @amterp in https://github.com/amterp/rad/pull/38
* feat: add function 'matches' for regex matching by @amterp in https://github.com/amterp/rad/pull/39
* fix: convert arg constraint targets to external name format by @amterp in https://github.com/amterp/rad/pull/40
* feat: add global flag '--rad-args-dump' by @amterp in https://github.com/amterp/rad/pull/41
* feat(repl): implement REPL by @amterp in https://github.com/amterp/rad/pull/43

---

# Historical Releases

Releases before automated GitHub releases. Notable changes only.

---

## v0.5 - 2025-02-12

- Replaced handwritten lexer/parser with tree sitter
- `defer` statements
- Further shell command support improvements, critical shell commands
- Emoji support
- Basic syntax highlighter
- Reworked JSON field extraction algo
- `errdefer`
- Reworked string character escaping
- Improved rad block sorting operation, added matching `sort` function
- Added more functions: `confirm`, `range`, `split` etc
- Removed Cobra
- Reworked strings (`RslString`, attributes, colors)
- http functions
- parsing functions e.g. parse_int, parse_float
- `.dot.syntax` for map key access
- Truthy/falsy logic
- Raw strings
- Multiline strings
- Arg constraints - enum, regex
- Modulo operator `%`
- `++`/`--` operators

---

## v0.4 - 2024-10-28

- `exit` function
- Allow output pass-through in `rad` blocks
- `rad` field modifiers: `truncate`, `color`
- Reworked arrays: all arrays now allow mixed types
- maps
- collection entry assignment
- `del`
- `in` and `not in`
- Json algo: allow capturing json nodes as maps
- Added list/string slicing
- Improved indexing, including negative indexing
- Added ternary expressions
- Added inline expressions for string interpolation, including formatting
- Implemented shell command invocation

---

## v0.3 - 2024-09-29

- Improved shell embedding
- Improved table-to-terminal size adjustment
- Good unit testing
- Compound assignments
- Allow mocking responses `--MOCK-RESPONSE`
- Json algo: add `*` wildcard capture
- `rad` sort statements
- Colorized headers
- Switch from `int` to `int64` representation of ints
- Add `pick` functions, including `pick_from_resource`
- Add list comprehensions
- `request` and `display` blocks

---

## v0.2 - 2024-09-09

- Added Apache License 2.0
- Arg defaults
- std functions: date functions, replace, join, upper/lower, etc
- 'Single quote' strings

---

## v0.1 - 2024-09-08

- Initial version
- Newest notable feature was `--STDIN` and output shell export commands.
