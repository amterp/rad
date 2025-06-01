---
title: Args
---

This section covers syntax for defining arguments that your script can accept.

## Arg Declarations

Rad takes a declarative approach to arguments.
You simply declare what arguments your script accepts, and let Rad take care of the rest, including parsing user input.

Arguments are declared as part of an **args block**.

Here's an example script we'll call `printwords` that prints an input word some number of times:

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
args:
    word string
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
      --word string
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
args:
    words string[]                         # Words to join together.
    joiner j string = "-"                  # Joiner for the words.
    should_capitalize "capitalize" c bool  # If true, capitalize the words.

if should_capitalize:
    words = words[upper(w) for w in words]

print(join(words, joiner))
```

If we run `-h` on this one:

```shell
./wordjoin -h
```

<div class="result">
```
Usage:
  wordjoin <words> [joiner] [-c, --capitalize]

Script args:
      --words string,string    Words to join together.
      -j, --joiner string      Joiner for the words. (default -)
      -c, --capitalize         If true, capitalize the words.
```
</div>

Let's break down each declaration to see what's going on here.

[//]: # (todo points below all show up as 1)

1. `words string[]  # Words to join together.`
    - We declare an arg `words` which is a list of strings. Note that `int[]`, `float[]` and `bool[]` can be used for int, float, and bool lists respectively.
    - We also define an arg comment to make the usage string include a description of what the argument is.

2. `joiner j string = "-"  # Joiner for the words.`
    - We declare a second argument, this one a string called `joiner`. We also define a shorthand flag `j`, allowing users to specify the arg with a simple `-j` flag.
    - After that, we define a **default** value `-` for this parameter that will be used if the user doesn't provide one. We finish with another arg comment.

3. `should_capitalize "capitalize" c bool  # If true, capitalize the words.`
    - We declare our final argument `should_capitalize`. We rename it with `"capitalize"`, which will be what users see exposed to them, instead of the initial variable name.
`should_capitalize` will remain the name of the variable to be referenced throughout the script. We define a shorthand `c`, and specify the parameter is a `bool` before finally giving it an arg comment.

Bool args are always false by default.

---

To bring it all together, this is the anatomy of an arg declaration (`<angle brackets>` represent it's required, `[square brackets]` indicate it's optional):

`<name> [rename] [shorthand flag] <type> [= default] [# arg comment]`

Feel free to go back up and check this against the example scripts we wrote, you'll see how each one fits this mold.

[//]: # (todo TIP on when to use rename)

## Constraints

In addition to declaring the arguments themselves, Rad also allows you to declare constraints on those arguments, such as what kinds of values are valid.

By doing this in the args block, Rad can use this information to validate input for you, and automatically include in the information in your script's usage string.

If a user gives an input which doesn't meet one of the listed constraints, Rad will print:

1. The specific error and constraint that was violated.

2. The usage string.

### Enums

If you have a string argument where you really only want to accept some limited number of known values, you can use an **enum constraint**.

Let's use a simple example, we'll call the script `colors`:

```rad linenums="1" hl_lines="0"
#!/usr/bin/env rad
args:
    color string
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
      --color string    Valid values: [red, green, blue].
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
      --color string    Valid values: [red, green, blue].
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

### Regex

If you'd like input strings to match a certain pattern, you can do that via a **regex constraint**.

```rad
args:
    name string
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
  file string
  url string

  file mutually excludes url

if is_defined("file"):
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

#### Requirement

Use the `requires` keyword to indicate that, when one argument is defined, so must another argument.

Consider a script that can authenticate either by using a token or by providing a username/password pair. 

If the user provides a username, the password is also required.

```rad title="File: auth"
args:
  token string
  username string
  password string

  username mutually requires password
  token mutually excludes username, password

if is_defined("token"):
    print("Authenticating with token:", token)
else:
    print("Authenticating user:", username)
```

Valid usage examples:

```
> ./auth --token abc123
Authenticating with token: abc123

> ./auth --username alice --password secret
Authenticating user: alice
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

- You can apply constraints to arguments inside the arg block, such as `enum`, `regex`, and relational constraints.
- Details in the arg block are used by Rad to provide a better usage/help string.

## Next

Nice, let's now look at another Rad feature which makes it uniquely suited to certain types of scripting: [Rad Blocks](./rad-blocks.md).
