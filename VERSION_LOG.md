# Version Log

Only for major & minor version releases. Contains only notable items.

## 0.5

- `defer` statements
- Further shell command support improvements, critical shell commands
- Emoji support
- Basic syntax highlighter

## 0.4

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

## 0.3

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

## 0.2

- Added Apache License 2.0
- Arg defaults
- std functions: date functions, replace, join, upper/lower, etc
- 'Single quote' strings

## 0.1

- Initial version
- Newest notable feature was `--STDIN` and output shell export commands.
