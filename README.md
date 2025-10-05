# 🤙 Rad

**A lightweight, modern CLI scripting language that's familiar, clean, and readable.**

Effortlessly write high-quality scripts without the quirks of Bash.
Rad makes command-line scripting simple and intuitive — whether automating tasks, processing text, or interacting with systems and APIs.

Example script (`greet`):

```
#!/usr/bin/env rad
---
Greets someone by name, potentially a lot of times!
---
args:
  name str                # Name of the person to greet.
  times int = 1           # How many times to greet them.
  shout s bool            # Enable to shout at them!
  greeting str    = "Hi"  # How to greet the person.
  
  times range (0, 10]
  greeting enum ["Hi", "Hello", "Hey"]

for _ in range(times):
  text = "{greeting}, {name}!"
  if shout:
    text = upper(text)
  print(text)
```

Generated help string:

```
> ./greet -h
Greets someone by name, potentially a lot of times!

Usage:
  greet <name> [times] [shout] [greeting]

Script args:
      --name str          Name of the person to greet.
      --times int         How many times to greet them. Range: (0, 10] (default 1)
  -s, --shout             Enable to shout at them!
      --greeting str      How to greet the person. Valid values: [Hi, Hello, Hey]. (default Hi)
```

Example invocation:

```
> ./greet bob 3 -s
HI, BOB!
HI, BOB!
HI, BOB!
```

## Installation

### macOS (Homebrew)

```shell
brew install amterp/rad/rad
```

### Go (from source, all platforms)

```shell
go install github.com/amterp/rad@latest
```

Installs directly from source using the Go (v1.17+) toolchain into your Go bin directory. Make sure that it's on your PATH.

**Note:** You will need to run `go install` yourself to upgrade Rad as new versions are released.
For automated updates, install via one of the package managers that support it.

### Binary Downloads (all platforms)

