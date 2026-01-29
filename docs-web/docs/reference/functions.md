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

- `code: string` - An [error code](./errors.md) indicating the type of error.
- `msg: string` - A description of the error.

Lastly, you may also see `number` referenced as a type -- this just means `int | float`, i.e. any numeric type.

---

## Output

### print

Prints zero or more items to stdout, separated by a delimiter.

```rad
print(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

```rad
print("Hello!")                    // -> Hello!
print()                            // -> (just newline)
print("Hello", "world")            // -> Hello world
print(1, 2, 3, sep=", ")           // -> 1, 2, 3
print("No newline", end="")        // -> No newline
```

### print_err

Behaves like [`print`](#print) but outputs to stderr instead of stdout.

```rad
print_err(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

### pprint

Pretty prints data in JSON format with indentation and colors.

```rad
pprint(_item: any?) -> void
```

```rad
item = { "name": "Alice", "age": 30 }
pprint(item)
// Output:
// {
//   "name": "Alice", 
//   "age": 30
// }
```

### debug

Behaves like [`print`](#print) but only outputs when debug mode is enabled via `--debug` flag.

```rad
debug(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

## Misc

### sleep

Pauses execution for the specified duration.

```rad
sleep(_duration: int|float|str, *, title: str?) -> void
```

Integer and float values are treated as seconds. String values support Go duration format like "2h45m", "1.5s", "500ms".
If `title` is provided, it's printed before sleeping.

**Duration string suffixes:**

| Suffix       | Description  |
|--------------|--------------|
| `h`          | Hours        |
| `m`          | Minutes      |
| `s`          | Seconds      |
| `ms`         | Milliseconds |
| `us` or `µs` | Microseconds |
| `ns`         | Nanoseconds  |

**Examples:**

```rad
sleep(2.5)              // -> Sleep for 2.5 seconds
sleep("1h30m")          // -> Sleep for 1 hour 30 minutes  
sleep("500ms")          // -> Sleep for 500 milliseconds
sleep(5, title="Waiting...") // -> Prints "Waiting..." then sleeps 5 seconds
```

### len

Returns the length of a string, list, or map.

```rad
len(input: str|list|map) -> int
```

```rad
len("hello")        // -> 5
len([1, 2, 3, 4])   // -> 4
len({"a": 1, "b": 2}) // -> 2
```

### range

Generates a list of numbers in a specified range. Useful in for loops.

```rad
range(_arg1: float|int, _arg2: float?|int?, _step: float|int = 1) -> list[float|int]
```

Single argument generates 0 to `_arg1` (exclusive). Two arguments generate `_arg1` to `_arg2` (exclusive). Step cannot
be zero. Returns float list if any argument is float, otherwise int list.

**Examples:**

```rad
range(5)            // -> [0, 1, 2, 3, 4]
range(2, 5)         // -> [2, 3, 4]
range(0.5, 3)       // -> [0.5, 1.5, 2.5]
range(10, 5, -2)    // -> [10, 8, 6]
```

### join

Joins a list into a string with separator, prefix, and suffix.

```rad
join(_list: list, *, sep: str = "", prefix: str = "", suffix: str = "") -> str
```

```rad
join([1, 2, 3], sep=", ")           // -> "1, 2, 3"
join(["a", "b"], prefix="[", suffix="]")  // -> "[ab]"
join(["x", "y", "z"], sep="-", prefix="(", suffix=")")  // -> "(x-y-z)"
```

### zip

Combines multiple lists into a list of lists, pairing elements by index.

```rad
zip(*lists: list, *, strict: bool = false, fill: any?) -> list[list]|error
```

**Parameters:**

| Parameter | Type           | Description                              |
|-----------|----------------|------------------------------------------|
| `*lists`  | `list`         | Variable number of lists to zip together |
| `strict`  | `bool = false` | If true, error on different list lengths |
| `fill`    | `any?`         | Value to fill shorter lists (optional)   |

- By default, truncates to the shortest list length
- Cannot use `strict=true` with `fill` parameter (mutually exclusive)
- Returns error if `strict=true` and lists have different lengths

**Examples:**

```rad
// Basic usage
zip([1, 2, 3], ["a", "b", "c"])           // -> [[1, "a"], [2, "b"], [3, "c"]]
zip([1, 2, 3, 4], ["a", "b"])             // -> [[1, "a"], [2, "b"]]

// With fill value for unequal lengths
zip([1, 2, 3, 4], ["a", "b"], fill="-")   // -> [[1, "a"], [2, "b"], [3, "-"], [4, "-"]]

// Strict mode (errors on length mismatch)  
zip([1, 2, 3], ["a", "b"], strict=true)   // -> Error: Lists must have the same length
```

### unique

Returns a list with duplicate values removed, preserving first occurrence order.

```rad
unique(_list: list[any]) -> list[any]
```

```rad
unique([2, 1, 2, 3, 1, 3, 4])  // -> [2, 1, 3, 4]
unique(["a", "b", "a", "c"])    // -> ["a", "b", "c"]
```

### sort

Sorts a list or string. When multiple lists are provided, performs parallel sorting where additional lists are reordered
to match the primary list's sort permutation.

```rad
sort(_primary: list|str) -> list|str
sort(_primary: list|str, *_others: list, *, reverse: bool = false) -> list[list]
```

**Parameters:**

| Parameter  | Type           | Description                                  |
|------------|----------------|----------------------------------------------|
| `_primary` | `list\|str`    | Primary data to sort (determines sort order) |
| `*_others` | `list`         | Additional lists to reorder in parallel      |
| `reverse`  | `bool = false` | Sort in descending order                     |

**Parallel Sorting Behavior:**

- The first list (`_primary`) determines the sort order
- All other lists are reordered to match the same permutation
- All lists must be the same length
- Returns a list containing all sorted lists: `[sorted_primary, sorted_other1, sorted_other2, ...]`

**Examples:**

```rad
// Basic sorting
sort([3, 4, 2, 1])                    // -> [1, 2, 3, 4]
sort([3, 4, 2, 1], reverse=true)      // -> [4, 3, 2, 1]
sort([3, 4, "2", 1, true])            // -> [true, 1, 3, 4, "2"]
sort("hello")                         // -> "ehllo"

// Parallel sorting
numbers = [2, 1, 4, 3]
letters = ["a", "b", "c", "d"] 
bools = [true, false, true, false]
sorted_nums, sorted_letters, sorted_bools = sort(numbers, letters, bools)
// -> [1, 2, 3, 4], ["b", "a", "d", "c"], [false, true, false, true]
```

### type_of

Returns the type of a value as a string.

```rad
type_of(_var: any) -> str
```

```rad
type_of("hi")    // -> "str"
type_of([2])     // -> "list" 
type_of(42)      // -> "int"
type_of(3.14)    // -> "float"
type_of({"a": 1}) // -> "map"
```

### str

Converts any value to a string representation.

```rad
str(_var: any) -> str
```

```rad
str(42)        // -> "42"
str(3.14)      // -> "3.14"
str([1, 2])    // -> "[1, 2]"
str(true)      // -> "true"
```

### int

Converts a value to an integer. Does not work on strings - use [`parse_int`](#parse_int) for string parsing.

```rad
int(_var: any) -> int|error
```

```rad
int(3.14)     // -> 3
int(true)     // -> 1
int(false)    // -> 0
int("42")     // -> Error: cannot convert string
```

### float

Converts a value to a float. Does not work on strings - use [`parse_float`](#parse_float) for string parsing.

```rad
float(_var: any) -> float|error
```

```rad
float(42)      // -> 42.0
float(true)    // -> 1.0
float(false)   // -> 0.0  
float("3.14")  // -> Error: cannot convert string
```

### is_defined

Checks if a variable with the given name exists in the current scope.

```rad
is_defined(_var: str) -> bool
```

```rad
name = "Alice"
is_defined("name")     // -> true
is_defined("age")      // -> false
```

### map

Applies a function to every element of a list or entry of a map.

```rad
map(_coll: list|map, _fn: fn(any) -> any | fn(any, any) -> any) -> list|map
```

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)`.

**Examples:**

```rad
map([1, 2, 3], fn(x) x * 2)              // -> [2, 4, 6]
map({"a": 1, "b": 2}, fn(k, v) v * 10)   // -> {"a": 10, "b": 20}
```

### filter

Applies a predicate function to filter elements of a list or map. Keeps only elements where the function returns true.

```rad
filter(_coll: list|map, _fn: fn(any) -> bool | fn(any, any) -> bool) -> list|map
```

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)`.

**Examples:**

```rad
filter([1, 2, 3, 4], fn(x) x % 2 == 0)      // -> [2, 4]
filter({"a": 1, "b": 2}, fn(k, v) v > 1)    // -> {"b": 2}
```

### flat_map

Flattens a list of lists, or applies a mapping function that returns lists and flattens the results.

```rad
flat_map(_coll: list|map, _fn: any?) -> list
```

**For lists without function:** All elements must be lists. Flattens one level.

**With function:** The function must return a list. Results are flattened.

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)` and is required.

**Examples:**

```rad
// Flatten list of lists (all elements must be lists)
[[1, 2], [3, 4]].flat_map()              // -> [1, 2, 3, 4]
[[], [1], []].flat_map()                 // -> [1]

// Only one level
[[[1]], [[2]]].flat_map()                // -> [[1], [2]]

// Map then flatten (function must return a list)
["a-b", "c-d"].flat_map(fn(e) e.split("-"))  // -> ["a", "b", "c", "d"]
[1, 2].flat_map(fn(x) [x, x * 10])           // -> [1, 10, 2, 20]
[1, 2].flat_map(fn(x) range(x))              // -> [0, 0, 1]

// Map collection - function required, must return list
{"a": [1, 2], "b": [3, 4]}.flat_map(fn(k, v) v)  // -> [1, 2, 3, 4]
{"a": 1, "b": 2}.flat_map(fn(k, v) [k, v])       // -> ["a", 1, "b", 2]

// Errors:
// [1, [2], 3].flat_map()           // Error: element 0 is not a list
// [1, 2].flat_map(fn(x) x * 2)     // Error: function must return a list
```

### load

Loads a value into a map using lazy evaluation. If key exists, returns cached value; otherwise runs loader function.

```rad
load(_map: map, _key: any, _loader: fn() -> any) -> any|error
load(_map: map, _key: any, _loader: fn() -> any, *, reload: bool = false) -> any|error
load(_map: map, _key: any, _loader: fn() -> any, *, override: any?) -> any|error
```

**Parameters:**

| Parameter  | Type           | Description                              |
|------------|----------------|------------------------------------------|
| `_map`     | `map`          | Map to store/retrieve cached values      |
| `_key`     | `any`          | Key to lookup in the map                 |
| `_loader`  | `fn() -> any`  | Function to call if key doesn't exist    |
| `reload`   | `bool = false` | Force reload even if key exists          |
| `override` | `any?`         | Use this value instead of calling loader |

If key doesn't exist, `_loader` is called and result is cached. Cannot use `reload=true` with `override` (mutually
exclusive).

