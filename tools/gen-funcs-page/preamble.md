---
title: Functions
---

This page aims to concisely document *all* in-built Rad functions.

## How to Read This Document

### Function Signatures

You'll see notation like this for function signatures (below are not real functions in Rad; just examples):

```
greet(name: str, times: int = 10) -> string
```

This means the function `greet` takes one required string argument `name`, and an optional int argument `times` which
defaults to 10 if not specified. It returns a string.

```
greet_many(names: list[string] | ...string) -> none
```

This means that `greet_many` can be called in two ways: either with a single argument that is a list of strings, or `|`
a variable number of string arguments.
In both cases, the function returns nothing.

```
do_something(input: any, log: string?) -> any, error?!
```

This means the function `do_something` takes a required argument `input` which can be of *any* type.
It also has an optional argument `log` which will default to `null` if left unspecified.

The values it returns depends on how the function is called. If it's being assigned to two variables e.g.

```
foo, bar = do_something(myvar)
```

then it will return some `any` value for `foo`, and it returns a nullable `error` for `bar`.

The exclamation point `!` signifies that, if the call is only assigned to one variable e.g.

```
foo = do_something(myvar)
```

and the function *fails* i.e. *would* return a non-`null` `error` value, then it will instead panic and exit the script
with said error.

### `error`

`error` may be referenced as a return type for some functions. `error` is really a `map` with the following keys:

- `code: string` - An error code (e.g. `RAD20003`). Use `rad docs <code>` to learn more.
- `msg: string` - A description of the error.

Lastly, you may also see `number` referenced as a type -- this just means `int | float`, i.e. any numeric type.

---