Pre-built binaries are available for macOS, Linux, and Windows on the [releases page](https://github.com/amterp/rad/releases).

Download the appropriate binary for your platform:

- **macOS**: `rad_darwin_arm64.tar.gz` (Apple Silicon) or `rad_darwin_amd64.tar.gz` (Intel)
- **Linux**: `rad_linux_arm64.tar.gz` or `rad_linux_amd64.tar.gz`
- **Windows**: `rad_windows_amd64.zip`

Extract the binary and add it to your PATH.

### From Source

See [here](./CONTRIBUTING.md#setup) for instructions on how to build from source.

### Visual Studio Code Extension

The VS Code extension for Rad can be found [here](https://marketplace.visualstudio.com/items?itemName=amterp.rad-extension). Source [here](./vsc-extension).

It provides syntax highlighting and integrates with the Rad language server for error detection. 

The language server is currently only available on macOS & Linux. Source [here](lsp-server).

![vsc-example.png](./assets/vsc-example.png)

## [Documentation](https://amterp.github.io/rad/)

Docs are still a work in progress, but there's enough to get you started!

Feel free to check out the [**Getting Started**](https://amterp.github.io/rad/guide/getting-started/) guide :)

## Status 📊

⚠️ **Rad is still in early development!** ⚠️

Rad is a working CLI tool and interpreter that can run useful Rad scripts.

It's complete enough to be useful, but do expect the following:

- Major, potentially script-breaking changes
- Rough edges
- Bugs
- Missing features

That said, please do give it a try, I'd love to hear your experience and any feedback :)

### What's being worked on 🚧

- Language features (there's a long list!)
- LSP language server ([RLS](lsp-server))

### What's planned 🌱 

- Many more language features
- Polished syntax error feedback
- `rad` script management features & helpers
- JetBrains IDE plugin
- Support for platforms other than macOS: Linux, Windows. 

## About

### What problem does Rad solve? 🎯

Shell languages like Bash are powerful, but often difficult to use.
Bash has unusual syntax for simple things like if statements that makes them hard to remember, and common patterns like
argument parsing can be tedious to implement.
Basic data structures like lists, maps, and even strings can be difficult to deal with, if they're available at all.

What's needed is a higher-level language, bells and whistles included, that's tailored to writing scripts.
Knowing what's commonly needed in scripts, it needs to make implementing these things as easy as possible,
and provide all the necessary utilities out of the box.

### How does Rad solve it? 🛠️

Rad is a language and interpreter, purpose-built for this exact problem.

- Rad is **familiar**, drawing on popular languages like Python.
- Rad **knows its domain** - it has unique syntax which makes writing scripts as easy as possible, such as its declarative approach to script args.
- Rad has **batteries included** - it aims to offer everything you need in a single installation that lets you write the scripts you want.

### Example: Printing a table of a repo's commits

An example for a type of script that Rad makes very easy to write, is one which queries a JSON API, 
extracts some fields, and prints the results in a table.

Let's see a concrete example script (`commits`):

```
args:
    repo str       # The repo to query. Format: user/project
    limit int = 20 # The max commits to return.

url = "https://api.github.com/repos/{repo}/commits?per_page={limit}"

Time = json[].commit.author.date
Author = json[].commit.author.name
SHA = json[].sha

rad url:
    fields Time, Author, SHA
```

Example invocation:

```
> rad commits spf13/cobra 3
Querying url: https://api.github.com/repos/spf13/cobra/commits?per_page=3
Time                  Author          SHA
2025-03-07T14:53:22Z  styee           4f9ef8cdbbc88c5302be95e0e67fd78ebbfa9dd2
2025-02-21T12:46:14Z  Fraser Waters   1995054b003053cc1e404bccfbf6d168e8731509
2025-02-17T19:16:17Z  Yedaya Katsman  f98cf4216d3cb5235e6e0cd00ee00959deb1dc65
```

1. This script takes two args: a repo string and an optional limit (defaults to 20).
    - The `#` comments are read by Rad and used to generate helpful docs / usage strings for the script.
2. It uses string interpolation to resolve the url we will query, based on the supplied args.
3. It defines the fields to extract from the JSON response.
4. It executes the query, extracting the specified fields, and displays the resulting data as a table.
    - Note the `rad url` syntax: "rad" actually stands for "request and display", which is what this built-in syntax does.

We keep this example somewhat minimal - there are Rad features we could use to improve this, but it's kept simple here.

Some alternative valid invocations for this example:

- `rad commits amterp/rad`
- `rad commits --repo amterp/rad --limit 5`
- `rad commits --limit 5 --repo amterp/rad`
- `rad commits amterp/rad --limit 5`

### Alternatives 📚

- **Bash**
  - Bash is great, especially for simple scripts that only need to invoke a series of system commands.
  - That said, as soon as you need to do anything more complex, Bash's syntax quickly becomes cumbersome and unproductive.
    - Crucially, arg parsing in Bash is good for simple cases, but for anything more complex, it quickly gets unwieldy.
  - Rad addresses these problems directly, while also being a great choice for simple scripts.
- **Python, Ruby, JavaScript, Rust, Go, etc**
  - These are general-purpose languages and very flexible, which can be great.
  - However, if your goal is just to write scripts, then they're not as focused as Rad.
    - They require boilerplate, maybe additional installations (modules, libraries), perhaps compilation, etc.
  - Rad will generally require less code to achieve great scripts that do what you want.

### Why Rad? 🚀

- Rad is **tailored to writing scripts**. You can write better scripts, in fewer lines of code.
- Rad is familiar, with a low learning curve. It's **simple and easy to pick up**.
- Rad offers inbuilt syntax that guides you towards writing **user-friendly scripts**, with helpful usage strings (available with `--help`).
- **Shell integration** - Rad offers built-in syntax for invoking shell commands, so you can still reach for Bash when needed.

### Why *not* Rad? ⚠️

Rad is **optimized for the majority of scripts**, but for extremely complex cases, a general-purpose language **may be more appropriate**.

When should you reach for something else?
- If your script **outgrows Rad** and becomes a full application.
- If you need **high-performance computation**, beyond typical scripting needs.
- If your script requires **specialized libraries** that aren't built into Rad.

That said, **Rad aims to handle 99% of CLI scripting needs** - so most of the time, it's the right tool for the job.