**Examples:**

```rad
cache = {}
load(cache, "data", fn() expensive_calculation())    // -> Runs loader, caches result
load(cache, "data", fn() expensive_calculation())    // -> Returns cached value

// Force reload
load(cache, "data", fn() new_calculation(), reload=true)

// Override with specific value  
load(cache, "data", fn() ignored(), override="forced")
```

## Input

### input

Gets a line of text input from the user with optional prompt, default, hint, and secret mode.

```rad
input(prompt: str = "> ") -> str|error
input(prompt: str = "> ", *, hint: str = "", default: str = "", secret: bool = false) -> str|error
```

**Parameters:**

| Parameter | Type           | Description                                  |
|-----------|----------------|----------------------------------------------|
| `prompt`  | `str = "> "`   | The text prompt to display to the user       |
| `hint`    | `str = ""`     | Placeholder text shown in input field        |
| `default` | `str = ""`     | Default value if user doesn't enter anything |
| `secret`  | `bool = false` | If true, hides input (useful for passwords)  |

If `secret` is true, input is hidden (useful for passwords). The `hint` parameter has no effect when `secret` is
enabled.

**Examples:**

```rad
// Basic input
name = input("What's your name? ")                    // -> Prompts and waits for input

// With default value
color = input("Favorite color? ", default="blue")     // -> Returns "blue" if user presses enter

// With hint text
email = input("Email: ", hint="user@example.com")     // -> Shows placeholder text

// Hidden input for passwords
password = input("Password: ", secret=true)           // -> Hides typed characters
```

