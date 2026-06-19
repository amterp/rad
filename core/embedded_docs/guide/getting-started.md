## What is Rad?

*If you want to jump straight to programming, feel free to skip forwards to Installation.*

**Rad is a modern programming language and interpreter designed to make writing CLI scripts delightful.** It has a familiar, Python-like syntax with
built-in concepts like declarative arguments, JSON handling, first-class shell integration, and user prompts - making common scripting tasks easy and
fast to implement.

If you've ever written scripts and felt like your tools were fighting you, you're not alone:

**With Bash**, you might find yourself:

- Wrestling with `while getopts` boilerplate just to parse a few arguments
- Doing `curl | jq | awk` gymnastics to process JSON
- Skipping the "nice-to-haves" like `--help` or input validation because they're too painful to implement
- Delivering scripts that work, but are hard to read and harder to maintain

**With Python**, you get better syntax, but:

- You need argparse boilerplate before you even start solving your problem, or reach for additional libraries
- Sharing scripts means coordinating runtimes and dependencies (virtualenvs, lockfiles)
- Simple CLI tasks feel heavier than they should

**With Rad**, the language is purpose-built for CLI scripting:

- **Delightful to write** - familiar syntax without the footguns
- **Self-documenting** - automatic `--help` generation from your code comments
- **Declarative arguments** - type checking, validation, and parsing built into the language
- **Easy to share** - batteries included, no external dependencies needed

If you ever write a CLI script, **Rad is the language to do it in**.

**Project Status:** Rad is in early development but useful today for real scripts. Core features work well, though expect breaking changes between
versions as the language evolves. Your feedback directly shapes Rad's future!

Let's get into it.

## Installation

### macOS (Homebrew)

```bash
brew install rad
```

### Go (from source, all platforms)

```bash
go install github.com/amterp/rad@latest
go install github.com/amterp/rad/radls@latest
```

The first command installs the `rad` interpreter. The second installs `radls`, the language server for editor support.

**Note:** You will need to run `go install` yourself to upgrade Rad as new versions are released.
For automated updates, use one of the supported package managers that allow it.

### Binary Downloads

