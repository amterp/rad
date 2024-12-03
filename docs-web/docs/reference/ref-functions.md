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

### range

```rsl
range(end int|float) -> int|float[]
range(start int|float, end int|float) -> int|float[]
range(start int|float, end int|float, step int|float) -> int|float[]
```

```rsl
range(5)         -> [0, 1, 2, 3, 4]
range(5.5)       -> [0, 1, 2, 3, 4, 5]
range(0.5, 5)    -> [0.5, 1.5, 2.5, 3.5, 4.5]
range(10, 5, -2) -> [10, 8, 6]
```

### confirm

```rsl
confirm() -> bool
confirm(prompt string) -> bool
```

```rsl title="Example 1"
if confirm():
    print("Confirmed!")
else:
    print("Not confirmed!")
```

```title="Example 1 Output"
Confirm? [y/n] y
Confirmed!
```

```rsl title="Example 2"
if confirm("Are you sure? > "):
    print("You're sure!")
else:
    print("Unsure!")
```

```title="Example 2 Output"
Are you sure? > n
Unsure!
```

### join

```rsl
join(input any[], joiner string, prefix string|int|float|bool?, suffix string|int|float|bool?) -> string
```

### unique

```rsl
unique(input any[]) -> any[]
```

```rsl
unique([2, 1, 2, 3, 1, 3, 4])  // [2, 1, 3, 4]
```

### sort

```rsl
sort(input any[], reverse=bool?)
```

```rsl
sort([3, 4, 2, 1])                 // [1, 2, 3, 4]
sort([3, 4, 2, 1], reversed=true)  // [4, 3, 2, 1]
sort([3, 4, "2", 1, true])         // [true, 1, 3, 4, "2"]
```

## Parsing

### int

```rsl
int(input str) -> int
```

### parse_json

```rsl
parse_json(input string) -> any
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

- Preserves string color attributes.

```rsl
upper(input any) -> string
```

### lower

- Preserves string color attributes.

```rsl
lower(input any) -> string
```

### replace

- Does *not* preserve string color attributes.

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

### split

- Does *not* preserve string color attributes.

```rsl
split(input string, delimiter_regex string) -> string[]
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
pick(options string[], filter string?) -> string
```

Named args:
- `prompt`

### pick_kv

```rsl
pick_kv(keys string[], values string[], filter string?) -> string
```

Named args:
- `prompt`

### pick_from_resource

```rsl
pick_from_resource(resource_path string, filter string?) -> any...
```

## HTTP

Map outputs contain the following keys:
- `status_code`
- `body`

Failed queries (e.g. invalid url, no response) will result in an error and script exit.

### http_get

```rsl
http_get(url string, headers map?) -> map
```

### http_post

```rsl
http_post(url string, body any?, headers map?) -> map
```

