# RSL TextMate Bundle Generator

Plan is to generate a textmate bundle from [RTS](../NAVIGATE.md#amterprts).

Not yet implemented.

## Grammar

- Global scope:
  - Shebang at top
  - Comments with `#` and `//`
  - Basic assignment: `a = 2`
  - String delimiters: """ , ', ", `
    - Has interpolation, unless delimiter prefixed with r
  - Types: string, int, float, bool
  - Bools: true, false
  - Lambda syntax e.g. a -> a * 2
  - Brackets: {}, [], ()
  - Keywords: args, if, for, in, else, json, unsafe, quiet, defer, errdefer, fail, recover, continue, break, case, del, rad, request, display
  - Comparison: ==, !=, <, <=, >, >=, not, in, not in
  - Logic: and, or
  - Math: +, -, *, /, %
  - Compound assignments: +=, -=, *=, /=, %= 
  - Functions: identifier followed by parentheses
  - Shell command: anything prefixed with $ or $!. *Can* be a string, or an identifier.
  - `args` scope:
    - Keywords: regex, enum
  - `rad`, `request`, `display` scope:
    - Keywords: fields, sort, map, color
