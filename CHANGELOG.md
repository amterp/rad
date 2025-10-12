# Version Log

Only for major & minor version releases. Contains only notable items.

## 0.7 (ongoing)

- Added support for thousands separators in number formatting e.g. `"{n:,.2f}"` -> `1,234.56`
- Added support for arg int incrementing via short clusters e.g. `verbose v int` allows `-vvv` -> `verbose == 3`
- Added functions `read_stdin` and `has_stdin` to allow Unix pipe-compatible Rad scripts
- Added global config file (default `~/.rad/config.toml`)
- Added opt-in invocation logging feature (leveraging global config file)
- Added `rad check --from-logs` flag for checking all Rad files found in invocation logs

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
