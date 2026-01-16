# Version Log

Only for major & minor version releases. Contains only notable items.

## 0.8 (ongoing)

## 0.7 (2026-01-15)

### Breaking Changes
- **For-loop syntax redesign**: Replaced implicit index unpacking with explicit context access
  - Old: `for idx, item in items:` (no longer works)
  - New: `for item in items with loop:` then use `loop.idx`, `loop.src`

### New Major Features
- **Script commands**: Define subcommands with their own argument blocks
  - `command deploy:` with positional args, flags, constraints, and callbacks
  - Script-level args become shared flags across all commands
- **Revamped shell syntax**: Named shell output assignment
  - `stderr = $\`cmd\`` captures just stderr (recognizes `code`, `stdout`, `stderr`)
- **Revamped error handling**: Unified catch suffix syntax for both shell and non-shell
  - `result = parse_int(s) catch:` block syntax
  - `??` fallback operator: `port = parse_int(s) ?? 8080`
- **Global config file**: `~/.rad/config.toml` for user-wide settings
- **Invocation logging**: Opt-in feature to log script invocations
- **`rad check --from-logs`**: Check all Rad files found in invocation logs

### New Functions
- `flat_map`: Flatten and map collections in one operation
- `multipick`: Multi-select version of `pick` for choosing multiple items
- `read_stdin`, `has_stdin`: Unix pipe-compatible stdin reading

### Enhancements
- Number formatting: Thousands separators via `{n:,.2f}` → `1,234.56`
- Arg int incrementing: `-vvv` → `verbose == 3` for short cluster flags
- Rad block `filter` field modifier for filtering displayed rows
- Rad block context support in map/filter lambdas
- `pick()`: Added `prefer_exact` named arg for exact match preference
- `parse_epoch()`: Now accepts floats
- `get_path()`: Additional fields like `modified_millis`
- `min()`/`max()`: Accept varargs of numbers
- `colorize()`: Accepts non-str types
- HTTP URL encoding: Improved automatic encoding behavior
- Syntax errors: More specific diagnostic messages
- Int defaults: Support scientific notation (e.g., `1e6`)

### Tooling
- VSCode extension: LSP server (radls) now discovered via PATH instead of bundled
- Checker: Warning for undefined command callback references
- Checker: Error when hoisted functions shadow args
- Checker: Recognize hoisted functions in unknown function checks

## 0.6 (2025-09-27)

- Relational arg constraints: `requires`, `excludes`
- Custom functions: `fn name():` definitions and lambda functions
- Function typing system with runtime type checking
- Variadic arguments (`*files str` syntax)
- REPL mode for interactive scripting
- Macros: `@enable_args_block`, `@enable_global_flags`
- New functions: `matches` (regex), `pow` (power), `get_env` (environment variables)
- Enhanced HTTP functions: added `json` named parameter
- `sort()` function: parallel array sorting capability
- `colorize` function: `skip_if_single` flag for conditional coloring
- Automatic underscore-to-hyphen conversion in argument names
- Fixed shell command race conditions
- Removed "Running:" prefix from shell commands
- New commands: `rad check` (diagnostics), `rad new` (script generation), `rad stash` (state management)
- Infrastructure: replaced pflag with custom Ra argument parsing library

## 0.5 (2025-02-12)

- Replaced handwritten lexer/parser with tree sitter.
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

## 0.4 (2024-10-28)

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

## 0.3 (2024-09-29)

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

## 0.2 (2024-09-09)

- Added Apache License 2.0
- Arg defaults
- std functions: date functions, replace, join, upper/lower, etc
- 'Single quote' strings

## 0.1 (2024-09-08)

- Initial version
- Newest notable feature was `--STDIN` and output shell export commands.
