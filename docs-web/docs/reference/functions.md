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

### sleep

```rsl
sleep(seconds int)
sleep(seconds float)
sleep(duration string)  // e.g. sleep("2h45m")
```

Allows for a named arg `title=str".

See the following table for valid `duration` string formats:

| Suffix   | Description  |
|----------|--------------|
| h        | Hours        |
| m        | Minutes      |
| s        | Seconds      |
| ms       | Milliseconds |
| us or µs | Microseconds |
| ns       | Nanoseconds  |


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

### join

```rsl
join(input list, joiner string, prefix string|int|float|bool?, suffix string|int|float|bool?) -> string
```

### zip

Combines multiple lists into a list of lists, pairing elements by index.

```rsl
zip(lists... list, strict bool?, fill any?)
```

If lists are of unequal length:

- By default, truncates to the shortest.
- If `fill` is provided, extends shorter lists to the longest, using the fill value.
- If `strict=true`, raises an error if lengths differ.
  - `strict` cannot be true while `fill` is defined.

Examples:

```
zip([1, 2, 3], ["a", "b", "c"])          // [[1, "a"], [2, "b"], [3, "c"]]
zip([1, 2, 3, 4], ["a", "b"])            // [[1, "a"], [2, "b"]]
zip([1, 2, 3], ["a", "b"], strict=true)  // Error: Lists must have the same length
zip([1, 2, 3, 4], ["a", "b"], fill="-")  // [[1, "a"], [2, "b"], [3, "-"], [4, "-"]]
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

### now

Returns the current time in the machine's local timezone, accessible in various forms.

```rsl
now() -> map
```

Map values:

| Accessor         | Description                           | Type   | Example             |
|------------------|---------------------------------------|--------|---------------------|
| `.date`          | Current date YYYY-MM-DD               | string | 2019-12-13          |
| `.year`          | Current calendar year                 | int    | 2019                |
| `.month`         | Current calendar month                | int    | 12                  |
| `.day`           | Current calendar day                  | int    | 13                  |
| `.hour`          | Current clock hour (24h)              | int    | 14                  |
| `.minute`        | Current minute of the hour            | int    | 15                  |
| `.second`        | Current second of the minute          | int    | 16                  |
| `.epoch.seconds` | Seconds since 1970-01-01 00:00:00 UTC | int    | 1576246516          |
| `.epoch.millis`  | Millis since 1970-01-01 00:00:00 UTC  | int    | 1576246516123       |
| `.epoch.nanos`   | Nanos since 1970-01-01 00:00:00 UTC   | int    | 1576246516123456789 |

### type_of

Returns the type of an input variable as a string.

```rsl
type_of(variable any)
```

```rsl
type_of("hi")  // string
type_of([2])   // list
```

### str

```
str(any) -> string
```

Converts any input to a string.

## Input

### input

Get a line of text input from the user.

```rsl
input(prompt string?, default=string?, hint=string?) -> string
```

### confirm

Get a boolean confirmation from the user.

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

## Parsing

### parse_int

```rsl
parse_int(input str) -> int, err
```

### parse_float

```rsl
parse_float(input str) -> float, err
```

### parse_json

```rsl
parse_json(input string) -> any
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

### count

```
count(input string, substring string) -> int
```

Counts the number of non-overlapping instances of `substring` in `input`.

### trim

```
trim(text string, chars string = " \t\n") -> string
```

Trims the start and end of an input string.
If `chars` is left unspecified, then it will default to whitespace characters i.e. spaces, tabs, and newlines.

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

RSL offers a function for each of the 9 [HTTP method types](https://en.wikipedia.org/wiki/HTTP#Request_methods). Respectively:

- `http_get`
- `http_post`
- `http_put`
- `http_patch`
- `http_delete`
- `http_head`
- `http_options`
- `http_trace`
- `http_connect`

Their inputs and outputs are the same - the only difference between them is the HTTP method in the request.
We'll use `http_post` as an example.

```rsl
http_post(url string) -> map
http_post(url string, body=string|map, headers=map) -> map
```

Keys in the `headers` map must be strings, and values may be either strings or lists of strings.

The **output** map contains the following entries (`?` signifies it may not be present, depending on the result):

```
"success" -> bool
"duration_seconds" -> float
"status_code"? -> int
"body"? -> any
"error"? -> string
```

## Math

### abs

```rsl
abs(int) -> int
abs(float) -> float
```

### sum

```rsl
sum(list[number]) -> float
```

Sums the input list of numbers to a resulting float.

## System

### exit

```rsl
exit(code int = 0)
```
