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

[//]: # (todo point 1 should clarify *FOR USERS*)

- All arguments can be defined positionally or via a flag.
- The positional ordering of args follows the order of declaration in the block.
- Flags are automatically generated and can be used by users to pass values for that argument, instead of doing it positionally.

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

[//]: # (todo TIP on when to use rename)

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

Bool arguments are **never positional** - they must be passed via flags. You can set them either by presence or with an explicit value:

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
- All args can be specified positionally or via a flag from the user.
- Anatomy of an arg declaration:

    `<name> [rename] [shorthand flag] <type> [= default] [# arg comment]`

- Rad supports various argument types:
    - Basic types: `int`, `float`, `str`, `bool`
    - List types: `str[]`, `int[]`, `float[]`, `bool[]` (passed via repeated flags)
    - Variadic: `*args <type>` (collects remaining positional arguments)
    - Optional: `<type>?` (value is `null` when not provided)
- Short flags support clustering (`-vdq`) and counting for ints (`-vvv` = 3)
- You can apply constraints to arguments inside the arg block:
    - `enum` for discrete values
    - `range` for numeric bounds (using `[` for inclusive, `(` for exclusive)
    - `regex` for pattern matching
    - Relational constraints (`requires`, `excludes`)
- Details in the arg block are used by Rad to provide a better usage/help string.

## Next

Nice, let's now look at another Rad feature which makes it uniquely suited to certain types of scripting: [Rad Blocks](./rad-blocks.md).
