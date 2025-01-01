Dollar sign?

```rsl
greeting = "hello {name}, you live on {address}"
greeting = "hello ${name}, you live on ${address}"
```

---

Opt in?

```rsl
greeting = f"hello {name}, you live on {address}"
greeting = f"hello ${name}, you live on ${address}"
```

---

raw?

```rsl
greeting = r"hello {name}, you live on {address}"  // no interpolation

greeting = "hello {name}, you live on {address}"
```

---

blocks?
with and without interpolation
with and without escaping for e.g. \n or \t

```rsl

// regular single line strings, will interpolate, and put name on own line
text = "hi\n{name}"
text = 'hi\n{name}'
text = `hi\n{name}`

// raw single line strings, will not interpolate, will not put name on own line 
text = r"hi\n{name}"
text = r'hi\n{name}'
text = r`hi\n{name}`

// block comments, will interpolate, will put 'whoa' on own line
text = """
These are some
contents {name}\nwhoa!
"""
```

// raw block comment, will not interpolate, will not put 'whoa' on own line
text = r"""
These are some
contents {name}\nwhoa!
"""

## Blocks

- At least 3 double quotes.
- Contents may not contain an equal or more # of consecutive quotes as the multiline delimiter, unless escaped.
  - i.e. if you use 3 quotes as the delimiter, you can only have 2 or less consecutive quotes in your string. If you use 4, you can have 3 or less.
  - C#.
- Opening quotes must be followed by only whitespace or newline (or comments).
  - i.e. content begins on the next line.
- The ending delimiter must be on its own line.
  - This means a newline separating the contents and the ending delimiter. This newline does *not* get included in the contents, since it's required for language purposes.
- Any whitespace before the ending delimiter will get removed from the start of each line in the 
  - C#, Swift
- Ending a line with `\` will collapse the *next* line onto this one. So in code you can spread it onto several lines, but contents when printed will be on one.
  - Swift
