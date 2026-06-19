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

## Crypto

### decode_base16

Decodes Base16 (hexadecimal) text back to original string.

```rad
decode_base16(_content: str) -> error|str
```

```rad
decode_base16("48656c6c6f")   // -> "Hello"
decode_base16("414243")       // -> "ABC"

// Error handling
result = decode_base16("invalid hex")
if result.error:
    print("Invalid hex string")
```

### decode_base64

Decodes Base64 text back to original string.

```rad
decode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> error|str
```

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

**Parameters:**

| Parameter  | Type           | Description                                     |
| ---------- | -------------- | ----------------------------------------------- |
| `_content` | `str`          | Base64 text to decode                           |
| `url_safe` | `bool = false` | Expect URL-safe encoding (`-_` instead of `+/`) |
| `padding`  | `bool = true`  | Expect padding characters (`=`)                 |

Settings must match those used for encoding.

### encode_base16

Encodes text to Base16 (hexadecimal) format.

```rad
encode_base16(_content: str) -> str
```

```rad
encode_base16("Hello")        // -> "48656c6c6f"
encode_base16("ABC")          // -> "414243"
```

### encode_base64

Encodes text to Base64 format.

```rad
encode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str
```

```rad
encode_base64("Hello World")                      // -> "SGVsbG8gV29ybGQ="
encode_base64("Hello World", url_safe=true)       // -> URL-safe version
encode_base64("Hello World", padding=false)       // -> "SGVsbG8gV29ybGQ"
```

**Parameters:**

| Parameter  | Type           | Description                                  |
| ---------- | -------------- | -------------------------------------------- |
| `_content` | `str`          | Text to encode                               |
| `url_safe` | `bool = false` | Replace `+/` with `-_` for URL-safe encoding |
| `padding`  | `bool = true`  | Include `=` padding characters               |

Use `url_safe=true` to replace `+/` with `-_` for URL-safe encoding. Use `padding=false` to omit `=` padding.

### gen_fid

