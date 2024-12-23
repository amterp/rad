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

## String Attributes

- Not all strings are just plain text. They may have attributes such as color.
- This means that RSL contains logic on how to handle attributes when strings are combined or operated on
    - e.g. concatenation, slicing, replace functions, etc
- The following operations maintain color attributes:
    - concatenation
    - index lookup
- The following *do not*, and just return a plain string:
    - slicing (to be added)
    - functions: `replace`, `split`
- Attributes do *not* impact things like equality or comparing strings.
    - A green string "Alice" and a yellow string "Alice" will be considered 'equal'.

## String Interpolation

### Formatting

- Float formatting does *not* require a `f` at the end.
    - Correct: `{myFloat:.2}`
    - Incorrect: `{myFloat:.2f}`

Examples:

```rsl
"{myString:20}"
"{myString:<20}"
"{myString:>20}"
"{myFloat:.2}"
```