### confirm

Gets a boolean confirmation from the user (y/n prompt).

```rad
confirm(prompt: str = "Confirm? [y/n] > ") -> bool|error
```

```rad
if confirm():                        // -> Uses default "Confirm? [y/n] > " prompt
    print("Confirmed!")

if confirm("Delete file? [y/n] "):   // -> Custom prompt
    print("File deleted")
```

## Parsing

### parse_int

Parses a string to an integer.

```rad
parse_int(_str: str) -> int|error
```

```rad
parse_int("42")    // -> 42
parse_int("3.14")  // -> Error: invalid syntax
parse_int("abc")   // -> Error: invalid syntax
```

### parse_float

Parses a string to a float.

```rad
parse_float(_str: str) -> float|error
```

```rad
parse_float("3.14")  // -> 3.14
parse_float("42")    // -> 42.0
parse_float("abc")   // -> Error: invalid syntax
```

### parse_json

Parses a JSON string into Rad data structures.

```rad
parse_json(_str: str) -> any|error
```

```rad
parse_json('{"name": "Alice", "age": 30}')  // -> {"name": "Alice", "age": 30}
parse_json('[1, 2, 3]')                     // -> [1, 2, 3]
parse_json('invalid json')                  // -> Error: invalid JSON
```

## Text

### upper

Converts a string to uppercase. Preserves color attributes.

```rad
upper(_val: str) -> str
```

```rad
upper("hello")          // -> "HELLO"
upper("Hello World")    // -> "HELLO WORLD"
```

### lower

Converts a string to lowercase. Preserves color attributes.

```rad
lower(_val: str) -> str
```

```rad
lower("HELLO")          // -> "hello"
lower("Hello World")    // -> "hello world"
```

### replace

Replaces text using regex patterns. Does not preserve string color attributes.

```rad
replace(_original: str, _find: str, _replace: str) -> str
```

The `_find` parameter is a regex pattern. The `_replace` parameter can use regex capture groups like `$1`.

**Examples:**

```rad
replace("hello world", "world", "Rad")        // -> "hello Rad"
replace("Name: Charlie Brown", "Charlie (.*)", "Alice $1")  // -> "Name: Alice Brown"
replace("abc123def", "\\d+", "XXX")           // -> "abcXXXdef"
```

### starts_with

Checks if a string starts with a given substring.

```rad
starts_with(_val: str, _start: str) -> bool
```

```rad
starts_with("hello world", "hello")  // -> true
starts_with("hello world", "world")  // -> false
```

### ends_with

Checks if a string ends with a given substring.

```rad
ends_with(_val: str, _end: str) -> bool
```

```rad
ends_with("hello world", "world")    // -> true
ends_with("hello world", "hello")    // -> false
```

### truncate

Truncates a string to a maximum length. Returns error if length is negative.

```rad
truncate(_str: str, _len: int) -> str|error
```

```rad
truncate("hello world", 5)   // -> "hello"
truncate("short", 10)        // -> "short"
truncate("test", -1)         // -> Error: Requires a non-negative int
```

### split

Splits a string using regex pattern as delimiter. Does not preserve string color attributes.

```rad
split(_val: str, _sep: str) -> list[str]
```

The `_sep` parameter is treated as a regex pattern if valid, otherwise as literal string.

```rad
split("a,b,c", ",")            // -> ["a", "b", "c"]
split("word1 word2", "\\s+")   // -> ["word1", "word2"]
split("abc123def", "\\d+")     // -> ["abc", "def"]
```

### count

Counts the number of non-overlapping instances of substring in string.

```rad
count(_str: str, _substr: str) -> int
```

```rad
count("hello world", "l")     // -> 3
count("banana", "na")         // -> 2
count("test", "xyz")          // -> 0
```

### trim

Trims characters from both the start and end of a string.

```rad
trim(_subject: str, _to_trim: str = " \t\n") -> str
```

```rad
trim("  hello  ")            // -> "hello"
trim("***hello***", "*")     // -> "hello"
trim("abcHELLOabc", "abc")   // -> "HELLO"
```

### trim_prefix

Trims characters from only the start of a string.

```rad
trim_prefix(_subject: str, _to_trim: str = " \t\n") -> str
```

```rad
trim_prefix("  hello  ")         // -> "hello  "
trim_prefix("***hello***", "*")  // -> "hello***"
```

### trim_suffix

Trims characters from only the end of a string.

```rad
trim_suffix(_subject: str, _to_trim: str = " \t\n") -> str
```

```rad
trim_suffix("  hello  ")         // -> "  hello"
trim_suffix("***hello***", "*")  // -> "***hello"
```

### reverse

Reverses a string or list.

```rad
reverse(_val: str|list) -> str|list
```

```rad
reverse("hello")           // -> "olleh"
reverse([1, 2, 3, 4])      // -> [4, 3, 2, 1]
reverse("racecar")         // -> "racecar"
```

### hyperlink

Creates a clickable hyperlink in supporting terminals.

```rad
hyperlink(_val: any, _link: str) -> str
```

Converts text into a terminal hyperlink that can be clicked in supported terminals.

```rad
hyperlink("Visit Google", "https://google.com")    // -> Clickable "Visit Google" link
hyperlink("localhost", "http://localhost:3000")    // -> Clickable "localhost" link
hyperlink(42, "https://example.com")               // -> Clickable "42" link
```

### Colors & Attributes

Rad offers several functions to format text with colors and style attributes. All functions follow the same pattern:

```rad
color_or_style(_item: any) -> str
```

```rad
red("Hello")           // -> "Hello" (in red)
blue(42)               // -> "42" (in blue) 
bold("Important")      // -> "Important" (in bold)
```

