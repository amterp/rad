---
title: Functions
---

## Output

### print

**Description**:

Prints the given input. Includes a newline after. Stringifies whatever is given to it.

```rsl
print(items ...any?)
```

**Parameters**:

- `items: ...any?`: Zero or more items to print. If several are given, they get printed separated by spaces.

**Examples**:

```rsl
print("Hello!")
```

```rsl
name = "Alice"
print("Hello", name)  // prints "Hello Alice"
```

```rsl
print()  // prints a newline
```

```rsl
numbers = [1, 20, 300]
print(numbers)  // prints "[1, 20, 300]"
```

### pprint

**Description**:

Pretty prints the given input. Mainly useful for maps so they get printed in a json-style.

```rsl
pprint(item any?)
```

**Parameters**:

- `input: any?`: Zero or one item to pretty print. If zero, just prints a newline.

**Examples**:

```rsl title="Example 1"
item = { "name": "Alice", age: 30 }
pprint(item)
```

```json title="Example 1 Output"
{
  "name": "Alice",
  "age": 30
}
```

### debug

Behaves like `print` but only prints if debug is enabled via the `--DEBUG` flag.

```rsl
debug(items ...any?)
```

## Misc

### exit

```rsl
exit(code int = 0)
```

### sleep

```rsl
sleep(seconds int)
sleep(seconds float)
sleep(duration string)
```

### len

```rsl
len(input string) -> int
len(input any[]) -> int
len(input map) -> int
```

## Time

### now_date

```rsl
now_date() -> string  // e.g. "2006-11-25"
```

### now_year

```rsl
now_year() -> int  // e.g. 2006
```

### now_month

```rsl
now_month() -> int  // e.g. 11
```

### now_day

```rsl
now_day() -> int  // e.g. 25
```

### now_hour

```rsl
now_hour() -> int  // e.g. 14
```

### now_minute

```rsl
now_minute() -> int  // e.g. 31
```

### now_second

```rsl
now_second() -> int  // e.g. 35
```

### epoch_seconds

```rsl
epoch_seconds() -> int  // e.g. 1731063226
```

### epoch_millis

```rsl
epoch_millis() -> int  // e.g. 1731063226123
```

### epoch_nanos

```rsl
epoch_nanos() -> int  // e.g. 1731063226123456789
```

## Text

### upper

```rsl
upper(input any) -> string
```

### lower

```rsl
lower(input any) -> string
```

### replace

**Parameters**:

- `input: string`
- `old: string`: Regex pattern of what text to replace.
- `new: string`: Regex pattern of what to replace matches *with*.

```rsl
replace(input string, old string, new string) -> string
```

**Examples**:

```rsl title="Example 1"
input = "Name: Charlie Brown"
replace(input, "Charlie (.*)", "Alice $1") 
```

```rsl title="Example 1 Output"
"Alice Brown" 
```

### join

```rsl
join(input any[], prefix string|int|float|bool?, suffix string|int|float|bool?) -> string
```

### starts_with

```rsl
starts_with(input string, substring string) -> bool
```

### ends_with

```rsl
ends_with(input string, substring string) -> bool
```

### truncate

```rsl
truncate(input string, length int) -> string
```

## Maps

### keys

```rsl
keys(input map) -> any[]
```

### values

```rsl
values(input map) -> any[]
```

## Random

### rand

```rsl
rand() -> float
```

### rand_int

```rsl
rand_int(max int) -> int
rand_int(min int, max int) -> int
```

### seed_random

```rsl
seed_random(seed int)
```

## Picking

### pick

```rsl
pick(options string[], filter string?)
```

### pick_kv

```rsl
pick_kv(keys string[], values string[], filter string?)
```

### pick_from_resource

```rsl
pick_from_resource(resource_path string, filter string?)
```