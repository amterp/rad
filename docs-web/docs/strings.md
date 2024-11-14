---
title: Strings
---

## Escaping

- `"double quote strings"` and `'single quote strings'` have the same rules around escaping.
- `` `backtick strings` `` have slightly different rules (less escaping).

### Double & Single Quotes

- `\` will escape:
    - `{` (to prevent string interpolation)
    - `\n` new line
    - `\t` tab
    - `\` i.e. itself, so you can write backslashes
    - The respective quote char itself, so `"\""` and `'\''`
        - However, it's advised to instead mix string delimiters instead, especially with backticks. So respectively: `` `"` ``, `` `'` ``.

### Backticks

- `\` will escape:
    - `{` (to prevent string interpolation)
    - `` ` `` to allow backticks in the string