**Available colors:**

- `plain`, `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `orange`, `pink`

**Available style attributes:**

- `bold`, `italic`, `underline`

### color_rgb

Applies RGB coloring to input text. RGB values must be in range [0, 255]. Not all terminals support this.

```rad
color_rgb(_val: any, *, red: int, green: int, blue: int) -> str|error
```

**Parameters:**

| Parameter | Type  | Description             |
|-----------|-------|-------------------------|
| `_val`    | `any` | Value to apply color to |
| `red`     | `int` | Red component (0-255)   |
| `green`   | `int` | Green component (0-255) |
| `blue`    | `int` | Blue component (0-255)  |

RGB values must be in range [0, 255]. Not all terminals support this.

```rad
color_rgb("Hello", red=255, green=0, blue=0)     // -> "Hello" (in bright red)
color_rgb(42, red=0, green=255, blue=128)        // -> "42" (in green-cyan)
color_rgb("test", red=300, green=0, blue=0)      // -> Error: RGB values must be [0, 255]
```

### colorize

Assigns consistent colors to values from a set of possible values. The same value always gets the same color within the
same set.

```rad
colorize(_val: any, _enum: any[], *, skip_if_single: bool = false) -> str
```

**Parameters:**

| Parameter        | Type           | Description                                    |
|------------------|----------------|------------------------------------------------|
| `_val`           | `any`          | Value to colorize                              |
| `_enum`          | `any[]`        | Set of possible values for consistent coloring |
| `skip_if_single` | `bool = false` | Don't colorize if only one value in set        |

Useful for automatically coloring table data or distinguishing values in lists.

**Examples:**

```rad
names = ["Alice", "Bob", "Charlie"]
colorize("Alice", names)     // -> "Alice" (in consistent color)
colorize("Bob", names)       // -> "Bob" (in different consistent color)

// In display blocks
names = ["Alice", "Bob", "Charlie", "David"]
display:
    fields names
    names:
        map fn(n) colorize(n, names)
