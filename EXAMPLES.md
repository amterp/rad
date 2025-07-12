# Examples

---

## Positional

```sh
mycmd aaa bbb
```

| Declared Arg | Value |
|--------------|-------|
| `arg1: str`  | `aaa` |
| `arg2: str`  | `bbb` |

---

## Flags

```sh
mycmd aaa --arg2 bbb -c ddd -f
```

| Declared Arg            | Value  |
|-------------------------|--------|
| `arg1: str`             | `aaa`  |
| `arg2: str`             | `bbb`  |
| `arg3: str`, short `c`  | `ddd`  |
| `arg4: bool`, short `f` | `true` |

---

## Positional slots assigned left to right

```sh
mycmd --arg1=aaa bbb  # assigns 'bbb' to 'arg2' because 'arg1' already assigned
```

| Declared Arg | Value |
|--------------|-------|
| `arg1: str`  | `aaa` |
| `arg2: str`  | `bbb` |

```sh
mycmd aaa --arg1=bbb  # 'aaa' first gets assigned positionally to 'arg1', and then overridden by 'bbb'.
```

| Declared Arg | Value                                                          |
|--------------|----------------------------------------------------------------|
| `arg1: str`  | `bbb`                                                          |
| `arg2: str`  | `undefined`, error because both inputs go to assigning `arg1`. |

---

## Short clusters (bools)

```sh
mycmd -bcd aaa  # since flags are bools, 'aaa' gets interpreted as the first positional arg
```

| Declared Arg            | Value   |
|-------------------------|---------|
| `arg1: str`             | `aaa`   |
| `arg2: bool`, short `b` | `true`  |
| `arg3: bool`, short `c` | `true`  |
| `arg4: bool`, short `d` | `true`  |
| `arg5: bool`, short `e` | `false` |

---

## Short clusters can end with non-bool

```sh
mycmd -abc ddd  # last flag 'c' is a non-bool and so will read 'ddd'
```

| Declared Arg            | Value  |
|-------------------------|--------|
| `arg1: bool`, short `a` | `true` |
| `arg2: bool`, short `b` | `true` |
| `arg3: str`, short `c`  | `ddd`  |

---

## Incrementing int shorts

- Always treat these as normal int args, with int behavior.
- Only changes:
  - If short flag specified at the end, even once, it does not require an arg; its value will be its count.
  - If specified 2 or more times in a cluster, we will not try to parse the *next* arg as part of the int arg.

^^ TODO maybe not... maybe we just need an IsCounter toggle which makes it never interact with neighboring args

```sh
mycmd -aaa
```

| Declared Arg                        | Value |
|-------------------------------------|-------|
| `arg1: int`, short `a`, default `0` | `3`   |

```sh
mycmd -a 3
```

| Declared Arg                        | Value |
|-------------------------------------|-------|
| `arg1: int`, short `a`, default `0` | ?     |
| `arg2: str`                         | ?     |

```sh
mycmd -aa 3
```

| Declared Arg                        | Value |
|-------------------------------------|-------|
| `arg1: int`, short `a`, default `0` | ?     |
| `arg2: str`                         | ?     |

```sh
mycmd -a bbb
```

| Declared Arg                        | Value |
|-------------------------------------|-------|
| `arg1: int`, short `a`, default `0` | ?     |
| `arg2: str`                         | ?     |

---

## Negative numbers

```sh
mycmd -1 --arg2 -2 -3.4
```

| Declared Arg  | Value  |
|---------------|--------|
| `arg1: int`   | `-1`   |
| `arg2: int`   | `-2`   |
| `arg3: float` | `-3.4` |

---

## Number shorts enables 'number shorts mode'

If any arg defines an int short, we enter 'number shorts mode'.
This means that any standalone int flags are *always* interpreted as flags, not negative ints.
In this mode, you can only pass negative ints as values to arguments by using `=`.

```sh
mycmd --arg1=-2 -2 aaa -a bbb ccc
```

| Declared Arg           | Value |
|------------------------|-------|
| `arg1: int`            | `-2`  |
| `arg2: str`, short `2` | `aaa` |
| `arg3: int`            | `ccc` |
| `arg4: str`, short `a` | `bbb` |

Take the below alternative invocation.

```sh
mycmd --arg1 -2 -2 aaa -a bbb ccc
```

This is invalid. Because the `-2` is interpreted as a short flag in both instances, the initial `--arg1` doesn't
have a corresponding value, nor does the `-2` that follows.

In other words, number shorts mode is a more restrictive mode to avoid ambiguities and mistakes caused by the
conflicting use cases of having int flags and passing negative ints as args to a script.

When no int shorts are defined, number shorts mode remains disabled, meaning any "int flags" are interpreted as
negative integer values.

---

## Positional variadic

The last positional arg can be a variadic.
If nothing is read into it, it will be an empty slice.

```sh
mycmd aaa
```

| Declared Arg          | Value |
|-----------------------|-------|
| `arg1: str`           | `aaa` |
| `arg2: str`, variadic | `[ ]` |

```sh
mycmd aaa bbb
```

| Declared Arg          | Value     |
|-----------------------|-----------|
| `arg1: str`           | `aaa`     |
| `arg2: str`, variadic | `[ bbb ]` |

```sh
mycmd aaa bbb ccc
```

| Declared Arg          | Value          |
|-----------------------|----------------|
| `arg1: str`           | `aaa`          |
| `arg2: str`, variadic | `[ bbb, ccc ]` |

---

## Variadic flags

```sh
mycmd aaa --arg2  # Can be specified as a flag, even with no arguments
```

| Declared Arg          | Value |
|-----------------------|-------|
| `arg1: str`           | `aaa` |
| `arg2: str`, variadic | `[ ]` |

```sh
mycmd aaa --arg2 bbb ccc
```

| Declared Arg          | Value          |
|-----------------------|----------------|
| `arg1: str`           | `aaa`          |
| `arg2: str`, variadic | `[ bbb, ccc ]` |

```sh
mycmd --arg2 aaa bbb --arg1 ccc  # Vararg reads until the next flag 
```

| Declared Arg          | Value          |
|-----------------------|----------------|
| `arg1: str`           | `ccc`          |
| `arg2: str`, variadic | `[ aaa, bbb ]` |

---

## Multiple Variadics

Input args will be assigned to the first var arg declaration the parser encounters, or until it reaches
a new flag. A new flag could be another variadic arg, repeating the same logic again.

```sh
mycmd aaa bbb --arg2 ccc ddd -e fff
```

| Declared Arg            | Value          |
|-------------------------|----------------|
| `arg1: str`, variadic   | `[ aaa, bbb ]` |
| `arg2: str`, variadic   | `[ ccc, ddd]`  |
| `arg3: bool`, short `e` | `true`         |
| `arg4: str`             | `fff`          |

Here, we consume `ccc` and `ddd` and assign them to `arg2`, but its variadic is then ended by `-e`. We then read `fff`,
check to see our next unfilled positional arg (`arg4`), and assign to that.
