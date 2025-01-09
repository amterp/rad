# UFCS

https://www.reddit.com/r/ProgrammingLanguages/comments/1htwxi6

## Pipes With Placeholder

Assume result gets passed into first parameter by default.

```
my_list | unique | len | print
```

Otherwise, use a placeholder.

Underscore example below. Instead of `qux(z, bar(foo(x, y)))`:

```
x | foo(_, y) | bar | qux(z, _)
```

Or `$$`:

```
x | foo($$, y) | bar | qux(z, $$)
```

```
if my_list | len > 10:
    print("whoa long list")
```

^ idk, this doesn't seem all that appealing. a simple UFCS would look better:

```
if my_list.len():
    print("whoa long list")
```

Pipes and UFCS aren't mutually exclusive.

## Standard UFCS

```
x.foo(y).bar() // can't do this one because we want qux(z, $$)!

// or, could do this?

x.foo(y).bar().qux(z, $$)

// I think this looks potentially confusing to new users tho...
```

```
my_list.unique().len().print()

// alternatively, don't require () for funcs with no additional arguments

my_list.unique.len.print

// but idk how you differentiate that from map dot.syntax, so this is probably a no-go.
```