```

## Maps

### keys

Returns all keys from a map as a list.

```rad
keys(_map: map) -> list[any]
```

```rad
keys({"a": 1, "b": 2, "c": 3})  // -> ["a", "b", "c"]
keys({})                        // -> []
```

### values

Returns all values from a map as a list.

```rad
values(_map: map) -> list[any]
```

```rad
values({"a": 1, "b": 2, "c": 3})  // -> [1, 2, 3]
values({})                         // -> []
```

## Random

### rand

Returns a random float between 0.0 (inclusive) and 1.0 (exclusive).

```rad
rand() -> float
```

```rad
rand()     // -> 0.7394832
rand()     // -> 0.2847293
```

### rand_int

Returns a random integer in a specified range.

```rad
rand_int(_arg1: int = 9223372036854775807, _arg2: int?) -> int
```

With one argument, returns random int from 0 to `_arg1` (exclusive). With two arguments, returns random int from `_arg1`
to `_arg2` (exclusive). Min must be less than max.

```rad
rand_int(10)        // -> Random int from 0-9
rand_int(5, 15)     // -> Random int from 5-14
rand_int(10, 5)     // -> Error: min (10) must be less than max (5)
```

### seed_random

Seeds the random number generator used by [`rand`](#rand) and [`rand_int`](#rand_int).

```rad
seed_random(_seed: int) -> void
```

```rad
seed_random(42)
rand()              // -> Same sequence every time with seed 42
rand_int(10)        // -> Same sequence every time with seed 42
```

### uuid_v4

Generates a random V4 UUID.

```rad
uuid_v4() -> str
```

```rad
uuid_v4()  // -> "f47ac10b-58cc-4372-a567-0e02b2c3d479"
```

### uuid_v7

Generates a random V7 UUID (time-ordered).

```rad
uuid_v7() -> str
```

```rad
uuid_v7()  // -> "01234567-89ab-7def-8123-456789abcdef"
```

### gen_fid

Generates a random [flex ID](https://github.com/amterp/flexid) (fid) - a time-ordered, URL-safe identifier.

```rad
gen_fid() -> str|error
gen_fid(*, alphabet: str?, tick_size_ms: int?, num_random_chars: int?) -> str|error
```

**Parameters:**

| Parameter          | Type                       | Description                            |
|--------------------|----------------------------|----------------------------------------|
| `alphabet`         | `str? = "[0-9][A-Z][a-z]"` | Characters to use (base-62 by default) |
| `tick_size_ms`     | `int? = 1`                 | Time precision in milliseconds         |
| `num_random_chars` | `int? = 6`                 | Number of random characters to append  |

Defaults: `alphabet` is base-62 (`[0-9][A-Z][a-z]`), `tick_size_ms` is 1ms, `num_random_chars` is 6.

```rad
gen_fid()                                    // -> "1a2b3c4d5e"
gen_fid(alphabet="0123456789")               // -> "1234567890"
gen_fid(num_random_chars=3)                  // -> "1a2b3c"
```

## Picking

### pick

Presents an interactive menu for selecting from a list of options.

```rad
pick(_options: list[str], _filter: str?|list[str]?, *, prompt: str = "Pick an option", prefer_exact: bool = false) -> str
```

Shows a fuzzy-searchable menu. Filter can be a string or list of strings to pre-filter options.

When `prefer_exact=true`, exact key matches (case-insensitive) are prioritized: if exactly one option exactly matches a filter, it's selected immediately; if multiple match exactly, only those are shown.

```rad
pick(["apple", "banana", "cherry"])                        // -> Interactive menu
pick(["red", "green", "blue"], "r")                        // -> Fuzzy-filtered to "red", "green"
pick(["grape", "g"], "g", prefer_exact=true)                 // -> Immediately picks "g" (exact match)
pick(["one", "two", "three"], prompt="Choose:")            // -> Custom prompt
```

### pick_kv

Presents an interactive menu showing keys but returns corresponding values.

```rad
pick_kv(keys: list[str], values: list[any], _filter: str?|list[str]?, *, prompt: str = "Pick an option", prefer_exact: bool = false) -> any
```

Displays keys in the menu but returns the value at the same index when selected.

When `prefer_exact=true`, exact key matches (case-insensitive) are prioritized: if exactly one key exactly matches a filter, its value is returned immediately; if multiple match exactly, only those are shown.

```rad
names = ["Alice", "Bob", "Charlie"]
ages = [25, 30, 35]
pick_kv(names, ages)                                        // -> Shows names, returns age
pick_kv(["Red", "Green"], ["#ff0000", "#00ff00"])           // -> Shows colors, returns hex
pick_kv(["grape", "g"], [1, 2], "g", prefer_exact=true)       // -> Returns 2 (exact match)
```

### pick_from_resource

Loads options from a resource file and presents an interactive menu.

```rad
pick_from_resource(path: str, _filter: str?, *, prompt: str = "Pick an option", prefer_exact: bool = true) -> any
```

Loads data from a JSON file and presents it as selectable options. Returns the selected item(s).

With `prefer_exact=true` (the default), exact key matches (case-insensitive) are prioritized: if exactly one entry has a key that exactly matches the filter, it's selected immediately; if multiple match exactly, only those are shown. Set `prefer_exact=false` to disable this and use pure fuzzy matching.

```rad
pick_from_resource("servers.json")                    // -> Menu from file
pick_from_resource("configs.json", "prod")            // -> Pre-filtered, exact match priority
pick_from_resource("data.json", prompt="Select:")     // -> Custom prompt
pick_from_resource("data.json", "x", prefer_exact=false) // -> Pure fuzzy matching
```

### multipick

Presents an interactive menu for selecting multiple options from a list.

```rad
multipick(_options: str[], *, prompt: str?, min: int = 0, max: int?) -> str[]
```

Shows an interactive multi-select menu where users can select zero or more options. =
Unlike `pick`, which returns a single selection, `multipick` returns a list of all selected items.

**Parameters:**

| Parameter  | Type      | Description                                                                   |
|------------|-----------|-------------------------------------------------------------------------------|
| `_options` | `str[]`   | List of options to display in the menu                                        |
| `prompt`   | `str?`    | Custom prompt text. If not provided, automatically generated based on min/max |
| `min`      | `int = 0` | Minimum number of selections required (default 0 allows empty selection)      |
| `max`      | `int?`    | Maximum number of selections allowed (optional, unlimited if not set)         |

The `prompt` parameter has smart defaults that adjust based on the min/max constraints.

**Example:**

```rad
fruits = ["apple", "banana", "cherry", "date"]
selected = multipick(fruits)
// selected equals e.g. [ "apple", "cherry" ]
```

## HTTP

Rad provides functions for all HTTP methods. All functions have identical signatures and return the same response
format.

### HTTP Functions

**Available methods:**

- `http_get`, `http_post`, `http_put`, `http_patch`, `http_delete`
- `http_head`, `http_options`, `http_trace`, `http_connect`

```rad
http_method(url: str, *, body: any?, json: any?, headers: map?) -> map
```

**Parameters:**

| Parameter | Type   | Description                                                               |
|-----------|--------|---------------------------------------------------------------------------|
| `url`     | `str`  | The target URL                                                            |
| `body`    | `any?` | Request body content (sent as-is)                                         |
| `json`    | `any?` | Request body content (JSON-serialized)                                    |
| `headers` | `map?` | Map of HTTP headers - optional. Values can be strings or lists of strings |

- **Body vs JSON**: The `body` parameter sends content as-is using string representation, while `json` automatically
  JSON-serializes the content and sets `Content-Type: application/json` header only if no `headers` are provided at all.
- **Mutually exclusive**: Cannot use both `body` and `json` parameters together - you must choose one or the other.

**URL Encoding:**

Rad automatically normalizes URLs to ensure proper encoding:

- **Spaces**: Encoded as `%20` everywhere (path and query parameters)
- **Special characters**: Properly percent-encoded per RFC 3986

This means you can write URLs naturally with spaces and special characters:

```rad
// URLs with spaces work naturally
http_get("https://api.example.com/search?query=hello world")
// Sent as: https://api.example.com/search?query=hello%20world

// Literal plus signs are preserved
http_get("https://api.example.com?formula=a+b")
// Sent as: https://api.example.com?formula=a%2Bb

// Parameter order is preserved
http_get("https://api.example.com?zebra=1&alpha=2")
// Sent as written (not reordered alphabetically)
```

**Response map contains:**

- `success: bool` - Whether request succeeded
- `duration_seconds: float` - Request duration
- `status_code?: int` - HTTP status code (if response received)
- `body?: any` - Response body parsed as JSON if possible (if present)
- `error?: str` - Error message (if request failed)

**Examples:**

```rad
// Simple GET request
response = http_get("https://api.example.com/users")
if response.success:
    users = response.body

// POST with JSON body (automatic serialization and Content-Type header)
data = {"name": "Alice", "email": "alice@example.com"}
response = http_post("https://api.example.com/users", json=data)

// POST with raw body content (sent as-is)
response = http_post("https://api.example.com/webhook", body="raw text data")

// With custom headers
headers = {"Authorization": "Bearer token123"}
response = http_get("https://api.example.com/data", headers=headers)

// JSON with custom headers (Content-Type automatically added)
response = http_post("https://api.example.com/users", json=data, headers={"Authorization": "Bearer token123"})

// Error handling
response = http_get("https://invalid-url")
if not response.success:
    print("Request failed:", response.error)

