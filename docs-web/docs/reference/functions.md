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

Prints the given input. Includes a newline after. Stringifies whatever is given to it.

```rad
print(items ...any?, sep: string = " ", end: string = "\n")
```

**Parameters**

| Parameter | Type            | Description                                                                    |
|-----------|-----------------|--------------------------------------------------------------------------------|
| `items`   | `...any?`       | Zero or more items to print. If several are given, they are separated by `sep. |
| `sep`     | `string = " "`  | Delimiter between `items`.                                                     |
| `end`     | `string = "\n"` | Appended to the output after all `items`.                                      |

**Examples**

```rad
print("Hello!")
print()              // prints a newline
print([1, 20, 300])  // prints "[ 1, 20, 300 ]"

name = "Alice"
print("Hello", name) // prints "Hello Alice"
```

### print_err

Behaves like [`print`](#print) but always goes to stderr instead of stdout.

```rad
print_err(items ...any?, sep: string = " ", end: string = "\n")
```

### pprint

**Description**:

Pretty prints the given input. Mainly useful for maps so they get printed in a json-style.

```rad
pprint(item any?)
```

**Parameters**:

- `input: any?`: Zero or one item to pretty print. If zero, just prints a newline.

**Examples**:

```rad title="Example 1"
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

Behaves like [`print`](#print) but only prints if debug is enabled via the `--DEBUG` flag.

```rad
debug(items ...any?, sep: string = " ", end: string = "\n")
```

## Misc

### sleep

```rad
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

```rad
len(input string) -> int
len(input any[]) -> int
len(input map) -> int
```

### range

Generates a list of numbers in a specified range. Useful in for loops.

```rad
range(end number) -> [number]
range(start number, end number, step: number = 1) -> [number]
```

```rad
range(5)         -> [0, 1, 2, 3, 4]
range(5.5)       -> [0, 1, 2, 3, 4, 5]
range(0.5, 5)    -> [0.5, 1.5, 2.5, 3.5, 4.5]
range(10, 5, -2) -> [10, 8, 6]
```

### join

```rad
join(input list, joiner string, prefix string|int|float|bool?, suffix string|int|float|bool?) -> string
```

### zip

Combines multiple lists into a list of lists, pairing elements by index.

```rad
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

```rad
unique(input any[]) -> any[]
```

```rad
unique([2, 1, 2, 3, 1, 3, 4])  // [2, 1, 3, 4]
```

### sort

```rad
sort(input any[], reverse=bool?)
```

```rad
sort([3, 4, 2, 1])                 // [1, 2, 3, 4]
sort([3, 4, 2, 1], reversed=true)  // [4, 3, 2, 1]
sort([3, 4, "2", 1, true])         // [true, 1, 3, 4, "2"]
```

### type_of

Returns the type of an input variable as a string.

```rad
type_of(variable any)
```

```rad
type_of("hi")  // string
type_of([2])   // list
```

### str

Converts any input to a string.

```
str(input: any) -> string
```

### int

Try to convert an input to an int.

Does not work on strings. If you want to parse a string to an int, use [`parse_int`](#parse_int).

```
int(input: any) -> int
```

### float

Try to convert an input to a float.

Does not work on strings. If you want to parse a string to a float, use [`parse_float`](#parse_float).

```
str(input: any) -> string
```

### is_defined

Checks if a variable exists.

```
is_defined(var: string) -> bool
```

### map

Applies a given lambda to every element of a list or entry of a map.

```
map(list, fn(v) -> any) -> list[any]
map(map, fn(k, v) -> any) -> list[any]
```

### filter

Applies a given lambda predicate to every element of a list or entry of a map. Keeps only elements that return true.

```
filter(list, fn(v) -> bool) -> list
filter(map, fn(k, v) -> bool) -> map
```

### load

Loads a value into a map. Returns the mapped value.

```
load(map, key, loader: fn() -> any, reload: bool?, override: any?) -> any
```

- Examples:
    - `load(m, k, loader)`
        - If `m` does not contain `k`, `loader` is run to calculate a value. This value is put into the map under `k`
          and returned.
        - If `m` contains `k`, `loader` is ignored, and the existing value is returned.
    - `load(m, k, loader, reload=true)`
        - Regardless of if `m` already contains `k`, `loader` is invoked and its value is put into the map for `k` and
          returned.
    - `load(m, k, loader, override=myvalue)`
        - Regardless of if `m` already contains `k`, if `myvalue` is a truthy value, then it is put into `m` under `k`
          and returned and `loader` is ignored.

`reload` cannot be true with `override` is truthy, that will return an error.

[//]: # (TODO Update that 'truthy' doc when we add nulls and its only null that causes that)

## Input

### input

Get a line of text input from the user.

```rad
input(prompt string?, default=string?, hint=string?, secret=bool?) -> string
```

**Parameters**

| Parameter | Type     | Description                                                                   |
|-----------|----------|-------------------------------------------------------------------------------|
| `prompt`  | `string` | The text prompt to display to the user.                                       |
| `default` | `string` | Default value if the user doesn't enter anything.                             |
| `hint`    | `string` | Placeholder text shown in the input field. Has no impact if `secret` enabled. |
| `secret`  | `bool`   | If true, hides the input (useful for passwords).                              |

**Return Values**

Returns the user's input as a string. If the user doesn't enter anything, returns the default value.

**Examples**

```rad
// Basic input
name = input("What's your name? ")

// With default value
color = input("Favorite color? ", default="blue")

// With hint
email = input("Email address: ", hint="example@domain.com")

// Password input
password = input("Enter password: ", secret=true)
```

### confirm

Get a boolean confirmation from the user.

```rad
confirm() -> bool
confirm(prompt string) -> bool
```

```rad title="Example 1"
if confirm():
    print("Confirmed!")
else:
    print("Not confirmed!")
```

```title="Example 1 Output"
Confirm? [y/n] y
Confirmed!
```

```rad title="Example 2"
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

```rad
parse_int(input str) -> int, err
```

### parse_float

```rad
parse_float(input str) -> float, err
```

### parse_json

```rad
parse_json(input string) -> any
```

## Text

### upper

- Preserves string color attributes.

```rad
upper(input any) -> string
```

### lower

- Preserves string color attributes.

```rad
lower(input any) -> string
```

### replace

- Does *not* preserve string color attributes.

**Parameters**:

- `input: string`
- `old: string`: Regex pattern of what text to replace.
- `new: string`: Regex pattern of what to replace matches *with*.

```rad
replace(input string, old string, new string) -> string
```

**Examples**:

```rad title="Example 1"
input = "Name: Charlie Brown"
replace(input, "Charlie (.*)", "Alice $1") 
```

```rad title="Example 1 Output"
"Alice Brown" 
```

### starts_with

```rad
starts_with(input string, substring string) -> bool
```

### ends_with

```rad
ends_with(input string, substring string) -> bool
```

### truncate

```rad
truncate(input string, length int) -> string
```

### split

- Does *not* preserve string color attributes.

```rad
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

### trim_prefix

```
trim_prefix(text string, chars string = " \t\n") -> string
```

Trims the start of an input string.
If `chars` is left unspecified, then it will default to whitespace characters i.e. spaces, tabs and newlines.

### trim_suffix

```
trim_suffix(text string, chars string = " \t\n") -> string
```

Trims the end of an input string.
If `chars` is left unspecified, then it will default to whitespace characters i.e. spaces, tabs and newlines.

### Colors & Attributes

Rad offers several functions to format text, including colors and modifiers like bold, italics, etc.
As an example, let's look at the `red` function:

```
red(string) -> string  // output string will be red
```

Complete list:

- `plain`
- `black`
- `red`
- `green`
- `yellow`
- `blue`
- `magenta`
- `cyan`
- `white`
- `orange`
- `pink`
- `bold`
- `italic`
- `underline`

### color_rgb

Apply RGB coloring to some input text. Not all terminals support this.

```
color_rgb(input: any, red: int, green: int, blue: int) -> string
```

### colorize

Given some universe of possible, enumerable values, and a value from that list of values,
color the given value with some RGB color assigned to it.
The same value will always be given the same color, given the same list of possible values.

Can use to quickly assign best-effort 'unique' colors to values in an enumerable set. Nice for coloring tables, etc.

```
colorize(value: any, possibleValues: list[any]) -> string
```

Example demonstrating use in a `display` block:

```
names = ["Alice", "Bob", "Charlie", "David"]
display:
    fields names
    names:
        map fn(n) n.colorize(names)
```

## Maps

### keys

Returns all keys from an input map as a list.

```rad
keys(input: map) -> any[]
```

### values

Returns all values from an input map as a list.

```rad
values(input: map) -> any[]
```

### get_default

Gets the value for the given key in the given map, if the key is in the map. Otherwise, returns the supplied default.

```rad
get_default(input: map, key: any, default: any) -> any
```

## Random

### rand

```rad
rand() -> float
```

### rand_int

```rad
rand_int(max int) -> int
rand_int(min int, max int) -> int
```

### seed_random

Seed the random number generator used by [rand](#rand) and [rand_int](#rand_int).

```rad
seed_random(seed: int)
```

### uuid_v4

Generate a random V4 UUID.

```rad
uuid_v4() -> string
```

### uuid_v7

Generate a random V7 UUID.

```rad
uuid_v7() -> string
```

### gen_fid

Generate a random [flex ID](https://github.com/amterp/flexid) (fid).

```rad
gen_fid(alphabet: string, tick_size_ms: int = 100, num_random_chars: int = 5) -> string
```

`alphabet` defaults to base-62 (`[0-9] [A-Z] [a-z]`).

## Picking

### pick

```rad
pick(options string[], filter string?) -> string
```

Named args:

- `prompt`

### pick_kv

```rad
pick_kv(keys string[], values string[], filter string?) -> string
```

Named args:

- `prompt`

### pick_from_resource

```rad
pick_from_resource(resource_path string, filter string?) -> any...
```

## HTTP

Rad offers a function for each of the 9 [HTTP method types](https://en.wikipedia.org/wiki/HTTP#Request_methods).
Respectively:

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

```rad
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

```rad
abs(int) -> int
abs(float) -> float
```

### sum

```rad
sum(list[number]) -> float
```

Sums the input list of numbers to a resulting float.

### round

```rad
round(number float|int, precision int = 0) -> float
```

Rounds the input number to the specified precision.
If `precision` is unspecified it defaults to 0 and rounds to the nearest integer.

### floor

```rad
floor(num float|int) -> float
```

Rounds the given input down to the next integer.

```
floor(1.89) -> 1
```

### ceil

```rad
ceil(num float|int) -> float
```

Rounds the given input up to the next integer.

```
ceil(1.21) -> 2
```

### min

```rad
min(nums list[float|num]) -> float
```

Returns the minimum number in the provided list.

```
min([1, 2, 3, 4]) -> 1
```

### max

```rad
max(nums list[float|num]) -> float
```

Returns the maximum number in the provided list.

```
max([1, 2, 3, 4]) -> 4
```

### clamp

```rad
clamp(val, min, max float|int) -> float
```

Clamps the given number `val` between `min` and `max`.

Specifically, `clamp` returns:

- `val` if it is between the provided min and max.
- `min` if `val` is lower than `min`.
- `max` if `val` is greater `max`.

`min` must be <= `max`, else the function will error.

```
clamp(25, 20, 30) -> 25
clamp(10, 20, 30) -> 20
clamp(40, 20, 30) -> 30
```

## Hashing & Encode/Decode

### hash

Hash some input text. Can choose between hashing algorithms.

```
hash(content: string, algo: string = "sha1") -> string
```

Supported algos: `sha1` (default), `sha256`, `sha512`, `md5`.

**Note**: The default `sha1` is **not cryptographically secure**.
If you need security, specify a secure algorithm such as `sha512`.

### encode_base64

Base64 encode some text.

```
encode_base64(content: string, url_safe: bool = false, padding: bool = true) -> string
```

- Enable `url_safe` to replaces url-unsafe characters in standard base64 encoding with url-safe ones.
- Disable `padding` to leave out `=` padding from the base64 encoding.

### decode_base64

Base64 decode some text.

```
decode_base64(content: string, url_safe: bool = false, padding: bool = true) -> string
```

- `url_safe` and `padding` settings should match what was used when *encoding* to ensure correct decoding.

### encode_base16

Base16 encode some text. Also known as "hex encoding".

```
encode_base16(content: string) -> string
```

### decode_base16

Base16 decode some text. Also known as "hex decoding".

```
decode_base16(content: string) -> string
```

## System & Files

### exit

Exits the script with the given exit code

```rad
exit(code int = 0)
```

### read_file

Reads the contents of a file with the given path.

```
read_file(path: string) -> map
read_file(path: string, mode: string = "text") -> map
```

- `mode` is a named arg which is `"text"` by default.
    - Decodes the contents as UTF-8 and makes it available as string.
- Other valid value is `"bytes"`.
    - Reads the bytes and makes them available as a list of ints.

The returned `map` contains two keys:

- `size_bytes -> int`
- `content -> string | list[int]` (depending on `mode`)

[//]: # (todo should the key be 'contents'? I find myself wanting to write plural frequently...)

### write_file

Writes a string to the given file path. Creates the file if it does not exist.

```
write_file(path: string, content: string, append: bool = false) -> map, map?
```

`append` is a named arg controlling whether we should append to the file instead of overriding existing data.

The first returned `map` contains two keys:

- `bytes_written -> int`
- `path -> string`

The second map is an error map which can be assigned. If it's not assigned and the function fails, it will instead error
exit.

### get_path

Gets information about a file or directory at the specified path.

```
get_path(path: string) -> map
```

The map will always contain these entries:

- `full_path -> string`
- `exists -> bool`

Only if the path exists, will it also contain the following entries:

- `base_name -> string`
- `permissions -> string`
- `type -> string` (`dir` or `file`)
- `size_bytes -> int` (Entry only defined if it's a file)

### find_paths

Returns a list of paths under the given target directory.

```
find_paths(target: string, depth: int = -1, relative: string = "target") -> list[string]
```

- `depth` defaults to `-1`, indicating no depth limit. Set a positive number to limit how deep the included paths should
  be.
- `relative` defaults to `"target"` and defines to where the resulting paths should be relative.
    - `"target"` (*default*): relative to the input target path
    - `"cwd"`: relative to the user's current working directory
    - `"absolute"`: return absolute paths

### get_env

Retrieves the value of an environment variable.

```
get_env(name: string) -> string
```

**Parameters**

| Parameter | Type     | Description                           |
|-----------|----------|---------------------------------------|
| `name`    | `string` | The name of the environment variable. |

**Return Values**

Returns the value of the environment variable as a string. If the environment variable doesn't exist, returns an empty
string.

**Examples**

```rad
// Get an environment variable
home_dir = get_env("HOME")

// Use with 'or' operator to provide a default value
api_key = get_env("API_KEY") or "default_key"
```

### get_rad_home

Returns rad's home directory.

```
get_rad_home() -> string
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

### `save_state`

Saves the given state to the stash's state file.

Requires a stash ID to have been defined.

```
save_state(state: map)
```

**Parameters**

| Parameter | Type  | Description               |
|-----------|-------|---------------------------|
| `state`   | `map` | The state object to save. |

### `load_stash_file`

Loads the contents of a file under the script's stash.
If the file doesn't exist, it gets created with the given default contents.

Requires a stash ID to have been defined.

```
load_stash_file(subpath: string, default_contents: string) -> map, bool
```

**Parameters**

| Parameter          | Type     | Description                                       |
|--------------------|----------|---------------------------------------------------|
| `subpath`          | `string` | The subpath to the file under the script's stash. |
| `default_contents` | `string` | Default contents if the file doesn't exist.       |

**Return Values**

1. `map` containing the following keys:
    - `path: string` - full path to the script
    - `content: string` - loaded contents
2. `bool` indicating if the loaded file already existed or had to be created. `true` if existed.

### `write_stash_file`

Write text to a file under the script's stash.

Requires a stash ID to have been defined.

```
write_stash_file(subpath: string, contents: string) -> string, error?!
```

**Parameters**

| Parameter  | Type     | Description                                        |
|------------|----------|----------------------------------------------------|
| `subpath`  | `string` | The subpath for the file under the script's stash. |
| `contents` | `string` | Contents to write.                                 |

**Return Values**

1. `string` for the full path to the written file.
2. `error` map containing errors if the write was unsuccessful; otherwise `null`.

## Time

### now

Returns the current time in the machine's local timezone, accessible in various forms.

```
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
| `.time`          | Current time in "hh:mm:ss" format     | string | 14:15:16            |
| `.epoch.seconds` | Seconds since 1970-01-01 00:00:00 UTC | int    | 1576246516          |
| `.epoch.millis`  | Millis since 1970-01-01 00:00:00 UTC  | int    | 1576246516123       |
| `.epoch.nanos`   | Nanos since 1970-01-01 00:00:00 UTC   | int    | 1576246516123456789 |

### parse_epoch

Given a Unix epoch timestamp, parse it into various other ready-to-use formats in the form of a map.

```
parse_epoch(epoch: int, unit: string = "auto", tz: string = "default") -> map?, error?!
```

**Parameters**

| Parameter | Type                 | Description                                                                                                                                                    |
|-----------|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `epoch`   | `int`                | The Unix epoch timestamp to parse.                                                                                                                             |
| `unit`    | `string = "auto"`    | The unit of the epoch e.g. seconds, milliseconds, microseconds, or nanoseconds.<br/>Default `auto` will try to derive it from the number of digits in `epoch`. |
| `tz`      | `string = "default"` | The time zone to use for local-time-formatted fields e.g. clock time, date, etc.<br/>Defaults to the system default.                                           |

**Return Values**

1. The returned map contains the same fields as [`now`](#now). It is `null` if there was an error parsing.
2. A nullable error map.

**Examples**

```rad
// Parse seconds epoch
time, err = parse_epoch(1712345678)  // err will be null

// Millis with TZ
time, err = parse_epoch(1712345678123, tz="America/Chicago")  // err will be null

// Invalid time zone
time, err = parse_epoch(1712345678, tz="not real time zone")  // time will be null, but err defined
```
