---
title: Getting Started 
---

## Installation

### macOS

```shell
brew tap amterp/rad
brew install rad
```

Other than building from source, rad is not available for other platforms/package managers (yet).

### Checking Installation

After you've installed rad, you can check your installation:

```shell
rad -V
```

If this prints rad's version, you're set!

## Your First RSL Script - Hello World

Let's write the classic "Hello, World!" program. We'll then modify it to give it a bit of an RSL twist, demoing a couple of features rad provides.

!!! note "Terminology: Rad & RSL"

    Rad is the name of the CLI tool and interpreter capable of running RSL scripts.

    **Rad** stands for "Request And Display", and **RSL** stands for "Rad Scripting Language".

First, create a file, let's call it simply `hello`, and give it the following contents:

```rsl
print("Hello, World!") 
```

!!! tip "`print()`"

    `print()` is the go-to function for outputting to stdout. It behaves a lot like `print()` in Python.

    You can read more about it [in the reference](../reference/functions.md#print).

!!! info "RSL Extensions"

    If you want to give your RSL scripts an extension, you can follow `.rsl` as a convention.

Then, run the script from your CLI by invoking `rad` on it, and you should see it print out like so:

```sh
> rad ./hello
```

<div class="result">
```
Hello, World!
```
</div>

Nice! Next, let's spruce it up with a few RSL features.

## Adding In Some RSL Features

One of the selling points of rad is that it makes defining arguments to your script super simple, using a declarative style.

Let's modify the script to greet a name you input from command line.

```rsl
args:
  name string
  
print("Hello, {name}!")
```

A couple of things to note here:

1. We define an "args block". Right now it contains just the one line, but [you can do lots of things in here](../reference/args.md).
2. The modified `print()` utilizes [string interpolation](../reference/strings.md#string-interpolation). String interpolation in RSL behaves a lot like it does in Python (you'll see this is a pattern).

Now, let's try invoking the script again, and this time, input your (or someone's) name:

```sh
> rad ./hello Alex
```

<div class="result">
```
Hello, Alex!
```
</div>

Cool! What if we invoke *without* a name?

```sh
> rad ./hello
```

<div class="result">
```
Usage:
  hello <name>

Script args:
      --name string
```
</div>

If you run an RSL script without providing *any* args to a script which expects at least one, rad will print out the script usage, interpreting your invocation similar to if you had passed `--help`.

This shows a little of the automatic script usage string generation that rad gives you. Let's explore that a bit more.

## Improving Script Usage String

RSL facilitates writing well-documented and easy-to-use scripts, in part through unique syntax that it offers. Let's use a couple of those now.

First, we'll add a **file header** to your script.

```rsl
---
Prints a polite greeting using an input name.
---
args:
  name string
  
print("Hello, {name}!")
```

!!! tip "File Headers"

    File headers, as the name suggests, go at the top of RSL scripts (with the exception of shebangs, to be covered later). 
    They allow you to write a description for your script in between two `---` lines. The contents will get printed as part of the script's usage string.

    Some ideas on what to cover in your file headers:

    - A brief description of what the script does.
    - Why you might want to use the script.
    - Examples of valid invocations and what they do.

Second, we can add **comments** to args that a script declares. Let's do that too:

```rsl
---
Prints a polite greeting using an input name.
---
args:
  name string # The name of someone to greet.
  
print("Hello, {name}!")
```

!!! info "Note on `#` vs. `//`"

    RSL uses `#` to denote a *arg* comments in RSL.
    `#` comments are special and **do get passed** to RSL's parser and can affect script behavior (namely in this case, adding information to the script usage string). 

    Standard code comments in RSL use `//`, similar to Java or C/C++. These are stripped prior to parsing and don't impact script behavior.

    You can use code comments on your arg comments, if you so choose e.g.

    ```rsl
    name string # A name.  // todo make this more descriptive
    ```

Now, when someone reads the script, it's pretty clear what the script does and what the expected arguments are.

But it gets better! Let's try invoking the script's usage string again (this time let's try using the `-h` flag explicitly, though it's not necessary):

```sh
> rad ./hello -h
```

<div class="result">
```
Prints a polite greeting using an input name.

Usage:
  hello <name>

Script args:
      --name string   The name of someone to greet.
```
</div>

Not only is the script now easier to maintain for developers, it's also easier for users to understand!

## Shebang

Last thing, as part of this introduction guide.

Needing to manually invoke `rad` each time you want to run an RSL script can be a little cumbersome. Thankfully, Unix kernels provide a mechanism known as a "shebang".

Let's add one to our script. It has to go on the very first line.

```rsl
#!/usr/bin/env rad
---
Prints a polite greeting using an input name.
---
args:
  name string # The name of someone to greet.
  
print("Hello, {name}!")
```

Then, make the script executable using the following command:

```sh
chmod +x ./hello
```

Now, you can invoke the script directly:

```sh
> ./hello Bob
```

<div class="result">
```
Hello, Bob!
```
</div>

Basically, when you invoke an executable script this way, the Kernel scans for a shebang (`#!`) in the first line.
If it finds a path to an interpreter (in this case, it will find `rad` if you've correctly put it in your `PATH`),
it will invoke said interpreter on the script (equivalent to `rad ./hello` like we were doing before).

## Learnings Summary

- We learned how to print, and saw an example of **string interpolation**.
- We were introduced to the **args block**
- We saw how we can write self-documenting scripts that also help our users by leveraging **file headers** and **arg comments**.
- We saw how we can leverage **shebangs** to make our scripts more convenient to run.

!!! info "Note on RSL file contents ordering"

    Rad expects a certain order between shebangs, file headers, arg blocks, and the rest of your code.

    **It's important to adhere to the following ordering in RSL scripts**, or you'll see errors:
    
    1) Shebang (if present)
  
    2) File header (if present)
    
    3) Args block (if present)
    
    4) Rest of the file

## Next

Great job on getting this far! You've gotten a peek at what rad and RSL have to offer.

From here, you have two options:

1. Continue your RSL journey: dive into more details with the next section: [Basics](./basics.md).

2. If you're interested instead in seeing additional unique RSL features, feel free to skip ahead to any of these sections:
    - [Args](./args.md)
    - [Rad Blocks](./rad-blocks.md)
    - [Shell Commands](./shell-commands.md)

[//]: # (TODO pick_from_resource ^)