// Cannot use both body and json together - this will error:
// response = http_post("url", body="data", json={"key": "value"})  // -> Error
```

## Math

### abs

Returns the absolute value of a number.

```rad
abs(_num: int|float) -> int|float
```

```rad
abs(-5)      // -> 5
abs(3.14)    // -> 3.14
abs(-2.7)    // -> 2.7
```

### sum

Sums all numbers in a list.

```rad
sum(_nums: list[float]) -> float|error
```

```rad
sum([1, 2, 3, 4])        // -> 10.0
sum([1.5, 2.5, 3.0])     // -> 7.0
sum([])                  // -> 0.0
sum([1, "text", 3])      // -> Error: requires list of numbers
```

### round

Rounds a number to the specified decimal precision.

```rad
round(_num: float, _decimals: int = 0) -> int|float|error
```

**Parameters:**

| Parameter   | Type      | Description                                     |
|-------------|-----------|-------------------------------------------------|
| `_num`      | `float`   | Number to round                                 |
| `_decimals` | `int = 0` | Number of decimal places (must be non-negative) |

With precision 0, returns an integer. With precision > 0, returns a float. Precision must be non-negative.

```rad
round(3.14159)           // -> 3 (integer)
round(3.14159, 2)        // -> 3.14 (float)
round(2.7)               // -> 3 (integer)
round(3.14, -1)          // -> Error: precision must be non-negative
```

### floor

Rounds a number down to the next integer.

```rad
floor(_num: float) -> int
```

```rad
floor(1.89)    // -> 1
floor(-1.2)    // -> -2
floor(5.0)     // -> 5
```

### ceil

Rounds a number up to the next integer.

```rad
ceil(_num: float) -> int
```

```rad
ceil(1.21)     // -> 2
ceil(-1.8)     // -> -1
ceil(5.0)      // -> 5
```

### min

Returns the minimum value from a list of numbers or from variadic arguments.

```rad
min(_nums: float|float[]) -> float|error
```

Accepts either a single list of numbers or multiple number arguments.

```rad
min([1, 2, 3, 4])        // -> 1.0
min(1, 2, 3, 4)          // -> 1.0
min(5.5, 2.1, 8.9)       // -> 2.1
min(5)                   // -> 5.0
min([])                  // -> Error: cannot find minimum of empty list
min([1, "text"])         // -> Error: requires list of numbers
```

### max

Returns the maximum value from a list of numbers or from variadic arguments.

```rad
max(_nums: float|float[]) -> float|error
```

Accepts either a single list of numbers or multiple number arguments.

```rad
max([1, 2, 3, 4])        // -> 4.0
max(1, 2, 3, 4)          // -> 4.0
max(5.5, 2.1, 8.9)       // -> 8.9
max(5)                   // -> 5.0
max([])                  // -> Error: cannot find maximum of empty list
max([1, "text"])         // -> Error: requires list of numbers
```

### clamp

Constrains a value between minimum and maximum bounds.

```rad
clamp(val: float, min: float, max: float) -> float|error
```

**Parameters:**

| Parameter | Type    | Description        |
|-----------|---------|--------------------|
| `val`     | `float` | Value to constrain |
| `min`     | `float` | Minimum bound      |
| `max`     | `float` | Maximum bound      |

Returns `val` if between min and max, otherwise returns the nearest bound. Min must be ≤ max.

```rad
clamp(25, 20, 30)    // -> 25.0
clamp(10, 20, 30)    // -> 20.0
clamp(40, 20, 30)    // -> 30.0
clamp(15, 30, 20)    // -> Error: min must be <= max
```

### pow

Raises `base` to the power of `exponent`. Useful for exponentiation, square roots, and cube roots.

```rad
pow(base: float, exponent: float) -> float
```

```rad
pow(2, 3)      // -> 8
pow(4, 0.5)    // -> 2.0 (square root)
pow(8, 1/3)    // -> 2.0 (cube root)  
pow(2, -2)     // -> 0.25
pow(-2, 3)     // -> -8
```

## Hashing & Encode/Decode

### hash

Generates a hash of the input text using various algorithms.

```rad
hash(_val: str) -> str
hash(_val: str, *, algo: ["sha1", "sha256", "sha512", "md5"] = "sha1") -> str
```

**Parameters:**

| Parameter | Type                                           | Description              |
|-----------|------------------------------------------------|--------------------------|
| `_val`    | `str`                                          | Text to hash             |
| `algo`    | `["sha1", "sha256", "sha512", "md5"] = "sha1"` | Hashing algorithm to use |

The default `sha1` is **not cryptographically secure**. Use `sha256` or `sha512` for security.

```rad
hash("hello world")                    // -> "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
hash("hello world", algo="sha256")     // -> "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
hash("sensitive data", algo="sha512")  // -> Long SHA-512 hash
```

### encode_base64

Encodes text to Base64 format.

```rad
encode_base64(_content: str) -> str
encode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str
```

**Parameters:**

| Parameter  | Type           | Description                                  |
|------------|----------------|----------------------------------------------|
| `_content` | `str`          | Text to encode                               |
| `url_safe` | `bool = false` | Replace `+/` with `-_` for URL-safe encoding |
| `padding`  | `bool = true`  | Include `=` padding characters               |

Use `url_safe=true` to replace `+/` with `-_` for URL-safe encoding. Use `padding=false` to omit `=` padding.

```rad
encode_base64("Hello World")                      // -> "SGVsbG8gV29ybGQ="
encode_base64("Hello World", url_safe=true)       // -> URL-safe version
encode_base64("Hello World", padding=false)       // -> "SGVsbG8gV29ybGQ"
```

### decode_base64

Decodes Base64 text back to original string.

```rad
decode_base64(_content: str) -> str|error
decode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str|error
```

**Parameters:**

| Parameter  | Type           | Description                                     |
|------------|----------------|-------------------------------------------------|
| `_content` | `str`          | Base64 text to decode                           |
| `url_safe` | `bool = false` | Expect URL-safe encoding (`-_` instead of `+/`) |
| `padding`  | `bool = true`  | Expect padding characters (`=`)                 |

Settings must match those used for encoding.

```rad
encoded = encode_base64("Hello World")
decoded = decode_base64(encoded)           // -> "Hello World"

// URL-safe decoding
url_encoded = encode_base64("test", url_safe=true)
decoded = decode_base64(url_encoded, url_safe=true)

// Error handling
result = decode_base64("invalid base64!")
if result.error:
    print("Decode failed:", result.error)