Generates a random flex ID (https://github.com/amterp/flexid) (fid) - a time-ordered, URL-safe identifier.

```rad
gen_fid(*, alphabet: str?, tick_size_ms: int?, num_random_chars: int?) -> error|str
```

```rad
gen_fid()                                    // -> "1a2b3c4d5e"
gen_fid(alphabet="0123456789")               // -> "1234567890"
gen_fid(num_random_chars=3)                  // -> "1a2b3c"
```

**Parameters:**

| Parameter          | Type                       | Description                            |
| ------------------ | -------------------------- | -------------------------------------- |
| `alphabet`         | `str? = "[0-9][A-Z][a-z]"` | Characters to use (base-62 by default) |
| `tick_size_ms`     | `int? = 1`                 | Time precision in milliseconds         |
| `num_random_chars` | `int? = 6`                 | Number of random characters to append  |

Defaults: `alphabet` is base-62 (`[0-9][A-Z][a-z]`), `tick_size_ms` is 1ms, `num_random_chars` is 6.

### hash

Generates a hash of the input text using various algorithms.

```rad
hash(_val: str, algo: ["sha1", "sha256", "sha512", "md5"] = "sha1") -> str
```

```rad
hash("hello world")                    // -> "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
hash("hello world", algo="sha256")     // -> "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
hash("sensitive data", algo="sha512")  // -> Long SHA-512 hash
```

**Parameters:**

| Parameter | Type                                           | Description              |
| --------- | ---------------------------------------------- | ------------------------ |
| `_val`    | `str`                                          | Text to hash             |
| `algo`    | `["sha1", "sha256", "sha512", "md5"] = "sha1"` | Hashing algorithm to use |

The default `sha1` is **not cryptographically secure**. Use `sha256` or `sha512` for security.

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

## Formatting

### black

Wraps its argument in the ANSI escape codes for black text.

```rad
black(_item: any) -> str
```

```rad
black("Hello")        // -> "Hello" wrapped in the black escape
black(42)             // -> "42" wrapped in the black escape
```

See also: `red`, `blue`, `color_rgb`

### blue

Wraps its argument in the ANSI escape codes for blue text.

```rad
blue(_item: any) -> str
```

```rad
blue("Hello")        // -> "Hello" wrapped in the blue escape
blue(42)             // -> "42" wrapped in the blue escape
```

See also: `cyan`, `magenta`, `color_rgb`

### bold

Wraps its argument in the ANSI escape codes for bold text.

```rad
bold(_item: any) -> str
```

```rad
bold("Hello")        // -> "Hello" wrapped in the bold escape
bold(42)             // -> "42" wrapped in the bold escape
```

See also: `dim`, `italic`, `underline`

### cyan

Wraps its argument in the ANSI escape codes for cyan text.

```rad
cyan(_item: any) -> str
```

```rad
cyan("Hello")        // -> "Hello" wrapped in the cyan escape
cyan(42)             // -> "42" wrapped in the cyan escape
```

See also: `blue`, `green`, `color_rgb`

### dim

Wraps its argument in the ANSI escape codes for dimmed text - the reverse of bold, useful for de-emphasising less important output.

```rad
dim(_item: any) -> str
```

```rad
dim("Hello")        // -> "Hello" wrapped in the dim escape
dim(42)             // -> "42" wrapped in the dim escape
```

See also: `bold`, `italic`

### green

Wraps its argument in the ANSI escape codes for green text.

```rad
green(_item: any) -> str
```

```rad
green("Hello")        // -> "Hello" wrapped in the green escape
green(42)             // -> "42" wrapped in the green escape
```

See also: `red`, `yellow`, `color_rgb`

### italic

Wraps its argument in the ANSI escape codes for italic text. Not every terminal renders italics; some show inverse or coloured text instead.

```rad
italic(_item: any) -> str
```

```rad
italic("Hello")        // -> "Hello" wrapped in the italic escape
italic(42)             // -> "42" wrapped in the italic escape
```

See also: `bold`, `underline`

### magenta

Wraps its argument in the ANSI escape codes for magenta text.

```rad
magenta(_item: any) -> str
```

```rad
magenta("Hello")        // -> "Hello" wrapped in the magenta escape
magenta(42)             // -> "42" wrapped in the magenta escape
```

See also: `red`, `blue`, `color_rgb`

### orange

Wraps its argument in the ANSI escape codes for orange text. Rendered via the closest 256-colour palette entry on terminals that don't support 24-bit colour.

```rad
orange(_item: any) -> str
```

```rad
orange("Hello")        // -> "Hello" wrapped in the orange escape
orange(42)             // -> "42" wrapped in the orange escape
```

See also: `yellow`, `red`, `color_rgb`

### pink

Wraps its argument in the ANSI escape codes for pink text. Rendered via the closest 256-colour palette entry on terminals that don't support 24-bit colour.

```rad
pink(_item: any) -> str
```

```rad
pink("Hello")        // -> "Hello" wrapped in the pink escape
pink(42)             // -> "42" wrapped in the pink escape
```

See also: `magenta`, `red`, `color_rgb`

### plain

Returns its argument as a plain string with no terminal colour or style applied. Useful for stripping styling back out of an expression where some branches have it and others don't.

```rad
plain(_item: any) -> str
```

```rad
plain("Hello")        // -> "Hello" wrapped in the plain escape
plain(42)             // -> "42" wrapped in the plain escape
```

See also: `colorize`, `color_rgb`

### red

Wraps its argument in the ANSI escape codes for red text.

```rad
red(_item: any) -> str
```

```rad
red("Hello")        // -> "Hello" wrapped in the red escape
red(42)             // -> "42" wrapped in the red escape
```

See also: `green`, `yellow`, `color_rgb`

### strikethrough

Wraps its argument in the ANSI escape codes for strikethrough text. Renders with a line through it on terminals that support the attribute.

```rad
strikethrough(_item: any) -> str
```

```rad
strikethrough("Hello")        // -> "Hello" wrapped in the strikethrough escape
strikethrough(42)             // -> "42" wrapped in the strikethrough escape
```

See also: `underline`, `dim`

### underline

Wraps its argument in the ANSI escape codes for underlined text.

```rad
underline(_item: any) -> str
```

```rad
underline("Hello")        // -> "Hello" wrapped in the underline escape
underline(42)             // -> "42" wrapped in the underline escape
```

See also: `bold`, `italic`

### white

Wraps its argument in the ANSI escape codes for white text.

```rad
white(_item: any) -> str
```

```rad
white("Hello")        // -> "Hello" wrapped in the white escape
white(42)             // -> "42" wrapped in the white escape
```

See also: `black`, `plain`, `color_rgb`

### yellow

Wraps its argument in the ANSI escape codes for yellow text.

```rad
yellow(_item: any) -> str
```

```rad
yellow("Hello")        // -> "Hello" wrapped in the yellow escape
yellow(42)             // -> "42" wrapped in the yellow escape
```

See also: `red`, `green`, `color_rgb`

## HTTP

### http_connect

Sends an HTTP CONNECT to `url` and returns the response as a map. Typically used for tunnelling through a proxy.

```rad
http_connect(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_connect("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_delete

Sends an HTTP DELETE to `url` and returns the response as a map.

```rad
http_delete(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_delete("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_get

Sends an HTTP GET to `url` and returns the response as a map. See `## Notes` for the response shape.

```rad
http_get(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_get("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_head

Sends an HTTP HEAD to `url` and returns the response as a map. The server returns headers without a body; the response map's `body` is omitted.

```rad
http_head(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_head("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_options

Sends an HTTP OPTIONS request to `url` and returns the response as a map. Typically used to discover the methods supported by a resource.

```rad
http_options(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_options("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_patch

Sends an HTTP PATCH to `url` and returns the response as a map. Use `body` for raw payloads or `json` for automatic JSON serialisation.

```rad
http_patch(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_patch("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_post

Sends an HTTP POST to `url` and returns the response as a map. Use `body` for raw payloads or `json` for automatic JSON serialisation.

```rad
http_post(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_post("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_put

Sends an HTTP PUT to `url` and returns the response as a map. Use `body` for raw payloads or `json` for automatic JSON serialisation.

```rad
http_put(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_put("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

### http_trace

Sends an HTTP TRACE to `url` and returns the response as a map. Often disabled at the server for security; expect failures against modern endpoints.

```rad
http_trace(url: str, *, body: any?, json: any?, headers: map?, insecure: bool = false) -> { "success": bool, "status_code"?: int, "headers": map, "body"?: any, "error"?: str, "duration_seconds": float }
```

```rad
r = http_trace("https://api.example.com/resource")
if r.success:
    print(r.body)
```

**Response map keys:**

- `success: bool` - whether the request succeeded.
- `duration_seconds: float` - total request time.
- `status_code?: int` - present when a response was received.
- `headers: map` - response headers.
- `body?: any` - response body, JSON-decoded when possible.
- `error?: str` - error message when `success` is false.

**Body vs JSON:** `body` is sent as-is; `json` is JSON-serialised and sets `Content-Type: application/json` when no headers are supplied. The two are mutually exclusive.

**Insecure:** pass `insecure=true` to skip TLS certificate verification.

## IO

### confirm

Gets a boolean confirmation from the user (y/n prompt). Accepts "y", "yes", or Enter (empty input) as
confirmation.

```rad
confirm(prompt: str = "Confirm? [Y/n] > ") -> error|bool
```

```rad
if confirm():                        // -> Uses default "Confirm? [Y/n] > " prompt
    print("Confirmed!")

if confirm("Delete file? [Y/n] "):   // -> Custom prompt
    print("File deleted")
```

### debug

Behaves like `print` but only outputs when debug mode is enabled via `--debug` flag.

```rad
debug(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

```rad
debug("entering loop")             // -> nothing unless --debug is on
debug("x =", x, "y =", y)          // -> debug-only diagnostics
```

### delete_path

Deletes a file or directory at the specified path.

```rad
delete_path(_path: str) -> bool
```

```rad
delete_path("temp.txt")         // -> true (if file existed and was deleted)
delete_path("missing.txt")      // -> false (file didn't exist)
delete_path("directory/")       // -> true (if directory existed and was deleted)
```

Returns `true` if the path was successfully deleted, `false` if it didn't exist or couldn't be deleted.

A leading `~` in `_path` is expanded to your home directory.

### find_paths

Returns a list of all paths under a directory.

```rad
find_paths(_path: str, *, depth: int = -1, relative: ["target", "cwd", "absolute"] = "target") -> error|str[]
```

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

**Parameters:**

| Parameter  | Type                                       | Description                            |
| ---------- | ------------------------------------------ | -------------------------------------- |
| `_path`    | `str`                                      | Directory to search                    |
| `depth`    | `int = -1`                                 | Max depth to search (-1 for unlimited) |
| `relative` | `["target", "cwd", "absolute"] = "target"` | How to format returned paths           |

- `"target"` - Relative to input path (default)
- `"cwd"` - Relative to current directory
- `"absolute"` - Full absolute paths

A leading `~` in `_path` is expanded to your home directory.

### input

Gets a line of text input from the user with optional prompt, default, hint, and secret mode.

```rad
input(prompt: str = "> ", *, hint: str = "", default: str = "", secret: bool = false) -> error|str
```

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

**Parameters:**

| Parameter | Type           | Description                                  |
| --------- | -------------- | -------------------------------------------- |
| `prompt`  | `str = "> "`   | The text prompt to display to the user       |
| `hint`    | `str = ""`     | Placeholder text shown in input field        |
| `default` | `str = ""`     | Default value if user doesn't enter anything |
| `secret`  | `bool = false` | If true, hides input (useful for passwords)  |

If `secret` is true, input is hidden (useful for passwords). The `hint` parameter has no effect when `secret` is
enabled.

### multipick

Presents an interactive menu for selecting multiple options from a list.

```rad
multipick(_options: str[], *, prompt: str?, min: int = 0, max: int?) -> str[]
```

```rad
fruits = ["apple", "banana", "cherry", "date"]
selected = multipick(fruits)
// selected equals e.g. [ "apple", "cherry" ]
```

Shows an interactive multi-select menu where users can select zero or more options.
Unlike `pick`, which returns a single selection, `multipick` returns a list of all selected items.

**Parameters:**

| Parameter  | Type      | Description                                                                   |
| ---------- | --------- | ----------------------------------------------------------------------------- |
| `_options` | `str[]`   | List of options to display in the menu                                        |
| `prompt`   | `str?`    | Custom prompt text. If not provided, automatically generated based on min/max |
| `min`      | `int = 0` | Minimum number of selections required (default 0 allows empty selection)      |
| `max`      | `int?`    | Maximum number of selections allowed (optional, unlimited if not set)         |

The `prompt` parameter has smart defaults that adjust based on the min/max constraints.

### pick

Presents an interactive menu for selecting from a list of options.

```rad
pick(_options: str[], _filter: (str|str[])?, *, prompt: str = "Pick an option", prefer_exact: bool = false) -> str
```

```rad
pick(["apple", "banana", "cherry"])                        // -> Interactive menu
pick(["red", "green", "blue"], "r")                        // -> Fuzzy-filtered to "red", "green"
pick(["grape", "g"], "g", prefer_exact=true)                 // -> Immediately picks "g" (exact match)
pick(["one", "two", "three"], prompt="Choose:")            // -> Custom prompt
```

Shows a fuzzy-searchable menu. Filter can be a string or list of strings to pre-filter options.

When `prefer_exact=true`, exact key matches (case-insensitive) are prioritized: if exactly one option exactly matches a
filter, it's selected immediately; if multiple match exactly, only those are shown.

### pick_from_resource

Loads options from a resource file and presents an interactive menu.

```rad
pick_from_resource(path: str, _filter: str?, *, prompt: str = "Pick an option", prefer_exact: bool = true) -> any
```

```rad
pick_from_resource("servers.json")                    // -> Menu from file
pick_from_resource("configs.json", "prod")            // -> Pre-filtered, exact match priority
pick_from_resource("data.json", prompt="Select:")     // -> Custom prompt
pick_from_resource("data.json", "x", prefer_exact=false) // -> Pure fuzzy matching
```

Loads data from a JSON file and presents it as selectable options. Returns the selected item(s).

With `prefer_exact=true` (the default), exact key matches (case-insensitive) are prioritized: if exactly one entry has a
key that exactly matches the filter, it's selected immediately; if multiple match exactly, only those are shown. Set
`prefer_exact=false` to disable this and use pure fuzzy matching.

### pick_kv

Presents an interactive menu showing keys but returns corresponding values.

```rad
pick_kv(keys: str[], values: any[], _filter: (str|str[])?, *, prompt: str = "Pick an option", prefer_exact: bool = false) -> any
```

```rad
names = ["Alice", "Bob", "Charlie"]
ages = [25, 30, 35]
pick_kv(names, ages)                                        // -> Shows names, returns age
pick_kv(["Red", "Green"], ["#ff0000", "#00ff00"])           // -> Shows colors, returns hex
pick_kv(["grape", "g"], [1, 2], "g", prefer_exact=true)       // -> Returns 2 (exact match)
```

Displays keys in the menu but returns the value at the same index when selected.

When `prefer_exact=true`, exact key matches (case-insensitive) are prioritized: if exactly one key exactly matches a
filter, its value is returned immediately; if multiple match exactly, only those are shown.

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

### print

Writes its arguments to stdout, separated by a space, followed by a
newline. The default workhorse for output. For error output, use
`print_err`; for structured pretty-printing, use `pprint`.

```rad
print(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

```rad
print("hello", "world")    // -> hello world
print(1, 2, 3, sep=", ")   // -> 1, 2, 3
print("no newline", end="") // -> no newline
```

See also: `print_err`, `pprint`, `debug`

### print_err

Behaves like `print` but outputs to stderr instead of stdout.

```rad
print_err(*_items: any, *, sep: str = " ", end: str = "\n") -> void
```

```rad
print_err("failed to load config")     // -> writes to stderr
print_err("error:", err.msg)           // -> "error: <msg>" to stderr
```

### read_file

Reads the contents of a file.

```rad
read_file(_path: str, *, mode: ["text", "bytes"] = "text") -> error|{ "size_bytes": int, "content": str|int[] }
```

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

**Parameters:**

| Parameter | Type                         | Description                     |
| --------- | ---------------------------- | ------------------------------- |
| `_path`   | `str`                        | Path to the file to read        |
| `mode`    | `["text", "bytes"] = "text"` | Read as UTF-8 text or raw bytes |

In text mode, decodes as UTF-8 and returns a string. In bytes mode, returns a list of integers.

A leading `~` in `_path` is expanded to your home directory.

**Return map contains:**

- `size_bytes: int` - File size in bytes
- `content: str|list[int]` - File contents (type depends on mode)

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
lines = content.split_lines() // Process stdin line-by-line
```

### write_file

Writes content to a file. Creates the file if it doesn't exist.

```rad
write_file(_path: str, _content: str, *, append: bool = false) -> error|{ "bytes_written": int, "path": str }
```

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

**Parameters:**

| Parameter  | Type           | Description                                       |
| ---------- | -------------- | ------------------------------------------------- |
| `_path`    | `str`          | Path where to write the file                      |
| `_content` | `str`          | Content to write                                  |
| `append`   | `bool = false` | Append to existing content instead of overwriting |

By default overwrites the file. Use `append=true` to append to existing content.

A leading `~` in `_path` is expanded to your home directory.

**Return map contains:**

- `bytes_written: int` - Number of bytes written
- `path: str` - Full path to the written file

## Lists

### filter

Applies a predicate function to filter elements of a list or map. Keeps only elements where the function returns true.

```rad
filter(_coll: map|list, _fn: fn(any) -> bool | fn(any, any) -> bool) -> map|list
```

```rad
filter([1, 2, 3, 4], fn(x) x % 2 == 0)      // -> [2, 4]
filter({"a": 1, "b": 2}, fn(k, v) v > 1)    // -> {"b": 2}
```

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)`.

### flat_map

Flattens a list of lists, or applies a mapping function that returns lists and flattens the results.

```rad
flat_map(_coll: map|list, _fn: any?) -> list
```

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

**For lists without function:** All elements must be lists. Flattens one level.

**With function:** The function must return a list. Results are flattened.

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)` and is required.

### join

Joins a list into a string with separator, prefix, and suffix.

```rad
join(_list: list, sep: str = "", prefix: str = "", suffix: str = "") -> str
```

```rad
join([1, 2, 3], sep=", ")           // -> "1, 2, 3"
join(["a", "b"], prefix="[", suffix="]")  // -> "[ab]"
join(["x", "y", "z"], sep="-", prefix="(", suffix=")")  // -> "(x-y-z)"
```

### keys

Returns all keys from a map as a list.

```rad
keys(_map: map) -> any[]
```

```rad
keys({"a": 1, "b": 2, "c": 3})  // -> ["a", "b", "c"]
keys({})                        // -> []
```

### len

Returns the number of elements in a string, list, or map. For
strings this is the rune count (not byte count), so unicode characters
contribute one each.

```rad
len(_val: str|list|map) -> int
```

```rad
len("hello")              // -> 5
len([1, 2, 3])            // -> 3
len({"a": 1, "b": 2})     // -> 2
len("héllo")              // -> 5 (rune count, not byte count)
```

See also: `sort`, `keys`, `values`

### map

Applies a function to every element of a list or entry of a map.

```rad
map(_coll: map|list, _fn: fn(any) -> any | fn(any, any) -> any) -> map|list
```

```rad
map([1, 2, 3], fn(x) x * 2)              // -> [2, 4, 6]
map({"a": 1, "b": 2}, fn(k, v) v * 10)   // -> {"a": 10, "b": 20}
```

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)`.

### sort

Returns a new sorted list (or string with characters sorted). The
input is not mutated. With `reverse=true`, sorts in descending order.
Multiple lists / strings can be passed - they're sorted in lockstep
using the first as the key.

```rad
sort(_primary: list|str, *_others: list|str, *, reverse: bool = false) -> list|str
```

```rad
sort([3, 1, 2])             // -> [1, 2, 3]
sort([3, 1, 2], reverse=true)  // -> [3, 2, 1]
sort("dcba")                // -> "abcd"

ages = [30, 25, 28]
names = ["alice", "bob", "carol"]
sort(ages, names)           // -> [25, 28, 30]
                            //    names is now ["bob", "carol", "alice"]
```

See also: `len`, `reverse`

### unique

Returns a list with duplicate values removed, preserving first occurrence order.

```rad
unique(_list: any[]) -> any[]
```

```rad
unique([2, 1, 2, 3, 1, 3, 4])  // -> [2, 1, 3, 4]
unique(["a", "b", "a", "c"])    // -> ["a", "b", "c"]
```

### values

Returns all values from a map as a list.

```rad
values(_map: map) -> any[]
```

```rad
values({"a": 1, "b": 2, "c": 3})  // -> [1, 2, 3]
values({})                         // -> []
```

### zip

Combines multiple lists into a list of lists, pairing elements by index.

```rad
zip(*_lists: list, *, fill: any?, strict: bool = false) -> error|list[]
```

```rad
// Basic usage
zip([1, 2, 3], ["a", "b", "c"])           // -> [[1, "a"], [2, "b"], [3, "c"]]
zip([1, 2, 3, 4], ["a", "b"])             // -> [[1, "a"], [2, "b"]]

// With fill value for unequal lengths
zip([1, 2, 3, 4], ["a", "b"], fill="-")   // -> [[1, "a"], [2, "b"], [3, "-"], [4, "-"]]

// Strict mode (errors on length mismatch)  
zip([1, 2, 3], ["a", "b"], strict=true)   // -> Error: Lists must have the same length
```

**Parameters:**

| Parameter | Type           | Description                              |
| --------- | -------------- | ---------------------------------------- |
| `*lists`  | `list`         | Variable number of lists to zip together |
| `strict`  | `bool = false` | If true, error on different list lengths |
| `fill`    | `any?`         | Value to fill shorter lists (optional)   |

- By default, truncates to the shortest list length
- Cannot use `strict=true` with `fill` parameter (mutually exclusive)
- Returns error if `strict=true` and lists have different lengths

## Math

### abs

Returns the absolute value of a number. The result's type matches
the input - `int` in, `int` out; `float` in, `float` out.

```rad
abs(_num: int|float) -> int|float
```

```rad
abs(-5)        // -> 5
abs(5)         // -> 5
abs(-3.14)     // -> 3.14
abs(0)         // -> 0
```

See also: `floor`, `ceil`, `round`

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

### clamp

Constrains a value between minimum and maximum bounds.

```rad
clamp(val: int|float, min: int|float, max: int|float) -> error|int|float
```

```rad
clamp(25, 20, 30)    // -> 25
clamp(10, 20, 30)    // -> 20
clamp(40, 20, 30)    // -> 30
clamp(5, 1.0, 10)    // -> 5.0 (float because 1.0 is float)
clamp(15, 30, 20)    // -> Error: min must be <= max
```

**Parameters:**

| Parameter | Type          | Description        |
| --------- | ------------- | ------------------ |
| `val`     | `int | float` | Value to constrain |
| `min`     | `int | float` | Minimum bound      |
| `max`     | `int | float` | Maximum bound      |

Returns `val` if between min and max, otherwise returns the nearest bound. Min must be ≤ max.
The return type preserves the input type: returns `int` if all inputs are integers, `float` if any input is a float.

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

### max

Returns the maximum value from a list of numbers or from variadic arguments.

```rad
max(*_nums: float|float[]) -> int|float|error
```

```rad
max([1, 2, 3, 4])        // -> 4
max(1, 2, 3, 4)          // -> 4
max(5.5, 2.1, 8.9)       // -> 8.9
max(1, 2.0, 3)           // -> 3.0 (float because 2.0 is float)
max(5)                   // -> 5
max([])                  // -> Error: cannot find maximum of empty list
max([1, "text"])         // -> Error: requires list of numbers
```

Accepts either a single list of numbers or multiple number arguments.
The return type preserves the input type: returns `int` if all inputs are integers, `float` if any input is a float.

### min

Returns the minimum value from a list of numbers or from variadic arguments.

```rad
min(*_nums: float|float[]) -> int|float|error
```

```rad
min([1, 2, 3, 4])        // -> 1
min(1, 2, 3, 4)          // -> 1
min(5.5, 2.1, 8.9)       // -> 2.1
min(1, 2.0, 3)           // -> 1.0 (float because 2.0 is float)
min(5)                   // -> 5
min([])                  // -> Error: cannot find minimum of empty list
min([1, "text"])         // -> Error: requires list of numbers
```

Accepts either a single list of numbers or multiple number arguments.
The return type preserves the input type: returns `int` if all inputs are integers, `float` if any input is a float.

### pow

Raises `base` to the power of `exponent`. Useful for exponentiation, square roots, and cube roots.

```rad
pow(_base: float, _exponent: float) -> float
```

```rad
pow(2, 3)      // -> 8
pow(4, 0.5)    // -> 2.0 (square root)
pow(8, 1/3)    // -> 2.0 (cube root)  
pow(2, -2)     // -> 0.25
pow(-2, 3)     // -> -8
```

### range

Returns a list of numbers covering the half-open interval
`[start, stop)`. With one argument, `start` defaults to 0. The list
type matches the inputs - all ints produces an `int[]`, any float
produces a `float[]`.

```rad
range(_arg1: float|int, _arg2: (float|int)?, _step: float|int = 1) -> float[]|int[]
```

```rad
range(5)              // -> [0, 1, 2, 3, 4]
range(1, 5)           // -> [1, 2, 3, 4]
range(0, 1, 0.25)     // -> [0.0, 0.25, 0.5, 0.75]
range(10, 0, -2)      // -> [10, 8, 6, 4, 2]
```

See also: `for`, `len`

### round

Rounds a number to the specified decimal precision.

```rad
round(_num: float, _decimals: int = 0) -> error|int|float
```

```rad
round(3.14159)           // -> 3 (integer)
round(3.14159, 2)        // -> 3.14 (float)
round(2.7)               // -> 3 (integer)
round(3.14, -1)          // -> Error: precision must be non-negative
```

**Parameters:**

| Parameter   | Type      | Description                                     |
| ----------- | --------- | ----------------------------------------------- |
| `_num`      | `float`   | Number to round                                 |
| `_decimals` | `int = 0` | Number of decimal places (must be non-negative) |

With precision 0, returns an integer. With precision > 0, returns a float. Precision must be non-negative.

### sum

Sums all numbers in a list.

```rad
sum(_nums: float[]) -> error|int|float
```

```rad
sum([1, 2, 3, 4])        // -> 10
sum([1.5, 2.5, 3.0])     // -> 7.0
sum([1, 2.0, 3])         // -> 6.0 (float because 2.0 is float)
sum([])                  // -> 0
sum([1, "text", 3])      // -> Error: requires list of numbers
```

The return type preserves the input type: returns `int` if all inputs are integers, `float` if any input is a float.

## Parsing

### convert_duration

Converts a numeric value in the specified unit. Supports `nanos`, `micros`, `millis`, `seconds`, `minutes`, `hours`, and
`days`.

```rad
convert_duration(_value: int|float, _unit: ["nanos", "micros", "millis", "seconds", "minutes", "hours", "days"]) -> error|{ "nanos": int, "micros": float, "millis": float, "seconds": float, "minutes": float, "hours": float, "days": float }
```

```rad
convert_duration(90, "seconds").minutes   // -> 1.5
convert_duration(1, "days").hours         // -> 24.0
convert_duration(1.5, "hours").minutes    // -> 90.0
convert_duration(1500, "millis").seconds  // -> 1.5
```

### float

Converts a value to a float. Does not work on strings - use `parse_float` for string parsing.

```rad
float(_var: any) -> float|error
```

```rad
float(42)      // -> 42.0
float(true)    // -> 1.0
float(false)   // -> 0.0  
float("3.14")  // -> Error: cannot convert string
```

### int

Converts a value to an integer. Does not work on strings - use `parse_int` for string parsing.

```rad
int(_var: any) -> int|error
```

```rad
int(3.14)     // -> 3
int(true)     // -> 1
int(false)    // -> 0
int("42")     // -> Error: cannot convert string
```

### parse_duration

Parses a human-readable duration string into a map of time units. Supports all standard suffixes (`ns`, `us`/`µs`, `ms`,
`s`, `m`, `h`) plus `d` for days (1d = 24h) and `w` for weeks (1w = 7d). Spaces are stripped, and a leading `-` negates
the whole duration.

```rad
parse_duration(_duration: str) -> error|{ "nanos": int, "micros": float, "millis": float, "seconds": float, "minutes": float, "hours": float, "days": float }
```

```rad
parse_duration("5m23s")         // -> { nanos: 323000000000, micros: 323000000.0, millis: 323000.0, seconds: 323.0, minutes: 5.3833..., hours: 0.0897..., days: 0.00374... }
parse_duration("1d12h").hours   // -> 36.0
parse_duration("1w2d3h").hours  // -> 219.0
parse_duration("300ms").millis  // -> 300.0
parse_duration("-5m").minutes   // -> -5.0
parse_duration("5m 30s")        // -> same as "5m30s"
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

### parse_json

Parses a JSON string into Rad data structures.

```rad
parse_json(_str: str) -> any|error
```

```rad
parse_json(r'{"name": "Alice", "age": 30}')  // -> {"name": "Alice", "age": 30}
parse_json('[1, 2, 3]')                      // -> [1, 2, 3]
parse_json('invalid json')                   // -> Error: invalid JSON
```

Use a raw string (`r'...'`) when the JSON contains `{` or `}` - plain
single- and double-quoted strings interpolate `{expr}`, which makes
JSON literals trip the interpolator. Raw strings are also natural for
JSON pasted verbatim from a sample.

### str

Converts any value to a string representation. Useful when you need to concatenate non-string values with `+`, though
interpolation (`"value: {x}"`) is generally preferred.

```rad
str(_var: any) -> str
```

```rad
str(42)        // -> "42"
str(3.14)      // -> "3.14"
str([1, 2])    // -> "[1, 2]"
str(true)      // -> "true"
```

### to_json

Serializes a Rad value into a JSON string. The inverse of `parse_json`.

```rad
to_json(_val: any, *, indent: int = 0) -> str
```

```rad
to_json({"name": "Alice", "age": 30})  // -> '{"age":30,"name":"Alice"}'
to_json([1, 2, "x"])                   // -> '[1,2,"x"]'
to_json("hi")                          // -> '"hi"'
to_json({"a": 1}, indent=2)            // -> multi-line, 2-space indented
```

**Parameters:**

| Parameter | Type      | Description                                          |
| --------- | --------- | ---------------------------------------------------- |
| `_val`    | `any`     | Value to serialize                                   |
| `indent`  | `int = 0` | If > 0, pretty-print with that many spaces of indent |

Unlike string interpolation or `str()`, the output is guaranteed to be valid
JSON: strings are escaped and quoted, including top-level ones. HTML characters
(`<`, `>`, `&`) are not escaped. `null` serializes to `"null"`.

Map keys are currently emitted in alphabetical order, not insertion order.

See also: `parse_json`, `pprint`

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

```rad
rand_int(10)        // -> Random int from 0-9
rand_int(5, 15)     // -> Random int from 5-14
rand_int(10, 5)     // -> Error: min (10) must be less than max (5)
```

With one argument, returns random int from 0 to `_arg1` (exclusive). With two arguments, returns random int from `_arg1`
to `_arg2` (exclusive). Min must be less than max.

### seed_random

Seeds the random number generator used by `rand` and `rand_int`.

```rad
seed_random(_seed: int) -> void
```

```rad
seed_random(42)
rand()              // -> Same sequence every time with seed 42
rand_int(10)        // -> Same sequence every time with seed 42
```

## Stash

### get_rad_home

Returns the path to rad's home folder on the user's machine.

**Return Values**

Defaults to `$HOME/.rad`, or `$RAD_HOME` if it's defined.

```rad
get_rad_home() -> str
```

```rad
home = get_rad_home()              // -> "/Users/me/.rad" (or $RAD_HOME)
```

### get_stash_path

Returns the full path to the script's stash directory, with the given subpath if specified.

Requires a stash ID to have been defined.

**Return Values**

- Without subpath defined: `<rad home>/stashes/<stash id>`
- With subpath defined: `<rad home>/stashes/<stash id>/<subpath>`

```rad
get_stash_path(_sub_path: str = "") -> error|str
```

```rad
root = get_stash_path()                // -> "<rad-home>/stashes/<stash-id>"
cache = get_stash_path("cache.json")   // -> "<rad-home>/stashes/<stash-id>/cache.json"
```

### load

Loads a value into a map using lazy evaluation. If key exists, returns cached value; otherwise runs loader function.

```rad
load(_map: map, _key: any, _loader: fn() -> any, *, reload: bool = false, override: any?) -> error|any
```

```rad
cache = {}
load(cache, "data", fn() expensive_calculation())    // -> Runs loader, caches result
load(cache, "data", fn() expensive_calculation())    // -> Returns cached value

// Force reload
load(cache, "data", fn() new_calculation(), reload=true)

// Override with specific value  
load(cache, "data", fn() ignored(), override="forced")
```

**Parameters:**

| Parameter  | Type           | Description                              |
| ---------- | -------------- | ---------------------------------------- |
| `_map`     | `map`          | Map to store/retrieve cached values      |
| `_key`     | `any`          | Key to lookup in the map                 |
| `_loader`  | `fn() -> any`  | Function to call if key doesn't exist    |
| `reload`   | `bool = false` | Force reload even if key exists          |
| `override` | `any?`         | Use this value instead of calling loader |

If key doesn't exist, `_loader` is called and result is cached. Cannot use `reload=true` with `override` (mutually
exclusive).

### load_stash_file

Loads a file from the script's stash directory, creating it with default content if it doesn't exist.

```rad
load_stash_file(_path: str, _default: str = "") -> error|{ "full_path": str, "created": bool, "content"?: str }
```

```rad
result = load_stash_file("config.txt", "default config")
if result.success:
    if result.created:
        print("Created new config file")
    content = result.content
```

**Return map contains:**

- `full_path: str` - Full path to the file
- `created: bool` - Whether the file was just created
- `content?: str` - File contents (if successfully loaded)

### load_state

Loads the script's stashed state. Creates it if it doesn't already exist.

Requires a stash ID to have been defined.

**Return Values**

1. `map` containing the saved state. Starts empty, before anything is saved to it.
2. `bool` representing if the state existed before the load, or if it was just created.

```rad
load_state() -> error|map
```

```rad
state = load_state()                   // -> map containing previous state
state["count"] = (state["count"] or 0) + 1
save_state(state)
```

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

## Strings

### color_rgb

Applies RGB coloring to input text. RGB values must be in range [0, 255]. Not all terminals support this.

```rad
color_rgb(_val: any, red: int, green: int, blue: int) -> error|str
```

```rad
color_rgb("Hello", red=255, green=0, blue=0)     // -> "Hello" (in bright red)
color_rgb(42, red=0, green=255, blue=128)        // -> "42" (in green-cyan)
color_rgb("test", red=300, green=0, blue=0)      // -> Error: RGB values must be [0, 255]
```

**Parameters:**

| Parameter | Type  | Description             |
| --------- | ----- | ----------------------- |
| `_val`    | `any` | Value to apply color to |
| `red`     | `int` | Red component (0-255)   |
| `green`   | `int` | Green component (0-255) |
| `blue`    | `int` | Blue component (0-255)  |

RGB values must be in range [0, 255]. Not all terminals support this.

### colorize

Assigns consistent colors to values from a set of possible values. The same value always gets the same color within the
same set.

```rad
colorize(_val: any, _enum: any[], *, skip_if_single: bool = false) -> str
```

```rad
names = ["Alice", "Bob", "Charlie"]
colorize("Alice", names)     // -> "Alice" (in consistent color)
colorize("Bob", names)       // -> "Bob" (in different consistent color)

// In rad blocks
names = ["Alice", "Bob", "Charlie", "David"]
rad:
    fields names
    names:
        map fn(n) colorize(n, names)
```

**Parameters:**

| Parameter        | Type           | Description                                    |
| ---------------- | -------------- | ---------------------------------------------- |
| `_val`           | `any`          | Value to colorize                              |
| `_enum`          | `any[]`        | Set of possible values for consistent coloring |
| `skip_if_single` | `bool = false` | Don't colorize if only one value in set        |

Useful for automatically coloring table data or distinguishing values in lists.

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

### ends_with

Checks if a string ends with a given substring.

```rad
ends_with(_val: str, _end: str) -> bool
```

```rad
ends_with("hello world", "world")    // -> true
ends_with("hello world", "hello")    // -> false
```

### hyperlink

Creates a clickable hyperlink in supporting terminals.

```rad
hyperlink(_val: any, _link: str) -> str
```

```rad
hyperlink("Visit Google", "https://google.com")    // -> Clickable "Visit Google" link
hyperlink("localhost", "http://localhost:3000")    // -> Clickable "localhost" link
hyperlink(42, "https://example.com")               // -> Clickable "42" link
```

Converts text into a terminal hyperlink that can be clicked in supported terminals.

### index_of

Finds the index of a target value within a string or list. Returns `null` if not found.

```rad
index_of(_subject: str|list, _target: any, *, n: int = 0, start: int = 0) -> int?
```

```rad
// String search
"hello world hello".index_of("hello")           // -> 0
"hello world hello".index_of("hello", n=1)      // -> 12
"hello world hello".index_of("hello", n=-1)     // -> 12
"hello".index_of("xyz")                         // -> null
"hello".index_of("xyz") ?? (-1)                 // -> -1
"hello".index_of("")                             // -> null (empty target)

// List search
["a", "b", "c", "b", "a"].index_of("b")        // -> 1
["a", "b", "c", "b", "a"].index_of("b", n=-1)  // -> 3
[1, 2, 3].index_of(99)                          // -> null
```

**Parameters:**

| Parameter  | Type       | Description                                           |
| ---------- | ---------- | ----------------------------------------------------- |
| `_subject` | `str|list` | The string or list to search within                   |
| `_target`  | `any`      | The value to search for                               |
| `n`        | `int = 0`  | Which occurrence to find (0=first, 1=second, -1=last) |
| `start`    | `int = 0`  | Position to start searching from                      |

### lower

Converts a string to lowercase. Preserves color attributes.

```rad
lower(_val: str) -> str
```

```rad
lower("HELLO")          // -> "hello"
lower("Hello World")    // -> "hello world"
```

### matches

Tests whether `_str` matches the regular-expression `_pattern`. By default the pattern must match the whole string; pass `partial=true` to match any substring. Returns an `error` when the pattern is malformed.

```rad
matches(_str: str, _pattern: str, *, partial: bool = false) -> bool|error
```

```rad
matches("hello", "h.+o")               // -> true
matches("hello world", "world")        // -> false (default is full-string match)
matches("hello world", "world", partial=true)  // -> true
matches("abc", "(")                    // -> error: invalid regex
```

See also: `replace`, `split`

### replace

Replaces text using regex patterns. Does not preserve string color attributes.

```rad
replace(_original: str, _find: str, _replace: str) -> str
```

```rad
replace("hello world", "world", "Rad")        // -> "hello Rad"
replace("Name: Charlie Brown", "Charlie (.*)", "Alice $1")  // -> "Name: Alice Brown"
replace("abc123def", "\\d+", "XXX")           // -> "abcXXXdef"
```

The `_find` parameter is a regex pattern. The `_replace` parameter can use regex capture groups like `$1`.

### reverse

Reverses a string or list. Preserves color attributes for strings.

```rad
reverse(_val: str|list) -> str|list
```

```rad
reverse("hello")           // -> "olleh"
reverse([1, 2, 3, 4])      // -> [4, 3, 2, 1]
reverse("racecar")         // -> "racecar"
```

### split

Splits a string using regex pattern as delimiter. Does not preserve string color attributes.

```rad
split(_val: str, _sep: str, *, limit: int?) -> str[]
```

```rad
split("a,b,c", ",")               // -> ["a", "b", "c"]
split("word1 word2", "\\s+")      // -> ["word1", "word2"]
split("abc123def", "\\d+")        // -> ["abc", "def"]
split("key=val=ue", "=", limit=1) // -> ["key", "val=ue"]
split("a,b,c,d", ",", limit=2)    // -> ["a", "b", "c,d"]
```

The `_sep` parameter is treated as a regex pattern if valid, otherwise as literal string.

When `limit` is provided, it caps the number of splits performed. The final element contains
the unsplit remainder. `limit` must be >= 1.

### split_lines

Splits a string by line endings. Handles all common styles: `\n` (Unix), `\r\n` (Windows), and `\r` (legacy Mac).

```rad
split_lines(_val: str) -> str[]
```

```rad
"a\nb\nc".split_lines()          // -> ["a", "b", "c"]
content = read_file("data.txt").content
for line in content.split_lines():
    print(line)
```

Use this instead of `split("\n")` when processing text that may come from different platforms.

Trailing line endings are stripped - `"a\nb\n".split_lines()` returns `["a", "b"]`, not `["a", "b", ""]`.

### starts_with

Checks if a string starts with a given substring.

```rad
starts_with(_val: str, _start: str) -> bool
```

```rad
starts_with("hello world", "hello")  // -> true
starts_with("hello world", "world")  // -> false
```

### trim

Strips all matching characters from both ends of a string. Preserves color attributes.

```rad
trim(_subject: str, _chars: str = " \t\n") -> str
```

```rad
trim("  hello  ")            // -> "hello"
trim("***hello***", "*")     // -> "hello"
trim("abcHELLOabc", "abc")   // -> "HELLO"
```

### trim_left

Strips all matching characters from the start of a string. Preserves color attributes.

```rad
trim_left(_subject: str, _chars: str = " \t\n") -> str
```

```rad
trim_left("  hello  ")          // -> "hello  "
trim_left("***hello***", "*")   // -> "hello***"
trim_left("aaabbb", "a")        // -> "bbb"
```

### trim_prefix

Removes a literal prefix from the start of a string (once). Preserves color attributes.

```rad
trim_prefix(_subject: str, _prefix: str) -> str
```

```rad
trim_prefix("hello world", "hello ")  // -> "world"
trim_prefix("aaabbb", "a")            // -> "aabbb" (one 'a' removed)
trim_prefix("test", "x")              // -> "test" (no match)
```

### trim_right

Strips all matching characters from the end of a string. Preserves color attributes.

```rad
trim_right(_subject: str, _chars: str = " \t\n") -> str
```

```rad
trim_right("  hello  ")         // -> "  hello"
trim_right("***hello***", "*")  // -> "***hello"
trim_right("aaabbb", "b")       // -> "aaa"
```

### trim_suffix

Removes a literal suffix from the end of a string (once). Preserves color attributes.

```rad
trim_suffix(_subject: str, _suffix: str) -> str
```

```rad
trim_suffix("hello world", " world")  // -> "hello"
trim_suffix("aaabbb", "b")            // -> "aaabb" (one 'b' removed)
trim_suffix("test", "x")              // -> "test" (no match)
```

### truncate

Truncates a string to a maximum length, adding an ellipsis if truncated. Requires length of at least 1.

```rad
truncate(_str: str, _len: int) -> error|str
```

```rad
truncate("hello world", 8)   // -> "hello w…"
truncate("short", 10)        // -> "short" (no truncation needed)
truncate("test", 0)          // -> Error: Requires at least 1
```

### upper

Converts a string to uppercase. Preserves color attributes.

```rad
upper(_val: str) -> str
```

```rad
upper("hello")          // -> "HELLO"
upper("Hello World")    // -> "HELLO WORLD"
```

## System

### error

Creates an error object with the given message.

```rad
error(_msg: str) -> error
```

```rad
fn validate(x: int):
    if x < 0:
        return error("Something went wrong")
    return x

result = validate(-1)  // -> Script will exit with this error message
```

`return` at the top level isn't legal Rad - wrap it in a `fn` and
return the error from there, or assign it and propagate via `??` /
`catch:`.

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

### get_args

Returns the raw command-line arguments passed to the script.

```rad
get_args() -> str[]
```

```rad
// If script was called: rad myscript.rad arg1 arg2 --flag
raw_args = get_args()  // -> ["./myscript.rad", "arg1", "arg2", "--flag"]
```

Returns all arguments after the script name. Unlike parsed args, this gives you raw access to all arguments.

The name `args` is reserved (it's the args-block keyword), so the
local has to be called something else.

### get_env

Retrieves the value of an environment variable.

```rad
get_env(_var: str) -> str
```

```rad
home_dir = get_env("HOME")                    // -> "/Users/username"
api_key = get_env("API_KEY") or "default"     // -> Uses default if not set
missing = get_env("NONEXISTENT")              // -> ""
```

Returns the environment variable value, or empty string if not set.

### get_path

Gets information about a file or directory path.

```rad
get_path(_path: str) -> { "exists": bool, "full_path": str, "base_name"?: str, "permissions"?: str, "type"?: str, "size_bytes"?: int, "modified_millis"?: int, "accessed_millis"?: int }
```

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

**Always returns:**

- `exists: bool` - Whether the path exists
- `full_path: str` - Absolute path (a leading `~` in `_path` is expanded to your home directory)

**When path exists, also returns:**

- `base_name?: str` - File/directory name
- `permissions?: str` - Permission string (e.g., "rwxr-xr-x")
- `type?: str` - Either "file" or "dir"
- `size_bytes?: int` - File size (only for files)
- `modified_millis?: int` - Modification time as epoch milliseconds
- `accessed_millis?: int` - Access time as epoch milliseconds (Unix/macOS only)

### get_pid

Returns the process ID (PID) of the running Rad process.

```rad
get_pid() -> int
```

```rad
pid = get_pid()    // -> e.g. 12345
print("My PID: {pid}")
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

### signal_ignore

Installs OS-level `SIG_IGN` for one or more signals, so the process is not woken
when they fire. The disposition is inherited by subprocesses. Distinct from a
no-op `signal_trap` handler, which still wakes the process on every delivery.

```rad
signal_ignore(_signal: ["sigint", "sigterm", "sighup", "sigusr1", "sigusr2", "sigpipe", "sigwinch"] | ["sigint", "sigterm", "sighup", "sigusr1", "sigusr2", "sigpipe", "sigwinch"][]) -> void
```

```rad
// Don't crash when a downstream pipe consumer (e.g. `head`) closes early
signal_ignore("sigpipe")

for i in range(0, 1000000):
    print(i)  // safe to pipe into `head` now
```

The primary use case is `sigpipe`: a script that pipes its output (e.g.
`rad script.rad | head`) would otherwise terminate when the consumer closes the
pipe. Ignoring it at the OS level keeps the script alive.

On Windows, only `sigint` and `sigterm` are supported.

### signal_trap

Registers a function to run when one of the named signals is delivered to the
script. The handler receives a single map argument with the signal name and the
conventional `128 + sig` exit code. Re-registering replaces the previous handler.

```rad
signal_trap(_signal: ["sigint", "sigterm", "sighup", "sigusr1", "sigusr2", "sigpipe", "sigwinch"] | ["sigint", "sigterm", "sighup", "sigusr1", "sigusr2", "sigpipe", "sigwinch"][], _handler: fn(any) -> any) -> void
```

```rad
// Cleanup on Ctrl+C
signal_trap("sigint", fn(ctx):
    print_err("Cancelled, cleaning up")
    exit(ctx.exit_code)  // 130
)

// One handler for several signals, dispatching on the name
signal_trap(["sigterm", "sighup"], fn(ctx):
    if ctx.signal == "sighup":
        reload_config()  // no exit() - control continues
    else:
        exit(ctx.exit_code)
)

// Status dump - script keeps running after the handler returns
signal_trap("sigusr1", fn(ctx):
    print("progress: {progress}/{total}")
)
```

The handler is invoked with one **map** argument (the parameter name `ctx` is
conventional; call it whatever you like):

| Field       | Type | Description                                              |
| ----------- | ---- | -------------------------------------------------------- |
| `signal`    | str  | The signal that fired (e.g. `"sigint"`).                 |
| `exit_code` | int  | The conventional `128 + sig` exit code (130 for SIGINT). |

After the handler returns, **execution always continues**. To terminate, the
handler must explicitly call `exit()`. This matches Bash `trap`, Python's
`signal.signal`, Ruby's `Signal.trap`, and Node's `process.on`. The
always-continue rule applies only to a clean return: an **unhandled error in the
handler aborts the script** (exit 1) and runs `defer`/`errdefer`, just like an
unhandled error anywhere else.

There is currently no built-in way to restore the platform default; once a
signal is trapped, it stays trapped for the lifetime of the script.

**Supported signals:**

| Signal     | Trigger                               | Default action  |
| ---------- | ------------------------------------- | --------------- |
| `sigint`   | Ctrl+C from terminal                  | terminate (130) |
| `sigterm`  | `kill <pid>`, systemd/Docker shutdown | terminate (143) |
| `sighup`   | Terminal hangup; convention: "reload" | terminate (129) |
| `sigusr1`  | User-defined; convention: "status"    | terminate (138) |
| `sigusr2`  | User-defined; convention: "toggle"    | terminate (140) |
| `sigpipe`  | Write to closed pipe                  | terminate (141) |
| `sigwinch` | Terminal resized                      | ignore          |

**Caveats:**

- Handlers run at the next statement boundary - they do not preempt
  mid-statement. A signal arriving during a long computation only fires when
  control returns between statements.
- A second SIGINT while a SIGINT handler is in progress force-exits with code
  130, skipping defers - the escape hatch when a handler hangs.
- Subprocesses started with `$` share Rad's process group, so Ctrl+C reaches the
  subprocess directly and it terminates before the Rad handler runs. The handler
  still runs (so temp-file cleanup works), but it cannot influence the already-
  killed subprocess.
- On Windows, only `sigint` and `sigterm` are supported.

### sleep

Pauses execution for the specified duration.

```rad
sleep(_duration: int|float|str, *, title: str?) -> void
```

```rad
sleep(2.5)              // -> Sleep for 2.5 seconds
sleep("1h30m")          // -> Sleep for 1 hour 30 minutes
sleep("500ms")          // -> Sleep for 500 milliseconds
sleep("1d12h")          // -> Sleep for 1 day 12 hours
sleep(5, title="Waiting...") // -> Prints "Waiting..." then sleeps 5 seconds
```

Integer and float values are treated as seconds. String values support duration format like "2h45m", "1.5s", "500ms".
Spaces are allowed in duration strings (e.g. `"5m 30s"`).
If `title` is provided, it's printed before sleeping.

**Duration string suffixes:**

| Suffix       | Description  |
| ------------ | ------------ |
| `d`          | Days         |
| `h`          | Hours        |
| `m`          | Minutes      |
| `s`          | Seconds      |
| `ms`         | Milliseconds |
| `us` or `µs` | Microseconds |
| `ns`         | Nanoseconds  |

### type_of

Returns the type of a value as a string.

```rad
type_of(_var: any) -> ["int", "str", "list", "map", "float", "bool", "null", "error", "function"]
```

```rad
type_of("hi")            // -> "str"
type_of([2])             // -> "list"
type_of(42)              // -> "int"
type_of(3.14)            // -> "float"
type_of({"a": 1})        // -> "map"
type_of(true)            // -> "bool"
type_of(null)            // -> "null"
type_of(fn() 1)          // -> "function"
// Builtins that may fail return an `error` value:
// type_of(parse_int("xx")) // -> "error"
```

## Time

### now

Returns the current time with various accessible formats.

```rad
now(*, tz: str = "local") -> error|{ "date": str, "year": int, "month": int, "day": int, "weekday": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }
```

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

**Parameters:**

| Parameter | Type            | Description                               |
| --------- | --------------- | ----------------------------------------- |
| `tz`      | `str = "local"` | Timezone (e.g., "UTC", "America/Chicago") |

Map values:

| Accessor         | Description                           | Type   | Example             |
| ---------------- | ------------------------------------- | ------ | ------------------- |
| `.date`          | Current date YYYY-MM-DD               | string | 2019-12-13          |
| `.year`          | Current calendar year                 | int    | 2019                |
| `.month`         | Current calendar month                | int    | 12                  |
| `.day`           | Current calendar day                  | int    | 13                  |
| `.weekday`       | ISO day of week (Monday=1..Sunday=7)  | int    | 5                   |
| `.hour`          | Current clock hour (24h)              | int    | 14                  |
| `.minute`        | Current minute of the hour            | int    | 15                  |
| `.second`        | Current second of the minute          | int    | 16                  |
| `.time`          | Current time in "hh:mm:ss" format     | string | 14:15:16            |
| `.epoch.seconds` | Seconds since 1970-01-01 00:00:00 UTC | int    | 1576246516          |
| `.epoch.millis`  | Millis since 1970-01-01 00:00:00 UTC  | int    | 1576246516123       |
| `.epoch.nanos`   | Nanos since 1970-01-01 00:00:00 UTC   | int    | 1576246516123456789 |

### parse_date

Parses a date string into the same time map format as `now()` and `parse_epoch()`.

```rad
parse_date(_date: str, *, format: str?, tz: str = "local") -> error|{ "date": str, "year": int, "month": int, "day": int, "weekday": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }
```

```rad
// Auto-detect common formats
time = parse_date("2026-03-22")
print(time.date)              // -> "2026-03-22"
print(time.epoch.seconds)     // -> epoch at midnight local time

time = parse_date("2026-03-22T14:30:00Z")
print(time.hour)              // -> 14 (or local equivalent)

// Custom format for non-standard date strings
time = parse_date("22/03/2026", format="DD/MM/YYYY")
print(time.date)              // -> "2026-03-22"

time = parse_date("22.03.2026 14:30", format="DD.MM.YYYY HH:mm")
print(time.hour, time.minute) // -> 14 30

// Timezone conversion
time = parse_date("2026-03-22T14:30:00Z", tz="America/Chicago")
print(time.hour)              // -> 9 (CDT = UTC-5)

// Error handling
time = parse_date("bad input") catch:
    print(time)  // error message with format hints
```

**Parameters:**

| Parameter | Type            | Description                                                     |
| --------- | --------------- | --------------------------------------------------------------- |
| `_date`   | `str`           | The date string to parse                                        |
| `format`  | `str?`          | Format string using tokens (see below). Auto-detects if omitted |
| `tz`      | `str = "local"` | Timezone (e.g., "UTC", "America/Chicago")                       |

**Auto-detected formats** (when `format` is omitted):

- `YYYY-MM-DD` (e.g., `2026-03-22`)
- `YYYY-MM-DDTHH:mm:ss` (e.g., `2026-03-22T14:30:00`)
- `YYYY-MM-DDTHH:mm:ssZ` or with offset (e.g., `2026-03-22T14:30:00+05:00`)
- `YYYY-MM-DD HH:mm:ss` (space-separated)
- All of the above with optional fractional seconds (e.g., `.123456`)

**Format tokens** (for the `format` parameter):

| Token  | Meaning               | Example |
| ------ | --------------------- | ------- |
| `YYYY` | 4-digit year          | `2026`  |
| `MM`   | 2-digit month (01-12) | `03`    |
| `DD`   | 2-digit day (01-31)   | `22`    |
| `HH`   | 2-digit hour, 24h     | `14`    |
| `mm`   | 2-digit minute        | `30`    |
| `ss`   | 2-digit second        | `00`    |

All other characters in the format string are treated as literal separators. Format strings should
contain only tokens and separators - avoid embedding prose text, as tokens like `mm` and `ss` will be
matched inside words (e.g., "co**mm**it" or "acce**ss**").

Note: `MM` (uppercase) is **month**, `mm` (lowercase) is **minute**. Mixing these up will produce
wrong results or parse errors.

For strings without timezone information, the time is interpreted in the `tz` timezone (local by
default). For strings with timezone info (e.g., `Z`, `+05:00`), the time is parsed in that timezone
and then converted to the output `tz`.

### parse_epoch

Parses a Unix epoch timestamp into various time formats.

```rad
parse_epoch(_epoch: int|float, *, tz: str = "local", unit: ["auto", "seconds", "millis", "micros", "nanos", "milliseconds", "microseconds", "nanoseconds"] = "auto") -> error|{ "date": str, "year": int, "month": int, "day": int, "weekday": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }
```

```rad
// Parse seconds epoch (auto-detected)
time = parse_epoch(1712345678)
print(time.date, time.time)  // -> "2024-04-05 22:01:18"

// Parse milliseconds with timezone
time = parse_epoch(1712345678123, tz="America/Chicago")
print(time.hour)  // -> Hour in Chicago timezone

// Explicit unit specification
time = parse_epoch(1712345678000, unit="millis")

// Float epoch with sub-second precision
time = parse_epoch(1712345678.5)  // 1712345678 seconds + 500ms
print(time.epoch.millis)  // -> 1712345678500

// Float with explicit unit (sub-millisecond precision)
time = parse_epoch(1712345678123.25, unit="millis")
print(time.epoch.nanos)  // -> 1712345678123250000

// Error handling
time, err = parse_epoch(1712345678, tz="Invalid/Timezone")
if err:
    print("Invalid timezone:", err.msg)
```

**Parameters:**

| Parameter | Type                                                        | Description                                         |
| --------- | ----------------------------------------------------------- | --------------------------------------------------- |
| `_epoch`  | `int|float`                                                 | Unix epoch timestamp (float for sub-unit precision) |
| `tz`      | `str = "local"`                                             | Timezone (e.g., "UTC", "America/Chicago")           |
| `unit`    | `["auto", "seconds", "millis", "micros", "nanos"] = "auto"` | Timestamp unit (auto-detects by default)            |

Converts an epoch timestamp to the same format as `now()`. Auto-detects units from digit count, or specify
explicitly. When using a float, the fractional part provides sub-unit precision (e.g., `1712345678.5` seconds includes
500 milliseconds).
