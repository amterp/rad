---
title: Args
---

This section covers syntax for declaring arguments that your script can accept.

## Arg Declarations

Rad takes a declarative approach to arguments.
You simply declare what arguments your script accepts, you can define some constraints you want for them,
and let Rad take care of the rest, including parsing user input and validation.

Arguments are declared as part of an **args block**.

Here's an example script we'll call `printwords` that prints an input word some number of times:

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
args:
    word str
    repeats int
    
for _ in range(repeats):
    print(word)
```

We can print its usage string using the `-h` flag:

```shell
./printwords -h
```

<div class="result">
```
Usage:
  printwords <word> <repeats>

Script args:
      --word str
      --repeats int
```
</div>

This script defines two mandatory arguments: `word` that is expected to be a string, and `repeats` which is expected to be an integer.

Some important things to note:

- **Every argument you declare works both as a positional parameter and as a named flag - automatically.** You don't have to choose between them, and your users can mix and match. See [How Argument Parsing Works](#how-argument-parsing-works) below for the full details.
- The positional ordering of args follows the order of declaration in the args block.
- Flags (like `--word` and `--repeats`) are generated for you based on each argument's name.

Let's look at a more complex example to demonstrate some more features. Let's call it `wordjoin`.

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
---
Given some words, joins them together, and optionally capitalizes them.
---
args:
    words str[]                            # Words to join together.
    joiner j str = "-"                     # Joiner for the words.
    should_capitalize "capitalize" c bool  # If true, capitalize the words.

if should_capitalize:
    words = [upper(w) for w in words]

print(join(words, joiner))
```

If we run `-h` on this one:

```shell
./wordjoin -h
```

<div class="result">
```
Given some words, joins them together, and optionally capitalizes them.

Usage:
  wordjoin <words> [joiner] [OPTIONS]

Script args:
      --words strs   Words to join together.
  -j, --joiner str   Joiner for the words. (default -)
  -c, --capitalize   If true, capitalize the words.
```
</div>

Let's break down each declaration to see what's going on here.

1. `words str[]  # Words to join together.`
    - We declare an arg `words` which is a list of strings. Note that `int[]`, `float[]` and `bool[]` can be used for int, float, and bool lists respectively.
    - We also define an arg comment to make the usage string include a description of what the argument is.

2. `joiner j str = "-"  # Joiner for the words.`
    - We declare a second argument, this one a string called `joiner`. We also define a shorthand flag `j`, allowing users to specify the arg with a simple `-j` flag.
    - After that, we define a **default** value `-` for this parameter that will be used if the user doesn't provide one. We finish with another arg comment.

3. `should_capitalize "capitalize" c bool  # If true, capitalize the words.`
    - We declare our final argument `should_capitalize`. We rename it with `"capitalize"`, which will be what users see exposed to them, instead of the initial variable name.
`should_capitalize` will remain the name of the variable to be referenced throughout the script. We define a shorthand `c`, and specify the parameter is a `bool` before finally giving it an arg comment.

Bool args are always false by default.

---

To bring it all together, this is the anatomy of an arg declaration (`<angle brackets>` mean it's required, `[square brackets]` indicate it's optional):

`<name> [rename] [shorthand flag] <type> [= default] [# arg comment]`

Feel free to go back up and check this against the example scripts we wrote, you'll see how each one fits this mold.

!!! tip "Underscores become hyphens"

    If you name an argument with underscores (e.g., `my_arg`), Rad automatically converts it to hyphens for the CLI
    flag: `--my-arg`. Inside your script, you still reference the variable as `my_arg`.

