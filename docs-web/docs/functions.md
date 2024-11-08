---
title: Functions
---

## General Overview

## Function Reference

### Output

#### print

**Description**:

Prints the given input. Includes a newline after. Stringifies whatever is given to it.

```py
print(items ...any?)
```

**Parameters**:

- `items: ...any?`: Zero or more items to print. If several are given, they get printed separated by spaces.

**Examples**:

```py
print("Hello!")
```

```py
name = "Alice"
print("Hello", name)  // prints "Hello Alice"
```

```py
print()  // prints a newline
```

```py
numbers = [1, 20, 300]
print(numbers)  // prints "[1, 20, 300]"
```

#### pprint

**Description**:

Pretty prints the given input. Mainly useful for maps so they get printed in a json-style.

```py
pprint(item any?)
```

**Parameters**:

- `input: any?`: Zero or one item to pretty print. If zero, just prints a newline.

**Examples**:

```py title="Example 1"
item = { "name": "Alice", age: 30 }
pprint(item)
```

```json title="Example 1 Output"
{
  "name": "Alice",
  "age": 30
}
```

#### debug

Behaves like `print` but only prints if debug is enabled via the `--DEBUG` flag.

```py
debug(items ...any?)
```

### Misc

#### exit

```py
exit(code int = 0)
```

#### sleep

```py
sleep(seconds int)
sleep(seconds float)
sleep(duration string)
```

#### len

```py
len(input string) -> int
len(input any[]) -> int
len(input map) -> int
```

### Time

#### now_date

```py
now_date() -> string  // e.g. "2006-11-25"
```

#### now_year

```py
now_year() -> int  // e.g. 2006
```

#### now_month

```py
now_month() -> int  // e.g. 11
```

#### now_day

```py
now_day() -> int  // e.g. 25
```

#### now_hour

```py
now_hour() -> int  // e.g. 14
```

#### now_minute

```py
now_minute() -> int  // e.g. 31
```

#### now_second

```py
now_second() -> int  // e.g. 35
```

#### epoch_seconds

```py
epoch_seconds() -> int  // e.g. 1731063226
```

#### epoch_millis

```py
epoch_millis() -> int  // e.g. 1731063226123
```

#### epoch_nanos

```py
epoch_nanos() -> int  // e.g. 1731063226123456789
```

### Text

#### upper

```py
upper(input any) -> string
```

#### lower

```py
lower(input any) -> string
```

#### replace

**Parameters**:

- `input: string`
- `old: string`: Regex pattern of what text to replace.
- `new: string`: Regex pattern of what to replace matches *with*.

```py
replace(input string, old string, new string) -> string
```

**Examples**:

```py title="Example 1"
input = "Name: Charlie Brown"
replace(input, "Charlie (.*)", "Alice $1") 
```

```py title="Example 1 Output"
"Alice Brown" 
```

#### join

```py
join(input any[], prefix string|int|float|bool?, suffix string|int|float|bool?) -> string
```

#### starts_with

```py
starts_with(input string, substring string) -> bool
```

#### ends_with

```py
ends_with(input string, substring string) -> bool
```

#### truncate

```py
truncate(input string, length int) -> string
```

### Maps

#### keys

```py
keys(input map) -> any[]
```

#### values

```py
values(input map) -> any[]
```

### Random

#### rand

```py
rand() -> float
```

#### rand_int

```py
rand_int(max int) -> int
rand_int(min int, max int) -> int
```

#### seed_random

```py
seed_random(seed int)
```

### Picking

#### pick

```py
pick(options string[], filter string?)
```

#### pick_kv

```py
pick_kv(keys string[], values string[], filter string?)
```

#### pick_from_resource

```py
pick_from_resource(resource_path string, filter string?)
```