```

### encode_base16

Encodes text to Base16 (hexadecimal) format.

```rad
encode_base16(_content: str) -> str
```

```rad
encode_base16("Hello")        // -> "48656c6c6f"
encode_base16("ABC")          // -> "414243"
```

### decode_base16

Decodes Base16 (hexadecimal) text back to original string.

```rad
decode_base16(_content: str) -> str|error
```

```rad
decode_base16("48656c6c6f")   // -> "Hello"
decode_base16("414243")       // -> "ABC"

// Error handling
result = decode_base16("invalid hex")
if result.error:
    print("Invalid hex string")
```

## System & Files

### exit

Exits the script with the given exit code.

```rad
exit(_code: int|bool = 0) -> void
```

```rad
exit()          // -> Exits with code 0
exit(1)         // -> Exits with code 1
exit(true)      // -> Exits with code 1 (bool conversion)
exit(false)     // -> Exits with code 0 (bool conversion)
```

### read_file

Reads the contents of a file.

```rad
read_file(_path: str, *, mode: ["text", "bytes"] = "text") -> map|error
```

**Parameters:**

| Parameter | Type                         | Description                     |
|-----------|------------------------------|---------------------------------|
| `_path`   | `str`                        | Path to the file to read        |
| `mode`    | `["text", "bytes"] = "text"` | Read as UTF-8 text or raw bytes |

In text mode, decodes as UTF-8 and returns a string. In bytes mode, returns a list of integers.

**Return map contains:**

- `size_bytes: int` - File size in bytes
- `content: str|list[int]` - File contents (type depends on mode)

**Examples:**

```rad
// Read text file
result = read_file("config.txt")
if result.success:
    content = result.content  // -> string
    
// Read binary file
result = read_file("image.png", mode="bytes")
if result.success:
    bytes = result.content    // -> list[int]
    
// Handle errors
result = read_file("missing.txt")
if not result.success:
    print("Error:", result.error)
```

### write_file

Writes content to a file. Creates the file if it doesn't exist.

```rad
write_file(_path: str, _content: str, *, append: bool = false) -> map|error
```

**Parameters:**

| Parameter  | Type           | Description                                       |
|------------|----------------|---------------------------------------------------|
| `_path`    | `str`          | Path where to write the file                      |
| `_content` | `str`          | Content to write                                  |
| `append`   | `bool = false` | Append to existing content instead of overwriting |

By default overwrites the file. Use `append=true` to append to existing content.

**Return map contains:**

- `bytes_written: int` - Number of bytes written
- `path: str` - Full path to the written file

**Examples:**

```rad
// Write new file
result = write_file("output.txt", "Hello world")
print("Wrote", result.bytes_written, "bytes")

// Append to existing file
write_file("log.txt", "\nNew entry", append=true)

// Error handling
result, err = write_file("/readonly/file.txt", "data")
if err:
    print("Write failed:", err.msg)
```

### read_stdin

Reads all data from stdin.

```rad
read_stdin() -> str?|error
```

```rad
read_stdin()                  // -> "piped content" (if piped)
read_stdin()                  // -> null (if not piped)
read_stdin()                  // -> Error 20026 if read fails
content = read_stdin()
lines = content.split("\n")   // Process stdin line-by-line
```

### has_stdin

Checks if stdin is piped to the script.

```rad
has_stdin() -> bool
```

```rad
has_stdin()                     // -> true (if piped)
has_stdin()                     // -> false (if not piped)
if has_stdin():
  content = read_stdin()        // Conditional read
```

### get_path

Gets information about a file or directory path.

```rad
get_path(_path: str) -> map
```

**Always returns:**

- `exists: bool` - Whether the path exists
- `full_path: str` - Absolute path

**When path exists, also returns:**

- `base_name?: str` - File/directory name
- `permissions?: str` - Permission string (e.g., "rwxr-xr-x")
- `type?: str` - Either "file" or "dir"
- `size_bytes?: int` - File size (only for files)
- `modified_millis?: int` - Modification time as epoch milliseconds
- `accessed_millis?: int` - Access time as epoch milliseconds (Unix/macOS only)

**Examples:**

```rad
info = get_path("config.txt")
if info.exists:
    print("File size:", info.size_bytes, "bytes")
    print("Type:", info.type)
else:
    print("File not found")

// Working with timestamps using parse_epoch()
info = get_path("data.json")
if info.exists:
    mtime = info.modified_millis.parse_epoch()
    print("Last modified:", mtime.date, mtime.time)
```

### find_paths

Returns a list of all paths under a directory.

```rad
find_paths(_path: str) -> list[str]|error
find_paths(_path: str, *, depth: int = -1, relative: ["target", "cwd", "absolute"] = "target") -> list[str]|error
```

**Parameters:**

| Parameter  | Type                                       | Description                            |
|------------|--------------------------------------------|----------------------------------------|
| `_path`    | `str`                                      | Directory to search                    |
| `depth`    | `int = -1`                                 | Max depth to search (-1 for unlimited) |
| `relative` | `["target", "cwd", "absolute"] = "target"` | How to format returned paths           |

- `"target"` - Relative to input path (default)
- `"cwd"` - Relative to current directory
- `"absolute"` - Full absolute paths

**Examples:**

```rad
// Find all files in directory
paths = find_paths("src/")
for path in paths:
    print(path)  // -> "file1.txt", "subdir/file2.txt", etc.

// Limit depth
paths = find_paths("src/", depth=1)  // -> Only direct children

