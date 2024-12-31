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

## Escaping

- `\` will escape:
    - `{` (to prevent string interpolation)
    - `\n` new line
    - `\t` tab
    - `\` i.e. itself, so you can write backslashes
    - The respective string delimiter itself, so `\"`, `\'`, or `` \` ``, epending on the string you're using. 
        - However, it's advised to instead mix string delimiters instead, especially with backticks. So respectively: `` `"` ``, `` `'` ``.

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
