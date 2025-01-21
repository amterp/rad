---
title: Strings (Advanced)
---

In this section, we'll cover some more advanced string concepts:

- string interpolation
- formatting
- multiline strings
- raw strings

## String Interpolation

RSL allows embedding expressions inside your strings that will get evaluated and replaced to produce the 'final' string.

Some examples:

```rsl
name = "Alice"
print("Hi, {name}!")

print("Uppercase: {upper(name)}")

print("Conditions: {len(name) > 5 ? 'long name' : 'short name'}!")
```

<div class="result">
```
Hi, Alice!
Uppercase: ALICE
Conditions: short name!
```
</div>

String interpolation expressions can be as simple as just an identifier, or can involve function calls, math, list comprehensions, etc (though you should consider extracting complex expressions into named variables beforehand for the sake of clarity).

Note the use of single quote `'` strings inside the ternary example's expression. Using double quotes would've closed the "outer" string prematurely, but using another delimiter allows us to avoid that without also needing to escape anything.

## Formatting

You can format expression results while doing string interpolation.
To do so, follow your expression with `:` and then the relevant syntax for the formatting you want to do. We'll demonstrate through some examples:

```rsl
pi = 3.14159265359

print("Pi: {pi}")       // no formatting
print("Pi: {pi:20}")    // left-pad to 20 places
print("Pi: {pi:>20}")   // equivalent to above left-pad, > is redundant
print("Pi: {pi:<20}")   // right-pad to 20 places
print("Pi: {pi:.3}")    // print to 3 decimal places
print("Pi: {pi:10.2}")  // left-pad to 10 places, including 2 decimal places
```

<div class="result">
```
Pi: 3.14159265359
Pi:             3.141593
Pi:             3.141593
Pi: 3.141593            
Pi: 3.142
Pi:       3.14
```
</div>

The decimal place formatting is only relevant to expressions that result in numbers. If it results in a string, then the formatting will error.

[//]: # (todo update here when comma, formatting added)
[//]: # (todo update here if we add 0 padding)

## Multiline Strings

Sometimes you want to write strings that contain several lines. These strings may themselves also contain string delimiters e.g. `"` or `'`.
For these scenarios, RSL offers `"""` multiline string syntax. To demonstrate:

```rsl
text = """
This is an
example of text
that "may contain quotes"!
"""
print(text)
```

<div class="result">
```
This is an
example of text
that "may contain quotes"!
```
</div>

Some things to note:

1. The opening `"""` must not be followed by any non-comment tokens on the same line.
2. The newline after the opening `"""` is *excluded* from the contents of the string. Contents begin on the next line.
3. The closing `"""` must not be preceded by any non-whitespace characters on that same line.
4. Whitespace preceding the closing `"""` will get removed from the front of each line in the string block.
    - In other words, you can use the indentation of the closing `"""` to control the desired indentation of your contents.
    - If the closing `"""` is preceded by more whitespace than exists on any line of string contents, that means we cannot remove that amount of whitespace from the line, leading to an error.

Below, we demonstrate the 4th point. Note that to make the "whitespaces" more visible, I've replaced them with dots, but keep in mind they *do represent spaces*:

```rsl
text = """
....This is an
.....example of text
..that "may contain quotes"!
.."""  // < 2 preceding spaces. will get removed from each line in the contents.
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

RSL also supports **raw strings**.
Raw strings don't have any sort of processing done on them (such as string interpolation) and do not allow any escaping (including the delimiter used to create them).
Use them when you want your contents to remain as "raw" and unprocessed as possible.

To use them, just prefix the delimiter of your choice (single/double quotes or backticks) with `r`.

```rsl
text = r"Hello\n{name}"
print(text)
```

<div class="result">
```
Hello\n{name}
```
</div>

Notice the printed string is exactly as written in code - the newline character and string interpolation are left as-is.

You could use any of the three delimiters for raw strings:

```rsl
text = r"Hello\n{name}"
text = r'Hello\n{name}'
text = r`Hello\n{name}`
```

!!! tip "Raw strings for file paths"

    Raw strings can be quite handy for file paths, especially Windows-style ones that use backslashes:

    ```rsl
    path = r"C:\Users\Documents\notes.txt"
    ```

!!! info "You cannot escape the raw string's own delimiter"

    RSL raw strings behave more like their equivalent in Go than Python.
    In Python, you can escape the delimiter used to make the raw string i.e. `r"quote: \"!"`. If printed, this will
    display as `quote: \"!` i.e. the escape character is also printed. There are lots of discussions online about this
    somewhat odd behavior, which is why RSL (and Go) opted to instead keep the rules very simple and not allow escaping
    in raw strings of any kind.
    
    Instead, if you try the same thing in RSL, you will get an error because the quote following `\` will close the
    string, leaving a dangling `!"` at the end, which is invalid syntax.

## Learnings Summary

TBC

## Next

TBC