Pre-built binaries for macOS, Linux, and Windows: releases page (https://github.com/amterp/rad/releases)

Each release archive includes both `rad` (the interpreter) and `radls` (the language server).

### Checking Installation

After you've installed Rad and ensured it's on your PATH, you can check your installation:

```sh
rad -h
```

If this prints the help string for Rad, you're set!

## Editor Support

### Visual Studio Code

Rad has a VS Code extension here (https://marketplace.visualstudio.com/items?itemName=amterp.rad-extension) which provides:

- **Syntax highlighting** - makes your Rad code easy to read
- **Real-time validation** - catches syntax errors as you type
- **Diagnostics** - warnings and hints from the Rad Language Server

Installing it is highly recommended!

!vsc-example.png

**Note: Language Server Requirement**

    The VS Code extension uses the Rad Language Server (`radls`) for real-time validation and diagnostics.
    This is installed automatically when you install Rad via Homebrew or `go install`.

    If you see "Rad Language Server not found", make sure `rad` and `radls` are on your PATH.

### Other Editors

The Rad Language Server (`radls`) can be used with any editor that supports the Language Server Protocol (LSP):

- **Neovim**: Configure with `nvim-lspconfig` - point it to the `radls` binary
- **Sublime Text**: Use the LSP package with the `radls` command

For syntax highlighting without LSP, Rad's TextMate grammar can be used in editors that support it.

## Your First Rad Script - Hello World

Let's write the classic "Hello, World!" program. We'll then modify it to give it a bit of a Rad twist, demoing a few features that Rad provides.

First, create a file, let's call it simply `hello`, and give it the following contents:

```rad
print("Hello, World!") 
```

**Tip: `print()`**

    `print()` is the go-to function for outputting to stdout. It behaves a lot like `print()` in Python.

    You can read more about it in the reference (rad docs print).

**Info: File Extension**

    If you want to give your Rad scripts an extension, you can use `.rad` as a convention.

Then, run the script from your CLI by invoking `rad` on it, and you should see it print out like so:

```sh
> rad hello
```

```
Hello, World!
```

Nice! Now, let's spruce it up with a few Rad features.

## Adding In Some Rad Features

One of Rad's primary features is its simple, declarative style to defining script arguments.

Let's modify the script to greet a name you input from command line.

```rad
args:
  name str

print("Hello, {name}!")
```

A couple of things to note here:

1. We define an "args block". Right now it contains just the one line, but you can do lots of things in here (rad docs guide/args).
2. The modified `print()` utilizes string interpolation (rad docs guide/strings-advanced). String interpolation in Rad behaves a lot like it does in Python (you'll see this is a pattern).

Now, let's try invoking the script again, and this time, input a name:

```sh
> rad hello Alex
```

```
Hello, Alex!
```

Cool! What happens if we invoke *without* a name?

```sh
> rad hello
```

```
Usage:
  hello <name>

Script args:
      --name str
```

If you run a Rad script without providing *any* args to a script which expects at least one, Rad will print out the script usage, interpreting your invocation as if you had passed `--help`.

**Tip: Positional vs Flag Arguments**

    Arguments can be passed positionally or via flags. For example, both `./hello Alice` and `./hello --name Alice` work identically. The help text shows both forms: `<name>` for positional usage and `--name str` for the flag form.

This shows a little of the automatic script usage that Rad generates for you. Let's explore that a bit more.

## Improving The Script Usage String

Rad facilitates writing well-documented and easy-to-use scripts, in part through unique syntax. Let's use a couple of those now.

First, we'll add a **file header** to your script.

```rad
---
Prints a polite greeting using an input name.
---
args:
  name str
  
print("Hello, {name}!")
```

**Tip: File Headers**

    File headers, as the name suggests, go at the top of Rad scripts (with the exception of shebangs, to be covered later). 
    They allow you to write a description for your script in between two `---` lines. The contents will get printed as part of the script's usage string.

    A couple of ideas on what to cover in your file headers:

    - A brief description of what the script does and how it works
    - Examples of valid invocations and what they do.

Second, we can add **comments** to args that a script declares. Let's do that too:

```rad
---
Prints a polite greeting using an input name.
---
args:
  name str  # The name of someone to greet.
  
print("Hello, {name}!")
```

**Info: Note on `#` vs. `//`**

    Rad uses `#` to denote *arg* comments.
    `#` comments are special and **do get passed** to Rad's parser and can affect script behavior (namely in this case, adding information to the script usage string). 

    Standard code comments in Rad use `//`, similar to Java or C/C++, and **don't** impact script behavior.

Now, when someone reads the script, it's pretty clear what the script does and what the expected arguments are.

But it gets better! Let's try invoking the script's usage string again (this time let's try using the `-h` flag explicitly, though it's not necessary):

```sh
> rad hello -h
```

```
Prints a polite greeting using an input name.

Usage:
  hello <name>

Script args:
      --name str   The name of someone to greet.
```

Not only is the script now easier to maintain for developers, it's also easier for users to understand!

## Shebang

Needing to manually invoke `rad` each time you want to run a Rad script can be a little cumbersome. Thankfully, Unix kernels provide a mechanism known as a "shebang".

Let's add one to our script. It has to go on the very first line.

```rad
#!/usr/bin/env rad
---
Prints a polite greeting using an input name.
---
args:
  name str  # The name of someone to greet.
  
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

```
Hello, Bob!
```

When you invoke an executable script this way, the kernel scans for a shebang (`#!`) in the first line.
If it finds a path to an interpreter (in this case, it will find `rad`, if set up correctly in your `PATH`),
then it will invoke said interpreter on the script (equivalent to `rad hello` like we were doing before).

**Tip: rad new**

    Rad has a command `rad new <script>` which saves you repeating these steps.
    It creates a ready-to-edit executable file with a `rad` shebang on the top.

**Tip: Tab completion**

    Once you have scripts on your PATH, you can enable tab completion for them.
    See Shell Completion (rad docs guide/shell-completion) for setup instructions.

## A Complete Example

Now that you've seen the basics, let's skip ahead a little with our example, to get a better idea of what Rad can really do.
Don't worry - we'll cover everything in detail in later sections. For now, just observe what Rad gives you for free.

Update your `hello` script to this:

```rad
#!/usr/bin/env rad
---
Prints a polite greeting using an input name.
---
args:
  name str                # The name of someone to greet.
  times t int = 1         # How many times to greet them.
  style s str = "normal"  # Greeting style.

  times range (0, 10]
  style enum ["normal", "excited", "formal"]

greeting = switch style:
  case "excited" -> "HEY"
  case "formal" -> "Good day"
  default -> "Hello"

for i in range(times):
  print("{greeting}, {name}!")
```

**What we added:**

- Two more arguments with **default values** (`times` and `style`)
- **Short flag aliases** (`-t` and `-s`) for convenience
- **Argument constraints**
    - `times` must be greater than 0, and max 10 (inclusive)
    - `style` must be one of three values
- A **switch expression** to map style to greeting text
- A **for loop** to repeat the greeting

All of this in ~20 lines, with zero validation boilerplate.

### Try it out

Normal usage with short flags:

```sh
> ./hello Alice -t 3 -s excited
```

```
HEY, Alice!
HEY, Alice!
HEY, Alice!
```

Check the automatically generated help:

```sh
> ./hello -h
```

```
Prints a polite greeting using an input name.

Usage:
  hello <name> [times] [style] [OPTIONS]

Script args:
      --name str    The name of someone to greet.
  -t, --times int   How many times to greet them. Range: (0, 10] (default 1)
  -s, --style str   Greeting style. Valid values: [normal, excited, formal] (default normal)
```

Note how the constraints appear in the help automatically.

### What about invalid input?

Try providing an invalid range:

```sh
> ./hello Bob -t 15
```

```
'times' value 15 is > maximum 10
```

Try an invalid enum value:

```sh
> ./hello Bob -s casual
```

```
Invalid 'style' value: casual (valid values: normal, excited, formal)
```

**You didn't write any of this validation code.**
Rad enforced the constraints you declared and generated helpful error messages automatically.

## Summary

In this section, we were introduced to several key features:

- **String interpolation** with `{variable}` syntax for readable output
- **args blocks** for declarative argument definitions
- **File headers** and **arg comments** for self-documenting scripts
- **Shebangs** to make scripts more convenient to run
- **Argument constraints** (range, enum) with automatic validation - no manual validation needed
- **Short flag aliases** for better user experience
- **Automatic `--help` generation** from your code comments

Features you'd normally skip in Bash, or need external dependencies for in other languages, come built-in with Rad's declarative syntax. 🤙

**Info: Note on Rad script content ordering**

    Rad expects a certain order between shebangs, file headers, arg blocks, and the rest of your code.

    **It's important to adhere to the following ordering in Rad scripts**, or you'll see errors:

    1) Shebang (if present)

    2) File header (if present)

    3) Args block (if present)

    4) Rest of the file

## Next

Great job on getting this far! You've gotten a small peek at what Rad has to offer.

From here, you have a couple of options:

1. Continue your Rad journey: dive into more details with the next section: Basics (rad docs guide/basics).

2. If you'd like to see additional unique Rad features, here are some sections you can jump to:
    - Args (rad docs guide/args)
    - Rad Blocks (rad docs guide/rad-blocks)
    - Shell Commands (rad docs guide/shell-commands)
    - Example: brewi (rad docs examples/brewi)