[//]: # (todo TIP on when to use rename)

## How Argument Parsing Works

Most CLI frameworks force you to choose: either an argument is positional, or it's a named flag.
Rad doesn't make you choose - every argument you declare automatically works as *both*.

This section explains that model and a few related parsing behaviors that are useful to know.

### The Dual-Nature Model

When you declare an argument, Rad generates both a positional slot and a named flag for it.
Your users can then invoke your script whichever way is most convenient - or mix both styles in the same call.

Consider this script:

```rad
args:
    src str       # Source file
    dest str      # Destination path
    verbose v bool
```

All of these invocations are equivalent:

```shell
# Pure positional
./script myfile.txt /backup/

# Pure flags
./script --src myfile.txt --dest /backup/

# Mixed - positional for src, flag for dest
./script myfile.txt --dest /backup/
```

This means script authors get a great CLI experience for free. Users who know the positional order can be terse;
users who want clarity can use flags; and anyone can mix and match as needed.

### Parsing Order

Rad processes arguments left-to-right. The rules are simple:

1. If an argument starts with `--` or `-`, it's matched as a flag by name (or short name).
2. If an argument is a bare value (no `-` prefix), it's assigned to the next unfilled positional slot, in the order you declared your args.
3. Flags can appear anywhere - before, after, or interspersed with positional values.

The key insight: when a flag "fills" an argument, that positional slot is skipped. For example:

```rad
args:
    first str
    second str
    third str
```

```shell
# --second fills the second slot via flag, so the two bare values
# fill "first" and "third" (the remaining unfilled slots, in order).
./script alpha --second beta gamma
```

Here, `first` = `"alpha"`, `second` = `"beta"`, `third` = `"gamma"`.

### The `--` Separator

Rad supports the standard Unix `--` convention: everything after a bare `--` is treated as a positional argument, even if it starts with `-`.

This is useful when you need to pass values that look like flags:

```shell
./script -- --not-a-flag
```

Here, `--not-a-flag` is treated as a positional string value, not as a flag.

### Negative Numbers

Passing negative numbers can look like flags since they start with `-`. There are a few ways to handle this:

```rad
args:
    count int
```

```shell
# Use flag + value
./script --count -5

# Use equals syntax
./script --count=-5

# Use -- to force positional
./script -- -5
```

### Equals Syntax

Both `--name=value` and `--name value` work for all flag types. For bool flags specifically,
`--flag` sets the value to `true` by presence, while `--flag=false` explicitly sets it to `false`.

Equals syntax is especially useful when you need to be unambiguous - for example, when passing negative numbers or values that start with `-`.

## Argument Types and User Input

This section explains the different argument types you can declare and how users pass values for each type.

### Basic Types

For basic types (`int`, `float`, `str`), users can pass values either positionally or via flags:

```rad
#!/usr/bin/env rad
args:
    name str
    age int
    height float

print("{name} is {age} years old and {height}m tall")
```

Both invocations work identically:

```shell
./script Alice 25 1.65
./script --name Alice --age 25 --height 1.65
./script Alice 25 --height 1.65
```

Type validation happens automatically. If a user provides `--age abc` when age is an `int`, Rad will show an error.

### Bool Flags

Bool arguments are **never positional** - they must always be passed via flags. You can set bool flags either by presence or with an explicit value:

```rad
args:
    verbose v bool
    debug d bool

print("verbose: {verbose}, debug: {debug}")
```

```shell
./script --verbose           # verbose=true, debug=false
./script -v                  # Same as above (short flag)
./script -v -d               # Both true
./script --verbose=true      # Explicit value
./script --verbose=false     # Explicitly set to false
```

Bool args always default to `false` unless you give them a different default:

```rad
args:
    verbose bool = true  # Defaults to true
```

### List Types

For list arguments (`str[]`, `int[]`, `float[]`, `bool[]`), users pass values by **repeating the flag**:

```rad
args:
    files f str[]
    counts int[]

print("files: {files}")
print("counts: {counts}")
```

```shell
./script -f hello.txt -f world.txt --counts 1 --counts 2 --counts 3
```

<div class="result">
```
files: [ "hello.txt", "world.txt" ]
counts: [ 1, 2, 3 ]
```
</div>

Both long and short flags can be used interchangeably and repeated as needed.

### Variadic Arguments

Variadic arguments use a `*` prefix and collect any number of positional values into a list:

```rad
args:
    command str
    *files str
    verbose v bool

print("command: {command}")
print("files: {files}")
```

```shell
./script build file1.txt file2.txt file3.txt --verbose
```

<div class="result">
```
command: build
files: [ "file1.txt", "file2.txt", "file3.txt" ]
```
</div>

Variadic args can have defaults and work with any type (`*numbers int`, `*values float`, etc.):

```rad
args:
    *items str = ["default.txt"]
```

You can have multiple variadic sections separated by flags - the flags act as delimiters:

```rad
args:
    *section1 str
    *section2 int
    flag f bool

print("section1: {section1}")
print("section2: {section2}")
```

```shell
./script a b c --flag 1 2 3
```

<div class="result">
```
section1: [ "a", "b", "c" ]
section2: [ 1, 2, 3 ]
```
</div>

!!! note "List types vs variadic"

    List args and variadic args both produce lists, but they're filled differently. List args are primarily filled by **repeating the flag** (`--files a.txt --files b.txt`), though a single positional value also works. If you want to collect multiple bare positional values without repeating flags, use a variadic argument (`*args`) instead - that's exactly what they're for.

### Optional Arguments

Mark arguments as optional with `?` suffix. When not provided, their value is `null`:

```rad
args:
    name str
    role str?
    year int?

if role == null:
    print("{name} has no role assigned")
else:
    print("{name} is a {role}")
```

```shell
./script Alice
```

<div class="result">
```
Alice has no role assigned
```
</div>

You can check for null values with `== null` or use truthy/falsy logic: `if role:` will be false when role is null.

!!! info "Defaults vs Optional"

    There are three states an argument can be in:

    - **Required** (e.g., `name str`): The user must provide a value. No default, not nullable.
    - **Has a default** (e.g., `joiner str = "-"`): Optional from the user's perspective - they can omit it and the default is used. The value is never `null`.
    - **Nullable** (e.g., `role str?`): If the user doesn't provide it, the value is `null`.

    Use a default when you have a sensible fallback value. Use `?` when the *absence* of a value is meaningful and you want to handle it explicitly in your script.

### Short Flag Clustering

When you define short flags (single letters), users can cluster multiple bool flags together:

```rad
args:
    verbose v bool
    debug d bool
    quiet q bool
```

```shell
./script -vdq    # Same as: -v -d -q
```

All three flags are set to true with this single clustered argument.

### Int Flag Counting

For `int` arguments with short flags, repeating the flag increments the count:

```rad
args:
    verbosity v int
```

```shell
./script -vvv           # verbosity = 3
./script -vvvvv         # verbosity = 5
```

This is useful for verbosity levels. If an explicit value is provided, it overrides the counting:

```shell
./script -vvv=10        # verbosity = 10 (not 3)
```

## Constraints

In addition to declaring the arguments themselves, Rad also allows you to declare constraints on those arguments, such as what kinds of values are valid.

By doing this in the args block, Rad can use this information to validate input for you, and automatically include the information in your script's usage string.

If a user gives an input which doesn't meet one of the listed constraints, Rad will print:

1. The specific error and constraint that was violated.

2. The usage string.

### Enums

If you have a string argument where you really only want to accept some limited number of known values, you can use an **enum constraint**.

Let's use a simple example, we'll call the script `colors`:

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
args:
    color str
    color enum ["red", "green", "blue"]
print("You like {color}!")
```

If we print the usage string, you can see it tells users what values are valid:

```shell
./colors -h
```

<div class="result">
```
Usage:
  colors <color>

Script args:
      --color str    Valid values: [red, green, blue].
```
</div>

If we invoke this script with a value outside the listed valid values:

```rad
./colors yellow
```

<div class="result">
```
Invalid 'color' value: yellow (valid values: red, green, blue)
Usage:
  colors <color>

Script args:
      --color str    Valid values: [red, green, blue].
```
</div>

Whereas using a valid value will run the script as intended:

```shell
./colors green
```

<div class="result">
```
You like green!
```
</div>

### Range

For numeric arguments (`int` or `float`), you can enforce minimum and maximum values using a **range constraint**.

Range constraints use `[` for inclusive bounds and `(` for exclusive bounds:

```rad
#!/usr/bin/env rad
args:
    age int
    temperature float

    age range [0, 120]              // 0 to 120, both inclusive
    temperature range (-50.0, 50.0) // Between -50 and 50, both exclusive
```

You can also specify only a floor or only a ceiling:

```rad
args:
    count int
    score float

    count range [1,]      // Minimum of 1, no maximum
    score range (,100.0]  // No minimum, maximum of 100 (inclusive)
```

!!! info "Floor/Ceiling Syntax"

    When specifying only a floor or ceiling, both `[0,]` and `[0,)` are equivalent (both mean "minimum of 0, no maximum"). The closing delimiter doesn't affect the meaning when the value is omitted.

Let's look at a complete example with various range types:

```rad title="File: validator"
#!/usr/bin/env rad
args:
    age int              # Person's age
    score float          # Test score
    count int            # Item count

    age range [0, 120]
    score range (0, 100]
    count range [1,]
```

The help text automatically shows the constraints:

```shell
./validator -h
```

<div class="result">
```
Usage:
  validator <age> <score> <count> [OPTIONS]

Script args:
      --age int      Person's age. Range: [0, 120]
      --score float  Test score. Range: (0, 100]
      --count int    Item count. Range: [1, )
```
</div>

If a user provides an invalid value, they get a clear error:

```shell
./validator 150 50.0 5
```

<div class="result">
```
'age' value 150 is > maximum 120

Usage:
  validator <age> <score> <count> [OPTIONS]

Script args:
      --age int      Person's age. Range: [0, 120]
      --score float  Test score. Range: (0, 100]
      --count int    Item count. Range: [1, )
```
</div>

When using exclusive bounds, values exactly at the boundary are rejected:

```shell
./validator 25 0 5   # score of 0 is invalid (exclusive minimum)
```

<div class="result">
```
'score' value 0 is <= minimum (exclusive) 0
```
</div>

### Regex

If you'd like input strings to match a certain pattern, you can do that via a **regex constraint**.

```rad
args:
    name str
    name regex "[A-Z][a-z]*"
print("Hi, {name}")
```

In this example, a valid `name` value must start with a capital letter, and can then be followed by any number of lowercase letters.
No other characters will be accepted, so `Alice` will be a valid value, but `bob` or `John123` are not.

As with other constraints, Rad will validate input against this regex, and if it doesn't match, it will print an error. The constraint is also printed in the script's usage string.

### Relational

Relational constraints let you express logical relationships **between your script's arguments**. There are two types of constraints you can define:

- `excludes` (arguments can't appear together)
- `requires` (an argument depends on another argument also being provided)

You can optionally precede these with the `mutually` keyword to indicate that the constraint applies in both directions.

#### Exclusion

Use `excludes` to prevent arguments from being specified together. For example, consider a script that accepts either a file (`--file`) or a URL (`--url`), but not both:

```rad title="File: fetcher"
#!/usr/bin/env rad
args:
  file str
  url str

  file mutually excludes url

if file:
    print("Reading from file:", file)
else:
    print("Fetching from URL:", url)
```

You can then provide either argument:

```
> ./fetcher --file data.json
Reading from file: data.json

> ./fetcher --url https://example.com/data.json
Fetching from URL: https://example.com/data.json
```

If both are provided, Rad gives a clear error:

```
> ./fetcher --file data.json --url https://example.com/data.json
Invalid arguments: 'file' excludes 'url', but 'url' was given
```

Note in this example, that if e.g. `file` is provided, then `url` will be `null` (and vice versa).

#### Requirement

Use the `requires` keyword to indicate that, when one argument is defined, so must another argument.

Consider a script that can authenticate either by using a token or by providing a username/password pair. 

If the user provides a username, the password is also required.

```rad title="File: auth"
args:
  token str
  username str
  password str

  username mutually requires password
  token mutually excludes username, password

if token:
    print("Authenticating with token {token}")
else:
    print("Authenticating user {username} with password length {len(password)}")
```

Valid usage examples:

```
> ./auth --token abc123
Authenticating with token abc123

> ./auth --username alice --password secret
Authenticating user alice with password length 6
```

Invalid usage examples:

```
> ./auth --username alice
Invalid arguments: 'username' requires 'password', but 'password' was not provided

> ./auth --token abc123 --password secret
Invalid arguments: 'token' excludes 'password', but 'password' was given
```

## Summary

- Rad takes a *declarative* approach to args, and handles parsing user input.
- **Every argument works both positionally and as a named flag** - users can mix and match freely.
- Parsing is left-to-right: flags fill slots by name, bare values fill the next unfilled positional slot in declaration order.
- Anatomy of an arg declaration:

    `<name> [rename] [shorthand flag] <type> [= default] [# arg comment]`

- Argument names with underscores become hyphenated flags (e.g., `my_arg` becomes `--my-arg`).
- Rad supports various argument types:
    - Basic types: `int`, `float`, `str`, `bool`
    - List types: `str[]`, `int[]`, `float[]`, `bool[]` (passed via repeated flags)
    - Variadic: `*args <type>` (collects remaining positional arguments)
    - Optional: `<type>?` (value is `null` when not provided)
- Bool flags are never positional. All other types are both positional and flag-based.
- Short flags support clustering (`-vdq`) and counting for ints (`-vvv` = 3).
- Use `--` to force everything after it to be treated as positional (useful for values starting with `-`).
- Both `--flag=value` and `--flag value` work. Equals syntax is useful for negative numbers (`--count=-5`).
- You can apply constraints to arguments inside the arg block:
    - `enum` for discrete values
    - `range` for numeric bounds (using `[` for inclusive, `(` for exclusive)
    - `regex` for pattern matching
    - Relational constraints (`requires`, `excludes`)
- Details in the arg block are used by Rad to provide a better usage/help string.

## Next

Next, we'll look at another important concept in Rad: [Functions](./functions.md).