// Get absolute paths
paths = find_paths("src/", relative="absolute")
```

### get_env

Retrieves the value of an environment variable.

```rad
get_env(_var: str) -> str
```

Returns the environment variable value, or empty string if not set.

```rad
home_dir = get_env("HOME")                    // -> "/Users/username"
api_key = get_env("API_KEY") or "default"     // -> Uses default if not set
missing = get_env("NONEXISTENT")              // -> ""
```

### delete_path

Deletes a file or directory at the specified path.

```rad
delete_path(_path: str) -> bool
```

Returns `true` if the path was successfully deleted, `false` if it didn't exist or couldn't be deleted.

```rad
delete_path("temp.txt")         // -> true (if file existed and was deleted)
delete_path("missing.txt")      // -> false (file didn't exist)
delete_path("directory/")       // -> true (if directory existed and was deleted)
```

### get_rad_home

Returns Rad's home directory.

```rad
get_rad_home() -> str
```

```rad
home = get_rad_home()  // -> "/Users/username/.rad" or $RAD_HOME
```

### get_args

Returns the raw command-line arguments passed to the script.

```rad
get_args() -> list[str]
```

Returns all arguments after the script name. Unlike parsed args, this gives you raw access to all arguments.

```rad
// If script was called: rad myscript.rad arg1 arg2 --flag
args = get_args()  // -> ["./myscript.rad", "arg1", "arg2", "--flag"]
```

### error

Creates an error object with the given message.

```rad
error(_msg: str) -> error
```

```rad
err = error("Something went wrong")
return err  // -> Script will exit with this error message
```

## Home & Stash

### `get_rad_home`

Returns the path to rad's home folder on the user's machine.

```
get_rad_home() -> string
```

**Return Values**

Defaults to `$HOME/.rad`, or `$RAD_HOME` if it's defined.

### `get_stash_dir`

Returns the full path to the script's stash directory, with the given subpath if specified.

Requires a stash ID to have been defined.

[//]: # (TODO link to stash id docs, and for below)

```
get_stash_dir(subpath: string?) -> string
```

**Return Values**

- Without subpath defined: `<rad home>/stashes/<stash id>`
- With subpath defined: `<rad home>/stashes/<stash id>/<subpath>`

### `load_state`

Loads the script's stashed state. Creates it if it doesn't already exist.

Requires a stash ID to have been defined.

```
load_state() -> map, bool
```

**Return Values**

1. `map` containing the saved state. Starts empty, before anything is saved to it.
2. `bool` representing if the state existed before the load, or if it was just created.

### save_state

Saves the script's state to persistent stash storage.

```rad
save_state(_state: map) -> error?
```

```rad
state = {"counter": 42, "last_run": now().date}
save_state(state)
print("State saved")
```

### load_stash_file

Loads a file from the script's stash directory, creating it with default content if it doesn't exist.

```rad
load_stash_file(_path: str, _default: str = "") -> map|error
```

**Return map contains:**

- `full_path: str` - Full path to the file
- `created: bool` - Whether the file was just created
- `content?: str` - File contents (if successfully loaded)

```rad
result = load_stash_file("config.txt", "default config")
if result.success:
    if result.created:
        print("Created new config file")
    content = result.content
```

### write_stash_file

Writes content to a file in the script's stash directory.

```rad
write_stash_file(_path: str, _content: str) -> error?
```

```rad
write_stash_file("log.txt", "Script executed at " + now().time)
write_stash_file("data/results.json", json_data)
print("Data saved to stash")
```

## Time

### now

Returns the current time with various accessible formats.

```rad
now(*, tz: str = "local") -> map|error
```

**Parameters:**

| Parameter | Type            | Description                               |
|-----------|-----------------|-------------------------------------------|
| `tz`      | `str = "local"` | Timezone (e.g., "UTC", "America/Chicago") |

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
| `.time`          | Current time in "hh:mm:ss" format     | string | 14:15:16            |
| `.epoch.seconds` | Seconds since 1970-01-01 00:00:00 UTC | int    | 1576246516          |
| `.epoch.millis`  | Millis since 1970-01-01 00:00:00 UTC  | int    | 1576246516123       |
| `.epoch.nanos`   | Nanos since 1970-01-01 00:00:00 UTC   | int    | 1576246516123456789 |

**Examples:**

```rad
time = now()
print("Current date:", time.date)          // -> "2024-04-05"
print("Current time:", time.time)          // -> "14:30:25"
print("Year:", time.year)                  // -> 2024

// Use epoch for timestamps
timestamp = now().epoch.seconds
print("Timestamp:", timestamp)             // -> 1712345678

// Different timezone
utc_time = now(tz="UTC")
print("UTC time:", utc_time.time)          // -> Time in UTC
```

### parse_epoch

Parses a Unix epoch timestamp into various time formats.

```rad
parse_epoch(_epoch: int|float) -> map|error
parse_epoch(_epoch: int|float, *, tz: str = "local") -> map|error
parse_epoch(_epoch: int|float, *, unit: ["auto", "seconds", "milliseconds", "microseconds", "nanoseconds"] = "auto") -> map|error
parse_epoch(_epoch: int|float, *, tz: str = "local", unit: ["auto", "seconds", "milliseconds", "microseconds", "nanoseconds"] = "auto") -> map|error
```

**Parameters:**

| Parameter | Type                                                                          | Description                                         |
|-----------|-------------------------------------------------------------------------------|-----------------------------------------------------|
| `_epoch`  | `int\|float`                                                                  | Unix epoch timestamp (float for sub-unit precision) |
| `tz`      | `str = "local"`                                                               | Timezone (e.g., "UTC", "America/Chicago")           |
| `unit`    | `["auto", "seconds", "milliseconds", "microseconds", "nanoseconds"] = "auto"` | Timestamp unit (auto-detects by default)            |

Converts an epoch timestamp to the same format as [`now()`](#now). Auto-detects units from digit count, or specify
explicitly. When using a float, the fractional part provides sub-unit precision (e.g., `1712345678.5` seconds includes
500 milliseconds).

**Examples:**

```rad
// Parse seconds epoch (auto-detected)
time = parse_epoch(1712345678)
print(time.date, time.time)  // -> "2024-04-05 22:01:18"

// Parse milliseconds with timezone
time = parse_epoch(1712345678123, tz="America/Chicago")
print(time.hour)  // -> Hour in Chicago timezone

// Explicit unit specification
time = parse_epoch(1712345678000, unit="milliseconds")

// Float epoch with sub-second precision
time = parse_epoch(1712345678.5)  // 1712345678 seconds + 500ms
print(time.epoch.millis)  // -> 1712345678500

// Float with explicit unit (sub-millisecond precision)
time = parse_epoch(1712345678123.25, unit="milliseconds")
print(time.epoch.nanos)  // -> 1712345678123250000

// Error handling
time, err = parse_epoch(1712345678, tz="Invalid/Timezone")
if err:
    print("Invalid timezone:", err.msg)
```
