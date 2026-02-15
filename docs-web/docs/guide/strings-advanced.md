---
title: Strings (Advanced)
---

Strings are everywhere in scripting - from building messages to formatting output. In the [Basics](./basics.md#str) section, we covered simple string operations, but Rad offers much more powerful features for working with text.

In this section, we'll explore:

- **String interpolation** - embed expressions directly in strings
- **Formatting** - control how values are displayed (padding, fill characters, zero-padding, precision, thousands separators)
- **Multiline strings** - work with text spanning multiple lines
- **Raw strings** - disable interpolation and escaping when you need literal text
- **Escape sequences** - include special characters like newlines and tabs
- **String attributes** - add color, bold, and other terminal styling

These features make it easy to generate well-formatted output, build complex strings, and create polished CLI experiences.

## String Interpolation

Rad allows embedding expressions inside your strings that will get evaluated and replaced to produce the 'final' string.

Some examples:

```rad
name = "Alice"
print("Hi, {name}!")

print("Uppercase: {name.upper()}")

print("Conditions: {name.len() > 5 ? 'long name' : 'short name'}!")
```

<div class="result">
```
Hi, Alice!
Uppercase: ALICE
Conditions: short name!
```
</div>

String interpolation expressions can be as simple as just an identifier, or can involve function calls, math, list comprehensions, etc (though you should consider extracting complex expressions into named variables beforehand for the sake of clarity).

Note the use of single quote `'` strings inside the last line of the above example.
Using double quotes would've closed the "outer" string prematurely,
but using another delimiter allows us to avoid that without also needing to escape anything.

## Formatting

You can format expression results while doing string interpolation.
To do so, follow your expression with a colon `:` and then the relevant syntax for the formatting you want to do.
We'll demonstrate through some examples:

```rad
pi = 3.14159265359

print("Pi: {pi}_")       // no formatting
print("Pi: {pi:20}_")    // left-pad to 20 places (default)
print("Pi: {pi:<20}_")   // right-pad to 20 places
print("Pi: {pi:.3}_")    // print to 3 decimal places
print("Pi: {pi:10.2}_")  // left-pad to 10 places, including 2 decimal places
```

<div class="result">
```
Pi: 3.14159265359_
Pi:             3.141593_
Pi: 3.141593            _
Pi: 3.142_
Pi:       3.14_
```
</div>

### Fill Characters

By default, padding uses spaces. You can specify a different fill character by placing it before the alignment indicator (`>` or `<`):

```rad
name = "Alice"
num = 42

print("_{name:*>10}_")   // fill with *, right-align
print("_{name:.<10}_")   // fill with ., left-align
print("_{num:->8}_")     // fill with -, right-align
```

<div class="result">
```
_*****Alice_
_Alice....._
_------42_
```
</div>

The fill character can be any single character. The `>` and `<` after it control the alignment, just like without a fill character.

### Zero-Padding

You can use the zero-pad shorthand by placing `0` before the width. This works on any type:

```rad
name = "alice"
num = 42
neg = -7

print("_{name:08}_")     // pad string with zeros to width 8
print("_{num:05}_")      // pad number with zeros to width 5
print("_{neg:06}_")      // sign-aware for numbers: sign before zeros
print("_{num:08.2}_")    // combine with precision
```

<div class="result">
```
_000alice_
_00042_
_-00007_
_00042.00_
```
</div>

You can also combine zero-padding with thousands separators:

```rad
num = 7000
print("_{num:010,}_")
```

<div class="result">
```
_00,007,000_
```
</div>

For numbers, zero-pad is sign-aware (signs placed before zeros). For non-numbers, it simply prepends zeros (same as explicit zero fill `{x:0>5}`).

### Thousands Separator

For large numbers, you can add comma separators using `,` in your formatting:

```rad
population = 1234567
price = 1234.56

print("Population: {population:,}")
print("Price: {price:,.2}")
print("Large: {population:20,.0}")  // combine padding, comma, and precision
```

<div class="result">
```
Population: 1,234,567
Price: 1,234.56
Large:            1,234,567
```
</div>

### Number vs String Formatting

Decimal place formatting (`.X`) and thousands separators (`,`) only work on numbers. Using them on strings will cause an error:

```rad
name = "Alice"
print("{name:.2}")   // Error: cannot format string with decimal places
print("{name:,}")    // Error: cannot format string with thousands separator
```

However, padding, fill characters, and zero-padding work on all types:

```rad
print("{name:10}")     // "     Alice" (padded to 10 chars)
print("{name:*>10}")   // "*****Alice" (fill with *, any type)
print("{name:08}")     // "000Alice"   (zero-pad shorthand, any type)
```

## Multiline Strings

Sometimes you want to write strings that contain several lines. These strings may themselves also contain string delimiters e.g. `"` or `'`.
For these scenarios, Rad offers `"""` multiline string syntax. To demonstrate:

```rad
text = """
This is an
example of text
that "may contain quotes"!
It also supports interpolation:
One plus one equals {1 + 1}
"""
print(text)
```

<div class="result">
```
This is an
example of text
that "may contain quotes"!
One plus one equals 2
```
</div>

Multiline strings must follow some rules:

1. The opening `"""` must not be followed by any non-comment tokens on the same line.
2. The newline after the opening `"""` is *excluded* from the contents of the string. Contents begin on the next line.
3. The closing `"""` must not be preceded by any non-whitespace characters on that same line.
4. Whitespace preceding the closing `"""` will get removed from the front of each line in the string block.
    - In other words, you can use the indentation of the closing `"""` to control the desired indentation of your contents.
    - If the closing `"""` is preceded by more whitespace than exists on any line of string contents, that means we cannot remove that amount of whitespace from the line, leading to an error.

Below, we demonstrate the 4th point. Note that to make the "whitespaces" more visible, I've replaced them with dots, but keep in mind they *do represent spaces*:

```rad
text = """
....This is an
.....example of text
..that "may contain quotes"!
.."""  // < 2 preceding spaces. will get removed from each line in the contents.
print(text)
```

<div class="result">
```
..This is an
...example of text
that "may contain quotes"!
```
</div>

[//]: # (todo when n-""" delimiters are implemented, update this)

## Raw Strings

Rad also supports **raw strings**.
Raw strings don't perform string interpolation and do not allow any escaping (including the delimiter used to create them).
Use them when you want your contents to remain as "raw" and unprocessed as possible.

To use them, just prefix the delimiter of your choice (single/double quotes or backticks) with `r`.

```rad
text = r"Hello\n{name}"
print(text)
```

<div class="result">
```
Hello\n{name}
```
</div>

Notice the printed string is exactly as written in code - the newline character and string interpolation are left as-is.

You can use any of the string delimiters for raw strings, including multiline `"""`:

```rad
text = r"Hello\n{name}"
text = r'Hello\n{name}'
text = r`Hello\n{name}`
text = r"""
Hello\n{name}
"""
```

!!! tip "Common uses for raw strings"

    Raw strings can be quite handy for file paths, especially Windows-style ones that use backslashes:

    ```rad
    path = r"C:\Users\Documents\notes.txt"
    ```

    They can also be useful for text containing lots of braces `{}`, in order to disable string interpolation:

    ```rad
    json_str = r"{ 'my_key': { 'my_key2' : 3 } }"
    ```

## Escape Sequences

When you need special characters in your strings, you can use backslash `\` to escape them:

```rad
print("Line 1\nLine 2")      // newline
print("Col1\tCol2")          // tab
print("Path: C:\\Users")     // backslash
print("She said \"Hi!\"")    // quote (though prefer using a different delimiter, or raw strings)
```

<div class="result">
```
Line 1
Line 2
Col1	Col2
Path: C:\Users
She said "Hi!"
```
</div>

**Available escape sequences:**

- `\n` - newline
- `\t` - tab
- `\\` - literal backslash
- `\"` `\'` `` \` `` - the delimiter itself (though prefer using a different delimiter)
- `\{` - literal brace (prevents interpolation, but consider using raw strings)

## String Attributes

Strings in Rad can carry attributes like color, bold, italic, and underline. These attributes are preserved through string interpolation and concatenation:

```rad
name = "Alice".green()
print("Hello, {name}!")  // "Alice" appears green in terminal
print("Status: " + "ACTIVE".bold())  // "ACTIVE" appears bold
```

You can apply multiple attributes by chaining function calls:

```rad
title = "Important".underline().bold()
warning = "WARNING".bold().red()
print(title)
print(warning)
```

Rad provides color functions (`red`, `green`, `blue`, `yellow`, etc.), style functions (`bold`, `italic`, `underline`),
and the `hyperlink` function for creating clickable terminal links. See the [functions reference](../reference/functions.md#colors--attributes).

!!! info "When Attributes Are Preserved"

    - **Preserved**: Interpolation, concatenation, index lookup, slicing, `upper()`, `lower()`, `trim()`, `trim_prefix()`, `trim_suffix()`, `trim_left()`, `trim_right()`, and `reverse()`
    - **Not preserved**: `replace()`, `split()`

!!! tip "String Manipulation Functions"

    In addition to the syntax features covered here, Rad provides many built-in functions for working with strings.
    Use [UFCS](./functions.md#ufcs) (dot notation) for cleaner, more readable code:

    - `text.upper()`, `text.lower()` - change case
    - `text.replace(old, new)`, `text.split(sep)`, `items.join(sep)` - transform and combine strings
    - `text.trim()`, `text.trim_left()`, `text.trim_right()` - strip matching characters
    - `text.trim_prefix(prefix)`, `text.trim_suffix(suffix)` - remove a literal prefix/suffix
    - `text.starts_with(prefix)`, `text.ends_with(suffix)` - check string prefixes/suffixes
    - `text.count(substr)` - count substring occurrences
    - And many more!

    See the [Functions Reference](../reference/functions.md) for the complete list with examples.

## Summary

- We learned about **escape sequences** like `\n`, `\t`, and `\{` for including special characters in strings.
- We covered **string interpolation**, which lets us put expressions directly into strings for evaluation.
- We saw how to **format** interpolated expressions using padding, fill characters, zero-padding, precision, and more. Example: `{num:0>10,.3}`.
- We explored **multiline strings** using `"""` syntax, which support both quotes and interpolation.
- We learned about **raw strings** (prefixed with `r`) that prevent interpolation and escaping.
- We covered **string attributes** like color and bold that are preserved through interpolation and concatenation.
- Rad also provides many built-in string manipulation functions covered in the [Functions Reference](../reference/functions.md).

## Next

Next, let's look at another Rad feature which makes it uniquely suited to certain types of scripting: [Rad Blocks](./rad-blocks.md).
