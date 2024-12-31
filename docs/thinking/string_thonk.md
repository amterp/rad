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

```
