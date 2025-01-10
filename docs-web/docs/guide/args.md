---
title: Args
---

This section covers syntax for defining arguments that your script can accept.

## Arg Block

RSL takes a declarative approach to arguments.
You simply declare what arguments your script accepts, and let RSL take care of the rest, including parsing user input.

Arguments are declared as part of an **args block**.

Here's an example for a script we'll call `printwords` which prints an input word N number of times:

```rsl
args:
    word string
    repeats int
    
for _ in range(repeats):
    print(word)
```

We can print its usage string using the `--help` flag:

```shell
rad printwords --help
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

- All arguments can be defined positionally or via a flag.
- The positional ordering of args is defined by the order of args in the block.
- Flags are automatically generated and can be used by users to pass values for that argument, instead of doing it positionally.

Let's look at a more complex example to demonstrate some more features. Let's call it `wordjoin`.

```rsl
args:
    words string[]  # Words to join together.
    joiner j string = "-"  # Joiner for the words.
    should_capitalize "capitalize" c bool  # If true, capitalize the words.

if should_capitalize:
    words = words[upper(w) for w in words]

print(join(words, joiner))
```

If we run `--help` on this one:

```shell
rad wordjoin --help
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

Let's break down each declaration to see what's going on here!

[//]: # (todo points below all show up as 1)

1. `words string[]  # Words to join together.`

We declare an arg `words` which is a list of strings. Note that `int[]`, `float[]` and `bool[]` can be used for int, float, and bool lists respectively.
We also define an arg comment to make the usage string include a description of what the argument is.

2. `joiner j string = "-"  # Joiner for the words.`

We declare a second argument, this one a string `joiner`. We also define a shorthand flag `j`, allowing users to specify the arg with a simple `-j` flag.
After that, we define a **default** for this parameter that will be used if the user doesn't provide one. We finish with another arg comment.

3. `should_capitalize "capitalize" c bool  # If true, capitalize the words.`

We declare our final argument `should_capitalize`. We rename it with `"capitalize"`, which will be what users see exposed to them.
`should_capitalize` will remain the name of the variable to be referenced throughout the script. We define a shorthand `c`, and specify the parameter is a `bool` before finally giving it an arg comment.

Bools are implicitly always false by default.

To bring it all together, this is the anatomy of an arg declaration (`<angle brackets>` represent it's required, `[square brackets]` indicate it's optional):

`<name> [rename] [shorthand flag] <type> [= default] [# arg comment]`

Feel free to go back up and check this against the example scripts we wrote, you'll see how each one fits this mold.

### Constraints

[//]: # (- TBD)
[//]: # (- enums)
[//]: # (- regex &#40;when implemented&#41;)
