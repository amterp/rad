---
title: Strings
---

RSL has three delimiters for strings:

```rsl
"double quotes"
'single quotes'
`backticks`
```

All three of these behave the same way. RSL offers three so you have alternatives to pick between depending on the contents of your string.
For example, if you have a string which itself contains lots of single quotes, you may choose to use the double quotes delimiter.
Or, if your string has both single *and* double quotes, you can use backticks to delimit your string. Specific example:

```rsl
`Single quotes: 'Hi!', double quotes: "Hi!"`
```

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

## Escaping

- `\` will escape:
  - `{` (to prevent string interpolation)
  - `\n` new line
  - `\t` tab
  - `\` i.e. itself, so you can write backslashes
  - The respective string delimiter itself, so `\"`, `\'`, or `` \` ``, depending on the delimiter you're using.
    - However, it's advised you use delimiters that don't clash with the contents of your string, if possible.

## Raw Strings

Raw strings can be used when you want RSL to treat the string as it's written, rather than performing escaping, interpolation, etc.

Raw strings are created by prefixing an `r` to the opening delimiter of your string. For example:

```rsl
name = "alice"
text = r"Regards,\n{name}"
print(text)
```

<div class="result">
```
Regards,\n{name}
```
</div>

Notice RSL did not render the `\n` as a newline as it would in a regular string,
and that `{name}` is also left as-is i.e. no interpolation was performed.

Unlike Python, **you cannot escape *anything* in raw strings**, including the delimiter. For example:

```rsl
r"\""
```

is illegal because the backslash does *not* escape the following `"`, and so that quote actually ends the raw string.
Then we're left with a third and dangling `"` at the end, causing a syntax error.
